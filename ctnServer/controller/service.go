package controller

import (
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
			//rplStatusMap := obj.(map[string]int)
			rplStatusMap := obj.(map[string]string)
			for rplName, status := range rplStatusMap {
				//rpl := service.GetRpl(rplName)
				switch status {
				case ctn.DIRTY_POSITION_REMOVED: //被正常删除
					service.DelRpl(rplName) //在副本层已经将容器对象从容器对象池中删除，服务层仅需删除副本
				case ctn.DIRTY_POSITION_DOCKER: //单纯容器在docker服务端被非正常手段删除
					//执行删除操作，注意在agent端不需要执行docker操作，直接删除agent端的容器对象即可
					//rpl.SetTargetStat(RPL_TARGET_REMOVED)
					//go service.schedule(rplName) //重新调度
				case ctn.DIRTY_POSITION_AGENT: //server端与agent端不同步导致的容器对象信息过期
					//删除server端的信息：包括副本信息和容器对象池中的信息
					//要不要删除由用户决定，进行调度
					//go service.schedule(rplName) //重新调度
				case ctn.DIRTY_POSTION_IMAGE: //server端操作与操作结果不一致，在docker服务器中执行失败
					//执行删除操作，在agent端需要执行docker操作
					//由用户决定是否继续调度，进行调度
					//rpl.SetTargetStat(RPL_TARGET_REMOVED)
					//go service.schedule(rplName) //重新调度
				case ctn.DIRTY_POSTION_SERVER: //server端操作执行超时导致的容器对象过期，网络不通或超时时间设置过短
					//执行删除操作，仅需删除server端的容器对象，并进行调度
					//rpl.SetTargetStat(RPL_TARGET_REMOVED)
					//go service.schedule(rplName) //重新调度
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
		pChan <- svcStatusMap
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
		Mylog.Info("-----------------------创建服务-----------------------")
		pSvc.SvcStats = SVC_CREATED
		pSvc.CreateTime = headers.ToString(time.Now(), headers.TIME_LAYOUT_NANO) //创建时间
	}

	switch info {
	case "":
		info = infoString(pSvc.SvcName, pSvc.SvcStats)
	default:
		err = errors.New(info)
	}

	Mylog.Info(info)
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
		Mylog.Info("-----------------------启动服务-----------------------")
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

	Mylog.Info(info)
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
		Mylog.Info("-----------------------调整规模-----------------------")
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

	Mylog.Info(info)
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
		Mylog.Info("-----------------------停止-----------------------")
		pSvc.SvcStats = SVC_STOPPED
		pSvc.updateRpl() //根据具体情况增删副本
	}

	switch info {
	case "":
		info = infoString(pSvc.SvcName, pSvc.SvcStats)
	default:
		err = errors.New(info)
	}

	Mylog.Info(info)
	return
}

//对应用户删除操作
func (pSvc *SERVICE) Remove() (err error) {
	var info string //判断是否满足执行动作的前提条件
	Mylog.Info("-----------------------删除-----------------------")
	pSvc.SvcStats = SVC_REMOVED
	pSvc.updateRpl() //根据具体情况增删副本

	switch info {
	case "":
		info = infoString(pSvc.SvcName, pSvc.SvcStats)
	default:
		err = errors.New(info)
	}

	Mylog.Info(info)
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
