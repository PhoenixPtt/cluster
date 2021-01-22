package ctnS

import (
	"ctnCommon/ctn"
	"ctnCommon/headers"
	"ctnCommon/protocol"
	"unsafe"

	//"ctnCommon/headers"
	"ctnCommon/pool"
	"fmt"
)

type CTNS_WORK_POOL struct {
	pool.WORK_POOL
}

const DAQ = "DAQ"

//发送网络消息
func (workPool *CTNS_WORK_POOL) Send() {
	for obj := range workPool.GetSendChan() {
		pSaTruck, ok := obj.(*protocol.SA_TRUCK)
		if ok {
			byteStream, err := headers.Encode(pSaTruck) //打包
			if err != nil {
				errCode := "CTN：网络数据打包失败！"
				fmt.Println(errCode)
				continue
			}
			pool.CallbackSendCtn(pSaTruck.DesAddr, 1, 0, pSaTruck.Flag, byteStream, workPool.GetSendFunc()) //通知主线程发送数据
		}
	}
}

//处理网络消息
func (workPool *CTNS_WORK_POOL) Recv() {
	for {
		select {
		case obj := <-workPool.GetRecvChan():
			pSaTruck := obj.(*protocol.SA_TRUCK)
			switch pSaTruck.Flag {
			case ctn.FLAG_CTRL:
				if pSaTruck.Index > 0 {
					pool.AppendInt(pSaTruck.Index, pSaTruck)
				}
			}
			UpdateInfo(pSaTruck)
		}
	}
}

var (
	pCtnsWorkPool *CTNS_WORK_POOL
)

func init() {
	pWorkPool := pool.NewWorkPool()
	pCtnsWorkPool = (*CTNS_WORK_POOL)(unsafe.Pointer(pWorkPool))
}

func Config(sendObjFunc pool.SendObjFunc) {
	pCtnsWorkPool.Config(sendObjFunc)
	go pCtnsWorkPool.Send()
	go pCtnsWorkPool.Recv()
}

//获取网口发送通道
func GetSendChan() chan interface{} {
	return pCtnsWorkPool.GetSendChan()
}

//获取网口接收通道
func GetRecvChan() chan interface{} {
	return pCtnsWorkPool.GetRecvChan()
}

//获取回调函数
func GetSendFunc() (sendObjFunc pool.SendObjFunc) {
	return pCtnsWorkPool.GetSendFunc()
}
