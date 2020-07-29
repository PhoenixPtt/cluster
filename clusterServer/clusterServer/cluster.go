package clusterServer

import (
	header "clusterHeader"
	"math/rand"
	"time"
)

var nodes Nodes
var warings Warnings	// 所有的报警信息
var clstStats *header.ClstStats = new(header.ClstStats)

func init() {
	nodes.Init()
}
func updateClusterStats() {
	// 更新集群状态
	clstStats.Res = *nodes.ResCount()
	clstStats.NodeCount = uint32(nodes.Count())
	clstStats.ExecServiceCount = 0
	clstStats.RunState = true

	// 报警信息更新
	var wItem header.WarningItem
	iType := uint8(rand.Uint32() % 3)
	switch iType {
	case 0:
		wItem.Type = "CPU"
		wItem.Msg = "温度异常"
	case 1:
		wItem.Type = "内存"
		wItem.Msg = "剩余空间不足"
	default:
		wItem.Type = "硬盘"
		wItem.Msg = "坏道过多"
	}

	wItem.Time = time.Now().Format("2006-01-02 15:04:05.000000000")
	wItem.Level = uint8(rand.Uint32() % uint32(header.WARNING_LEVEL_COUNT))
	warings.Add(wItem)

	//wstr := header.JsonString(*warings.WarningInfo())
	//fmt.Println(wstr)
}


