package main

import (
	"fmt"
	"webserver/server"

	"clusterServer/clusterServer"
)

func main() {
	// 输出集群服务端启动信息，并启动集群服务端
	fmt.Println("集群服务端 准备启动...")
	go clusterServer.Start()
	fmt.Println("集群服务端 启动中...")

	// 初始化集群的web服务端
	if err := server.Init(); err != nil {
		fmt.Println("cluster web server init failure, error is : ", err)
	}
}
