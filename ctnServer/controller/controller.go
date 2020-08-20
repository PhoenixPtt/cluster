package controller

import (
	"ctnCommon/pool"
	"fmt"
)

//启动集群
func (controller *CONTROLLER) Start(watchServicesKey string, watchNodesKey string)  {
	go controller.WatchServices(watchServicesKey)//监视来自web的服务操作请求
	go controller.WatchNodes(watchNodesKey)//监视节点状态的变化
}

//停止集群
func (controller *CONTROLLER) Stop()  {
	controller.exitWatchServicesChan<-1
	controller.exitWatchNodesChan<-1
}

func (controller *CONTROLLER) PutService(pSvcOperTruck *SERVICE_OPER_TRUCK) {
	pChan := pool.GetPrivateChanStr(controller.watchServicesKey)
	pChan<-pSvcOperTruck
}

//集群工作协程
func (controller *CONTROLLER) WatchServices(watchServicesKey string)  {
	var err error
	controller.watchNodesKey = watchServicesKey
	pool.RegPrivateChanStr(watchServicesKey)
	for {
		select {
		case obj := <-pool.GetPrivateChanStr(watchServicesKey):
			pClstOper:=obj.(*SERVICE_OPER_TRUCK)
			switch pClstOper.OperType{
			case SCREATE:
				err=controller.CreateSvc(pClstOper.ConfigFileName, pClstOper.ConfigFileType)
			case SSTART:
				err=controller.StartSvc(pClstOper.SvcName)
			case SSTOP:
				err=controller.StopSvc(pClstOper.SvcName)
			case SRESTART:
				err=controller.StopSvc(pClstOper.SvcName)
				if err!=nil{
					continue
				}
				err=controller.StartSvc(pClstOper.SvcName)
			case SREMOVE:
				err=controller.RemoveSvc(pClstOper.SvcName)
			case SSCALE:
				err=controller.ScaleSvc(pClstOper.SvcName, pClstOper.ScaleNum)
			}
			//将执行结果返回给客户端
			fmt.Println()
			fmt.Printf("服务名称：%s\n",pClstOper.SvcName)
			fmt.Printf("服务操作：%s\n",pClstOper)
			fmt.Printf("应答错误：%s\n",err.Error())
		case <-controller.exitWatchServicesChan:
			pool.UnregPrivateChanStr(watchServicesKey)
			return
		default:
		}
	}
}

func (controller *CONTROLLER) PutNode(nodeName string, status bool) {
	nodeStatusMap := make(map[string]bool,1)
	pChan := pool.GetPrivateChanStr(controller.watchServicesKey)
	pChan<-nodeStatusMap
}

func (controller *CONTROLLER) WatchNodes(watchNodesKey string)  {
	controller.watchNodesKey = watchNodesKey
	pool.RegPrivateChanStr(watchNodesKey)
	for {
		select {
		case obj := <-pool.GetPrivateChanStr(watchNodesKey):
			pNodeStatus:=obj.(map[string]bool)
			for key,value:=range pNodeStatus{
				controller.SetNodeStatus(key, value)
			}
		case <-controller.exitWatchNodesChan:
			pool.UnregPrivateChanStr(watchNodesKey)
			return
		default:
		}
	}
}

func (controller *CONTROLLER) SetNodeStatus(nodeName string, status bool){
	_,ok:=controller.NodeStatusMap[nodeName]
	if ok{
		currStatus := controller.NodeStatusMap[nodeName]
		if currStatus!=status{//节点状态有变化
			controller.NodeStatusMap[nodeName] = status//更新节点状态
			//将节点状态通知所有服务
		}
	}else{
		controller.NodeStatusMap[nodeName] = status
	}

	for _, pService:=range controller.ServiceMap{
		pService.SetNodeStatus(nodeName, status)
	}
}

//获取全部节点或在线的节点
func (controller *CONTROLLER) GetNodeList(nodeSelector int) (nodeList []string){
	nodeList = make([]string,0,NODE_NUM)
	switch nodeSelector{
	case ALL_NODES:
		for node, _ := range controller.NodeStatusMap{
			nodeList = append(nodeList, node)
		}
	case ACTIVE_NODES:
		for node, status := range controller.NodeStatusMap{
			if status == true{
				nodeList = append(nodeList, node)
			}
		}
	}
	return
}


