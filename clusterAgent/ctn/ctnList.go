package ctn

import (
	//"time"

	"github.com/docker/docker/api/types"
	"sync"
)

const (
	RUN_CTN = iota
	ALL_CTN
)

var (
	mutex_ls sync.Mutex
)

//获取容器在容器列表中的序号
func CtnIndex(ctnID string, flag int) int {
	ctns, _ := CtnList(flag)

	mutex_ls.Lock()
	defer mutex_ls.Unlock()
	for index, val := range ctns {
		if ctnID == val.ID {
			return index
		}
	}
	return -1
}

//获取容器列表
func CtnList(flag int) ([]types.Container, error) {
	mutex_ls.Lock()
	defer mutex_ls.Unlock()
	var containers []types.Container

	switch flag {
	case RUN_CTN:
		//获取运行中的容器列表
		containers, err = Cli.ContainerList(Ctx, types.ContainerListOptions{})
	case ALL_CTN:
		//获取运行和停止的所有容器列表
		containers, err = Cli.ContainerList(Ctx, types.ContainerListOptions{
			All: true,
		})
	}

	return containers, err
}
