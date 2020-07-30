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

	// 测试网络请求

	//var respChan chan interface{}

	//time.Sleep(time.Second*1)
	//var imageData header.ImageData
	//imageData.DealType = header.FLAG_IMAG_LIST
	//respChan = make(chan interface{}, 1)
	//clusterServer.ResponseURL(header.FLAG_IMAG, "", imageData, respChan)
	//respDataInterface := <-respChan
	//respData := respDataInterface.(header.ImageData)
	//fmt.Printf("%#v",respData)

	//time.Sleep(time.Second*1)
	//var nodeData header.NODE
	//nodeData.Oper.Type = header.FLAG_NODE
	//respChan = make(chan interface{}, 1)
	//clusterServer.ResponseURL(header.FLAG_NODE, "", nodeData, respChan)
	//respNodeDataInterface := <-respChan
	//respNodeData := respNodeDataInterface.(header.NODE)
	//fmt.Printf("%#v",respNodeData)

	// 不退出 阻塞
	for {
		time.Sleep(time.Second)
	}
}
