package controller

import (
	header "clusterHeader"
	"context"
	"ctnCommon/ctn"
	"ctnCommon/pool"
	"ctnCommon/protocol"
	"ctnServer/ctnS"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
)

const UPLOAD = "UPLOAD"
const UPLOAD_EVT = "UPLOAD_EVT"
const UPLOAD_ERR = "UPLOAD_ERR"

//处理网络消息
func (pController *CONTROLLER) Daq(){
	pool.RegPrivateChanStr(UPLOAD, CHAN_BUFFER)
	pool.RegPrivateChanStr(ctnS.DAQ, CHAN_BUFFER)
	pool.RegPrivateChanStr(UPLOAD_EVT, CHAN_BUFFER)
	pool.RegPrivateChanStr(UPLOAD_ERR, CHAN_BUFFER)

	ctnInfoMap := make(map[string]types.Container, CTN_SIZE)
	ctnStatMap := make(map[string]ctn.CTN_STATS,CTN_SIZE)

	var ctx context.Context
	ctx,pController.CancelDaq=context.WithCancel(context.Background())

	for{
		select {
		case <-ctx.Done():
			fmt.Println("取消采集")
			pool.UnregPrivateChanStr(ctnS.DAQ)
			return
		case obj := <-pool.GetPrivateChanStr(ctnS.DAQ):
			pSaTruck := obj.(*protocol.SA_TRUCK)
			switch pSaTruck.Flag {
			case ctn.FLAG_CTRL:
				//不处理
			case ctn.FLAG_CTN: //更新容器信息
				for _, container := range pSaTruck.CtnList {
					ctnId := container.ID
					ctnInfoMap[ctnId] = container
				}
				//fmt.Println(pSaTruck.CtnList)
				pWebServices := ToWebService(pController, ctnInfoMap, ctnStatMap)
				select {
				case pool.GetPrivateChanStr(UPLOAD) <- pWebServices:
					//fmt.Println("111111111111111111111111111111111", pWebServices)
				default:
				}
			case ctn.FLAG_STATS: //更新资源使用情况
				for _, ctnStat := range pSaTruck.CtnStat {
					ctnId := ctnStat.ID
					ctnStatMap[ctnId] = ctnStat
				}
				//fmt.Println(pSaTruck.CtnStat)
			case ctn.FLAG_EVENT: //更新事件
				//这些信息都要返回给上层
				if len(pSaTruck.EvtMsg) > 0 {
					eventMsg := pSaTruck.EvtMsg[0]
					//fmt.Printf("%#v", eventMsg)
					pChan := pool.GetPrivateChanStr(UPLOAD_EVT)
					select {
					case pChan <- eventMsg:
					default:
					}
				}

				if len(pSaTruck.ErrMsg) > 0 {
					//更新错误信息
					eventErr := pSaTruck.ErrMsg[0]
					pChan := pool.GetPrivateChanStr(UPLOAD_ERR)
					select {
					case pChan <- eventErr:
					default:
					}
				}
			}
		}
	}
}

func (pController *CONTROLLER) WaitWebService() (pWebServices *header.SERVICE) {
	select {
	case obj := <-pool.GetPrivateChanStr(UPLOAD)://类型：header.SERVICE
		pWebServices=obj.(*header.SERVICE)
		fmt.Println("2222222222222222222222222222", pWebServices)
	}
	return
}


func (pController *CONTROLLER) WaitEventMsg() (evtMsg events.Message) {
	select {
	case obj:=<-pool.GetPrivateChanStr(UPLOAD_EVT)://类型：event.Message
		evtMsg=obj.(events.Message)
	}
	return
}


func (pController *CONTROLLER) WaitEventErr() (err error) {
	select {
	case obj:=<-pool.GetPrivateChanStr(UPLOAD_ERR)://类型：error
		err=obj.(error)
	}
	return
}

