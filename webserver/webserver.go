package main

import (
	"fmt"
	"webserver/server"
)

func main() {
	// 初始化集群的web服务端
	fmt.Println("集群web服务端 启动")
	if err := server.Init(); err != nil {
		fmt.Println("cluster web server init failure, error is : ", err)
		return
	}

	// 等待系统结束命令
	server.WaitForInterruptSignal()

	// 停止集群的web服务端
	fmt.Println("集群web服务端 正在停止...")
	server.Stop()


	fmt.Println("集群web服务端 调试程序已关闭")
}
