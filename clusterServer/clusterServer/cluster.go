package clusterServer

import (
	header "clusterHeader"
	"ctnServer/cluster"
	"ctnServer/ctnS"
	"github.com/shirou/gopsutil/host"
	"math/rand"
	"sync"
	"time"
)

var nodes Nodes
var warings Warnings	// 所有的报警信息
var clstStats *header.ClstStats = new(header.ClstStats)
var hostInfo *host.InfoStat // 主机信息

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

const (
	FLAG_CTRL  = "CTRL"
	FLAG_CTN   = "INFO"
	FLAG_STATS = "STAT"
	FLAG_EVENT = "EVTM"

	CLUSTER_NAME = "集群管理平台"
	SERVICE_WATCH = "集群管理平台"+"_"+"服务监视"
	NODE_WATCH = "集群管理平台"+"_"+"节点监视"
)

var (
	exit    bool
	mMutex  sync.Mutex
	g_cluster *cluster.CLUSTER
)

func init() {
	g_cluster = cluster.NewCluster(CLUSTER_NAME)
	ctnS.Config(writeAgentData)
	g_cluster.Start(SERVICE_WATCH,NODE_WATCH)
	//go cluster.MsgEvent()
}


