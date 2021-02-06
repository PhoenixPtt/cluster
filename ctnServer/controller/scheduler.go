package controller

import (
	"ctnCommon/headers"
	"ctnServer/ctnS"
	"errors"
	"fmt"
	"math"
	"math/rand"
)

func (pSvc *SERVICE) schedule(rplName string) (err error) {
	//判断是否满足执行动作的前提条件
	var info string
	switch pSvc.SvcStats {
	case SVC_DEFAULT:
		info = infoString(pSvc.SvcName, "未创建，无法进行调度操作。")
	case SVC_CREATED:
		info = infoString(pSvc.SvcName, "已创建但未启动，无法执行调度操作。")
	case SVC_STOPPED:
		info = infoString(pSvc.SvcName, "已停止，无法执行调度操作。")
	default:
		pSvc.mutex.Lock()

		ctnS.Mylog.Info("-----------------------调度-----------------------")
		pSvc.SvcStats = SVC_RUNNING
		//迁移到新的副本
		var agentAddrNumMap map[string]int
		agentAddrNumMap, err = pSvc.EScale(1)
		if err != nil {
			fmt.Println("调度失败…………………………………………………………………………………………………………………………")
			return
		}
		for addr, rplNum := range agentAddrNumMap {
			for i := 0; i < rplNum; i++ {
				pRpl := pSvc.NewRpl(pSvc.SvcName+"_"+headers.UniqueId(), pSvc.Image, addr)
				pRpl.SetTargetStat(RPL_TARGET_RUNNING) //设置副本的目标状态
			}
		}
		pSvc.mutex.Unlock()
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

/*服务扩容：增加副本的数量*/
func (pSvc *SERVICE) EScale(scaleNum int) (agentAddrNumMap map[string]int, err error) {
	var agentAddr []string
	//从agent列表中找到在线的agent进行分配
	for node, status := range pSvc.NodeStatusMap {
		if status == true {
			agentAddr = append(agentAddr, node)
		}
	}

	//随机选择agent
	agentAddrNumMap = make(map[string]int)
	agentAddrNumMap = RandomSelect(agentAddr, scaleNum)
	if len(agentAddrNumMap) == 0 {
		err = errors.New("没有找到可供容器运行的agent！")
		return
	}

	return
}

/*服务缩容：停止部分或全部副本*/
func (pSvc *SERVICE) DScale(scaleNum int) (rplNames []string, err error) {
	//随机选择运行中的副本
	rplNames = make([]string, 0, 100)
	for _, pRpl := range pSvc.getReplicas() {
		if scaleNum > 0 {
			pCtnS := ctnS.GetCtn(pRpl.CtnName)
			fmt.Println(pCtnS.CtnID, pCtnS.CtnName)

			if pCtnS != nil {
				if pCtnS.State == "running" {
					rplNames = append(rplNames, pRpl.RplName)
					scaleNum--
				}
			}
		}
	}
	return
}

/*计算需要扩容或者缩容的副本数量*/
func (pSvc *SERVICE) getScaleNum() (dir int, scaleNum int) {
	var activeCtnNum int = 0
	pSvc.mutex.Lock()
	replicas := pSvc.getReplicas()
	fmt.Println("服务包含的副本数量：", len(replicas))
	for _, rpl := range replicas {
		pCtn := ctnS.GetCtn(rpl.CtnName)
		if pCtn == nil {
			fmt.Printf("副本%s对应的容器不存在！容器名称：%s\n", rpl.RplName, rpl.CtnName)
			continue
		}
		if pCtn.Dirty == false {
			if pCtn.State == "running" {
				activeCtnNum++
			}
		}
	}
	pSvc.mutex.Unlock()

	switch pSvc.SvcStats {
	case SVC_DEFAULT, SVC_CREATED:
		return 0, 0
	case SVC_RUNNING:
		scaleNum = pSvc.SvcScale - activeCtnNum
	case SVC_STOPPED, SVC_REMOVED:
		scaleNum = -activeCtnNum
	}

	fmt.Println("运行的副本数量：", activeCtnNum, "|", "需要调整的副本数量：", scaleNum)
	switch {
	case scaleNum > 0:
		return 1, int(math.Abs(float64(scaleNum)))
	case scaleNum < 0:
		return -1, int(math.Abs(float64(scaleNum)))
	}
	return 0, 0
}

func RandomSelect(addrs []string, selectNum int) map[string]int {
	var (
		selectionMap map[string]int
	)

	aLen := len(addrs) //获得切片长度
	if aLen == 0 {
		fmt.Println("没有节点可供分配。")
		return selectionMap
	}

	var selection []string
	selection = make([]string, 0, selectNum)
	for i := 0; i < selectNum; i++ {
		index := rand.Intn(aLen) //生成切片序号的随机数
		selection = append(selection, addrs[index])
	}

	selectionMap = make(map[string]int)
	for _, val := range selection {
		_, ok := selectionMap[val]
		if !ok {
			selectionMap[val] = 1
		} else {
			selectionMap[val]++
		}
	}

	return selectionMap
}
