package ctnA

import (
	"context"
	"ctnCommon/ctn"
	"ctnCommon/headers"
	"ctnCommon/pool"
	"encoding/json"
	"io"

	"fmt"
	"github.com/docker/docker/client"
	"runtime"
	"time"
)

var (
	cli *client.Client
	cancelStatsMap map[string]context.CancelFunc
	cancelStatsAll context.CancelFunc
)

func init() {
	cancelStatsMap = make(map[string]context.CancelFunc)

	var err error
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Errorf("%s", err.Error())
	}
}

//获取所有容器的资源使用状态，并发送
func CtnStatsAll(distAddr string) {
	var ctx context.Context
	ctx,cancelStatsAll=context.WithCancel(context.Background())

	for{
		timer:=time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			ctns, err := CtnList(RUN_CTN)
			if err != nil {
				fmt.Errorf(err.Error())
			}

			//遍历容器列表，启动容器资源监控
			for _, ctn := range ctns {
				_,ok:=cancelStatsMap[ctn.ID]
				if !ok{
					go CtnStats(ctn.ID,distAddr)
				}
			}

			for ctnId,_:=range cancelStatsMap{
				if CtnIndex(ctnId, RUN_CTN) == -1 {
					CancelCtnStats(ctnId)
				}
			}
		}
	}
}

func CancelCtnStatsAll() {
	cancelStatsAll()
	for ctnId, _:= range cancelStatsMap{
		CancelCtnStats(ctnId)
	}
}

//放入数据
func CtnStats(ctnId string, distAddr string) {
	ctx_Stats, cancel_Stats := context.WithCancel(context.Background())
	_,ok:=cancelStatsMap[ctnId]
	if !ok{
		cancelStatsMap[ctnId]=cancel_Stats
	}else{
		return
	}

	stats, err := cli.ContainerStats(ctx_Stats, ctnId, true)
	if err != nil {
		fmt.Errorf("%s", err.Error())
	}

	var count uint64 //采集频率计数
	count = 0
	for {
		decoder := json.NewDecoder(stats.Body)
		var ctnStats ctn.CTN_STATS
		cpuNum := runtime.NumCPU()
		ctnStats.PercpuUsageCalc = make([]float64, cpuNum)
		for i := 0; i < cpuNum; i++ {
			ctnStats.CpuStats.CPUUsage.PercpuUsage = make([]float64, cpuNum)
			ctnStats.PrecpuStats.CPUUsage.PercpuUsage = make([]float64, cpuNum)
		}
		select {
		case <-ctx_Stats.Done():
			stats.Body.Close()
			fmt.Println("Stop CTN Stats",ctnId)
			return
		default:
			count++
			err = decoder.Decode(&ctnStats)
			if err == io.EOF {
				return
			} else if err != nil {
				cancel_Stats()
			}

			if count%uint64(freq) != 0 {
				break
			} else {
				base := 1024.00
				ctnStats.MemoryStats.Limit = ctnStats.MemoryStats.Limit / base / base / base
				ctnStats.MemoryStats.Stats.ActiveAnon = ctnStats.MemoryStats.Stats.ActiveAnon / base

				ctnStats.CPUUsageCalc = (ctnStats.CpuStats.CPUUsage.TotalUsage - ctnStats.PrecpuStats.CPUUsage.TotalUsage) * 100 / (ctnStats.CpuStats.SystemCPUUsage - ctnStats.PrecpuStats.SystemCPUUsage)
				for i := 0; i < cpuNum; i++ {
					ctnStats.PercpuUsageCalc[i] = (ctnStats.CpuStats.CPUUsage.PercpuUsage[i] - ctnStats.PrecpuStats.CPUUsage.PercpuUsage[i]) * 100 / (ctnStats.CpuStats.SystemCPUUsage - ctnStats.PrecpuStats.SystemCPUUsage)
				}

				//转时间为当地时间
				ctnStats.Read = headers.ToLocalTime(ctnStats.Read)
				ctnStats.Preread = headers.ToLocalTime(ctnStats.Preread)

				////直接发给server端
				//fmt.Println(ctnStats.ID[:10], ctnStats.Read, ctnStats.Preread, ctnStats.CPUUsageCalc, ctnStats.PercpuUsageCalc)
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

				GetSendChan() <- pSaTruck
			}
		}
	}
}

func CancelCtnStats(ctnId string) {
	_,ok:=cancelStatsMap[ctnId]
	if ok{
		fmt.Println("stop ctn stats all", ctnId)
		cancelStatsMap[ctnId]()
		delete(cancelStatsMap, ctnId)
	}
}
