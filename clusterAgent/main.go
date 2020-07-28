package main

import (
	"fmt"
	"time"

	"tcpSocket"

	"clusterAgent/agentImage"
)

func main() {
	fmt.Println("Agent端运行")

	// 启动容器引擎
	//启动容器事件监测
	go CtnEvents()   // 启动容器事件监测
	agentImage.ImageInit()

	// 发出广播通知server连接
	go clusterAgentDiscovery()

	// 设置采样频率，并启动监控
	StartMonitor()

	// 退出时关闭网络
	defer tcpSocket.AbortAll()
	defer StopMonitor()

	// 不退出 阻塞
	for {
		time.Sleep(time.Second)
	}
}
