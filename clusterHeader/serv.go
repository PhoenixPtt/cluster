package header

import (
	"ctnCommon/ctn"
	"github.com/docker/docker/api/types"
)

// 应用服务相关
const (
	FLAG_SERV       = "SVCS"
	FLAG_SERV_CTRL  = "SVCS_CTRL" // 启动服务 停止服务 删除服务 扩缩容 重启服务
 	FLAG_SERV_LIST  = "SVCS_LIST"   // 获取服务列表
	//FLAG_SERV_CREATE  = "SVCS_CRET" // 创建服务
	//FLAG_SERV_START   = "SVCS_STAT" // 启动服务
	//FLAG_SERV_STOP    = "SVCS_STOP" // 停止服务
	//FLAG_SERV_REMOVE  = "SVCS_REMV" // 删除服务
	//FLAG_SERV_SCALE   = "SVCS_SCAL" // 扩缩容
	//FLAG_SERV_RESTART = "SVCS_REST" // 重启服务
	FLAG_SERV_STATS = "SVCS_STAT"
	FLAG_SERV_INFO  = "SVCS_INFO"
)

// 容器操作
const (
	FLAG_CTNS   	 = "CTNS"
	FLAG_CTNS_CTRL   = "CTNS_CTRL"  // 启动容器 停止容器 强制停止容器 删除容器 获取容器日志 获取容器详细信息
	FLAG_CTNS_CRET   = "CTNS_CRET"  // 创建容器
	//FLAG_CTNS_START   = "CTNS_STAT" // 启动容器
	//FLAG_CTNS_STOP   = "CTNS_STOP"  // 停止容器
	//FLAG_CTNS_FCST   = "CTNS_FCST"  // 强制停止容器
	//FLAG_CTNS_REMV   = "CTNS_REMV"  // 删除容器
	//FLAG_CTNS_LOG    = "CTNS_LOG"   // 获取容器日志
	FLAG_CTNS_STATS  = "CTNS_STAT" // 获取容器状态
	//FLAG_CTNS_INFO   = "CTNS_INFO"  // 获取容器详细信息
)

// 服务结构体
type SERVICE struct {
	Oper
	Services
}

type Services struct {
	Count   uint32    // 应用服务数量
	Service []Service // 所有的应用服务
}

type Service struct {
	ServiceInfo
	Cfg     ServiceCfg     // 服务配置信息
	//Res     ResourceStatus // 服务的CPU使用率、内存使用量和内存使用率
	Replica []Replica      // 服务的所有副本
}

type ServiceInfo struct {
	Id           string // 服务Id
	State        string // 服务状态
	Scale        uint32 // 设定的副本数量
	ReplicaCount uint32 // 应用服务的当前副本数量
	CreateTime   string // 服务创建时间
	StartTime    string // 服务启动时间
	NameSpace    string //服务的命名空间
}

// 副本信息
type Replica struct {
	Id         string    // 副本ID
	CreateTime string    // 副本创建时间
	NodeId     string    // 所在节点
	State      uint32    // 副本的状态
	Ctn        types.Container       // 容器的副本信息
	CtnStats   ctn.CTN_STATS // 服务的CPU使用率、内存使用量和内存使用率
}

//// 服务静态配置结构体
//type ServiceCfg struct {
//	Name 			string 				// 服务名称
//	Image       	string 				// 服务的镜像
//	Scale 			uint32				// 服务规模
//	Cmd				string				// cmd入口指令
//	CmdPars			string				// cmd入口指令参数
//	EntryPoint		string				// 入口指令
//	EntryPointPars 	string				// 入口参数
//	Policy			Policy      		// 服务的调度策略
//}

//// 服务调度策略结构体
//type Policy struct {
//	Type        	int          		// 策略类型：  0:指定节点；1：根据资源使用情况动态分配；2：为每个节点分配一个副本
//	AssignNodes 	[]string     		// 策略类型0： 指定的节点列表
//	Algorithm   	int          		// 策略类型1： 算法类型
//	RcWeight 		RcWeight     		// 资源权重
//	RcEvaUsage  	RcUsage      		// 记录服务预估资源使用量
//}
type ServiceCfg struct {
	Version  string       // 服务版本
	ServicePar // 服务配置参数
}

// 服务配置参数
type ServicePar struct {
	Name           string 				// 服务名称
	Image          string 				// 镜像名称
	Cmd            string 				// cmd入口指令
	CmdPars        string 				// cmd入口指令参数
	EntryPoint     string 				// 入口指令
	EntryPointPars string 				// 入口参数
	Deploy Policy
}

// 服务部署策略
type Policy struct {
	Mode  			string 				//global（每个节点部署一个副本） replicated
	Replicas 		uint32 				// 设定的副本数量
	Placement 		Placement			// 服务placement
	Resources  		[]RcUsage			// 预估的服务资源使用量
	RcWeight  		RcWeight			// 资源权重结构体
}

// 服务placement
type Placement struct {
	Constraints  []string   //给服务设置的标签列表
}

// 资源权重结构体
type RcWeight struct {
	NodeWeight     float64 // 硬盘权重
	CpuWeight      float64 // CPU权重
	MemWeight      float64 // 内存权重
	HardDiskWeight float64 // 硬盘权重
}

// 预估的服务资源使用量
type RcUsage struct {
	Name          string  //limits 最高限制 requests 最低需求
	CpuUsage      float64 // CPU使用量
	MemUsage      float64 // 内存使用量
	HardDiskUsage float64 // 硬盘使用量
}
