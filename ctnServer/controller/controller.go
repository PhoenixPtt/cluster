package controller

import (
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
	pController.CancelWatchNode()
}

func (pController *CONTROLLER) PutService(pSvcOperTruck *SERVICE_OPER_TRUCK) {
	svcWatch := GetServiceWatchKey(pSvcOperTruck.SvcName)
	if pSvcOperTruck.OperType == SCREATE {
		pool.RegPrivateChanStr(svcWatch, CHAN_BUFFER)
		svcExitChan := make(chan int, 1)
		pController.svcExitChanMap[svcWatch] = svcExitChan
		go pController.WatchService(svcWatch, pController.svcExitChanMap[svcWatch])
	}
	pChan := pool.GetPrivateChanStr(svcWatch)
	pChan <- pSvcOperTruck
}

//集群工作协程
func (pController *CONTROLLER) WatchService(svcWatch string, exitChan chan int) {
	var err error
	for {
		select {
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
			case SSCALE:
				err = pController.ScaleSvc(pClstOper.SvcName, pClstOper.ScaleNum)
			}
			//将执行结果返回给客户端
			fmt.Println()
			fmt.Printf("服务名称：%s\n", pClstOper.SvcName)
			fmt.Printf("服务操作：%s\n", pClstOper)
			fmt.Printf("应答错误：%s\n", err.Error())
		case <-pController.svcExitChanMap[svcWatch]:
			pool.UnregPrivateChanStr(svcWatch)
			close(pController.svcExitChanMap[svcWatch])
			delete(pController.svcExitChanMap, svcWatch)
			return
		}
	}
}

func (pController *CONTROLLER) CancelWatchService(svcWatch string) {
	_, ok := pController.svcExitChanMap[svcWatch]
	if ok {
		svcExitChan := pController.svcExitChanMap[svcWatch]
		svcExitChan <- 1
	}
}

func (pController *CONTROLLER) WatchServiceStatus() {
	pController.exitWatchSvcStatusChan = make(chan int, 1)
	pool.RegPrivateChanStr(SERVICE_STATUS_WATCH, CHAN_BUFFER)
	for {
		select {
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
		case <-pController.exitWatchSvcStatusChan:
			pool.UnregPrivateChanStr(SERVICE_STATUS_WATCH)
			close(pController.exitWatchSvcStatusChan)
			return
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
	for {
		select {
		case obj := <-pool.GetPrivateChanStr(NODE_WATCH):
			pNodeStatus := obj.(map[string]bool)
			for key, value := range pNodeStatus {
				pController.SetNodeStatus(key, value)
			}
		case <-pController.exitWatchNodesChan:
			pool.UnregPrivateChanStr(NODE_WATCH)
			close(pController.exitWatchNodesChan)
			return
		}
	}
}

func (pController *CONTROLLER) CancelWatchNode() {
	pController.exitWatchNodesChan <- 1
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
