package ctnA

import (
	"context"
	"ctnCommon/ctn"
	"ctnCommon/headers"
	"ctnCommon/pool"
	"fmt"
	"github.com/docker/docker/api/types/events"
	"unsafe"

	"github.com/docker/docker/api/types"
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
	evtMsgChan, _ = cli.Events(ctx, options)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Stop CTN Stats")
			return
		default:
			//向Server端发送事件
			var pSaTruck ctn.SA_TRUCK
			pool.AddIndex()
			pSaTruck.Flag = ctn.FLAG_EVENT
			pSaTruck.Index = pool.GetIndex()
			pSaTruck.Addr = distAddr

			for val := range evtMsgChan {
				//更新本地容器池
				if val.Type == "container"{
					pCtn := GetCtnFromID(val.ID)
					if pCtn!=nil{
						pCtn.CtnAction = val.Action
						pCtn.CtnActionTime = headers.ToStringInt(val.TimeNano, headers.TIME_LAYOUT_NANO)
						pCtn.CtnActionTimeInt = val.TimeNano
					}
				}

				pSaTruck.EvtMsg = append(pSaTruck.EvtMsg, val)
			}

			pObj := (*interface{})(unsafe.Pointer(&pSaTruck))
			GetSendChan() <- pObj
		}
	}

	fmt.Println("exit CtnEvents")
}
