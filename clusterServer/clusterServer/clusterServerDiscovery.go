package clusterServer

import (
	"log"
	"net"
	"strconv"
	"time"
)

func clusterServerDiscovery() {
	laddr, err := net.ResolveUDPAddr("udp", ":"+strconv.FormatUint(uint64(d.ServerUdpPort), 10))
	raddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+strconv.FormatUint(uint64(d.AgentUdpPort), 10))

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		log.Println("clusterServerDiscovery Listen UDP error:", err)
		panic(err)
	}
	defer conn.Close()

	sendData := []byte("ClusterServer is ready, please connect!")
	log.Println(time.Now(), "clusterServerDiscovery", string(sendData))
	conn.WriteToUDP(sendData, raddr)

	data := make([]byte, 1024)
	for {
		read, raddr, _ := conn.ReadFromUDP(data)
		if read == 0 {
			continue
		}

		log.Println(time.Now(), "clusterServerDiscovery ReadFromUDP", read, raddr.IP.String(), raddr.Port, string(data[:read]))

		// 当接收到Agent发送的我是客户端消息后，服务端立即回复
		if string(data[:read]) == string("I'm clusterAgent!") {
			conn.WriteToUDP(sendData, raddr)
		}
	}

}
