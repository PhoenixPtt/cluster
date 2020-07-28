// main project main.go
package clusterServer

import (
	"fmt"
	"tcpSocket"
	"time"
)

var netListenHandle_Agent string  // 监听Agent的网络句柄
//var netListenHandle_Client string // 监听客户端的网络句柄

var nodes Nodes

func init() {
	nodes.Init()
}

func Start() {
	// 加载配置
	LoadCfg()

	fmt.Println("d.ServerTcpPortForListenAgent", d.ServerTcpPortForListenAgent)

	// 监听 Agent 连接
	netListenHandle_Agent = tcpSocket.Listen("0.0.0.0", int(d.ServerTcpPortForListenAgent), onAgentReadData, onAgentStateChanged)

	// 监听 Client 连接
	//netListenHandle_Client = tcpSocket.Listen("0.0.0.0", int(d.ServerTcpPortForListenClient), onClientReadData, onClientStateChanged)

	// 启动网络发现
	go clusterServerDiscovery()

	defer Quit()

	// 不退出 阻塞
	for {
		time.Sleep(time.Second)
	}
}

// 退出集群
func Quit() {

	//for ip, index := range NodeIds {
	//	node := &d.Node[index]
	//	if node.State == tcpSocket.TCP_CONNECT_SUCCESS {
	//		tcpSocket.Abort(ip)
	//	}
	//}

	tcpSocket.StopListen(netListenHandle_Agent)
	//tcpSocket.StopListen(netListenHandle_Client)

}
