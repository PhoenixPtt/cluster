// main project main.go
package main

import (
	"clusterServer/clusterServer"
	"fmt"
	"time"
)

func main() {
	fmt.Println("集群服务端")

	go clusterServer.Start()

	// 不退出 阻塞
	for {
		time.Sleep(time.Second)
	}
}
