package controller

import (
	header "clusterHeader"
	"context"
	"ctnCommon/ctn"
	"ctnCommon/headers"
	"ctnCommon/pool"
	"ctnServer/ctnS"
	"errors"
	"fmt"
	"time"
)

const (
	SVC_DEFAULT = "未创建"
	SVC_CREATED = "已创建"
	SVC_RUNNING = "运行中"
	SVC_STOPPED = "已停止"
	SVC_REMOVED = "已删除"

	SVC_STATUS_GODIRTY = iota
	SVC_STATUS_REMOVED
	SVC_STATUS_RUNNING
)

//服务接口
type SERVICE_BEHAVIOR interface {
	SetNodeStatus(nodeName string, status bool)
	Create()
	Start()
	Scale(scaleNum int)
	Stop()
	Remove()
	GetHealthDegree() float64 //获取服务健康度
	GetServiceStatus() string //获取服务状态
	GetServiceScal() int      //获取服务规模
}

func (service *SERVICE) SetNodeStatus(nodeName string, status bool) {
	_, ok := service.NodeStatusMap[nodeName]
	if ok {
		currStatus := service.NodeStatusMap[nodeName]
		if currStatus != status { //节点状态有变化
			service.NodeStatusMap[nodeName] = status //更新节点状态
		}
	} else {
		service.NodeStatusMap[nodeName] = status
	}

	for _, rpl := range service.Replicas { //通知关联副本
		rpl.SetNodeStatus(nodeName, status)
	}
}

func (service *SERVICE) WatchRpl() {
	pool.RegPrivateChanStr(service.SvcName, CHAN_BUFFER)
	var ctx context.Context
	ctx, service.CancelWatchRpl = context.WithCancel(context.Background())
	for {
		select {
		case <-ctx.Done():
			fmt.Println("&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&", service.SvcName, "退出")
			pool.UnregPrivateChanStr(service.SvcName)
			return
		case obj := <-pool.GetPrivateChanStr(service.SvcName):
			rplStatusMap := obj.(map[string]string)
			for rplName, status := range rplStatusMap {
				rpl := service.GetRpl(rplName)
				ctnS.Mylog.Debug(fmt.Sprintf("服务收到副本数据：	副本名称：%s|dirty状态：%s\n", rplName, status))
				switch status {
				case ctn.DIRTY_POSITION_REMOVED:
					service.DelRpl(rplName)
					ctnS.RemoveCtn(rpl.CtnName)
				case ctn.DIRTY_POSITION_ERR_BEFORE_RPL_OPER: //副本操作之前的合法性检查失败
					service.DelRpl(rplName) //删除副本
					ctnS.RemoveCtn(rpl.CtnName)
					go service.schedule(rplName) //重新调度
				case ctn.DIRTY_POSITION_RPL_OPER_TIMEOUT: //副本操作超时
					service.DelRpl(rplName) //删除副本
					ctnS.RemoveCtn(rpl.CtnName)
					go service.schedule(rplName) //重新调度
				case ctn.DIRTY_POSITION_DOCKER, ctn.DIRTY_POSTION_IMAGE_RUN_ERR, ctn.DIRTY_POSITION_CTN_EXIST_IN_SERVER_BUT_NOT_IN_AGENT:
					//单纯容器在docker服务端被非正常手段删除
					//server端操作与操作结果不一致，在docker服务器中执行失败
					pCtnS := ctnS.GetCtn(rpl.CtnName)
					//ctnS.Mylog.Debug(fmt.Sprintf("rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr333333333333333333333333333333333%v", pCtnS))
					if pCtnS == nil {
						continue
					} else {
						if pCtnS.CtnID == "" {
							//表示此容器已经被删除了，处理下server端参与信息
							service.DelRpl(rplName) //删除副本
							ctnS.RemoveCtn(rpl.CtnName)
							//ctnS.Mylog.Debug(fmt.Sprintf("oooooooooooooooooooooooooooooooooooooooooooooooooooooooo终于删掉了		%s, %s", rpl.CtnName, pCtnS.OperType))
						} else {
							rpl.SetTargetStat(RPL_TARGET_REMOVED) //执行删除操作
						}
					}

					//执行调度操作
					//ctnS.Mylog.Debug(fmt.Sprintf("%s %s %v %v", rplName, rpl.CtnName, rpl.IsSchedulering, pCtnS))
					if !rpl.IsSchedulering {
						go service.schedule(rplName) //重新调度
					}

					//ctnS.Mylog.Debug(fmt.Sprintf("%s %v", rpl.RplName, rpl.IsSchedulering))
					if !rpl.IsSchedulering {
						rpl.IsSchedulering = true
					}
				case ctn.DIRTY_POSTION_SERVER_LOST_CONNICTION:
					go service.schedule(rplName) //重新调度
				}
			}
		}

		//更新服务运行状态
		pChan := pool.GetPrivateChanStr(SERVICE_STATUS_WATCH)
		svcStatusMap := make(map[string]int)
		rplLen := len(service.Replicas)
		//实际副本数量与预定规模相等，都为0
		if (rplLen == 0) && (service.SvcScale == 0) {
			svcStatusMap[service.SvcName] = SVC_STATUS_REMOVED
		}

		select {
		case pChan <- svcStatusMap:
		default:

		}

	}
}

func (pSvc *SERVICE) NewRpl(name string, image string, agentAddr string) (rpl *REPLICA) {
	rpl = &REPLICA{}
	rpl.RplName = name
	rpl.SvcName = pSvc.SvcName
	rpl.RplImage = image
	rpl.AgentAddr = agentAddr
	rpl.AgentStatus = true
	rpl.RplTargetStat = RPL_TARGET_REMOVED
	rpl.CreateTime = headers.ToString(time.Now(), headers.TIME_LAYOUT_NANO) //启动时间
	rpl.Timeout = pSvc.Timeout
	rpl.AgentTryNum = pSvc.AgentTryNum
	rpl.IsSchedulering = false

	pSvc.Replicas = append(pSvc.Replicas, rpl)
	return
}

func (pSvc *SERVICE) DelRpl(rplName string) {
	pRpl := pSvc.GetRpl(rplName)
	pRpl.CancelWatchCtn()

	rIndex := pSvc.GetRplIndex(rplName)
	if rIndex == -1 {
		return
	}
	pSvc.Replicas = append(pSvc.Replicas[:rIndex], pSvc.Replicas[rIndex+1:]...)
}

//对应用户创建操作
func (pSvc *SERVICE) Create() (err error) {
	//判断是否满足执行动作的前提条件
	var info string
	if pSvc.SvcStats != SVC_DEFAULT { //服务只有处于初始（未创建）状态，才允许被创建
		info = infoString(pSvc.SvcStats, "已创建，本次创建操作将被忽略。")
	} else {
		ctnS.Mylog.Info("-----------------------创建服务-----------------------")
		pSvc.SvcStats = SVC_CREATED
		pSvc.CreateTime = headers.ToString(time.Now(), headers.TIME_LAYOUT_NANO) //创建时间
	}

	switch info {
	case "":
		info = infoString(pSvc.SvcName, pSvc.SvcStats)
	default:
		err = errors.New(info)
	}

	ctnS.Mylog.Info(info)
	return
}

//对应用户启动操作
func (pSvc *SERVICE) Start() (err error) {
	//判断是否满足执行动作的前提条件
	var info string
	switch pSvc.SvcStats {
	case SVC_DEFAULT:
		info = infoString(pSvc.SvcName, "未创建，无法执行启动操作。")
	case SVC_RUNNING:
		info = infoString(pSvc.SvcName, "正在稳定运行中，本次启动操作将被忽略。")
	default:
		ctnS.Mylog.Info("\n-----------------------启动服务-----------------------")
		pSvc.SvcStats = SVC_RUNNING
		pSvc.updateRpl()                                                        //根据具体情况增删副本
		pSvc.StartTime = headers.ToString(time.Now(), headers.TIME_LAYOUT_NANO) //启动时间
	}

	switch info {
	case "":
		info = infoString(pSvc.SvcName, pSvc.SvcStats)
	default:
		err = errors.New(info)
	}

	ctnS.Mylog.Info(info)
	return
}

//对应用户改变规模操作
func (pSvc *SERVICE) Scale(scaleNum int) (err error) {
	//判断是否满足执行动作的前提条件
	var info string
	switch pSvc.SvcStats {
	case SVC_DEFAULT:
		info = infoString(pSvc.SvcName, "未创建，无法执行调整操作。")
	case SVC_CREATED:
		info = infoString(pSvc.SvcName, "已创建但未启动，无法执行调整规模操作。")
	case SVC_STOPPED:
		info = infoString(pSvc.SvcName, "已停止，无法执行调整规模操作。")
	default:
		ctnS.Mylog.Info("-----------------------调整规模-----------------------")
		pSvc.SvcStats = SVC_RUNNING
		pSvc.SvcScale = scaleNum
		pSvc.updateRpl() //根据具体情况增删副本
	}

	switch info {
	case "":
		info = infoString(pSvc.SvcName, pSvc.SvcStats)
	default:
		err = errors.New(info)
	}

	ctnS.Mylog.Info(info)
	return
}

//对应用户停止操作
func (pSvc *SERVICE) Stop() (err error) {
	//判断是否满足执行动作的前提条件
	var info string
	switch pSvc.SvcStats {
	case SVC_DEFAULT:
		info = infoString(pSvc.SvcName, "未创建，无法执行停止操作。")
	case SVC_CREATED:
		info = infoString(pSvc.SvcName, "已创建但未启动，无法执行停止操作。")
	case SVC_STOPPED:
		info = infoString(pSvc.SvcName, "已经停止，本次停止操作将被忽略。")
	default:
		ctnS.Mylog.Info("-----------------------停止-----------------------")
		pSvc.SvcStats = SVC_STOPPED
		pSvc.updateRpl() //根据具体情况增删副本
	}

	switch info {
	case "":
		info = infoString(pSvc.SvcName, pSvc.SvcStats)
	default:
		err = errors.New(info)
	}

	ctnS.Mylog.Info(info)
	return
}

//对应用户删除操作
func (pSvc *SERVICE) Remove() (err error) {
	var info string //判断是否满足执行动作的前提条件
	ctnS.Mylog.Info("-----------------------删除-----------------------")
	pSvc.SvcStats = SVC_REMOVED
	pSvc.updateRpl() //根据具体情况增删副本

	switch info {
	case "":
		info = infoString(pSvc.SvcName, pSvc.SvcStats)
	default:
		err = errors.New(info)
	}

	ctnS.Mylog.Info(info)
	return
}

//改变规模
func (pSvc *SERVICE) updateRpl() {
	dir, scaleNum := pSvc.getScaleNum()
	switch {
	case dir > 0:
		//创建副本
		agentAddrNumMap, err := pSvc.EScale(scaleNum)
		if err != nil {
			return
		}
		for addr, rplNum := range agentAddrNumMap {
			for i := 0; i < rplNum; i++ {
				pRpl := pSvc.NewRpl(pSvc.SvcName+"_"+headers.UniqueId(), pSvc.Image, addr)
				pRpl.SetTargetStat(RPL_TARGET_RUNNING) //设置副本的目标状态
			}
		}
	case dir < 0:
		rplNames, _ := pSvc.DScale(scaleNum)
		for _, rplName := range rplNames {
			pRpl := pSvc.GetRpl(rplName)
			pRpl.SetTargetStat(RPL_TARGET_REMOVED) //设置副本的目标状态
		}
	}
	return
}

//获取服务规模
func (service *SERVICE) GetServiceScale() int {
	return service.SvcScale
}

//获取服务状态
func (service *SERVICE) GetServiceStatus() string {
	return service.SvcStats
}

//获取健康度
func (service *SERVICE) GetHealthDegree() float64 {
	return service.SvcHealthDegree
}

//设置服务规模
func (service *SERVICE) setServiceScale(scalNum int) {
	service.SvcScale = scalNum
}

//更新健康度
func (service *SERVICE) updateHealthDegree() (healthDegree float64) {
	var activeNum int = 0
	replicas := service.getReplicas()
	for _, replica := range replicas {
		pCtn := ctnS.GetCtn(replica.CtnName)
		if !pCtn.Dirty && pCtn.State == "running" {
			activeNum++
		}
	}
	service.SvcHealthDegree = float64(activeNum) / float64(service.SvcScale)
	return
}

//获取所有副本名称
func (service *SERVICE) GetRplNames() []string {
	var (
		rNames []string
	)
	rNames = make([]string, 0, 100)
	for _, val := range service.getReplicas() {
		rNames = append(rNames, val.RplName)
	}
	return rNames
}

//根据副本名称获取副本
func (service *SERVICE) GetRpl(rplName string) *REPLICA {
	for _, replica := range service.getReplicas() {
		if replica.RplName == rplName {
			return replica
		}
	}
	return nil
}

//获取副本序号
func (service *SERVICE) GetRplIndex(rplName string) int {
	for index, replica := range service.getReplicas() {
		if replica.RplName == rplName {
			return index
		}
	}
	return -1
}

//获取所有副本
func (service *SERVICE) getReplicas() []*REPLICA {
	return service.Replicas
}

//处理网络消息
func (service *SERVICE) Daq() {
	pool.RegPrivateChanStr(UPLOAD, CHAN_BUFFER)

	var ctx context.Context
	ctx, cancelDaq = context.WithCancel(context.Background())

	//定时器时间间隔
	G_samplingRate = 10
	var interval time.Duration = time.Second * time.Duration(G_samplingRate)
	var timer *time.Timer = time.NewTimer(interval)
	var pWebServices *header.SERVICE = &header.SERVICE{}
	for {
		select {
		case <-ctx.Done():
			fmt.Println("取消采集")
			pool.UnregPrivateChanStr(ctnS.DAQ)
			return
		case <-timer.C:
			pWebServices.Service = make([]header.Service, 0, SVC_NUM)
			pWebServices.Count = 1
			var webSvc header.Service
			//ServiceInfo
			webSvc.Id = service.SvcName
			webSvc.State = service.SvcStats
			webSvc.Scale = uint32(service.SvcScale)
			webSvc.ReplicaCount = uint32(len(service.Replicas))
			webSvc.CreateTime = service.CreateTime
			webSvc.StartTime = service.StartTime
			webSvc.NameSpace = service.NameSpace
			//ServiceCfg 服务配置信息 暂时不填充
			//副本信息
			for _, replica := range service.Replicas {
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

			//ctnS.Mylog.Debug(fmt.Sprintf("######向web前端发送数据		%s\n",infoStr))
			timer.Reset(interval)
		}
	}
}
