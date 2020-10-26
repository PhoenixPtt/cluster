package ctnA

import (
	"context"
	"ctnCommon/ctn"
	"ctnCommon/pool"
	"ctnCommon/protocol"

	// "fmt"
	"time"

	"github.com/docker/docker/api/types"
)

var (
	freq      int
	cancelInfo context.CancelFunc
)

func init() {
	freq = 1
}

//设置状态采集频率，包括容器状态、容器资源使用状态
//单位：秒
func SetFreq(curfreq int) {
	freq = curfreq
}

//容器状态自动上传
func CtnInfoAll(distAddr string) {
	var ctx context.Context
	ctx,cancelInfo=context.WithCancel(context.Background())

	for{
		timer:=time.NewTimer(time.Second * time.Duration(freq))
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			containers, _ := CtnList(ALL_CTN)

			//更新本地容器池
			for _, val := range containers {
				UpdateCtnInfo(val)
			}

			//向Server端发送容器信息
			var pSaTruck protocol.SA_TRUCK
			pool.AddIndex()
			pSaTruck.Flag = ctn.FLAG_CTN
			pSaTruck.Index = pool.GetIndex()
			pSaTruck.Addr = distAddr
			pSaTruck.CtnList = containers

			//还要判一下是否有容器池中有，但是实时获取不到的的容器，也做一下状态更新，打上时间标签
			for _, ctnName := range GetCtnNames() {
				pCtn := GetCtn(ctnName)
				ctnID := pCtn.ID
				bExisted := false
				for _, container := range containers {
					if container.ID == ctnID {
						bExisted = true
						break
					}
				}

				if !bExisted {
					var container types.Container
					pSaTruck.CtnList = append(pSaTruck.CtnList, container)
				}
			}

			//fmt.Printf("%#v\n", pSaTruck.CtnList)

			GetSendChan() <- &pSaTruck
		}
	}
}

//取消上传容器状态信息
func CancelCtnInfoAll() {
	cancelInfo()
}
