package controller

import (
	header "clusterHeader"
	"context"
	"ctnCommon/pool"
	"ctnServer/ctnS"
	"fmt"
	"github.com/docker/docker/api/types/events"
	"time"
)

//启动集群
func (pController *CONTROLLER) Start() {
	go pController.WatchNodes() //监视节点状态的变化
	go pController.WatchServiceStatus()
}

//停止集群
func (pController *CONTROLLER) Stop() {
	pController.CancelDaq()
	pController.CancelWatchNodes()
	pController.CancelWatchSvcStatus()
}

func (pController *CONTROLLER) PutService(pSvcOperTruck *SERVICE_OPER_TRUCK) {
	svcWatch := GetServiceWatchKey(pSvcOperTruck.SvcName)
	switch pSvcOperTruck.OperType {
	case SCREATE:
		pool.RegPrivateChanStr(svcWatch, CHAN_BUFFER)
		go pController.WatchService(svcWatch)
	}
	pChan := pool.GetPrivateChanStr(svcWatch)
	pChan <- pSvcOperTruck
}

//集群工作协程
func (pController *CONTROLLER) WatchService(svcWatch string) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ok := pController.CancelWatchSvcs[svcWatch]
	if !ok {
		pController.CancelWatchSvcs[svcWatch] = cancel
	} else {
		return
	}
	var err error
	for {
		select {
		case <-ctx.Done():
			pool.UnregPrivateChanStr(svcWatch)
			delete(pController.CancelWatchSvcs, svcWatch)
			return
		case obj := <-pool.GetPrivateChanStr(svcWatch):
			pClstOper := obj.(*SERVICE_OPER_TRUCK)
			switch pClstOper.OperType {
			case SCREATE:
				//err = pController.CreateSvcFromFile(pClstOper.ConfigFileName, pClstOper.ConfigFileType)
				err = pController.CreateSvc(&pClstOper.SvcCfg)
			case SSTART:
				err = pController.StartSvc(pClstOper.SvcName)
			case SSTOP:
				err = pController.StopSvc(pClstOper.SvcName)
			case SRESTART:
				err = pController.StopSvc(pClstOper.SvcName)
				if err != nil {
					continue
				}
				err = pController.StartSvc(pClstOper.SvcName)
			case SREMOVE:
				err = pController.RemoveSvc(pClstOper.SvcName)
				//关闭服务协程
				pController.CancelWatchSvcs[svcWatch]()
			case SSCALE:
				err = pController.ScaleSvc(pClstOper.SvcName, pClstOper.ScaleNum)
			}
			//将执行结果返回给客户端
			fmt.Println()
			fmt.Printf("服务名称：%s\n", pClstOper.SvcName)
			fmt.Printf("服务操作：%s\n", pClstOper)
			fmt.Printf("应答错误：%s\n", err.Error())
		}
	}
}

func (pController *CONTROLLER) CancelWatchService(svcWatch string) {
	_, ok := pController.CancelWatchSvcs[svcWatch]
	if ok {
		pController.CancelWatchSvcs[svcWatch]()
	}
}

func (pController *CONTROLLER) WatchServiceStatus() {
	pool.RegPrivateChanStr(UPLOAD, CHAN_BUFFER)
	pool.RegPrivateChanStr(SERVICE_STATUS_WATCH, CHAN_BUFFER)
	var ctx context.Context
	ctx, pController.CancelWatchSvcStatus = context.WithCancel(context.Background())
	var G_samplingRate int = 1
	//定时器时间间隔
	var interval time.Duration = time.Second * time.Duration(G_samplingRate)

	timer := time.NewTimer(interval)
	for {
		pWebServices := &header.SERVICE{}
		select {
		case <-ctx.Done():
			fmt.Println("\n取消监控服务状态\n")
			pool.UnregPrivateChanStr(SERVICE_STATUS_WATCH)
			return
		case <-timer.C:
			timer.Reset(interval)
		case obj := <-pool.GetPrivateChanStr(SERVICE_STATUS_WATCH):
			//svcStatusMap := obj.(map[string]int)
			//for svcName, status := range rplStatusMap {
			//	pSvc := pController.GetSvc(svcName)
			//	pSvc
			//}
			//
			//
			var webSvcs header.Services
			webSvcs.Service = make([]header.Service, 0, SVC_NUM)
			rplStatusMap := obj.(map[string]int)
			for svcName, status := range rplStatusMap {
				pSvc := pController.GetSvc(svcName)
				pWebSvc := &header.Service{}
				switch status {
				case SVC_STATUS_RUNNING: //服务在运行
					//更新服务信息
					//发送给web前段
					//服务的基本信息
					pWebSvc.Id = pSvc.SvcName // 服务Id
					//服务状态
					pWebSvc.Scale = uint32(pSvc.SvcScale)             // 设定的副本数量
					pWebSvc.ReplicaCount = uint32(len(pSvc.Replicas)) // 应用服务的当前副本数量
					pWebSvc.CreateTime = pSvc.CreateTime              // 服务创建时间
					pWebSvc.StartTime = pSvc.StartTime                // 服务启动时间
					pWebSvc.NameSpace = pSvc.NameSpace                //服务的命名空间
					//服务配置信息,空缺

					//服务的所有副本
					for _, pRpl := range pSvc.Replicas {
						pWebRpl := &header.Replica{}
						pWebRpl.Id = pRpl.RplName
						pWebRpl.CreateTime = pRpl.CreateTime
						pCtn := ctnS.GetCtn(pRpl.CtnName)
						if pCtn != nil {
							pWebRpl.Ctn = pCtn.Container
							pWebRpl.CtnStats = pCtn.CTN_STATS
						}
						pWebSvc.Replica = append(pWebSvc.Replica, *pWebRpl)
					}
				case SVC_STATUS_REMOVED:
					pSvc := pController.GetSvc(svcName)
					DelService(pSvc)
					delete(pController.ServiceMap, svcName)
					fmt.Println(svcName, "已被删除", "哈哈哈哈")
				}
				webSvcs.Service = append(webSvcs.Service, *pWebSvc)
			}
			webSvcs.Count = uint32(len(webSvcs.Service))
			pWebServices.Services = webSvcs
			timer.Reset(interval)
		case obj := <-pool.GetPrivateChanStr(ctnS.EVENT_MSG_WATCH):
			{
				evtMsg := obj.(events.Message)
				pWebServices.EventMsg = append(pWebServices.EventMsg, evtMsg)
			}
		case obj := <-pool.GetPrivateChanStr(ctnS.ERR_MSG_WATCH):
			{
				errMsg := obj.(error)
				pWebServices.ErrMsg = append(pWebServices.ErrMsg, errMsg.Error())
			}
		}
		select {
		case pool.GetPrivateChanStr(UPLOAD) <- pWebServices:
			//fmt.Println("111111111111111111111111111111111", pWebServices)
		default:
		}
	}
}

func (pController *CONTROLLER) PutNode(nodeName string, status bool) {
	nodeStatusMap := make(map[string]bool)
	nodeStatusMap[nodeName] = status
	pChan := pool.GetPrivateChanStr(NODE_WATCH)
	pChan <- nodeStatusMap
}

func (pController *CONTROLLER) WatchNodes() {
	pool.RegPrivateChanStr(NODE_WATCH, CHAN_BUFFER)
	var ctx context.Context
	ctx, pController.CancelWatchNodes = context.WithCancel(context.Background())
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n取消监控节点\n")
			pool.UnregPrivateChanStr(NODE_WATCH)
			return
		case obj := <-pool.GetPrivateChanStr(NODE_WATCH):
			pNodeStatus := obj.(map[string]bool)
			for key, value := range pNodeStatus {
				pController.SetNodeStatus(key, value)
			}
		}
	}
}

func (pController *CONTROLLER) SetNodeStatus(nodeName string, status bool) {
	_, ok := pController.NodeStatusMap[nodeName]
	if ok {
		currStatus := pController.NodeStatusMap[nodeName]
		if currStatus != status { //节点状态有变化
			pController.NodeStatusMap[nodeName] = status //更新节点状态
			//将节点状态通知所有服务
		}
	} else {
		pController.NodeStatusMap[nodeName] = status
	}

	for _, pService := range pController.ServiceMap {
		pService.SetNodeStatus(nodeName, status)
	}
}

//获取全部节点或在线的节点
func (pController *CONTROLLER) GetNodeList(nodeSelector int) (nodeList []string) {
	nodeList = make([]string, 0, NODE_NUM)
	switch nodeSelector {
	case ALL_NODES:
		for node, _ := range pController.NodeStatusMap {
			nodeList = append(nodeList, node)
		}
	case ACTIVE_NODES:
		for node, status := range pController.NodeStatusMap {
			if status == true {
				nodeList = append(nodeList, node)
			}
		}
	}
	return
}
