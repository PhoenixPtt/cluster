package ctnA

import (
	"bytes"
	"context"
	"ctnCommon/ctn"
	"ctnCommon/headers"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"runtime"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

//创建容器
func Create(ctx context.Context, ctnName string, imgName string) (err error) {
	var (
		obj  interface{}
		resp container.ContainerCreateCreatedBody
		pCtn *ctn.CTN
	)

	//判断容器名称是否存在
	if obj = G_ctnMgr.ctnObjPool.GetObj(ctnName); obj != nil {
		//容器名称已存在，禁止重复创建
	}

	//判断镜像是否存在
	if !IsImageExisted(ctx, imgName) {
		//如果镜像不存在，则从私有仓库中拉取
		if err = ImagePull(ctx, imgName); err != nil {
			//如果拉取失败，则返回
			return
		}
	}

	//创建容器
	resp, err = cli.ContainerCreate(ctx,
		&container.Config{
			Image: imgName,
		},
		&container.HostConfig{
			NetworkMode: "host",
		},
		nil,
		nil,
		"")
	if err != nil {
		return
	}

	//新建容器对象
	pCtn = &ctn.CTN{
		CtnName:   ctnName,
		ImageName: imgName,
		CtnID:     resp.ID,
	}

	//将容器对象添加到容器池中
	G_ctnMgr.ctnObjPool.AddObj(ctnName, pCtn)
	return
}

//启动容器
func Start(ctx context.Context, ctnName string) (err error) {
	var (
		obj  interface{}
		pCtn *ctn.CTN
	)

	//判断该容器是否存在
	if obj = G_ctnMgr.ctnObjPool.GetObj(ctnName); obj == nil {
		err = errors.New(fmt.Sprintf("容器：%s不存在", ctnName))
		return
	}
	pCtn = obj.(*ctn.CTN) //接口强制类型转换为容器对象类型

	//判断容器当前运行状态,如果容器正在运行，则直接返回
	if pCtn.State == "running" {
		return
	}

	//启动容器
	if err = cli.ContainerStart(ctx, pCtn.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	return
}

//运行容器
func Run(ctx context.Context, ctnName string, imgName string) (err error) {
	var (
		obj interface{}
	)

	//判断该容器是否存在
	if obj = G_ctnMgr.ctnObjPool.GetObj(ctnName); obj == nil {
		//容器不存在，则创建容器
		if err = Create(ctx, ctnName, imgName); err != nil {
			return
		}
	}

	//启动容器
	if err = Start(ctx, ctnName); err != nil {
		goto ERROR
	}

	return

ERROR:
	//运行失败，则将容器删除
	err = Remove(ctx, ctnName)
	return
}

//停止容器
func Stop(ctx context.Context, ctnName string) (err error) {
	var (
		obj  interface{}
		pCtn *ctn.CTN
	)

	//判断该容器是否存在
	if obj = G_ctnMgr.ctnObjPool.GetObj(ctnName); obj == nil {
		err = errors.New(fmt.Sprintf("容器：%s不存在", ctnName))
		return
	}
	pCtn = obj.(*ctn.CTN) //接口强制类型转换为容器对象类型

	//判断容器当前运行状态,如果容器不在运行，则直接返回
	if pCtn.State != "running" {
		return
	}

	//正常停止容器
	var timeout *time.Duration
	err = cli.ContainerStop(ctx, pCtn.ID, timeout)

	return
}

//强制停止容器
func Kill(ctx context.Context, ctnName string) (err error) {
	var (
		obj  interface{}
		pCtn *ctn.CTN
	)

	//判断该容器是否存在
	if obj = G_ctnMgr.ctnObjPool.GetObj(ctnName); obj == nil {
		err = errors.New(fmt.Sprintf("容器：%s不存在", ctnName))
		return
	}
	pCtn = obj.(*ctn.CTN) //接口强制类型转换为容器对象类型

	//判断容器当前运行状态,如果容器不在运行，则直接返回
	if pCtn.State != "running" {
		return
	}

	//正常停止容器
	err = cli.ContainerKill(ctx, pCtn.ID, "")

	return
}

//删除容器
func Remove(ctx context.Context, ctnName string) (err error) {
	var (
		obj  interface{}
		pCtn *ctn.CTN
	)

	//判断该容器是否存在
	if obj = G_ctnMgr.ctnObjPool.GetObj(ctnName); obj == nil {
		err = errors.New(fmt.Sprintf("容器：%s不存在", ctnName))
		return
	}
	pCtn = obj.(*ctn.CTN) //接口强制类型转换为容器对象类型

	//判断容器当前运行状态,如果容器正在运行，则kill容器
	if pCtn.State == "running" {
		if err = Kill(ctx, ctnName); err != nil {
			return
		}
	}

	if err = cli.ContainerRemove(ctx, pCtn.ID, types.ContainerRemoveOptions{}); err == nil {
		//容器删除成功，则删除容器池中对应的容器对象
		G_ctnMgr.ctnObjPool.RemoveObj(ctnName)
	}

	return
}

//获取容器日志
//注意：容器停止运行后无法获取容器日志
func GetLog(ctx context.Context, ctnName string) (logStr string, err error) {
	var (
		obj  interface{}
		pCtn *ctn.CTN
		logs io.ReadCloser
	)

	//判断该容器是否存在
	if obj = G_ctnMgr.ctnObjPool.GetObj(ctnName); obj == nil {
		err = errors.New(fmt.Sprintf("容器：%s不存在", ctnName))
		return
	}
	pCtn = obj.(*ctn.CTN) //接口强制类型转换为容器对象类型

	//判断容器当前运行状态,只有运行的容器才能获取到日志
	if pCtn.State == "running" {
		if logs, err = cli.ContainerLogs(ctx, pCtn.ID, types.ContainerLogsOptions{ShowStdout: true}); err != nil {
			return
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(logs)
		logStr = buf.String()
	}

	return
}

//查看容器详细信息
func Inspect(ctx context.Context, ctnName string) (ctnInspect ctn.CTN_INSPECT, err error) {
	var (
		obj           interface{}
		pCtn          *ctn.CTN
		ctnJson       types.ContainerJSON
		inspectStream []byte
	)

	//判断该容器是否存在
	if obj = G_ctnMgr.ctnObjPool.GetObj(ctnName); obj == nil {
		err = errors.New(fmt.Sprintf("容器：%s不存在", ctnName))
		return
	}
	pCtn = obj.(*ctn.CTN) //接口强制类型转换为容器对象类型

	//获取容器详细信息
	if ctnJson, err = cli.ContainerInspect(ctx, pCtn.ID); err != nil {
		return
	}

	//json序列化
	if inspectStream, err = json.Marshal(ctnJson); err != nil {
		return
	}

	//json反序列化
	err = json.Unmarshal(inspectStream, &ctnInspect)

	return
}

func CtnStats(ctx context.Context, ctnName string) {
	var (
		obj      interface{}
		err      error
		pCtn     *ctn.CTN
		ctnId    string
		stats    types.ContainerStats
		count    uint64 //采集频率计数
		decoder  *json.Decoder
		ctnStats ctn.CTN_STATS
		cpuNum   int
		base     float64
	)

	//初始化变量
	count = 0
	cpuNum = runtime.NumCPU()
	base = 1024.00

	//判断该容器是否存在
	if obj = G_ctnMgr.ctnObjPool.GetObj(ctnName); obj == nil {
		err = errors.New(fmt.Sprintf("容器：%s不存在", ctnName))
		return
	}
	pCtn = obj.(*ctn.CTN) //接口强制类型转换为容器对象类型
	ctnId = pCtn.CtnID

	//容器资源监控
	stats, err = cli.ContainerStats(ctx, ctnId, true)
	if err != nil {
		fmt.Errorf("%s", err.Error())
	}

	for {
		decoder = json.NewDecoder(stats.Body)
		ctnStats.PercpuUsageCalc = make([]float64, cpuNum)
		for i := 0; i < cpuNum; i++ {
			ctnStats.CpuStats.CPUUsage.PercpuUsage = make([]float64, cpuNum)
			ctnStats.PrecpuStats.CPUUsage.PercpuUsage = make([]float64, cpuNum)
		}
		select {
		case <-ctx.Done():
			stats.Body.Close()
			fmt.Println("Stop CTN Stats", ctnId)
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

				//直接发给server端
				//fmt.Println(ctnStats.ID[:10], ctnStats.Read, ctnStats.Preread, ctnStats.CPUUsageCalc, ctnStats.PercpuUsageCalc)
				//fmt.Printf("内存限值：%.2f\n", ctnStats.MemoryStats.Limit)
				//fmt.Printf("内存占有量：%.2f\n", ctnStats.MemoryStats.Stats.ActiveAnon)
				//sum := 0.0
				//for index, val := range ctnStats.PercpuUsageCalc {
				//	fmt.Printf("核序号：%d， 单核CPU占有率：%.2f\n", index, val)
				//	sum += val
				//}
				//fmt.Printf("单核累加：%.2f，总的CPU占有率：%.2f\n", sum, ctnStats.CPUUsageCalc)
			}
		}
	}
}
