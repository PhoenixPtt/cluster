package main

import (
	"clusterHeader"
	"context"
	"tcpSocket"
	"time"

	"github.com/docker/docker/api/types/events"

	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var (
	Cancel_Evt context.CancelFunc
)

// 获取容器事件
func CtnEvents() {
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
			for val := range evtMsgChan {
				fmt.Println(time.Unix(0, val.TimeNano).Format("2006-01-02 15:04:05.000"), val.Type, val.Status, val.Actor)
				byteStream, _ := header.Encode(&val)
				//发送容器资源状态数据
				tcpSocket.WriteData("", 1, 0, FLAG_EVENT, byteStream)
			}
		}

	}

	//发送容器资源状态数据
	//writeData("", tcpSocket.TCP_TYPE_MONITOR, 0, FLAG_EVENT, byteStream)
}

