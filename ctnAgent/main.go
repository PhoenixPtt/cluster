package main

import (
	"ctnAgent/ctnA"
	"ctnCommon/headers"
	"ctnCommon/protocol"
	"fmt"
	"tcpSocket"
	"time"
)

var (
	serverAddrs []string
)

func main() {
	serverAddrs = append(serverAddrs, "192.168.1.155")
	for index, _:=range serverAddrs{
		tcpSocket.ConnectToHost(serverAddrs[index], 10000, "192.168.1.155", 0, myReceiveData, myStateChange)
	}

	ctnA.InitCtnMgr(mySendCtn, serverAddrs)

	fmt.Println("kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk")
	for {
		time.Sleep(time.Second*time.Duration(1))
	}
}

func myReceiveData(h string, pkgId uint16, i string, s []byte) {
	ReceiveDataFromServer(h, 1, pkgId, i, s)
}

func ReceiveDataFromServer(h string, level uint8, pkgId uint16, i string, s []byte) {
	fmt.Println("从客户端收取数据")

	pSaTruck := new(protocol.SA_TRUCK)
	err := headers.Decode(s, pSaTruck)
	if err != nil {
		fmt.Errorf(err.Error())
		return
	}
	pSaTruck.SrcAddr=append(pSaTruck.SrcAddr, h)

	//卸货，将服务器端收到的数据给容器管理器
	ctnA.Unload(pSaTruck)
}

func myStateChange(id string, mystring uint8) {
	fmt.Println(id, mystring)
	serverAddrs = append(serverAddrs, id)
}

//向Server端发送容器操作命令
func mySendCtn(ip string, level uint8, pkgId uint16, flag string, data []byte) {
	//fmt.Println("mySendCtn:", ip, G_id)
	tcpSocket.WriteData(ip, level, pkgId, flag, data)
}

type SendObjFunc func(ip string, level uint8, pkgId uint16, flag string, data []byte)
