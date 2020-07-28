package main

import (
	"fmt"
	"webserver/server"
)

func main() {
	// 初始化集群的web服务器
	if err := server.Init(); err != nil {
		fmt.Println("cluster web server init failure, error is : ", err)
	}
}
