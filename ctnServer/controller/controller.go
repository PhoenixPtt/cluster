package controller

import (
	"context"
	"ctnCommon/pool"
	"fmt"
)

//启动集群
func (pController *CONTROLLER) Start() {
	go pController.Daq()
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
	ctx,cancel:=context.WithCancel(context.Background())
	_,ok:=pController.CancelWatchSvcs[svcWatch]
	if !ok{
		pController.CancelWatchSvcs[svcWatch] = cancel
	}else{
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
				err = pController.CreateSvcFromFile(pClstOper.ConfigFileName, pClstOper.ConfigFileType)
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
	_,ok:=pController.CancelWatchSvcs[svcWatch]
	if ok{
		fmt.Println("GGGGGGGGGGGGGGGGGGGGGGGGGGGGG","停止监控",svcWatch)
		pController.CancelWatchSvcs[svcWatch]()
	}
}

func (pController *CONTROLLER) WatchServiceStatus() {
	pool.RegPrivateChanStr(SERVICE_STATUS_WATCH, CHAN_BUFFER)
	var ctx context.Context
	ctx,pController.CancelWatchSvcStatus=context.WithCancel(context.Background())
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n取消监控服务状态\n")
			pool.UnregPrivateChanStr(SERVICE_STATUS_WATCH)
			return
		case obj := <-pool.GetPrivateChanStr(SERVICE_STATUS_WATCH):
			rplStatusMap := obj.(map[string]int)
			for svcName, status := range rplStatusMap {
				switch status {
				case SVC_STATUS_GODIRTY: //副本变脏
				case SVC_STATUS_REMOVED:
					pSvc := pController.GetSvc(svcName)
					DelService(pSvc)
					delete(pController.ServiceMap, svcName)
					fmt.Println(svcName, "已被删除", "哈哈哈哈")
				}
			}
		}
	}
}

func (pController *CONTROLLER) PutNode(nodeName string, status bool) {
	nodeStatusMap := make(map[string]bool, 1)
	nodeStatusMap[nodeName] = status
	pChan := pool.GetPrivateChanStr(NODE_WATCH)
	pChan <- nodeStatusMap
}

func (pController *CONTROLLER) WatchNodes() {
	pool.RegPrivateChanStr(NODE_WATCH, CHAN_BUFFER)
	var ctx context.Context
	ctx,pController.CancelWatchNodes=context.WithCancel(context.Background())
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
