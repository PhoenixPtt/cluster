package ctnA

import (
	"context"
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
			handleEventMessage(val)
		case errMsg:=<- errMsgChan:
			handleErrorMessage(errMsg)
		}
	}
	fmt.Println("exit CtnEvents")
}
