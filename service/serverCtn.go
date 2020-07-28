package clusterServer

import (
	header "clusterHeader"
	"clusterServer/ctn"
	"time"

	"fmt"

	"github.com/docker/docker/api/types/events"

	"clusterServer/method"
	"sync"
	"tcpSocket"
)

const (
	FLAG_CTRL  = "CTRL"
	FLAG_CTN   = "INFO"
	FLAG_STATS = "STAT"
	FLAG_EVENT = "EVTM"
)

var (
	cluster method.CLUSTER
	exit    bool
	mMutex  sync.Mutex
)

func init() {
}

func CtnInfo() {
	exit = true
	for exit {
		for index, _ := range ctn.Ctn_index_map {
			pCtn:=ctn.GetCtnFromIndex(index)
			fmt.Println("容器信息监控:", pCtn.ID, pCtn.State, pCtn.OperFlag, pCtn.OperIndex, pCtn.Err)
		}
		time.Sleep(time.Second)
	}
}

//容器操作
//func ProcessContainerFlagDataFromClient(ip string, pkgId uint16, data []byte) {
//
//	fmt.Println("容器操作列表：")
//	fmt.Println("1.创建容器")
//	fmt.Println("2.启动容器")
//	fmt.Println("3.停止容器")
//	fmt.Println("4.强制停止容器")
//	fmt.Println("5.删除容器")
//	fmt.Println("6.获取容器详细信息")
//	fmt.Println("7.获取容器内日志")
//	fmt.Println("8.启动监控")
//	fmt.Println("9.停止监控")
//	fmt.Println("10.退出")
//	fmt.Printf("请输入容器操作序号：")
//
//	var index int
//	fmt.Scanln(&index)
//
//	pCtn := &ctn.CTN{}
//
//	var service method.SERVICE
//	service.ServiceName = method.NO_SPECIFIED_SERVICE
//	cluster.AddService(&service)
//
//	switch index {
//	case 1: //创建容器
//		{
//			mapAddr := make(map[int]string, 50)
//			var addrIndex int
//			for {
//				var i int = 0
//				for {
//					i = 0
//					for index, _ := range cluster.AgentMap {
//						i++
//						fmt.Printf("%d:%s\n", i, index)
//						mapAddr[i] = index
//					}
//					fmt.Println("请输入agent端IP地址序号：\n")
//
//					fmt.Scanln(&addrIndex)
//					_, ok := mapAddr[addrIndex]
//					if ok {
//						goto TT
//					} else {
//					}
//				}
//			}
//		TT:
//			fmt.Println("创建容器:")
//			fmt.Println("请输入镜像名称：")
//			var imageName string
//			fmt.Scanln(&imageName)
//			fmt.Printf("%s\n", imageName)
//			pCtn.Image = imageName
//			pCtn.AgentAddr = mapAddr[addrIndex]
//			pCtn.PrepareData(ctn.CREATE)
//			ctn.AddCtn(*pCtn)
//		}
//	case 2: //启动容器
//		{
//			ctnId := chooseCtnId()
//			if ctnId == "" {
//				break
//			}
//
//			fmt.Printf("启动容器：%s\n", ctnId)
//			pCtn = ctn.GetCtn(ctnId)
//			pCtn.PrepareData(ctn.START)
//		}
//	case 3: //停止容器
//		{
//			ctnId := chooseCtnId()
//			if ctnId == "" {
//				break
//			}
//			fmt.Printf("停止容器：%s\n", ctnId)
//			pCtn = ctn.GetCtn(ctnId)
//			pCtn.PrepareData(ctn.STOP)
//		}
//	case 4: //强制停止容器
//		{
//			ctnId := chooseCtnId()
//			if ctnId == "" {
//				break
//			}
//			fmt.Printf("强制停止容器：%s\n", ctnId)
//			pCtn = ctn.GetCtn(ctnId)
//			pCtn.PrepareData(ctn.KILL)
//		}
//	case 5: //删除容器
//		{
//			ctnId := chooseCtnId()
//			if ctnId == "" {
//				break
//			}
//			fmt.Printf("删除容器：%s\n", ctnId)
//			pCtn = ctn.GetCtn(ctnId)
//			pCtn.PrepareData(ctn.REMOVE)
//		}
//	case 6: //获取容器信息
//		{
//			ctnId := chooseCtnId()
//			if ctnId == "" {
//				break
//			}
//			fmt.Printf("获取容器详细信息：%s\n", ctnId)
//			pCtn = ctn.GetCtn(ctnId)
//			pCtn.PrepareData(ctn.INSPECT)
//		}
//	case 7: //获取容器内日志
//		{
//			ctnId := chooseCtnId()
//			if ctnId == "" {
//				break
//			}
//			fmt.Printf("获取容器内日志：%s\n", ctnId)
//			pCtn = ctn.GetCtn(ctnId)
//			pCtn.PrepareData(ctn.GETLOG)
//		}
//	case 8: //启动监控
//		{
//			go CtnInfo()
//		}
//	case 9: //停止监控
//		{
//			CancelCtnInfo()
//		}
//
//	}
//
//	send(pCtn.AgentAddr, *pCtn)
//}

func chooseCtnId() string {
	for {
		var sliceAddr []string
		sliceAddr = make([]string, 0, 50)

		for _,val:=range ctn.Ctn_index_map{
			sliceAddr = append(sliceAddr, val.CtnId)
		}

		for i, value := range sliceAddr {
			fmt.Printf("%d:%s\n", i, value)
		}
		fmt.Println("\n请输入容器序号：\n")

		var addrIndex int
		fmt.Scanln(&addrIndex)
		if addrIndex < len(sliceAddr[addrIndex]) {
			return sliceAddr[addrIndex]
		} else {
			return ""
		}
	}
}


func ProcessTaskFlagDataFromAgent(h string, pkgId uint16, i string, s []byte) {
	switch i {
	case FLAG_CTRL: //容器操作控制信息
		goto CTRL
	case FLAG_CTN: //容器信息
		var cctn header.CTN
		err := header.Decode(s, &cctn)
		if err != nil {
			fmt.Println(err.Error())
		}

		var p_sctn *header.CTN
		p_sctn = ctn.GetCtn(cctn.ID)
		if p_sctn != nil {
			p_sctn.Container = cctn.Container
		}

		fmt.Printf("%#v", p_sctn)

		////显示下收到的信息
		//fmt.Println("容器信息", cctn.ID)
		////更新容器信息
		//for index, val := range ctn.Ctn_index_map {
		//	if cctn.ID == val.CtnId {
		//		pCtn:=ctn.GetCtnFromIndex(index)
		//		pCtn.Container=cctn.Container
		//	}
		//}
		////////////////////////////////////////////////////////////////////////////////////
		//需补充的内容
		//转发容器信息给B端

		////////////////////////////////////////////////////////////////////////////////////
	case FLAG_STATS: //容器资源使用状态
		var ctnStats header.CTN_STATS
		err := header.Decode(s, &ctnStats)
		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Println(ctnStats.ID, ctnStats.Read, ctnStats.CPUUsageCalc, ctnStats.PercpuUsageCalc)
		//更新容器资源使用情况
		////////////////////////////////////////////////////////////////////////////////////
		//需补充的内容
		//转发容器资源使用状态给B端

		////////////////////////////////////////////////////////////////////////////////////
	case FLAG_EVENT: //容器事件
		var events events.Message

		err := header.Decode(s, &events)
		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Printf("容器事件信息；%#v", events)
		////////////////////////////////////////////////////////////////////////////////////
		//需补充的内容
		//转发容器事件给B端

		////////////////////////////////////////////////////////////////////////////////////
	}

CTRL:
	var cctn header.CTN
	err := header.Decode(s, &cctn)
	if err != nil {
		fmt.Errorf(err.Error())
	}
	switch cctn.OperFlag {
	case ctn.CREATE:
		{
			//发和收的容器结构体中的index相匹配
			//if ctn.IsCtnExisted(cctn.OperIndex) { //由于是指令应答模式，因此如果不匹配的话，则收到的包丢弃
			//通过映射关系找到结构体,更新容器信息
			ctn.UpdateCtn(cctn)

			pCtn:=ctn.GetCtnFromIndex(cctn.OperIndex)
			fmt.Printf("\n创建容器应答\nOperIndex:%d\n容器ID:%s\n错误信息：%s\n",
				pCtn.OperIndex,pCtn.ID,pCtn.Err)

			//mMutex.Unlock()
			//如果容器有服务名，则判断服务当前执行状态
			if cctn.ServiceName != "" {
				service := cluster.GetService(cctn.ServiceName) //通过服务名称获取服务
				//遍历服务中所有容器，判断容器执行状态
				rlt := true
				//mMutex.Lock()
				for _, val := range ctn.GetCtnsInService(service.ServiceName) {
					if (*val).Err != "nil" {
						rlt = false
					}
				}
				if rlt {
					fmt.Printf("创建服务%s成功！\n", service.ServiceName)
				}
				//mMutex.Unlock()
			}
		}
		//}
	case ctn.START:
		{
			//通过映射关系找到结构体，更新容器信息
			ctn.UpdateCtn(cctn)
			pCtn:=ctn.GetCtnFromIndex(cctn.OperIndex)
			fmt.Printf("\n启动容器应答\nOperIndex:%d\n容器ID:%s\n错误信息：%s\n",
				pCtn.OperIndex,pCtn.ID[:10],pCtn.Err)
			//如果容器有服务名，则判断服务当前执行状态
			if cctn.ServiceName != "" {
				service := cluster.GetService(cctn.ServiceName) //通过服务名称获取服务
				//遍历服务中所有容器，判断容器执行状态
				rlt := true
				for _, val := range ctn.GetCtnsInService(service.ServiceName) {
					if (*val).Err != "nil" {
						rlt = false
					}
				}
				if rlt {
					fmt.Printf("启动服务%s成功！\n", service.ServiceName)
				}
			}
		}
	case ctn.STOP:
		{
			//通过映射关系找到结构体，更新容器信息
			ctn.UpdateCtn(cctn)
			pCtn:=ctn.GetCtnFromIndex(cctn.OperIndex)
			fmt.Printf("\n停止容器应答\nOperIndex:%d\n容器ID:%s\n错误信息：%s\n",
				pCtn.OperIndex,pCtn.ID[:10],pCtn.Err)
			//如果容器有服务名，则判断服务当前执行状态
			if cctn.ServiceName != "" {
				service := cluster.GetService(cctn.ServiceName) //通过服务名称获取服务
				//遍历服务中所有容器，判断容器执行状态
				rlt := true
				for _, val := range ctn.GetCtnsInService(service.ServiceName) {
					if (*val).Err != "nil" {
						rlt = false
					}
				}
				if rlt {
					fmt.Printf("停止服务%s成功！\n", service.ServiceName)
				}
			}
		}

	case ctn.KILL:
		{
			//通过映射关系找到结构体，更新容器信息
			ctn.UpdateCtn(cctn)
			pCtn:=ctn.GetCtnFromIndex(cctn.OperIndex)
			fmt.Printf("\n强制停止容器应答\nOperIndex:%d\n容器ID:%s\n错误信息：%s\n",
				pCtn.OperIndex,pCtn.ID[:10],pCtn.Err)
			if cctn.ServiceName != "" {
				service := cluster.GetService(cctn.ServiceName) //通过服务名称获取服务
				//遍历服务中所有容器，判断容器执行状态
				rlt := true
				for _, val := range ctn.GetCtnsInService(service.ServiceName) {
					if (*val).Err != "nil" {
						rlt = false
					}
				}
				if rlt {
					fmt.Printf("强制停止服务%s成功！\n", service.ServiceName)
				}
			}
		}

	case ctn.REMOVE:
		{
			//通过映射关系找到结构体，更新容器信息
			ctn.UpdateCtn(cctn)
			pCtn:=ctn.GetCtnFromIndex(cctn.OperIndex)
			fmt.Printf("\n删除容器应答\nOperIndex:%d\n容器ID:%s\n错误信息：%s\n",
				pCtn.OperIndex,pCtn.ID[:10],pCtn.Err)
			if cctn.ServiceName != "" {
				service := cluster.GetService(cctn.ServiceName) //通过服务名称获取服务
				//遍历服务中所有容器，判断容器执行状态
				rlt := true
				for _, val := range ctn.GetCtnsInService(service.ServiceName) {
					if (*val).Err != "nil" {
						rlt = false
					}
				}
				if rlt {
					fmt.Printf("删除服务%s成功！\n", service.ServiceName)
				}
			}
		}
	case ctn.GETLOG:
		{
			//通过映射关系找到结构体，更新容器信息
			ctn.UpdateCtn(cctn)
			//fmt.Printf("\n获取容器日志应答\nOperIndex:%d\n容器ID：%s\n错误信息：%s\n容器日志：%s\n结构体数据：%#v\n",
			//	ctn[cctn.OperIndex].OperIndex,
			//	ctn[cctn.OperIndex].ID[:10],
			//	headers.Bytes2str(ctn[cctn.OperIndex].Err),
			//	headers.Bytes2str(ctn[cctn.OperIndex].Logs),
			//	ctn[cctn.OperIndex])

			/////////////////////////////////////////////////////////////////

		}
	case ctn.INSPECT:
		{
			//通过映射关系找到结构体，更新容器信息
			ctn.UpdateCtn(cctn)
			//fmt.Printf("\n获取容器详情应答\nOperIndex:%d\n容器ID：%s\n错误信息：%s\n结构体数据：%#v\n",
			//	ctn[cctn.OperIndex].OperIndex,
			//	ctn[cctn.OperIndex].ID[:10],
			//	headers.Bytes2str(ctn[cctn.OperIndex].Err),
			//	ctn[cctn.OperIndex])

			//ctnByte, err := json.Marshal(ctn[cctn.OperIndex].CtnInspect)
			//if err != nil {
			//	cctn.Err = headers.Str2bytes(err.Error())
			//}
			//
			//fmt.Println(headers.Bytes2str(ctnByte))
		}

	}
}

func myStateChange(id string, mystring uint8) {
	ok := cluster.IsAgentExist(id)
	if !ok {
		agent := &method.NODE{
			Addr: id,
		}

		cluster.AddAgent(agent)

		fmt.Printf("向集群添加agnent:%s", id)
	}
}

func send(nodeName string, sctn header.CTN) {
	byteStream, err := header.Encode(sctn)
	if err != nil {
		return
	}

	fmt.Println("向客户端发送指令:", sctn.OperFlag, nodeName)
	tcpSocket.WriteData(nodeName, 1, 0, FLAG_CTRL, byteStream)
}

// func serviceStats() {
// 	for {
// 		for _, service := range cluster.ServiceMap {
// 			rlt := true
// 			for _, vals := range *service.Pctn {
// 				if vals.OperFlag != ctn.REMOVE || vals.Err != "nil" {
// 					rlt = false
// 				}
// 			}
// 			if rlt == true {
// 				delete(cluster.ServiceMap, service)
// 			}
// 		}
// 	}
// }
