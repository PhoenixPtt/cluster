package main

import (
	"ctnAgent/ctnA"
	"ctnCommon/headers"
	"ctnCommon/protocol"
	"fmt"
	"tcpSocket"
	"time"
)

func main() {
	var serverAddrs []string = []string{"192.168.1.155"} //agent可能为多台server服务
	var agentAddr string = "192.168.1.155"
	//初始化容器管理器
	ctnA.InitCtnMgr(mySendCtn, agentAddr)

	for _, serverAddr := range serverAddrs {
		ctnA.AddServer(serverAddr)
		tcpSocket.ConnectToHost(serverAddr, 10000, agentAddr, 0, myReceiveData, myStateChange)
	}

	for {
		time.Sleep(time.Second * time.Duration(1))
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
	pSaTruck.SrcAddr = h

	//卸货，将服务器端收到的数据给容器管理器
	ctnA.Unload(pSaTruck)
}

func myStateChange(id string, mystring uint8) {
	fmt.Println(id, mystring)

	//判断服务器ip是否在容器管理器列表中
	if ctnA.G_ctnMgr.IsServerExisted(id) {
		switch mystring {
		case 1: //在线
			ctnA.G_ctnMgr.UpdateServerOnlineStatus(id, true)
		case 2: //离线
			ctnA.G_ctnMgr.UpdateServerOnlineStatus(id, false)
		}
	}
}

//向Server端发送容器操作命令
func mySendCtn(ip string, level uint8, pkgId uint16, flag string, data []byte) {
	//fmt.Println("mySendCtn:", ip, G_id)
	tcpSocket.WriteData(ip, level, pkgId, flag, data)
}

type SendObjFunc func(ip string, level uint8, pkgId uint16, flag string, data []byte)
