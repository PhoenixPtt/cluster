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
	G_id string
)

func init() {
	ctnA.Config(mySendCtn)
}

func main() {
	//tcpSocket.ConnectToHost("192.168.43.100", 10000, "192.168.43.100", 0, myReceiveData, myStateChange)
	tcpSocket.ConnectToHost("192.168.1.166", 10000, "192.168.1.166", 0, myReceiveData, myStateChange)
	//tcpSocket.ConnectToHost("192.168.92.141", 10000, "192.168.92.141", 0, myReceiveData, myStateChange)

	//ctnA.InitCtnMgr()
	////启动容器事件监测
	//go ctnA.CtnEvents(G_id)
	//go ctnA.CtnInfoAll(G_id)
	//go ctnA.CtnStatsAll(G_id)
	//for {
	//	fmt.Println("请选择下列操作：")
	//	fmt.Println("1.实时上传所有容器资源使用状态")
	//	fmt.Println("2.取消实时上传所有容器资源使用状态")
	//	fmt.Println("3.实时上传所有容器信息")
	//	fmt.Println("4.取消实时上传所有容器信息")
	//
	//	//设置采样率
	//	ctnA.SetFreq(1)
	//	var val int
	//	fmt.Scanln(&val)
	//	switch val {
	//	case 1:
	//		go ctnA.CtnStatsAll(G_id)
	//	case 2:
	//		ctnA.CancelCtnStatsAll()
	//	case 3:
	//		go ctnA.CtnInfoAll(G_id)
	//	case 4:
	//		ctnA.CancelCtnInfoAll()
	//	}
	//
	//	time.Sleep(time.Second)
	//}
	////END:
	//fmt.Printf("主线程退出-------------------------------------------------------\n\n")
}

func myReceiveData(h string, pkgId uint16, i string, s []byte) {
	ReceiveDataFromServer(G_id, 1, pkgId, i, s)
}

func ReceiveDataFromServer(h string, level uint8, pkgId uint16, i string, s []byte) {
	fmt.Println("从客户端收取数据")

	pSaTruck := new(protocol.SA_TRUCK)
	err := headers.Decode(s, pSaTruck)
	if err != nil {
		fmt.Errorf(err.Error())
		return
	}
	pSaTruck.SrcAddr=h

	pRecvChan := ctnA.GetRecvChan()
	fmt.Println(pSaTruck)
	pRecvChan <- pSaTruck
}

func myStateChange(id string, mystring uint8) {
	fmt.Println(id, mystring)
	G_id = id
}

//向Server端发送容器操作命令
func mySendCtn(ip string, level uint8, pkgId uint16, flag string, data []byte) {
	//fmt.Println("mySendCtn:", ip, G_id)
	tcpSocket.WriteData(G_id, level, pkgId, flag, data)
}

type SendObjFunc func(ip string, level uint8, pkgId uint16, flag string, data []byte)
