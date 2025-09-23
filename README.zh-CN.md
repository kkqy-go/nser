# nser - IPv6 邻居请求工具

`nser` 是一个用于网络诊断和探索的命令行工具，专门用于制作和发送 IPv6 邻居请求 (Neighbor Solicitation, NS) 数据包。它使用 Go 语言构建，并依赖 `gopacket` 库。

该工具可帮助网络管理员和工程师排查 IPv6 网络连接问题、验证邻居发现配置以及探索网络设备的行为。

一些服务器的IPv6环境受限制，服务器必须以IPv6地址作为源地址向IPv6网关发送NS请求以后，IPv6网关才会转发这个IPv6的数据包，然后服务器才可以通过这个IPv6地址对外通信。
导致的现象是，服务器的默认IPv6地址能连通，但新添加的IPv6地址无法连通。
使用本工具能很方便地解决这种问题。

## 功能

-   **手动模式:** 通过指定源 IP、目标 IP 和网络接口，制作并发送自定义的邻居请求数据包。
-   **自动网关模式:** 自动发现默认的 IPv6 网关，并从指定接口上的所有可用 IPv6 地址向其发送 NS 数据包。此模式非常适合快速测试与网关的连接性。

## 环境要求

-   Go (1.23 或更高版本)

## 构建

1.  克隆仓库：
    ```sh
    git clone https://github.com/kkqy/nser.git
    cd nser
    ```

2.  构建可执行文件：
    ```sh
    go build
    ```

## 使用方法

该工具需要提升权限（管理员或 root）才能捕获和发送网络数据包。

### 1. 手动模式

发送一个具有特定源和目标的 NS 数据包。

*   **命令:**
    ```sh
    ./nser -iface "<接口名称>" -src "<你的源_ipv6>" -dst "<目标_ipv6>"
    ```
*   **示例:**
    ```sh
    # 在 Windows 上
    ./nser.exe -iface "Ethernet" -src "fe80::1" -dst "fe80::2"

    # 在 Linux 上
    sudo ./nser -iface "eth0" -src "fe80::1" -dst "fe80::2"
    ```

### 2. 自动网关模式

自动查找 IPv6 网关，并从指定接口上的每个 IPv6 地址向其发送 NS 数据包。

*   **命令:**
    ```sh
    ./nser -iface "<接口名称>" -gateway
    ```
*   **示例:**
    ```sh
    # 在 Windows 上
    ./nser.exe -iface "Ethernet" -gateway

    # 在 Linux 上
    sudo ./nser -iface "eth0" -gateway
    ```
