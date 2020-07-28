package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/client"
	"io"
	"runtime"
	"time"

	"clusterAgent/ctn"
	"clusterHeader"
)

//阻塞方式获取容器的资源使用状态

var (
	exit            bool
	exitFlagChanMap map[string]chan int //结束监听操作通道
	ctx_Stats 		context.Context
	cancel_Stats 	context.CancelFunc
	cli 			*client.Client
	)

func init() {
	exitFlagChanMap = make(map[string]chan int, 100)
	ctx_Stats, cancel_Stats = context.WithCancel(context.Background())

	var err error
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Errorf("%s", err.Error())
	}
}

//获取所有容器的资源使用状态，并发送
func CtnResStatsAll() {
	exit = true
	for exit {
		//获取所有运行中的容器
		ctns, err := ctn.CtnList(ctn.RUN_CTN)
		if err != nil {
			fmt.Errorf(err.Error())
		}

		//遍历容器列表，启动容器资源监控
		for _, ctn := range ctns {
			//初始化结束监听标志通道
			exitFlagChan, ok := exitFlagChanMap[ctn.ID]
			if !ok {
				exitFlagChanMap[ctn.ID] = exitFlagChan
				//为其分配存储空间
				exitFlagChanMap[ctn.ID] = make(chan int)
				fmt.Println("启动容器监控", ctn.ID)
				go CtnResStats(ctn.ID)
			}
		}

		for ctnId, _ := range exitFlagChanMap {
			if ctn.CtnIndex(ctnId, ctn.RUN_CTN) == -1 {
				CancelCtnResStats(ctnId)
			}
		}
		time.Sleep(time.Second)
	}
}

func CancelCtnResStatsAll() {
	exit = false
	for ctnId, _ := range exitFlagChanMap {
		CancelCtnResStats(ctnId)
	}
}

//放入数据
func CtnResStats(ctnId string) {
	stats, err := cli.ContainerStats(ctx_Stats, ctnId, true)
	if err != nil {
		fmt.Errorf("%s", err.Error())
	}

	var count int //采集频率计数
	count = 0

	decoder := json.NewDecoder(stats.Body)
	var ctnStats header.CTN_STATS
	cpuNum := runtime.NumCPU()
	ctnStats.PercpuUsageCalc = make([]float64, cpuNum)
	for i := 0; i < cpuNum; i++ {
		ctnStats.CpuStats.CPUUsage.PercpuUsage = make([]float64, cpuNum)
		ctnStats.PrecpuStats.CPUUsage.PercpuUsage = make([]float64, cpuNum)
	}
	for {
		select {
		case <-ctx_Stats.Done():
			stats.Body.Close()
			close(exitFlagChanMap[ctnId])
			delete(exitFlagChanMap, ctnId)
			fmt.Println("Stop CTN Stats")
			return
		case <-exitFlagChanMap[ctnId]:
			cancel_Stats()
		default:
			count++
			err := decoder.Decode(&ctnStats)
			if err == io.EOF {
				return
			} else if err != nil {
				cancel_Stats()
			}

			if count%statsFreq != 0 {
				break
			} else {
				base := 1024.00
				ctnStats.MemoryStats.Limit = ctnStats.MemoryStats.Limit / base / base / base
				ctnStats.MemoryStats.Stats.ActiveAnon = ctnStats.MemoryStats.Stats.ActiveAnon / base

				ctnStats.CPUUsageCalc = (ctnStats.CpuStats.CPUUsage.TotalUsage - ctnStats.PrecpuStats.CPUUsage.TotalUsage) * 100 / (ctnStats.CpuStats.SystemCPUUsage - ctnStats.PrecpuStats.SystemCPUUsage)
				for i := 0; i < cpuNum; i++ {
					ctnStats.PercpuUsageCalc[i] = (ctnStats.CpuStats.CPUUsage.PercpuUsage[i] - ctnStats.PrecpuStats.CPUUsage.PercpuUsage[i]) * 100 / (ctnStats.CpuStats.SystemCPUUsage - ctnStats.PrecpuStats.SystemCPUUsage)
				}

				//直接发给server端
				fmt.Println(ctnStats.ID, ctnStats.Read, ctnStats.CPUUsageCalc, ctnStats.PercpuUsageCalc)
				//fmt.Printf("内存限值：%.2f\n", ctnStats.MemoryStats.Limit)
				//fmt.Printf("内存占有量：%.2f\n", ctnStats.MemoryStats.Stats.ActiveAnon)
				//sum := 0.0
				//for index, val := range ctnStats.PercpuUsageCalc {
				//	fmt.Printf("核序号：%d， 单核CPU占有率：%.2f\n", index, val)
				//	sum += val
				//}
				//fmt.Printf("单核累加：%.2f，总的CPU占有率：%.2f\n", sum, ctnStats.CPUUsageCalc)

				byteStream, err := header.Encode(&ctnStats)
				if err != nil {
					fmt.Println(err.Error())
				}

				//发送容器资源状态数据
				fmt.Println("发送监控数据",  FLAG_STATS)
				writeData("", 1, 0, FLAG_STATS, byteStream)
			}
		}
	}
}

func CancelCtnResStats(ctnId string) {
	_, ok := exitFlagChanMap[ctnId]
	if ok {
		fmt.Println("stop ctn stats all", ctnId)
		exitFlagChanMap[ctnId] <- 1
	}
}

