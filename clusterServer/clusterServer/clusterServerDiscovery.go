package clusterServer

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

func clusterServerDiscovery() {
	laddr, err := net.ResolveUDPAddr("udp", ":"+strconv.FormatUint(uint64(d.ServerUdpPort), 10))
	raddr, err := net.ResolveUDPAddr("udp", "192.168.1.255:"+strconv.FormatUint(uint64(d.AgentUdpPort), 10))

	fmt.Println("clusterServerDiscovery()" , raddr.IP.String())

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05.000000"), "clusterServerDiscovery Listen UDP error:", err)
		return
	}
	defer conn.Close()

	sendData := []byte("ClusterServer is ready, please connect!")
	fmt.Println(time.Now().Format("2006-01-02 15:04:05.000000"), "clusterServerDiscovery", string(sendData))
	conn.WriteToUDP(sendData, raddr)

	data := make([]byte, 1024)
	for isRunning {
		read, raddr, _ := conn.ReadFromUDP(data)
		if read == 0 {
			continue
		}

		fmt.Println(time.Now().Format("2006-01-02 15:04:05.000000"), "clusterServerDiscovery ReadFromUDP", read, raddr.IP.String(), raddr.Port, string(data[:read]))

		// 当接收到Agent发送的我是客户端消息后，服务端立即回复
		if string(data[:read]) == "I'm clusterAgent!" {
			conn.WriteToUDP(sendData, raddr)
		}
	}

}
