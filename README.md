[中文文档](README.zh-CN.md)

# nser - IPv6 Neighbor Solicitation Tool

`nser` is a command-line utility for network diagnostics and exploration, specifically designed for crafting and sending IPv6 Neighbor Solicitation (NS) packets. It's built in Go and uses the `gopacket` library.

This tool is useful for network administrators and engineers to troubleshoot IPv6 network connectivity, verify neighbor discovery configurations, and explore the behavior of network devices.

Some servers have restricted IPv6 environments. The server must send a Neighbor Solicitation (NS) request to the IPv6 gateway using the IPv6 address as the source address. Only after this will the gateway forward packets from that IPv6 address, allowing the server to communicate externally with it.
As a result, the server’s default IPv6 address is reachable, but newly added IPv6 addresses are not.
This tool can easily solve such problems.

## Features

-   **Manual Mode:** Craft and send a custom Neighbor Solicitation packet by specifying the source IP, target IP, and network interface.
-   **Automatic Gateway Mode:** Automatically discover the default IPv6 gateway and send NS packets from all available IPv6 addresses on a specified interface. This is ideal for quickly testing connectivity to the gateway.

## Prerequisites

-   Go (1.23 or later)

## Building

1.  Clone the repository:
    ```sh
    git clone https://github.com/kkqy/nser.git
    cd nser
    ```

2.  Build the executable:
    ```sh
    go build
    ```

## Usage

The tool requires elevated privileges to capture and send network packets.

### 1. Manual Mode

Send a single NS packet with a specific source and destination.

*   **Command:**
    ```sh
    ./nser -iface "<interface_name>" -src "<your_source_ipv6>" -dst "<target_ipv6>"
    ```
*   **Example:**
    ```sh
    # On Windows
    ./nser.exe -iface "Ethernet" -src "fe80::1" -dst "fe80::2"

    # On Linux
    sudo ./nser -iface "eth0" -src "fe80::1" -dst "fe80::2"
    ```

### 2. Automatic Gateway Mode

Automatically find the IPv6 gateway and send NS packets to it from every IPv6 address on the specified interface.

*   **Command:**
    ```sh
    ./nser -iface "<interface_name>" -gateway
    ```
*   **Example:**
    ```sh
    # On Windows
    ./nser.exe -iface "Ethernet" -gateway

    # On Linux
    sudo ./nser -iface "eth0" -gateway
    ```
