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
	RPL_STATUS_GODIRTY = iota
	RPL_STATUS_REMOVED

	RPL_TARGET_RUNNING = "运行中"
	RPL_TARGET_REMOVED = "已删除"
	RPL_GETLOG         = "RPL_GETLOG"  //获取副本内容器日志
	RPL_INSPECT        = "RPL_INSPECT" //获取副本详情

	ERR_RPL_NOTEXIST    = "副本不存在"
	ERR_RPL_CTNNOTEXIST = "副本的容器不存在"
	ERR_RPL_COMMFAIL    = "副本的节点通信故障"
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
				if rpl.Dirty {
					go rpl.Remove()
				}
			case false: //下线
				if !rpl.Dirty {
					fmt.Println("")
					//rpl.Dirty = true
					pChan := pool.GetPrivateChanStr(rpl.SvcName) //通知服务进行调度
					var statusMap map[string]int
					statusMap = make(map[string]int)
					statusMap[rpl.RplName] = RPL_STATUS_GODIRTY
					pChan <- statusMap
				}
			}
		}
	}
}

func (rpl *REPLICA) WatchCtn() {
	var statusMap map[string]int
	statusMap = make(map[string]int)
	pool.RegPrivateChanStr(rpl.CtnName, 1)
	var ctx context.Context
	ctx, rpl.CancelWatchCtn = context.WithCancel(context.Background())
	for {
		select {
		case <-ctx.Done():
			log := fmt.Sprintf("停止监控副本%s。\n", rpl.RplName)
			Mylog.Info(log)
			pool.UnregPrivateChanStr(rpl.CtnName)
			return
		case obj := <-pool.GetPrivateChanStr(rpl.CtnName):
			if rpl.Dirty {
				if rpl.AgentStatus { //如果节点在线，则删除副本
					fmt.Println("删除有污点的副本", rpl.RplName, rpl.RplImage)
					go rpl.Remove()
				}
			} else {
				ctnStatus := obj.(string)
				switch ctnStatus {
				case ctnS.CTN_STATUS_RUNNING:
					switch rpl.RplTargetStat {
					case RPL_TARGET_RUNNING: //副本状态与容器状态一致
					case RPL_TARGET_REMOVED: //副本状态与容器状态不一致
						if !rpl.Dirty {
							fmt.Println("通知服务进行调度1", rpl.SvcName, rpl.RplName)
							pChan := pool.GetPrivateChanStr(rpl.SvcName) //通知服务进行调度
							statusMap[rpl.RplName] = RPL_STATUS_GODIRTY
							pChan <- statusMap
						}
					}
				case ctnS.CTN_STATUS_NOTRUNNING:
					switch rpl.RplTargetStat {
					case RPL_TARGET_RUNNING: //副本状态与容器状态不一致
						if !rpl.Dirty {
							fmt.Println("通知服务进行调度2", rpl.SvcName, rpl.RplName)
							if rpl.AgentStatus { //如果节点在线，则删除副本
								go rpl.Remove()
							}
							pChan := pool.GetPrivateChanStr(rpl.SvcName) //通知服务进行调度
							statusMap[rpl.RplName] = RPL_STATUS_GODIRTY
							pChan <- statusMap
						}
					case RPL_TARGET_REMOVED: //副本状态与容器状态一致
					}
				}
			}
		}
	}
}

//运行副本
func (rpl *REPLICA) Run() (errType string, err error) {
	var pCtnS *ctnS.CTNS
	var configMap map[string]string
	var log string
	var ctx context.Context
	var cancel context.CancelFunc

	errType, err = check(rpl, RPL_TARGET_RUNNING)
	if err != nil {
		goto Error
	}

	errType, err = rpl.checkNodeStatus()
	if err != nil {
		goto Error
	}

	if rpl.CtnName == "" {
		pCtnS = ctnS.NewCtnS(rpl.RplImage, rpl.AgentAddr, configMap)
		rpl.CtnName = pCtnS.CtnName
		go rpl.WatchCtn()
		ctnS.AddCtn(pCtnS)
	}

	//获取副本对应的容器
	pCtnS = ctnS.GetCtn(rpl.CtnName)
	errType, err = checkCtn(pCtnS, RPL_TARGET_RUNNING)
	if err != nil {
		goto Error
	}

	ctx, cancel = context.WithTimeout(context.TODO(), time.Second*time.Duration(rpl.Timeout))
	defer cancel()
	errType, err = pCtnS.Run(ctx)
	if err != nil {
		goto Error
	}

	log = fmt.Sprintf("%s执行Run操作执行成功。\n", rpl.RplName)
	Mylog.Info(log)
	return

Error:
	if !rpl.Dirty {
		pChan := pool.GetPrivateChanStr(rpl.SvcName) //通知服务进行调度
		statusMap := make(map[string]int)
		statusMap[rpl.RplName] = RPL_STATUS_GODIRTY
		pChan <- statusMap
	}
	log = fmt.Sprintf("%s执行Run操作执行失败。错误类型：%s；错误详情：%s\n", rpl.RplName, errType, errors.New(err.Error()))
	Mylog.Info(log)
	return
}

//删除副本
func (rpl *REPLICA) Remove() (errType string, err error) {
	var (
		pCtnS     *ctnS.CTNS
		log       string
		rplStatus map[string]int
		pChan     chan interface{}
		ctx       context.Context
		cancel    context.CancelFunc
	)
	rplStatus = make(map[string]int)

	errType, err = check(rpl, RPL_TARGET_REMOVED)
	if err != nil {
		goto Error
	}

	errType, err = rpl.checkNodeStatus()
	if err != nil {
		goto Error
	}

	//获取副本对应的容器
	pCtnS = ctnS.GetCtn(rpl.CtnName)
	errType, err = checkCtn(pCtnS, RPL_TARGET_REMOVED)
	if err != nil {
		goto Error
	}

	ctx, cancel = context.WithTimeout(context.TODO(), time.Second*time.Duration(rpl.Timeout))
	defer cancel()
	errType, err = pCtnS.Remove(ctx)
	if err != nil {
		goto Error
	}

	pChan = pool.GetPrivateChanStr(rpl.SvcName) //通知服务删除该副本
	rplStatus[rpl.RplName] = RPL_STATUS_REMOVED
	pChan <- rplStatus
	log = fmt.Sprintf("%s执行Remove操作执行成功。\n", rpl.RplName)
	Mylog.Info(log)
	return
Error:
	if !rpl.Dirty {
		pChan := pool.GetPrivateChanStr(rpl.SvcName) //通知服务进行调度
		statusMap := make(map[string]int)
		statusMap[rpl.RplName] = RPL_STATUS_GODIRTY
		pChan <- statusMap
	}
	log = fmt.Sprintf("%s执行Remove操作执行失败。错误类型：%s；错误详情：%s\n", rpl.RplName, errType, errors.New(err.Error()))
	Mylog.Info(log)
	return
}

//获取副本对应容器日志
func (rpl *REPLICA) GetLog() (log string, err error) {
	_, err = check(rpl, RPL_GETLOG)
	if err != nil {
		return
	}

	//获取副本对应的容器
	pCtnS := ctnS.GetCtn(rpl.CtnName)
	_, err = checkCtn(pCtnS, RPL_GETLOG)
	if err != nil {
		return
	}

	log, err = pCtnS.GetLog()
	if err != nil {
		return
	}
	return
}

//获取副本对应容器详细信息
func (rpl *REPLICA) Inspect() (ctnInspect ctn.CTN_INSPECT, err error) {
	_, err = check(rpl, RPL_INSPECT)
	if err != nil {
		return
	}

	//获取副本对应的容器
	pCtnS := ctnS.GetCtn(rpl.CtnName)
	_, err = checkCtn(pCtnS, RPL_INSPECT)
	if err != nil {
		return
	}

	ctnInspect, err = pCtnS.Inspect()
	if err != nil {
		return
	}
	return
}

func check(rpl *REPLICA, operType string) (errType string, err error) {
	err = nil
	if rpl == nil {
		errType = ERR_RPL_NOTEXIST
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
		errType = ERR_RPL_CTNNOTEXIST
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

func checkCtn(pCtn *ctnS.CTNS, operType string) (errType string, err error) {
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

func (rpl *REPLICA) checkNodeStatus() (errType string, err error) {
	err = nil
	if !rpl.AgentStatus { //通信故障
		errType = ERR_RPL_COMMFAIL
		err = errors.New(rpl.AgentAddr + ERR_RPL_COMMFAIL)
		return
	}
	return
}
