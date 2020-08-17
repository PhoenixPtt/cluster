// main project main.go
package main

import (
	"clusterServer/clusterServer"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("集群服务端")

	clusterServer.Init()

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

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-quit

	clusterServer.Stop()
}
