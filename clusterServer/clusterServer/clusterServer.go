// main project main.go
package clusterServer

import (
	"fmt"
	"tcpSocket"
	"time"
)

var netListenHandle_Agent string  // 监听Agent的网络句柄
//var netListenHandle_Client string // 监听客户端的网络句柄

var isRunning bool = false

func Init() error {
	return Start()
}

func Start() error{
	// 启动集群之前必须先退出集群
	if isRunning {
		Stop()
	}

	// 标志系统已经开始运行
	isRunning = true

	// 加载配置
	LoadCfg()

	// 监听 Agent 连接
	netListenHandle_Agent = tcpSocket.Listen("0.0.0.0", int(d.ServerTcpPortForListenAgent), onAgentReadData, onAgentStateChanged)

	// 监听 Client 连接
	//netListenHandle_Client = tcpSocket.Listen("0.0.0.0", int(d.ServerTcpPortForListenClient), onClientReadData, onClientStateChanged)

	// 启动网络发现
	go clusterServerDiscovery()

	// 自动按周期监控资源
	go autoUpdateClusterStats()

	go startController()

	return nil
}

// 关闭集群
func Stop() {

	// 如果系统还未开始运行，则直接返回，不执行任何退出操作
	if isRunning == false {
		return
	} else {
		isRunning = false
	}

	stopController()

	// 关闭所有在线节点的网络连接
	for _,h := range nodes.GetNodeIds() {
		if nodes.GetState(h) == true {
			tcpSocket.Abort(h)
		}
	}

	// 停止监听
	tcpSocket.StopListen(netListenHandle_Agent)
	//tcpSocket.StopListen(netListenHandle_Client)

	fmt.Println("-----------------------stop.................")
}

// 自动更新集群状态
func autoUpdateClusterStats() {
	// 周期更新状态
	for isRunning {
		// 睡眠
		time.Sleep(time.Second* time.Duration(d.ResSampleFeq))
		// 更新一次状态
		updateClusterStats()
	}
}

func startController() {
	g_controller.Start()
}

func stopController() {
	g_controller.Stop()
}

