package main

import (
	"clusterAgent/ctn"
	"clusterHeader"
	"fmt"
	"tcpSocket"
	"time"
)

const (
	FLAG_CTRL  = "CTRL"
	FLAG_CTN   = "INFO"
	FLAG_STATS = "STAT"
	FLAG_EVENT = "EVTM"
)

var (
	statsFreq int
	exit_info bool
)

func init() {
	statsFreq = 1
}

//设置状态采集频率，包括容器状态、容器资源使用状态
//单位：秒
func SetStatsFreq(freq int) {
	statsFreq = freq
}

//容器状态自动上传
func CtnStatsAll() {
	exit_info = true
	for exit_info {
		containers, _ := ctn.CtnList(ctn.ALL_CTN)
		for _, val := range containers {
			cctn := &ctn.CTN{
				Container: val,
				OperFlag:  FLAG_CTN,
			}

			fmt.Println("容器自身状态", cctn.ID, cctn.State, time.Now().Format("2006-01-02 15:04:05.000"))

			//上传容器自身运行状态
			byteStream, err := header.Encode(&cctn)
			if err != nil {
				cctn.Err = header.Str2bytes(err.Error())
			}

			//发送容器资源状态数据
			tcpSocket.WriteData("", 1, 0, FLAG_CTN, byteStream)
			time.Sleep(time.Second * time.Duration(statsFreq))
		}
	}
}

//取消上传容器状态信息
func CancelCtnStatsAll() {
	exit_info = false
}
