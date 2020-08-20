package main

import (
	"clusterAgent/agentImage"
	"clusterAgent/servers"
	header "clusterHeader"
	"ctnAgent/ctnA"
	"ctnCommon/ctn"
	"ctnCommon/headers"
	"encoding/json"
	"fmt"
	"log"
	"tcpSocket"
	"time"
)

//func SendContainerEvents(c <-chan events.Message) {
//	for {
//		v, ok := <-c
//		if ok {
//
//			ctno := header.CTNO{
//				OperType:  "CEVENT",
//				OperIndex: -1,
//				CtnEvtMsg: v,
//			}
//
//			byteStream, err := header.Encode(ctno)
//			if err != nil {
//				fmt.Println(err.Error())
//			}
//			writeData("", tcpSocket.TCP_TPYE_CONTROLLER, 0, "CTNM", byteStream)
//		} else {
//			break
//		}
//	}
//}

func onNetStateChanged(h string, state uint8) {
	if state == tcpSocket.TCP_CONNECT_SUCCESS {
		servers.Add(h)
		log.Println("Server:", h, "Connected success!")
	} else {
		s := servers.GetServer(h)
		if s != nil {
			s.SetState(false)
		}
		log.Println("Server:", h, "DisConnected!")
	}
}

func onNetReadData(ip string, pkgId uint16, flag string, data []byte) {
	log.Println(time.Now().Nanosecond(), "接收到Server数据：", ip, "flag=", flag, "dataLen=", len(data))

	switch flag {
	//case header.FLAG_HOST:
	//	hostInfo := SysResMonitor.GetHostInfo()
	//	writeData(ip, tcpSocket.TCP_TYPE_MONITOR, pkgId, flag, []byte(hostInfo.String()))
	//case header.FLAG_RFRE:
	//	fields := strings.Fields(string(data))
	//	if len(fields) > 1 {
	//		val, _ := strconv.Atoi(fields[1])
	//		resSampleFeq := time.Second * time.Duration(val)
	//		writeData("", tcpSocket.TCP_CONNECT_SUCCESS, pkgId, flag, data)
	//		SetRefreshFreq(resSampleFeq)
	//		log.Println("Client:", ip, "设置资源采样频率为", val, "秒")
	//	} else {
	//		// 错误
	//		log.Println("Client:", ip, "设置资源采样频率参数格式错误！", string(data))
	//	}

	case header.FLAG_CLST:
		var clst header.CLST
		json.Unmarshal(data, clst)
		switch clst.Oper.Type {
		case header.FLAG_CLST_CFG:
			if clst.Cfg.Name != "" {
				d.Name = clst.Cfg.Name
			}
			if clst.Cfg.TaskMigrateTimeFromAgent > 0 {
				d.TaskMigrateTimeFromAgent = clst.Cfg.TaskMigrateTimeFromAgent
			}
			if clst.Cfg.ResSampleFeq > 0 {
				d.ResSampleFeq = clst.Cfg.ResSampleFeq
			}

			var ans header.MESG
			ans.Oper.Type = header.FLAG_CLST
			ans.Oper.Progress = 100
			ans.Oper.Success = true
			ans.SubFlag = header.FLAG_CLST_CFG
			writeData("", tcpSocket.TCP_TPYE_CONTROLLER, pkgId, header.FLAG_MESG, header.JsonByteArray(ans) )
		default:
		}
	case header.FLAG_NODE:
		var nd header.NODE
		json.Unmarshal(data, nd)
		switch nd.Oper.Type {
		case header.FLAG_NODE_HOST:
			GetHostInfo()
		default:
		}
	case header.FLAG_IMAG:
		var imageData header.ImageData
		json.Unmarshal(data, &imageData)
		agentImage.RecieveDataFromServer(ip, pkgId, imageData)

	case header.FLAG_CTNS: // 容器相关
		// 处理接收到的任务相关的数据，然后返回结果
		pSaTruck := new(ctn.SA_TRUCK)
		err := headers.Decode(data, pSaTruck)
		if err != nil {
			fmt.Errorf(err.Error())
			return
		}

		pRecvChan := ctnA.GetRecvChan()
		fmt.Println(pSaTruck)
		pRecvChan <- pSaTruck

	}

}

func writeData(ip string, tcpType uint8, pkgId uint16, flag string, data []byte) {
	if len(ip) <= 0 { // 如果不指定ip，则给所有ip发
		for _,h := range servers.GetOnlineServerHandles() {
			tcpSocket.WriteData(h, tcpType, pkgId, flag, data)
		}
	} else { // 如果指定了ip则仅回复指定ip
		tcpSocket.WriteData(ip, tcpType, pkgId, flag, data)
	}
}

//func ProcessTaskFlag(h string, level uint8, pkgId uint16, i string, s []byte) {
//	//按照下面的结构体对网络数据解析
//	cctn := new(ctn.CTN)
//	err := header.Decode(s, cctn)
//	if err != nil {
//		fmt.Errorf(err.Error())
//		return
//	}
//
//	if (ctn.CtnIndex(cctn.ID, ctn.ALL_CTN) == -1) && (cctn.OperFlag != "CREATE") {
//		cctn.Err = header.Str2bytes(fmt.Sprintf("容器ID：%s不存在。\n", cctn.ID))
//
//		byteStream, _ := header.Encode(cctn)
//		tcpSocket.WriteData(h, level, pkgId, i, byteStream)
//
//		return //直接返回错误信息
//	}
//
//	var logs io.ReadCloser
//	var ctnInspect header.CTN_INSPECT
//
//	switch cctn.OperFlag {
//	case ctn.CREATE: //创建容器
//		err = cctn.Create()
//		// 如果容器创建成功，则添加到当前集群的容器列表中
//		if err == nil {
//			//addServerCtn(h, cctn.ID, false)
//		}
//	case ctn.START: //启动容器
//		err = cctn.Start()
//		if err == nil {
//			//updateClusterCtn(h, cctn.ID, true)
//		}
//	case ctn.STOP: //停止容器
//		err = cctn.Stop()
//		if err == nil {
//			//updateClusterCtn(h, cctn.ID, false)
//		}
//	case ctn.KILL: //强制停止容器
//		err = cctn.Kill()
//		if err == nil {
//			//updateClusterCtn(h, cctn.ID, false)
//		}
//	case ctn.REMOVE: //删除容器
//		err = cctn.Remove()
//		if err == nil {
//			//removeClusterCtn(h, cctn.ID)
//		}
//	case ctn.GETLOG: //获取容器日志
//		logs, err = cctn.GetLog()
//		cctn.Logs, err = ioutil.ReadAll(logs)
//	case ctn.INSPECT: //获取容器信息
//		ctnInspect, err = cctn.Inspect()
//		cctn.CtnInspect = ctnInspect
//	}
//
//	if err != nil {
//		cctn.Err = header.Str2bytes(err.Error())
//	} else {
//		cctn.Err = header.Str2bytes("nil")
//	}
//
//	byteStream, err := header.Encode(cctn)
//	tcpSocket.WriteData(h, level, pkgId, i, byteStream)
//}
