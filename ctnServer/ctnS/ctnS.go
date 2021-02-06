package ctnS

import (
	"context"
	"ctnCommon/ctn"
	"ctnCommon/pool"
	"ctnCommon/protocol"
	"errors"
	"sync"
)

const (
	ERR_CTN_NILANS  = "通信：收到空的应答"
	ERR_CTN_TIMEOUT = "通信：接收应答超时 "
)

//Server端结构体声明
type CTNS struct {
	OperMap      map[int]string //记录操作及响应情况
	OperMapMutex sync.Mutex
	ctn.CTN
}

//创建容器
func (pCtnS *CTNS) Create(ctx context.Context) (err error) {
	err = pCtnS.Oper(ctx, ctn.CREATE)
	return
}

//启动容器
func (pCtnS *CTNS) Start(ctx context.Context) (err error) {
	err = pCtnS.Oper(ctx, ctn.START)
	return
}

//运行容器
func (pCtnS *CTNS) Run(ctx context.Context) (err error) {
	err = pCtnS.Oper(ctx, ctn.RUN)
	return
}

//停止容器
func (pCtnS *CTNS) Stop(ctx context.Context) (err error) {
	err = pCtnS.Oper(ctx, ctn.STOP)
	return
}

//强制停止容器
func (pCtnS *CTNS) Kill(ctx context.Context) (err error) {
	err = pCtnS.Oper(ctx, ctn.KILL)
	return
}

//删除容器
func (pCtnS *CTNS) Remove(ctx context.Context) (err error) {
	err = pCtnS.Oper(ctx, ctn.REMOVE)
	return
}

//获取容器日志
//注意：容器被删除之后无法获取容器日志
func (pCtnS *CTNS) GetLog(ctx context.Context) (err error) {
	err = pCtnS.Oper(ctx, ctn.GETLOG)
	return
}

//查看容器详细信息
func (pCtnS *CTNS) Inspect(ctx context.Context) (err error) {
	err = pCtnS.Oper(ctx, ctn.INSPECT)
	return
}

//容器操作回路
func (pCtnS *CTNS) Oper(ctx context.Context, operFlag string) (err error) {
	pool.AddIndex()
	pCtnS.OperType = operFlag
	pCtnS.OperErr = ""
	pCtnS.Dirty = false
	pCtnS.DirtyPosition = ""
	pCtnS.OperMapMutex.Lock()
	pCtnS.OperMap[pool.GetIndex()] = pCtnS.OperType
	pCtnS.OperMapMutex.Unlock()

	pSaTruck := &protocol.SA_TRUCK{}
	pSaTruck.Flag = ctn.FLAG_CTRL
	pSaTruck.Index = pool.GetIndex()
	pSaTruck.DesAddr = pCtnS.AgentAddr
	pSaTruck.Req_Ans = make([]protocol.REQ_ANS, 1)
	pSaTruck.Req_Ans[0].CtnOper = operFlag
	pSaTruck.Req_Ans[0].CtnName = pCtnS.CtnName
	pSaTruck.Req_Ans[0].CtnImage = pCtnS.Image
	pSaTruck.Req_Ans[0].AgentTryNum = pCtnS.AgentTryNum

	pool.RegPrivateChanInt(pSaTruck.Index, 1)
	pPrivateChan := pool.GetPrivateChanInt(pSaTruck.Index)
	select {
	case GetSendChan() <- pSaTruck:
	default:

	}

	select {
	case <-ctx.Done():
		pool.UnregPrivateChanInt(pSaTruck.Index)
		pCtnS.OperMapMutex.Lock()
		_, ok := pCtnS.OperMap[pSaTruck.Index]
		if ok {
			delete(pCtnS.OperMap, pSaTruck.Index)
		}
		pCtnS.OperMapMutex.Unlock()
		return errors.New(ERR_CTN_TIMEOUT)
	case obj := <-pPrivateChan:
		pSaAnsTruck := obj.(*protocol.SA_TRUCK)
		if len(pSaAnsTruck.Req_Ans) < 1 {
			return errors.New(ERR_CTN_NILANS)
		}
		reqAns := pSaAnsTruck.Req_Ans[0]
		if reqAns.CtnErr == "nil" {
			err = nil
		} else {
			err = errors.New(reqAns.CtnErr)
		}
		return
	}

	return
}
