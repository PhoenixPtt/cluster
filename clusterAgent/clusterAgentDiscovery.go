package main

import (
	"log"
	"net"
	"strconv"
	"tcpSocket"
	"time"
)

func clusterAgentDiscovery() {
	laddr, err := net.ResolveUDPAddr("udp", ":"+strconv.FormatUint(uint64(d.AgentUdpPort), 10))
	raddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+strconv.FormatUint(uint64(d.ServerUdpPort), 10))

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		log.Println("clusterAgentDiscovery connect fail !", err)
		return
	}
	defer conn.Close()

	sendData := []byte("I'm clusterAgent!")
	log.Println(time.Now(), "clusterAgentDiscovery", string(sendData))
	conn.WriteToUDP(sendData, raddr)

	data := make([]byte, 1024)
	for {
		read, raddr, err := conn.ReadFromUDP(data)
		if err != nil {
			continue
		}

		log.Println(time.Now(), "clusterAgentDiscovery ReadFromUDP", read, raddr.IP.String(), raddr.Port, string(data[:read]))
		if string(data[:read]) == string("ClusterServer is ready, please connect!") {
			tcpSocket.ConnectToHost(raddr.IP.String(), int(d.ServerTcpPortForListenAgent),"0.0.0.0", 0, onNetReadData, onNetStateChanged)
			//tcpSocket.ConnectToHost(raddr.IP.String(), int(d.ServerTcpPortForListenAgent),"0.0.0.0", int(d.AgentTcpPort), onNetReadData, onNetStateChanged)
		}
	}
}
