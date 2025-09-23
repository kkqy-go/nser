package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/routing"
	"golang.org/x/net/ipv6"
)

func main() {
	// =================================================================================
	// Command-line parameter definitions
	// =================================================================================
	ifaceName := flag.String("iface", "", "(Required) Name of the network interface to use (e.g., Ethernet, eth0)")
	srcIPStr := flag.String("src", "", "(Manual Mode) Source IPv6 address")
	targetIPStr := flag.String("dst", "", "(Manual Mode) Target IPv6 address to query")
	useGateway := flag.Bool("gateway", false, "(Auto Mode) Use this parameter to automatically discover the gateway and send NS requests from all IPv6 addresses on the interface")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -iface <interface> [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Sends an IPv6 Neighbor Solicitation (NS) packet.\n\n")
		fmt.Fprintf(os.Stderr, "Modes:\n")
		fmt.Fprintf(os.Stderr, "  1. Manual Mode: %s -iface <interface> -src <source_ip> -dst <target_ip>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  2. Auto Gateway Mode: %s -iface <interface> --gateway\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Global Parameters:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Check for required parameters
	if *ifaceName == "" {
		log.Println("Error: -iface parameter is required.")
		flag.Usage()
		os.Exit(1)
	}

	// =================================================================================
	// Execute based on mode
	// =================================================================================
	if *useGateway {
		// --- Auto Gateway Mode ---
		fmt.Printf("[*] Entering Auto Gateway Mode for interface: %s\n", *ifaceName)

		// 1. Discover IPv6 Gateway
		r, err := routing.New()
		if err != nil {
			log.Fatalf("Failed to create router: %v", err)
		}
		_, gatewayIP, _, err := r.Route(net.IPv6zero)
		if err != nil {
			log.Fatalf("Failed to automatically discover gateway: %v", err)
		}
		fmt.Printf("[+] Successfully discovered IPv6 gateway: %s\n", gatewayIP)
		// 2. Get all IPv6 addresses on the interface
		iface, err := net.InterfaceByName(*ifaceName)
		if err != nil {
			log.Fatalf("Failed to find interface '%s': %v", *ifaceName, err)
		}
		addrs, err := iface.Addrs()
		if err != nil {
			log.Fatalf("Failed to get addresses for interface '%s': %v", *ifaceName, err)
		}

		var sourceIPs []net.IP
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && ipNet.IP.To4() == nil { // Make sure it's an IPv6 address
				sourceIPs = append(sourceIPs, ipNet.IP)
			}
		}

		if len(sourceIPs) == 0 {
			log.Fatalf("No IPv6 addresses found on interface '%s'", *ifaceName)
		}

		fmt.Printf("[+] Found %d IPv6 addresses on interface '%s'\n", len(sourceIPs), *ifaceName)

		// 3. Iterate through all source IPs and send NS requests
		for _, srcIP := range sourceIPs {
			fmt.Println("--------------------------------------------------")
			err := sendNS(*ifaceName, srcIP, gatewayIP)
			if err != nil {
				log.Printf("Failed to send NS request from %s: %v", srcIP, err)
			} else {
				fmt.Printf("-> Successfully sent NS request from %s to gateway %s\n", srcIP, gatewayIP)
			}
		}
	} else {
		// --- Manual Mode ---
		if *srcIPStr == "" || *targetIPStr == "" {
			log.Println("Error: In manual mode, -src and -dst parameters are required.")
			flag.Usage()
			os.Exit(1)
		}

		sourceIP := net.ParseIP(*srcIPStr)
		if sourceIP == nil {
			log.Fatalf("Invalid source IP address: %s", *srcIPStr)
		}
		targetIP := net.ParseIP(*targetIPStr)
		if targetIP == nil {
			log.Fatalf("Invalid target IP address: %s", *targetIPStr)
		}

		err := sendNS(*ifaceName, sourceIP, targetIP)
		if err != nil {
			log.Fatalf("Failed to send NS request: %v", err)
		}
	}
}

// sendNS builds and sends a Neighbor Solicitation packet using native Go raw sockets
func sendNS(ifaceName string, sourceIP, targetIP net.IP) error {
	// 1. Find the interface
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return fmt.Errorf("failed to get interface %s: %w", ifaceName, err)
	}

	// 2. Create an ICMPv6 connection. Protocol 58 is ICMPv6.
	// This requires running with root/administrator privileges.
	conn, err := net.ListenPacket("ip6:58", "::")
	if err != nil {
		return fmt.Errorf("failed to listen for icmpv6 packets: %w. Check for root privileges", err)
	}
	defer conn.Close()

	rawConn := ipv6.NewPacketConn(conn)

	// Set the Hop Limit to 255 as required by RFC 4861 for Neighbor Discovery.
	if err := rawConn.SetMulticastHopLimit(255); err != nil {
		return fmt.Errorf("failed to set multicast hop limit: %w", err)
	}

	// 3. Calculate the Solicited-Node multicast address (ff02::1:ffxx:xxxx)
	solicitedNodeAddr := net.IP{
		0xff, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x01, 0xff, targetIP[13], targetIP[14], targetIP[15],
	}

	// 4. Build the ICMPv6 Neighbor Solicitation packet (without IP or Ethernet layers)
	icmpv6Layer := &layers.ICMPv6{TypeCode: layers.CreateICMPv6TypeCode(layers.ICMPv6TypeNeighborSolicitation, 0)}
	// For checksum calculation, gopacket needs a pseudo-header from the network layer.
	ipv6LayerForChecksum := &layers.IPv6{
		SrcIP:      sourceIP,
		DstIP:      solicitedNodeAddr,
		NextHeader: layers.IPProtocolICMPv6,
	}
	icmpv6Layer.SetNetworkLayerForChecksum(ipv6LayerForChecksum)

	nsLayer := &layers.ICMPv6NeighborSolicitation{
		TargetAddress: targetIP,
		Options: []layers.ICMPv6Option{
			{Type: layers.ICMPv6OptSourceAddress, Data: iface.HardwareAddr},
		},
	}

	// 5. Serialize the ICMPv6 layer
	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
	err = gopacket.SerializeLayers(buffer, opts, icmpv6Layer, nsLayer)
	if err != nil {
		return fmt.Errorf("failed to serialize icmpv6 ns layer: %w", err)
	}
	packetData := buffer.Bytes()

	// 6. Send the packet
	wcm := ipv6.ControlMessage{
		Src:     sourceIP,
		IfIndex: iface.Index,
	}
	ipAddr := &net.IPAddr{IP: solicitedNodeAddr}

	_, err = rawConn.WriteTo(packetData, &wcm, ipAddr)
	if err != nil {
		return fmt.Errorf("failed to write packet: %w", err)
	}

	fmt.Printf("Successfully sent Neighbor Solicitation (NS) for %s to %s\n", targetIP, solicitedNodeAddr)
	fmt.Printf("  Source IP: %s\n", sourceIP)
	return nil
}
