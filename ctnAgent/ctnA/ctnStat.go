package ctnA

import (
	"context"
	"ctnCommon/ctn"
	"ctnCommon/pool"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"time"
	"unsafe"

	"github.com/docker/docker/client"
)

//阻塞方式获取容器的资源使用状态

var (
	exit            bool
	exitFlagChanMap map[string](chan int) //结束监听操作通道
	cli *client.Client
)

func init() {
	exitFlagChanMap = make(map[string](chan int), 100)

	var err error
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Errorf("%s", err.Error())
	}
}

//获取所有容器的资源使用状态，并发送
func CtnStatsAll(distAddr string) {
	exit = true
	for exit {
		//获取所有运行中的容器
		ctns, err := CtnList(RUN_CTN)
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
				go CtnStats(ctn.ID,distAddr)
			}
		}

		for ctnId, _ := range exitFlagChanMap {
			if CtnIndex(ctnId, RUN_CTN) == -1 {
				CancelCtnStats(ctnId)
			}
		}
		time.Sleep(time.Second)
	}
}

func CancelCtnStatsAll() {
	exit = false
	for ctnId, _ := range exitFlagChanMap {
		CancelCtnStats(ctnId)
	}
}

//放入数据
func CtnStats(ctnId string, distAddr string) {
	//ctx_Stats context.Context
	//cancel_Stats context.CancelFunc
	ctx_Stats, cancel_Stats := context.WithCancel(context.Background())
	stats, err := cli.ContainerStats(ctx_Stats, ctnId, true)
	if err != nil {
		fmt.Errorf("%s", err.Error())
	}

	var count int //采集频率计数
	count = 0

	decoder := json.NewDecoder(stats.Body)
	var ctnStats ctn.CTN_STATS
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
			fmt.Println("Stop CTN Stats",ctnId)
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

			if count%freq != 0 {
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

				pSaTruck := &ctn.SA_TRUCK{}
				pSaTruck.Flag = ctn.FLAG_STATS
				pool.AddIndex()
				pSaTruck.Index = pool.GetIndex()
				pSaTruck.Addr = distAddr
				pSaTruck.CtnStat = make([]ctn.CTN_STATS,0,1)
				pSaTruck.CtnStat = append(pSaTruck.CtnStat,ctnStats)

				pObj := (*interface{})(unsafe.Pointer(&pSaTruck))
				GetSendChan() <- pObj
			}
		}
	}
}

func CancelCtnStats(ctnId string) {
	_, ok := exitFlagChanMap[ctnId]
	if ok {
		fmt.Println("stop ctn stats all", ctnId)
		exitFlagChanMap[ctnId] <- 1
	}
}
