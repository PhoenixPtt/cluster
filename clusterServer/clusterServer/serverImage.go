package clusterServer

import (
	"log"
	"time"

	"clusterHeader"
	"tcpSocket"
)

func ReceiveDataFromAgent(handle string, pkgId uint16, sendbyte []byte) {

	log.Println("server端接收agent数据", handle, len(sendbyte))
	var imagedata header.ImageData
	err := header.Decode(sendbyte, &imagedata)
	if err != nil {
		log.Println("decode data false")
		return
	}

	//dealType := imagedata.DealType
	imageName := imagedata.ImageName
	tags := imagedata.Tags
	body := []byte(imagedata.ImageBody)
	result := imagedata.Result
	tiperr := imagedata.TipError
	//grade := tcpSocket.TCP_TPYE_CONTROLLER
	//
	//if dealType == header.FLAG_IMAG_REMO {
	//	grade = tcpSocket.TCP_TYPE_FILE
	//}

	_, ok := HandleFromPkgId(pkgId)
	if ok {
		//writeClientData(clientHandle, grade, header.FLAG_IMAG, sendbyte)
	} else {
		log.Println(time.Now().Nanosecond(), "server端接收agent镜像操作的返回信息,但是客户端IP丢失", imageName, tags, len(body), result)
	}

	log.Println("server端接收agent镜像操作的返回信息", imageName, tags, len(body), result , tiperr)
	// SERVER端将处理结果写入日志系统

}

func ProcessImageFlagDataFromClient(handle string, pkgId uint16, sendbyte []byte) {

	log.Println("server端接收client数据", handle, len(sendbyte))

	var imagedata header.ImageData
	err := header.Decode(sendbyte, &imagedata)
	if err != nil {
		log.Println("decode data false")
		return
	}

	dealType := imagedata.DealType
	var grade = tcpSocket.TCP_TPYE_CONTROLLER
	// log.Println("dealtype of ", dealType)

	switch dealType {
	case header.FLAG_IMAG_LIST, header.FLAG_IMAG_TGLS, header.FLAG_IMAG_REMO, header.FLAG_IMAG_UPDT: // 私有仓库操作
		//交给私有仓库处理
		registryReceiveDataFromClient(handle, pkgId, sendbyte)
	case header.FLAG_IMAG_DIST, header.FLAG_IMAG_PUSH, header.FLAG_IMAG_SAVE:
		grade = tcpSocket.TCP_TPYE_CONTROLLER
		writeAgentData("", grade, pkgId, header.FLAG_IMAG, sendbyte)
	case header.FLAG_IMAG_BUID, header.FLAG_IMAG_LOAD:
		grade = tcpSocket.TCP_TYPE_FILE
		writeAgentData("", grade, pkgId, header.FLAG_IMAG, sendbyte)
	default:
		log.Panicln("nothing match!")

	}

}
