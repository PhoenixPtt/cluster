package ctnA

import (
	"ctnCommon/ctn"
	"ctnCommon/headers"
	"ctnCommon/pool"
	"fmt"
	"unsafe"
)

type CTNA_WORKPOOL struct {
	pool.WORK_POOL
}

//发送网络消息
func (workPool *CTNA_WORKPOOL) Send()  {
	for obj:= range workPool.GetSendChan(){
		pSaTruck := obj.(*ctn.SA_TRUCK)
		byteStream, err := headers.Encode(pSaTruck)//打包
		if err != nil {
			errCode := "CTN：网络数据打包失败！"
			fmt.Println(errCode)
			continue
		}
		pool.CallbackSendCtn("", 1, 0, pSaTruck.Flag, byteStream, workPool.GetSendFunc())//通知主线程发送数据
	}
}

//处理网络消息
func (workPool *CTNA_WORKPOOL) Recv(){
	for{
		select {
		case obj:=<-workPool.GetRecvChan():
			pSaTruck := obj.(*ctn.SA_TRUCK)
			if pSaTruck.Flag!=ctn.FLAG_CTRL{//仅接收控制指令
				continue
			}

			if len(pSaTruck.Req_Ans)<1{
				continue
			}

			reqAns := pSaTruck.Req_Ans[0]
			reqAns.CtnErrType = make([]string, 1)

			//获取对应的容器
			pCtnA :=GetCtn(reqAns.CtnName)

			switch reqAns.CtnOper {
			case ctn.CREATE,ctn.RUN:
				if pCtnA == nil{
					pCtnA = &CTNA{}
					pCtnA.CtnName = reqAns.CtnName
					pCtnA.Image = reqAns.CtnImage
				}
				errType, err:=OperateWithStratgy(pCtnA, reqAns.CtnOper)
				reqAns.CtnID = make([]string, 1)
				reqAns.CtnID[0] = pCtnA.ID
				reqAns.CtnErrType[0] = errType
				reqAns.CtnErr = err

				if err==nil{
					AddCtn(pCtnA)
				}
			default:
				if pCtnA == nil{
					continue
				}
				switch reqAns.CtnOper {
				case ctn.START,ctn.STOP,ctn.KILL:
					errType, err:=OperateWithStratgy(pCtnA, reqAns.CtnOper)
					reqAns.CtnErrType[0] = errType
					reqAns.CtnErr = err
				case ctn.REMOVE:
					errType, err:=OperateWithStratgy(pCtnA, reqAns.CtnOper)
					reqAns.CtnErrType[0] = errType
					reqAns.CtnErr = err

					if err==nil{
						RemoveCtn(pCtnA.CtnName)
					}
				case ctn.GETLOG:
					log,err:=pCtnA.GetLog()
					reqAns.CtnLog = make([]string,1)
					reqAns.CtnLog[0]=log
					reqAns.CtnErr = err
				case ctn.INSPECT:
					ctnInspect,err:=pCtnA.Inspect()
					reqAns.CtnInspect = make([]ctn.CTN_INSPECT,1)
					reqAns.CtnInspect[0]=ctnInspect
					reqAns.CtnErr = err
				}
			}

			pSaTruck.Req_Ans[0] = reqAns
			pSendChan:=workPool.GetSendChan()
			pSendChan <- pSaTruck
		default:
		}
	}
}

var(
	pWorkPool *CTNA_WORKPOOL
)

func init() () {
	pCtnWorkPool := pool.NewWorkPool()
	pWorkPool = (*CTNA_WORKPOOL)(unsafe.Pointer(pCtnWorkPool))
}

func Config(sendObjFunc pool.SendObjFunc)  {
	pWorkPool.Config(sendObjFunc)
	go pWorkPool.Send()
	go pWorkPool.Recv()
}

//获取网口发送通道
func GetSendChan() chan interface{}{
	return pWorkPool.GetSendChan()
}

//获取网口接收通道
func GetRecvChan() chan interface{}  {
	return pWorkPool.GetRecvChan()
}

//获取回调函数
func GetSendFunc() (sendObjFunc pool.SendObjFunc) {
	return pWorkPool.GetSendFunc()
}




