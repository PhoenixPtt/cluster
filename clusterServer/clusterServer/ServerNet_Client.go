package clusterServer

import (
	"clusterHeader"
)

//func writeClientData(ip string, tcpType uint8, flag string, data []byte) {
//	if len(ip) <= 0 { // 如果不指定ip，则给所有ip发
//		for ipAddr, state := range clientStates {
//			if state == tcpSocket.TCP_CONNECT_SUCCESS {
//				tcpSocket.WriteData(ipAddr, tcpType, 0, flag, data)
//			}
//		}
//	} else { // 如果制定了ip则仅回复指定ip
//		tcpSocket.WriteData(ip, tcpType, 0, flag, data)
//	}
//}

//func onClientStateChanged(ip string, state uint8) {
//	if state == tcpSocket.TCP_CONNECT_SUCCESS {
//		clientStates[ip] = state
//		log.Println("Client:", ip, "Connected success!")
//	} else {
//		delete(clientStates, ip)
//		log.Println("Client:", ip, "DisConnected!")
//	}
//}

//func onClientReadData(ip string, pkgId uint16, flag string, data []byte) {
//	log.Println(time.Now().Nanosecond(), "接收到Client数据：", ip, flag, len(data))
//	switch flag {
//	case header.FLAG_CLST: //
//		fields := strings.Fields(string(data))
//		ok := false
//		if len(fields) > 1 {
//			val, err := strconv.ParseUint(fields[1], 10, 64)
//			if err == nil {
//				ok = true
//				d.ResSampleFeq = uint32(val)
//				writeAgentData("", tcpSocket.TCP_TPYE_CONTROLLER, 0, flag, data)
//				log.Println("Client:", ip, "设置资源采样频率为", val, "秒")
//			}
//		}
//
//		// 如果设置错误
//		if ok {
//			log.Println("Client:", ip, "设置资源采样频率参数格式错误！", string(data))
//		}
//
//	case header.FLAG_IMAG:
//		pkgId := GetPkgId()
//		PkgIdMap[pkgId] = ip
//		ProcessImageFlagDataFromClient(ip, pkgId, data)
//
//	case header.FLAG_FILE:
//		//turn byte to tar
//		ioReader := bytes.NewBuffer(data)
//		pwdStr := "/tmp/cluster/"
//		err := targz.Tar(ioReader, pwdStr+"test.tar")
//		if err != nil {
//			//use exec to tar file
//			os.Chdir(pwdStr)
//			err := header.TarFile("test")
//			if err != nil {
//				return
//			}
//		}
//
//
//		// 节点相关
//	case header.FLAG_NODE:
//
//		// 应用服务相关
//	case header.FLAG_SERV:
//		pkgId := GetPkgId()
//		PkgIdMap[pkgId] = ip
//		ProcessTaskFlagDataFromClient(ip, pkgId, data)
//
//		// 容器操作
//	case header.FLAG_CTNS:
//		//pkgId := GetPkgId()
//		//PkgIdMap[pkgId] = ip
//		//ProcessContainerFlagDataFromClient(ip, pkgId, data)
//	}
//	return
//}

func returnResultToClient(pkgId uint16, imagedata header.ImageData) {

	//log.Println("结果返回客户端")
	//dealType := imagedata.DealType
	//sendbyte, err := header.Encode(imagedata)
	//if err != nil {
	//	log.Println("encode data err", err)
	//}
	//log.Println("agent端返回给server端的数据", dealType, imagedata)
	//var grade = tcpSocket.TCP_TPYE_CONTROLLER
	//if dealType == header.FLAG_IMAG_SAVE {
	//	grade = tcpSocket.TCP_TYPE_FILE
	//}
	//
	//clientHandle, ok := PkgIdMap[pkgId]
	//if ok {
	//	writeClientData(clientHandle, grade, header.FLAG_IMAG, sendbyte)
	//} else {
	//	log.Println(time.Now().Nanosecond(), "server端接收agent镜像操作的返回信息,但是客户端IP丢失")
	//}
}
