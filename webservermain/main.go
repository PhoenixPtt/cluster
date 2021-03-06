package main

import (
	"clusterServer/clusterServer"
	"fmt"
	"time"
	"webserver/server"
	"webservermain/SysResMonitor"
)

func main() {
	// 输出集群服务端启动信息，并启动集群服务端
	fmt.Println("集群服务端 启动")
	if err := clusterServer.Init(); err != nil {
		fmt.Println("cluster server init failure, error is : ", err)
		return
	}

	// 初始化集群的web服务端
	fmt.Println("集群web服务端 启动")
	if err := server.Init(); err != nil {
		fmt.Println("cluster web server init failure, error is : ", err)
		return
	}

	// 监测主进程的CPU内存使用率
	go func() {
		// 获取系统的总内存量，单位MiB
		totalMem := SysResMonitor.GetMemStatus().Total.Val/102400/1024

		// 间隔固定时间进行CPU和内存的监控，并显示监控值
		for {
			curCpuUsed := SysResMonitor.GetCurProcessCpuUsedPercent()
			curMemUsed := SysResMonitor.GetCurProcessMemUsedPercent()
			fmt.Printf("当前进程资源占用率： cpu=%.2f%%\tmem=%.2f%%(%.2fMiB)\n", curCpuUsed, curMemUsed,
				totalMem*curMemUsed)

			time.Sleep(30e9)
		}
	}()

	// 等待系统结束命令
	server.WaitForInterruptSignal()

	// 停止集群的web服务端
	fmt.Println("集群web服务端 正在停止...")
	server.Stop()

	// 停止集群服务端
	fmt.Println("集群服务端 正在停止...")
	clusterServer.Stop()
}
