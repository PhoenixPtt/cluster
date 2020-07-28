// main project main.go
package clusterServer

import (
	"fmt"
	"tcpSocket"
	"time"
)

var netListenHandle_Agent string  // 监听Agent的网络句柄
//var netListenHandle_Client string // 监听客户端的网络句柄


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
		// 睡眠
		time.Sleep(time.Second* time.Duration(d.ResSampleFeq))

		updateClusterStats()
	}
}

// 退出集群
func Quit() {

	// 关闭所有在线节点的网络连接
	for _,h := range nodes.GetNodeIds() {
		if nodes.GetState(h) == true {
			tcpSocket.Abort(h)
		}
	}

	// 停止监听
	tcpSocket.StopListen(netListenHandle_Agent)
	//tcpSocket.StopListen(netListenHandle_Client)

}
