package main

import (
	"clusterAgent/SysResMonitor"
	header "clusterHeader"
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/host"
	"tcpSocket"
	"time"
)

// public

// 设置刷新间隔时间 单位：秒
func SetRefreshFreq(interval time.Duration) {
	refreshInterval = interval
}
// public
// 启动监控
func StartMonitor() {
	if isRunning {
		return
	}

	isRunning = true

	// 在协程中持续监控资源
	go run()
	//go CtnResStatsAll()
}

func StopMonitor() {
	isRunning = false
}

// private
var isRunning bool = false
var refreshInterval time.Duration = time.Second // 刷新时间间隔，单位s

// 获取主机信息，返回值为NODE - json
func GetHostInfo() {
	hInfo, _ := host.Info()

	var nd header.NODE
	nd.Oper.Type = header.FLAG_NODE_HOST
	nd.Nodes.Node = make([]header.Node,1)
	node := &nd.Nodes.Node[0]
	node.NodeInfo.HostInfo = *hInfo

	// 从网络发走数据
	data, _ := json.Marshal(nd)
	writeData("", tcpSocket.TCP_TYPE_MONITOR, 0, header.FLAG_NODE, data)
}



func run() {
	// 循环获取资源
	for isRunning {
		// 睡眠 RefreshInterval
		time.Sleep(refreshInterval)

		var nd header.NODE
		nd.Oper.Type = header.FLAG_NODE_RDAT
		nd.Nodes.Node = make([]header.Node,1)
		node := &nd.Nodes.Node[0]
		node.State = true
		node.Res.Time = time.Now().String()
		// 获取系统资源
		// 获取CPU使用率
		node.Res.Cpu = SysResMonitor.GetCpuStatus()

		// 获取内存状态
		node.Res.Mem = SysResMonitor.GetMemStatus()

		// 获取硬盘状态，仅获取根目录所在分区的状态
		node.Res.Disk = SysResMonitor.GetDiskStatus()

		// 从网络发走数据
		data, _ := json.Marshal(nd)
		writeData("", tcpSocket.TCP_TYPE_MONITOR, 0, header.FLAG_NODE, data)

		curCpuUsed := SysResMonitor.GetCurProcessCpuUsedPercent()
		curMemUsed := SysResMonitor.GetCurProcessMemUsedPercent()
		fmt.Printf("当前进程资源占用率： cpu=%.2f%%\tmem=%.2f%%\n", curCpuUsed, curMemUsed)

	}
}

// 获取所有容器的状态
//for i,h := range servers.GetOnlineServerHandles() {
//	containerCount := len(server.Ctn)
//	if containerCount > 0 {
//		Resource.Ctn = make([]ResStatus.ConResStatus, containerCount )
//
//		for ctnId, _ := range server.Ctn {
//			//初始化结束监听标志通道
//			exitFlagChan, ok := exitFlagChanMap[ctnId]
//			if !ok {
//				exitFlagChanMap[ctnId] = exitFlagChan
//				//为其分配存储空间
//				exitFlagChanMap[ctnId] = make(chan int)
//				fmt.Println("启动容器监控", ctnId)
//				go CtnResStats(ctnId)
//			}
//		}
//	} else { // 如果当前没有在运行的容器，则把数据缓存清空
//		if len(Resource.Ctn) != 0 {
//			Resource.Ctn = []ResStatus.ConResStatus{}
//		}
//	}
//}