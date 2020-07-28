package method

import (
	"context"
	"fmt"
	//"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const (
	RUN_CTN = iota
	ALL_CTN
)

//获取容器在容器列表中的序号
func CtnIndex(ctnID string, flag int) int {
	ctns, _ := CtnList(flag)

	for index, val := range ctns {
		if ctnID == val.ID {
			return index
		}
	}
	return -1
}

//获取容器列表
func CtnList(flag int) ([]types.Container, error) {
	var containers []types.Container

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Errorf(err.Error())
		return containers, err
	}

	switch flag {
	case RUN_CTN:
		//获取运行中的容器列表
		containers, err = cli.ContainerList(ctx, types.ContainerListOptions{})
	case ALL_CTN:
		//获取运行和停止的所有容器列表
		containers, err = cli.ContainerList(ctx, types.ContainerListOptions{
			All: true,
		})
	}

	return containers, err
}
