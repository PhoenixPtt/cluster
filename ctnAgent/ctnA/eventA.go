package ctnA

import (
	"context"
	"ctnCommon/ctn"
	"ctnCommon/headers"
	"ctnCommon/pool"
	"ctnCommon/protocol"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

var (
	Cancel_Evt context.CancelFunc
)

// 获取容器事件
func CtnEvents(distAddr string) {
	ctx := context.Background()
	ctx, Cancel_Evt = context.WithCancel(ctx)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Errorf(err.Error())
	}

	var options types.EventsOptions
	options.Since = "00"
	evtMsgChan := make(<-chan events.Message, 100)
	errMsgChan := make(<-chan error, 100)
	evtMsgChan, errMsgChan = cli.Events(ctx, options)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Stop CTN Stats")
			return
		case val:=<-evtMsgChan:
			//向Server端发送事件
			var pSaTruck protocol.SA_TRUCK
			pool.AddIndex()
			pSaTruck.Flag = ctn.FLAG_EVENT
			pSaTruck.Index = pool.GetIndex()
			pSaTruck.Addr = distAddr

			//更新本地容器池
			if val.Type == "container"{
				pCtn := GetCtnFromID(val.ID)
				if pCtn!=nil{
					pCtn.CtnAction = val.Action
					pCtn.CtnActionTime = headers.ToStringInt(val.TimeNano, headers.TIME_LAYOUT_NANO)
					pCtn.CtnActionTimeInt = val.TimeNano
				}
			}

			pSaTruck.EvtMsg = make([]events.Message,0,1)
			pSaTruck.EvtMsg = append(pSaTruck.EvtMsg, val)
			GetSendChan() <- &pSaTruck
		case errMsg:=<- errMsgChan:
			//向Server端发送事件
			var pSaTruck protocol.SA_TRUCK
			pool.AddIndex()
			pSaTruck.Flag = ctn.FLAG_EVENT
			pSaTruck.Index = pool.GetIndex()
			pSaTruck.Addr = distAddr

			pSaTruck.ErrMsg = make([]error,0,1)
			pSaTruck.ErrMsg = append(pSaTruck.ErrMsg, errMsg)
			GetSendChan() <- &pSaTruck
		}
	}
	fmt.Println("exit CtnEvents")
}
