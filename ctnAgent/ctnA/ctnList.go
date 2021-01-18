package ctnA

import (
	"sync"
)

//const (
//	RUN_CTN = iota
//	ALL_CTN
//)

var (
	mutex_ls sync.Mutex
)

////获取容器列表
//func CtnList(cli *client.Client, ctx context.Context, flag int) (containers []types.Container, err error) {
//	mutex_ls.Lock()
//	defer mutex_ls.Unlock()
//
//	switch flag {
//	case RUN_CTN:
//		//获取运行中的容器列表
//		containers, err = cli.ContainerList(ctx, types.ContainerListOptions{})
//	case ALL_CTN:
//		//获取运行和停止的所有容器列表
//		containers, err = cli.ContainerList(ctx, types.ContainerListOptions{
//			All: true,
//		})
//	}
//
//	return
//}

////获取指定容器ID的容器信息
//func GetCtnInfo(ctnId string) (types.Container,error)  {
//	var(
//		container types.Container
//		containers []types.Container
//		ctnListOption types.ContainerListOptions
//		err error
//		ctx context.Context
//	)
//
//	ctnListOption = types.ContainerListOptions{
//		All: true,
//	}
//	if containers,err = cli.ContainerList(ctx, ctnListOption); err!=nil{
//		return container,err
//	}
//
//	//遍历所有容器找到目标容器
//	for _,val:=range containers{
//		if val.ID==ctnId{
//			container=val
//		}
//	}
//
//	if container.ID==""{
//		err=errors.New(fmt.Sprintf("容器ID：%s的容器不存在，无法获取容器信息",ctnId))
//		return container, err
//	}
//
//	return container, err
//}

////获取容器在容器列表中的序号
//func CtnIndex(ctnID string, flag int) int {
//	ctns, _ := CtnList(flag)
//
//	mutex_ls.Lock()
//	defer mutex_ls.Unlock()
//	for index, val := range ctns {
//		if ctnID == val.ID {
//			return index
//		}
//	}
//	return -1
//}

