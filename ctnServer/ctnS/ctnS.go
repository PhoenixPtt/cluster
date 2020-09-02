package ctnS

import (
	"context"
	"ctnCommon/ctn"
	"ctnCommon/pool"
	"errors"
	"time"
)

const (
	ERR_CTN_NILANS  = "通信：收到空的应答"
	ERR_CTN_TIMEOUT = "通信：接收应答超时 "

	CTN_STATUS_RUNNING    = "running"
	CTN_STATUS_NOTRUNNING = "not running"
)

//Server端结构体声明
type CTNS struct {
	ctn.CTN
	TimeOut int //超时时间
}

//创建容器
func (pCtnS *CTNS) Create() (errType string, err error) {
	errType, err = pCtnS.Oper(ctn.CREATE)
	return
}

//启动容器
func (pCtnS *CTNS) Start() (errType string, err error) {
	errType, err = pCtnS.Oper(ctn.START)
	return
}

//运行容器
func (pCtnS *CTNS) Run() (errType string, err error) {
	errType, err = pCtnS.Oper(ctn.RUN)
	return
}

//停止容器
func (pCtnS *CTNS) Stop() (errType string, err error) {
	errType, err = pCtnS.Oper(ctn.STOP)
	return
}

//强制停止容器
func (pCtnS *CTNS) Kill() (errType string, err error) {
	errType, err = pCtnS.Oper(ctn.KILL)
	return
}

//删除容器
func (pCtnS *CTNS) Remove() (errType string, err error) {
	errType, err = pCtnS.Oper(ctn.REMOVE)
	return
}

//获取容器日志
//注意：容器被删除之后无法获取容器日志
func (pCtnS *CTNS) GetLog() (log string, err error) {
	pool.AddIndex()
	pSaTruck := &ctn.SA_TRUCK{}
	pSaTruck.Flag = ctn.FLAG_CTRL
	pSaTruck.Index = pool.GetIndex()
	pSaTruck.Addr = pCtnS.AgentAddr
	pSaTruck.Req_Ans = make([]ctn.REQ_ANS, 1)
	pSaTruck.Req_Ans[0].CtnOper = ctn.GETLOG
	pSaTruck.Req_Ans[0].CtnName = pCtnS.CtnName

	pool.RegPrivateChanInt(pSaTruck.Index,1)
	pPrivateChan := pool.GetPrivateChanInt(pSaTruck.Index)
	GetSendChan() <- pSaTruck
	ctx, cancel:=context.WithTimeout(context.Background(), time.Second * time.Duration(5))
	select {
	case <-ctx.Done():
		close(pPrivateChan)
		cancel()
		return "", errors.New(ERR_CTN_TIMEOUT)
	case obj := <-pPrivateChan:
		pSaAnsTruck := obj.(*ctn.SA_TRUCK)
		if len(pSaAnsTruck.Req_Ans) < 1 {
			return ERR_CTN_NILANS, errors.New(ERR_CTN_NILANS)
		}
		reqAns := pSaAnsTruck.Req_Ans[0]
		switch reqAns.CtnOper {
		case ctn.GETLOG:
			log, err = pCtnS.GetLog()
			return
		}
	}

	return
}

//查看容器详细信息
func (pCtnS *CTNS) Inspect() (ctnInspect ctn.CTN_INSPECT, err error) {
	pool.AddIndex()
	pSaTruck := &ctn.SA_TRUCK{}
	pSaTruck.Flag = ctn.FLAG_CTRL
	pSaTruck.Index = pool.GetIndex()
	pSaTruck.Addr = pCtnS.AgentAddr
	pSaTruck.Req_Ans = make([]ctn.REQ_ANS, 1)
	pSaTruck.Req_Ans[0].CtnOper = ctn.GETLOG
	pSaTruck.Req_Ans[0].CtnName = pCtnS.CtnName

	pool.RegPrivateChanInt(pSaTruck.Index, 1)
	pPrivateChan := pool.GetPrivateChanInt(pSaTruck.Index)
	GetSendChan() <- pSaTruck
	ctx, cancel:=context.WithTimeout(context.Background(), time.Second * time.Duration(5))
	select {
	case <-ctx.Done():
		close(pPrivateChan)
		cancel()
		return ctnInspect, errors.New(ERR_CTN_TIMEOUT)
	case obj := <-pPrivateChan:
		pSaAnsTruck := obj.(*ctn.SA_TRUCK)
		if len(pSaAnsTruck.Req_Ans) < 1 {
			return ctnInspect, errors.New(ERR_CTN_NILANS)
		}
		reqAns := pSaAnsTruck.Req_Ans[0]
		switch reqAns.CtnOper {
		case ctn.INSPECT:
			ctnInspect, err = pCtnS.Inspect()
		}
		return
	}

	return
}

//容器操作回路
func (pCtnS *CTNS) Oper(operFlag string) (errType string, err error) {
	pool.AddIndex()
	pSaTruck := &ctn.SA_TRUCK{}
	pSaTruck.Flag = ctn.FLAG_CTRL
	pSaTruck.Index = pool.GetIndex()
	pSaTruck.Addr = pCtnS.AgentAddr
	pSaTruck.Req_Ans = make([]ctn.REQ_ANS, 1)
	pSaTruck.Req_Ans[0].CtnOper = operFlag
	pSaTruck.Req_Ans[0].CtnName = pCtnS.CtnName
	pSaTruck.Req_Ans[0].CtnImage = pCtnS.Image

	pool.RegPrivateChanInt(pSaTruck.Index, 1)
	pPrivateChan := pool.GetPrivateChanInt(pSaTruck.Index)
	GetSendChan() <- pSaTruck
	ctx, cancel:=context.WithTimeout(context.Background(), time.Second * time.Duration(5))
	select {
	case <-ctx.Done():
		close(pPrivateChan)
		cancel()
		return ERR_CTN_TIMEOUT, errors.New(ERR_CTN_TIMEOUT)
	case obj := <-pPrivateChan:
		pSaAnsTruck := obj.(*ctn.SA_TRUCK)
		if len(pSaAnsTruck.Req_Ans) < 1 {
			return ERR_CTN_NILANS, errors.New(ERR_CTN_NILANS)
		}
		reqAns := pSaAnsTruck.Req_Ans[0]
		switch reqAns.CtnOper {
		case ctn.CREATE, ctn.RUN:
			UpdateCtnID(pCtnS, reqAns.CtnID[0])
			pCtnS.State = reqAns.CtnState
			errType = reqAns.CtnErrType[0]
			err = reqAns.CtnErr
		default:
			switch reqAns.CtnOper {
			case ctn.START, ctn.STOP, ctn.KILL:
				pCtnS.State = reqAns.CtnState
				errType = reqAns.CtnErrType[0]
				err = reqAns.CtnErr
			case ctn.REMOVE:
				pCtnS.State = reqAns.CtnState
				errType = reqAns.CtnErrType[0]
				err = reqAns.CtnErr

				if err == nil {
					RemoveCtn(pCtnS.CtnName)
				}
			}
		}
		return
	}

	return
}
