package ctn

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"

	// "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"

	"encoding/base64"
	"encoding/json"

	"clusterHeader"
)

const (
	DEFAULT_MAP_SIZE = 1000
)

const (
	CREATE  = "CREATE"
	START   = "START"
	STOP    = "STOP"
	KILL    = "KILL"
	REMOVE  = "REMOVE"
	GETLOG  = "GETLOG"
	INSPECT = "INSPECT"
	CTNEXIT = "EXIT"
)

const (
	iCREATE = iota
	iSTART
	iSTOP
	iKILL
	iREMOVE
	iGETLOG
	iINSPECT
	iSTATS
)

//容器结构体
type CTN struct {
	types.Container
	Logs        []byte      `json:"logs"`     //容器日志对应getlog
	CtnInspect  header.CTN_INSPECT `json:"ctn_json"` //容器信息对应inspect
	Err         []byte      `json:"err"`
	OperFlag    string      `json:"oper_flag"`
	OperIndex   int         `json:"oper_index"`
	AgentAddr   string      `json:"agentaddr"`
	ServiceName string      `json:"servicename"`
}

var (
	Ctx context.Context
	Cli *client.Client

	clis    []*client.Client
	err     error
	g_index int
)

func init() {
	g_index = 0
	Ctx = context.Background()

	Cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println(err.Error())
	}
}

//创建容器
func (ctn *CTN) Create() error {
	imageSummery, err := Cli.ImageList(Ctx, types.ImageListOptions{})
	if err != nil {
		return err
	}

	fmt.Println(ctn.Image)
	index_repo := -1
	index_image := -1
	fmt.Println("镜像仓库列表如下所示：")
	for i, repo := range imageSummery {
		fmt.Printf("镜像仓库序号：%d\t镜像仓库：%#v\n", i, repo)
		for j, v := range repo.RepoTags {
			fmt.Printf("\t镜像序号：%d\t镜像名称：%s\n", j, v)
			if v == ctn.Image { // 假设需要获取os.Args[k], k = 1
				index_image = j
				break
			}
		}
		if index_image != -1 {
			index_repo = i
			break
		}
	}

	if index_repo == -1 && index_image == -1 {
		fmt.Println("本地仓库中不存在镜像imageTag")
		//本地仓库不存在，从私有仓库下载
		auth, _ := registryAuth(true, "docker", "27MTjlJyZWD0XxLf7C_SxOLlYpaprdzURn-Ec10Ew-U")
		var options types.ImagePullOptions
		options.RegistryAuth = auth
		_, err := Cli.ImagePull(Ctx, ctn.Image, options)
		if err != nil {
			return err
		}
		fmt.Println("从私有仓库中Pull镜像成功")
	} else {
		fmt.Println("镜像在本地仓库已存在！")
		fmt.Printf("序号：[%d,%d]\n", index_repo, index_image)
	}

	resp, err := Cli.ContainerCreate(Ctx,
		&container.Config{
			Image: ctn.Image,
		},
		&container.HostConfig{
			NetworkMode: "host",
		},
		nil,
		"")
	if err != nil {
		return err
	}

	ctn.ID = resp.ID

	return err
}

//启动容器
func (ctn *CTN) Start() error {
	if err = Cli.ContainerStart(Ctx, ctn.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	return err
}

//停止容器
func (ctn *CTN) Stop() error {
	// cli = clis[iSTOP]
	//正常停止容器
	var timeout *time.Duration
	err = Cli.ContainerStop(Ctx, ctn.ID, timeout)
	if err != nil {
		//正常停止容器失败，强制停止容器
		err = ctn.Kill()
		if err != nil {
			return err
		}
		return err
	}
	fmt.Printf("container: %s\t\t%s\n", ctn.ID[:10], "normal stopped")
	return err
}

//强制停止容器
func (ctn *CTN) Kill() error {
	// cli = clis[iKILL]
	//正常停止容器
	err = Cli.ContainerKill(Ctx, ctn.ID, "")
	if err != nil {
		return err
	}
	fmt.Printf("container:%s\t\t%s\n", ctn.ID[:10], "force stopped\n")
	return err
}

//删除容器
func (ctn *CTN) Remove() error {
	// cli = clis[iREMOVE]
	err := ctn.Stop()
	if err != nil {
		return err
	}

	err = Cli.ContainerRemove(Ctx, ctn.ID, types.ContainerRemoveOptions{})
	if err != nil {
		return err
	}
	fmt.Print("container:", ctn.ID, "\tnormal remove\n")

	return err
}

//获取容器日志
//注意：容器被删除之后无法获取容器日志
func (ctn *CTN) GetLog() (io.ReadCloser, error) {
	// cli = clis[iGETLOG]
	var logs io.ReadCloser

	logs, err = Cli.ContainerLogs(Ctx, ctn.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return logs, err
	}

	return logs, err
}

//查看容器详细信息
func (ctn *CTN) Inspect() (header.CTN_INSPECT, error) {
	var ctnInspect header.CTN_INSPECT
	var ctnJson types.ContainerJSON

	ctnJson, err = Cli.ContainerInspect(Ctx, ctn.ID)
	if err != nil {
		return ctnInspect, err
	}

	var inspectStream []byte
	inspectStream, err = json.Marshal(ctnJson)
	if err != nil {
		return ctnInspect, err
	}

	err := json.Unmarshal(inspectStream, &ctnInspect)
	if err != nil {
		return ctnInspect, err
	}

	return ctnInspect, err
}

//单次获取容器资源使用状态
//func (ctn *CTN) StatsOneShot() error {
//	// cli = clis[iSTATS]
//	stats, err := Cli.ContainerStatsOneShot(Ctx, ctn.ID)
//	if err != nil {
//		fmt.Errorf("%s", err.Error())
//	}
//
//	//切片需分配内存空间
//	decoder := json.NewDecoder(stats.Body)
//	var ctnStats header.CTN_STATS
//	cpuNum := runtime.NumCPU()
//	ctnStats.PercpuUsageCalc = make([]float64, cpuNum)
//	for i := 0; i < cpuNum; i++ {
//		ctnStats.CpuStats.CPUUsage.PercpuUsage = make([]float64, cpuNum)
//		ctnStats.PrecpuStats.CPUUsage.PercpuUsage = make([]float64, cpuNum)
//	}
//
//	err = decoder.Decode(&ctnStats)
//	if err != nil {
//		return err
//	}
//
//	//做单位换算
//	base := 1024.00
//	ctnStats.MemoryStats.Limit = ctnStats.MemoryStats.Limit / base / base / base         //单位是GiB
//	ctnStats.MemoryStats.Stats.ActiveAnon = ctnStats.MemoryStats.Stats.ActiveAnon / base //单位是MiB
//
//	//计算CPU使用率
//	ctnStats.CPUUsageCalc = (ctnStats.CpuStats.CPUUsage.TotalUsage - ctnStats.PrecpuStats.CPUUsage.TotalUsage) * 100 / (ctnStats.CpuStats.SystemCPUUsage - ctnStats.PrecpuStats.SystemCPUUsage)
//	for i := 0; i < cpuNum; i++ {
//		ctnStats.PercpuUsageCalc[i] = (ctnStats.CpuStats.CPUUsage.PercpuUsage[i] - ctnStats.PrecpuStats.CPUUsage.PercpuUsage[i]) * 100 / (ctnStats.CpuStats.SystemCPUUsage - ctnStats.PrecpuStats.SystemCPUUsage)
//	}
//
//	fmt.Println("showCtnStats", ctnStats.ID, ctnStats.Read)
//	//fmt.Printf("内存限值：%.2f\n", ctnStats.MemoryStats.Limit)
//	//fmt.Printf("内存占有量：%.2f\n", ctnStats.MemoryStats.Stats.ActiveAnon)
//	//sum := 0.0
//	//for index, val := range ctnStats.PercpuUsageCalc {
//	//	fmt.Printf("核序号：%d， 单核CPU占有率：%.2f\n", index, val)
//	//	sum += val
//	//}
//	//fmt.Printf("单核累加：%.2f，总的CPU占有率：%.2f\n", sum, ctnStats.CPUUsageCalc)
//	return err
//}

//单次获取容器资源使用状态

func registryAuth(isRegisAuth bool, username string, password string) (string, bool) {
	//认证
	var authStr string
	if isRegisAuth {
		authConfig := types.AuthConfig{
			Username: username,
			Password: password,
		}
		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			panic(err)
			return authStr, false
		}
		authStr = base64.URLEncoding.EncodeToString(encodedJSON)
	}
	return authStr, true

}

////////////////////////////////以下是服务端容器操作////////////////////////////////////////////
//准备数据
func (sctn *CTN) PrepareData(operFlag string) int {
	g_index++
	sctn.OperFlag = operFlag
	sctn.OperIndex = g_index
	sctn.Err = header.Str2bytes("")
	return g_index
}
