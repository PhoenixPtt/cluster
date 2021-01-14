package ctnA

import (
	"context"
	"ctnCommon/ctn"
	"ctnCommon/headers"
	"encoding/json"
	"io"

	"fmt"
	"runtime"
)

func CtnStats(ctx context.Context, ctnId string) {
	stats, err := cli.ContainerStats(ctx, ctnId, true)
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
		case <-ctx.Done():
			stats.Body.Close()
			fmt.Println("Stop CTN Stats",ctnId)
			return
		default:
			count++
			err = decoder.Decode(&ctnStats)
			if err == io.EOF {
				return
			} else if err != nil {
				return
			}

			if count%uint64(G_samplingRate) != 0 {
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

				ctnStatMap[ctnId] = ctnStats

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

				//pSaTruck := &protocol.SA_TRUCK{}
				//pSaTruck.Flag = ctn.FLAG_STATS
				//pool.AddIndex()
				//pSaTruck.Index = pool.GetIndex()
				//pSaTruck.Addr = distAddr
				//pSaTruck.CtnStat = make([]ctn.CTN_STATS,0,1)
				//pSaTruck.CtnStat = append(pSaTruck.CtnStat,ctnStats)
				//
				//GetSendChan() <- pSaTruck
			}
		}
	}
}
