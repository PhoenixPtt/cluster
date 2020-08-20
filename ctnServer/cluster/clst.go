package cluster

import (
	"ctnCommon/pool"
	"fmt"
)

//导入集群配置
func (cluster *CLUSTER) ImportConfig()  {
}

//导出集群配置
func (cluster *CLUSTER) OutputConfig()  {
}

//启动集群
func (cluster *CLUSTER) Start(watchServicesKey string, watchNodesKey string)  {
	go cluster.WatchServices(watchServicesKey)//监视来自web的服务操作请求
	go cluster.WatchNodes(watchNodesKey)//监视节点状态的变化
}

//停止集群
func (cluster *CLUSTER) Stop()  {
	cluster.exitWatchServicesChan<-1
	cluster.exitWatchNodesChan<-1
}

//集群工作协程
func (cluster *CLUSTER) WatchServices(watchServicesKey string)  {
	var err error
	pool.RegPrivateChanStr(watchServicesKey)
	for {
		select {
		case obj := <-pool.GetPrivateChanStr(watchServicesKey):
			pClstOper:=obj.(*SERVICE_OPER_TRUCK)
			switch pClstOper.OperType{
			case SCREATE:
				err=cluster.CreateSvc(pClstOper.ConfigFileName, pClstOper.ConfigFileType)
			case SSTART:
				err=cluster.StartSvc(pClstOper.SvcName)
			case SSTOP:
				err=cluster.StopSvc(pClstOper.SvcName)
			case SRESTART:
				err=cluster.StopSvc(pClstOper.SvcName)
				if err!=nil{
					continue
				}
				err=cluster.StartSvc(pClstOper.SvcName)
			case SREMOVE:
				err=cluster.RemoveSvc(pClstOper.SvcName)
			case SSCALE:
				err=cluster.ScaleSvc(pClstOper.SvcName, pClstOper.ScaleNum)
			}
			//将执行结果返回给客户端
			fmt.Println()
			fmt.Printf("服务名称：%s\n",pClstOper.SvcName)
			fmt.Printf("服务操作：%s\n",pClstOper)
			fmt.Printf("应答错误：%s\n",err.Error())
		case <-cluster.exitWatchServicesChan:
			pool.UnregPrivateChanStr(watchServicesKey)
			return
		default:
		}
	}
}

func (cluster *CLUSTER) WatchNodes(watchNodesKey string)  {
	pool.RegPrivateChanStr(watchNodesKey)
	for {
		select {
		case obj := <-pool.GetPrivateChanStr(watchNodesKey):
			pNodeStatus:=obj.(map[string]bool)
			for key,value:=range pNodeStatus{
				cluster.SetNodeStatus(key, value)
			}
		case <-cluster.exitWatchNodesChan:
			pool.UnregPrivateChanStr(watchNodesKey)
			return
		default:
		}
	}
}

func (cluster *CLUSTER) SetNodeStatus(nodeName string, status bool){
	_,ok:=cluster.NodeStatusMap[nodeName]
	if ok{
		currStatus := cluster.NodeStatusMap[nodeName]
		if currStatus!=status{//节点状态有变化
			cluster.NodeStatusMap[nodeName] = status//更新节点状态
			//将节点状态通知所有服务
		}
	}else{
		cluster.NodeStatusMap[nodeName] = status
	}

	for _, pService:=range cluster.ServiceMap{
		pService.SetNodeStatus(nodeName, status)
	}
}

//获取全部节点或在线的节点
func (cluster *CLUSTER) GetNodeList(nodeSelector int) (nodeList []string){
	nodeList = make([]string,0,NODE_NUM)
	switch nodeSelector{
	case ALL_NODES:
		for node, _ := range cluster.NodeStatusMap{
			nodeList = append(nodeList, node)
		}
	case ACTIVE_NODES:
		for node, status := range cluster.NodeStatusMap{
			if status == true{
				nodeList = append(nodeList, node)
			}
		}
	}
	return
}


