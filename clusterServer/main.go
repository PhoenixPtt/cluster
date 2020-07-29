// main project main.go
package main

import (
	header "clusterHeader"
	"clusterServer/clusterServer"
	"fmt"
	"time"
)

func main() {
	fmt.Println("集群服务端")

	go clusterServer.Start()

	// 3秒后测试网络请求
	time.Sleep(time.Second*3)

	var imageData header.ImageData
	imageData.DealType = header.FLAG_IMAG_LIST

	var respChan chan interface{} = make(chan interface{}, 1)

	clusterServer.ResponseURL(header.FLAG_IMAG, "", imageData, respChan)

	respDataInterface := <-respChan

	respData := respDataInterface.(header.ImageData)

	fmt.Printf("%#v",respData)


	// 不退出 阻塞
	for {
		time.Sleep(time.Second)
	}
}
