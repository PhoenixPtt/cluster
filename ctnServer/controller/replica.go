package controller

import (
	"context"
	"ctnCommon/ctn"
	"ctnCommon/pool"
	"ctnServer/ctnS"
	"errors"
	"fmt"
	"time"
)

const (
	//RPL_STATUS_GODIRTY = iota
	//RPL_STATUS_REMOVED

	RPL_TARGET_RUNNING = "运行中"
	RPL_TARGET_REMOVED = "已删除"
	RPL_GETLOG         = "RPL_GETLOG"  //获取副本内容器日志
	RPL_INSPECT        = "RPL_INSPECT" //获取副本详情

	//ERR_RPL_NOTEXIST    = "副本不存在"
	//ERR_RPL_CTNNOTEXIST = "副本的容器不存在"
	ERR_RPL_COMMFAIL = "副本的节点通信故障"
)

func (rpl *REPLICA) SetTargetStat(targetStat string) {
	if rpl.RplTargetStat != targetStat {
		rpl.RplTargetStat = targetStat
		switch rpl.RplTargetStat {
		case RPL_TARGET_RUNNING:
			go rpl.Run()
		case RPL_TARGET_REMOVED:
			go rpl.Remove()
		}
	}
}

func (rpl *REPLICA) SetNodeStatus(nodeName string, status bool) {
	if nodeName == rpl.AgentAddr {
		if rpl.AgentStatus != status { //节点状态有变化
			rpl.AgentStatus = status //更新节点状态
			switch rpl.AgentStatus {
			case true: //上线
				//if rpl.Dirty {
				//	go rpl.Remove()
				//}
			case false: //下线
				//if !rpl.Dirty {
				fmt.Println("")
				pCtn := ctnS.GetCtn(rpl.CtnName)
				if pCtn != nil {
					pCtn.Dirty = true
					pCtn.DirtyPosition = ctn.DIRTY_POSTION_SERVER
					//rpl.Dirty = true
					//rpl= ctn.DIRTY_POSTION_SERVER
					pChan := pool.GetPrivateChanStr(rpl.SvcName) //通知服务进行调度
					var statusMap map[string]string
					statusMap = make(map[string]string)
					statusMap[rpl.RplName] = pCtn.DirtyPosition
					pChan <- statusMap
				}
				//}
			}
		}
	}
}

func (rpl *REPLICA) WatchCtn() {
	var statusMap map[string]string
	statusMap = make(map[string]string)
	//var statusMap map[string]int
	//statusMap = make(map[string]int)
	pool.RegPrivateChanStr(rpl.CtnName, 1)
	var ctx context.Context
	ctx, rpl.CancelWatchCtn = context.WithCancel(context.Background())
	//增加一个超时机制
	for {
		select {
		case <-ctx.Done():
			fmt.Println("&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&", rpl.RplName, "退出")
			pool.UnregPrivateChanStr(rpl.CtnName)
			return
		case obj := <-pool.GetPrivateChanStr(rpl.CtnName):
			//获取副本对应的容器对象
			pCtn := ctnS.GetCtn(rpl.CtnName)
			//判断容器对象信息是否过期
			if !pCtn.Dirty { //未过期
				ctnStatus := obj.(string) //获取容器状态
				//容器实际运行状态与副本目标状态做比较。如果二者不一致，需要做处理
				var ctnRunningStat bool
				if ctnStatus == "running" {
					ctnRunningStat = true
				} else {
					ctnRunningStat = false
				}

				var rplTargetRunningStat bool
				if rpl.RplTargetStat == RPL_TARGET_RUNNING {
					rplTargetRunningStat = true
				} else {
					rplTargetRunningStat = false
				}

				if ctnRunningStat != rplTargetRunningStat { //容器实际运行状态与副本目标状态不一致
					pCtn.Dirty = true                            //此时也认为容器已过期
					pCtn.DirtyPosition = ctn.DIRTY_POSTION_IMAGE //初步认为该容器的镜像有问题
					//容器过期，副本随之过期，将副本过期的消息传给其所属服务
					pChan := pool.GetPrivateChanStr(rpl.SvcName) //通知服务进行调度
					statusMap[rpl.RplName] = pCtn.DirtyPosition  //将容器信息过期的消息传给服务
					pChan <- statusMap
				}
			} else { //已过期
				//if rpl.AgentStatus { //如果节点在线，则删除副本
				//	//go rpl.Remove()
				//}
				//容器过期，副本随之过期，将副本过期的消息传给其所属服务
				pChan := pool.GetPrivateChanStr(rpl.SvcName) //通知服务进行调度
				switch rpl.RplTargetStat {
				case RPL_TARGET_REMOVED:
					if pCtn.DirtyPosition == ctn.DIRTY_POSITION_AGENT { //执行删除操作，并在agent端成功删除了
						pCtn.DirtyPosition = ctn.DIRTY_POSITION_REMOVED //正常删除
						//从容器对象池中删除
						pool.RemoveObj(pCtn.CtnName)
					}
				default:

				}

				statusMap[rpl.RplName] = pCtn.DirtyPosition //将容器信息过期的消息传给服务
				pChan <- statusMap
			}
		}
	}
}

//运行副本
func (rpl *REPLICA) Run() (err error) {
	var pCtnS *ctnS.CTNS
	var configMap map[string]string
	var log string
	var ctx context.Context
	var cancel context.CancelFunc

	err = check(rpl, RPL_TARGET_RUNNING)
	if err != nil {
		goto Error
	}

	err = rpl.checkNodeStatus()
	if err != nil {
		goto Error
	}

	if rpl.CtnName == "" {
		configMap = make(map[string]string)
		configMap[ctnS.AGENT_TRY_NUM] = fmt.Sprint(rpl.AgentTryNum)
		pCtnS = ctnS.NewCtnS(rpl.RplImage, rpl.AgentAddr, configMap)
		rpl.CtnName = pCtnS.CtnName
		go rpl.WatchCtn()
		ctnS.AddCtn(pCtnS)
	}

	//获取副本对应的容器
	pCtnS = ctnS.GetCtn(rpl.CtnName)
	err = checkCtn(pCtnS, RPL_TARGET_RUNNING)
	if err != nil {
		goto Error
	}

	ctx, cancel = context.WithTimeout(context.TODO(), time.Second*time.Duration(rpl.Timeout))
	defer cancel()
	err = pCtnS.Run(ctx)
	if err != nil {
		goto Error
	}

	log = fmt.Sprintf("%s执行Run操作执行成功。\n", rpl.RplName)
	Mylog.Info(log)
	return

Error:
	pCtnS.Dirty = true
	pCtnS.DirtyPosition = ctn.DIRTY_POSTION_SERVER
	pChan := pool.GetPrivateChanStr(rpl.SvcName) //通知服务进行调度
	statusMap := make(map[string]string)
	statusMap[rpl.RplName] = pCtnS.DirtyPosition
	pChan <- statusMap
	log = fmt.Sprintf("%s执行Run操作执行失败。错误信息：%s\n", rpl.RplName, errors.New(err.Error()))
	Mylog.Info(log)
	return
}

//删除副本
func (rpl *REPLICA) Remove() (err error) {
	var (
		pCtnS *ctnS.CTNS
		log   string
		//rplStatus map[string]int
		pChan  chan interface{}
		ctx    context.Context
		cancel context.CancelFunc
	)
	//rplStatus = make(map[string]int)

	err = check(rpl, RPL_TARGET_REMOVED)
	if err != nil {
		goto Error
	}

	err = rpl.checkNodeStatus()
	if err != nil {
		goto Error
	}

	//获取副本对应的容器
	pCtnS = ctnS.GetCtn(rpl.CtnName)
	err = checkCtn(pCtnS, RPL_TARGET_REMOVED)
	if err != nil {
		goto Error
	}

	ctx, cancel = context.WithTimeout(context.TODO(), time.Second*time.Duration(rpl.Timeout))
	defer cancel()
	err = pCtnS.Remove(ctx)
	if err != nil {
		goto Error
	}

	log = fmt.Sprintf("%s执行Remove操作执行成功。\n", rpl.RplName)
	Mylog.Info(log)
	return

Error:
	pCtnS.Dirty = true
	pCtnS.DirtyPosition = ctn.DIRTY_POSTION_SERVER
	pChan = pool.GetPrivateChanStr(rpl.SvcName) //通知服务进行调度
	statusMap := make(map[string]string)
	statusMap[rpl.RplName] = pCtnS.DirtyPosition
	pChan <- statusMap
	log = fmt.Sprintf("%s执行Remove操作执行失败。错误详情：%s\n", rpl.RplName, errors.New(err.Error()))
	Mylog.Info(log)
	return
}

//获取副本对应容器日志
func (rpl *REPLICA) GetLog() (err error) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	err = check(rpl, RPL_GETLOG)
	if err != nil {
		return
	}

	//获取副本对应的容器
	pCtnS := ctnS.GetCtn(rpl.CtnName)
	err = checkCtn(pCtnS, RPL_GETLOG)
	if err != nil {
		return
	}

	ctx, cancel = context.WithTimeout(context.TODO(), time.Second*time.Duration(rpl.Timeout))
	defer cancel()
	err = pCtnS.GetLog(ctx)
	if err != nil {
		return
	}
	return
}

//获取副本对应容器详细信息
func (rpl *REPLICA) Inspect() (err error) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	err = check(rpl, RPL_INSPECT)
	if err != nil {
		return
	}

	//获取副本对应的容器
	pCtnS := ctnS.GetCtn(rpl.CtnName)
	err = checkCtn(pCtnS, RPL_INSPECT)
	if err != nil {
		return
	}

	ctx, cancel = context.WithTimeout(context.TODO(), time.Second*time.Duration(rpl.Timeout))
	defer cancel()
	err = pCtnS.Inspect(ctx)
	if err != nil {
		return
	}
	return
}

func check(rpl *REPLICA, operType string) (err error) {
	err = nil
	if rpl == nil {
		var errStr string
		switch operType {
		case RPL_TARGET_RUNNING:
			errStr = "副本为空，无法创建并启动副本"
		case RPL_TARGET_REMOVED:
			errStr = "副本为空，无法停止并删除副本"
		case RPL_GETLOG:
			errStr = "副本为空，无法获取副本日志"
		case RPL_INSPECT:
			errStr = "副本为空，无法获取副本详细信息"
		}
		err = errors.New(errStr)
		return
	}

	if rpl.CtnName != "" {
		if operType == RPL_TARGET_RUNNING {
			errStr := "副本的容器已存在，无法再次创建并副本"
			err = errors.New(errStr)
			return
		}
	}

	if rpl.CtnName == "" {
		var errStr string
		switch operType {
		case RPL_TARGET_RUNNING:
		case RPL_TARGET_REMOVED:
			errStr = "副本的容器为空，无法停止并删除副本"
		case RPL_GETLOG:
			errStr = "副本的容器为空，无法获取副本日志"
		case RPL_INSPECT:
			errStr = "副本的容器为空，无法获取副本详细信息"
		}
		if errStr != "" {
			err = errors.New(errStr)
		}

		return
	}
	return
}

func checkCtn(pCtn *ctnS.CTNS, operType string) (err error) {
	err = nil
	if pCtn == nil {
		var errStr string
		switch operType {
		case RPL_TARGET_RUNNING:
		case RPL_TARGET_REMOVED:
			errStr = "容器为空，无法停止并删除副本"
		case RPL_GETLOG:
			errStr = "容器为空，无法获取副本日志"
		case RPL_INSPECT:
			errStr = "容器为空，无法获取副本详细信息"
		}
		if errStr != "" {
			err = errors.New(errStr)
		}
	}
	return
}

func (rpl *REPLICA) checkNodeStatus() (err error) {
	err = nil
	if !rpl.AgentStatus { //通信故障
		err = errors.New(rpl.AgentAddr + ERR_RPL_COMMFAIL)
		return
	}
	return
}
