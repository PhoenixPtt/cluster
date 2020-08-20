package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"tcpSocket"

	"clusterAgent/agentImage"
	"ctnAgent/ctnA"
)

func main() {
	fmt.Println("Agent端运行")

	// 启动容器引擎
	//启动容器事件监测
	ctnA.Config(writeData)
	go ctnA.CtnEvents("")   // 启动容器事件监测

	agentImage.ImageInit()

	// 发出广播通知server连接
	go clusterAgentDiscovery()

	// 设置采样频率，并启动监控
	StartMonitor()

	// 退出时关闭网络
	defer tcpSocket.AbortAll()
	defer StopMonitor()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-quit
}
