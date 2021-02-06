package controller

import (
	header "clusterHeader"
	"context"
	"ctnCommon/pool"
	"ctnServer/ctnS"
	"fmt"
	"time"
)

const UPLOAD = "UPLOAD"

var (
	G_samplingRate int
	G_controller   *CONTROLLER
	cancelDaq      context.CancelFunc
)

func InitCtnMgr(sendObjFunc pool.SendObjFunc) {
	G_controller = NewController(sendObjFunc)
	G_controller.Start()
}

//处理网络消息
func Daq() {
	pool.RegPrivateChanStr(UPLOAD, CHAN_BUFFER)

	//var ctx context.Context
	//ctx, cancelDaq = context.WithCancel(context.Background())
	//
	////定时器时间间隔
	//G_samplingRate = 10
	//var interval time.Duration = time.Second * time.Duration(G_samplingRate)
	//var timer *time.Timer = time.NewTimer(interval)
	//var pWebServices *header.SERVICE = &header.SERVICE{}
	//for {
	//	select {
	//	case <-ctx.Done():
	//		fmt.Println("取消采集")
	//		pool.UnregPrivateChanStr(ctnS.DAQ)
	//		return
	//	case <-timer.C:
	//		pWebServices.Service = make([]header.Service, 0, SVC_NUM)
	//		pWebServices.Count = 1
	//		var webSvc header.Service
	//		//ServiceInfo
	//		webSvc.Id = service.SvcName
	//		webSvc.State = service.SvcStats
	//		webSvc.Scale = uint32(service.SvcScale)
	//		webSvc.ReplicaCount = uint32(len(service.Replicas))
	//		webSvc.CreateTime = service.CreateTime
	//		webSvc.StartTime = service.StartTime
	//		webSvc.NameSpace = service.NameSpace
	//		//ServiceCfg 服务配置信息 暂时不填充
	//		//副本信息
	//		for _, replica := range service.Replicas {
	//			var webReplica header.Replica
	//			webReplica.Id = replica.RplName
	//			webReplica.CreateTime = replica.CreateTime
	//			webReplica.NodeId = replica.AgentAddr
	//			webReplica.State = 0
	//			pCtn := ctnS.GetCtn(replica.CtnName)
	//			if pCtn != nil {
	//				webReplica.Ctn = pCtn.Container
	//				webReplica.CtnStats = pCtn.CTN_STATS
	//				webReplica.Log = pCtn.CtnLog
	//				webReplica.CtnInspect = pCtn.CtnInspect
	//			}
	//			webSvc.Replica = append(webSvc.Replica, webReplica)
	//		}
	//		pWebServices.Service = append(pWebServices.Service, webSvc)
	//		pChan := pool.GetPrivateChanStr(UPLOAD)
	//		select {
	//		case pChan <- pWebServices:
	//		default:
	//		}
	//		ctnS.Mylog.Debug(fmt.Sprintf("######向web前端发送数据		%v\n", pWebServices))
	//	}
	//}

	var ctx context.Context
	ctx, cancelDaq = context.WithCancel(context.Background())

	//定时器时间间隔
	var interval time.Duration = time.Second * time.Duration(G_samplingRate)
	var timer *time.Timer = time.NewTimer(interval)
	var pWebServices *header.SERVICE = &header.SERVICE{}
	pWebServices.Service = make([]header.Service, 0, SVC_NUM)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("取消采集")
			pool.UnregPrivateChanStr(ctnS.DAQ)
			return
		case <-timer.C:
			pWebServices.Count = uint32(len(G_controller.ServiceMap))
			for svcName, currService := range G_controller.ServiceMap {
				var webSvc header.Service
				//ServiceInfo
				webSvc.Id = svcName
				webSvc.State = currService.SvcStats
				webSvc.Scale = uint32(currService.SvcScale)
				webSvc.ReplicaCount = uint32(len(currService.Replicas))
				webSvc.CreateTime = currService.CreateTime
				webSvc.StartTime = currService.StartTime
				webSvc.NameSpace = currService.NameSpace
				//ServiceCfg 服务配置信息 暂时不填充
				//副本信息
				for _, replica := range currService.Replicas {
					var webReplica header.Replica
					webReplica.Id = replica.RplName
					webReplica.CreateTime = replica.CreateTime
					webReplica.NodeId = replica.AgentAddr
					webReplica.State = 0
					pCtn := ctnS.GetCtn(replica.CtnName)
					if pCtn != nil {
						webReplica.Ctn = pCtn.Container
						webReplica.CtnStats = pCtn.CTN_STATS
						webReplica.Log = pCtn.CtnLog
						webReplica.CtnInspect = pCtn.CtnInspect
					}
					webSvc.Replica = append(webSvc.Replica, webReplica)
				}
				pWebServices.Service = append(pWebServices.Service, webSvc)
			}
			pChan := pool.GetPrivateChanStr(UPLOAD)
			select {
			case pChan <- pWebServices:
			default:

			}
			ctnS.Mylog.Debug(fmt.Sprintf("######向web前端发送数据		%v\n", pWebServices))

			//var infoStr string
			//var serviceNames []string
			//for _,service:=range pWebServices.Service{
			//	serviceNames = append(serviceNames, service.Id)
			//
			//	var serviceStr string
			//	serviceStr = fmt.Sprintf("%s\n%s\n", serviceStr, service.Id)
			//
			//	var replicaNames []string
			//
			//	for _, replica:=range service.Replica{
			//		replicaNames = append(replicaNames,replica.Id)
			//	}
			//
			//	var replicaNameStr string
			//	for _, replicaName:=range replicaNames{
			//		replicaNameStr = fmt.Sprintf("%s\n%s\n",replicaNameStr,replicaName)
			//	}
			//
			//	serviceStr = fmt.Sprintf("%s%s",serviceStr, replicaNameStr)
			//
			//	infoStr = fmt.Sprintf("%s%s",infoStr,serviceStr)
			//}
			//
			//ctnS.Mylog.Debug(fmt.Sprintf("######向web前端发送数据\n%s\n",infoStr))
		}
	}
}

func WaitWebService() (pWebServices *header.SERVICE) {
	select {
	case obj := <-pool.GetPrivateChanStr(UPLOAD): //类型：header.SERVICE
		pWebServices = obj.(*header.SERVICE)
		fmt.Println("2222222222222222222222222222", pWebServices)
	}
	return
}
