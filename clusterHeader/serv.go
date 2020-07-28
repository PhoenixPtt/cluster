package header

// 应用服务相关

const (
	FLAG_SERV   		= "SVCS"
	FLAG_SERV_CTRL   	= "SVCS_CTRL"
	FLAG_SERV_STATS   	= "SVCS_STATS"
	FLAG_SERV_INFO   	= "SVCS_INFO"
)

// 服务结构体
type SERVICE struct {
	Oper
	Services
}

type Services struct {
	Count			uint32 				// 应用服务数量
	Service			[]Service			// 所有的应用服务
}

type Service struct {
	ServiceInfo
	Cfg 			ServiceCfg      	// 服务配置信息
	Res     		ResourceStatus  	// 服务的CPU使用率、内存使用量和内存使用率
	NodeIds			[]string   			// 所在的节点
	Replica			[]Replica    		// 服务的所有副本
}

type ServiceInfo struct {
	Id				string				// 服务Id
	State			uint32				// 服务状态
	Scale 			uint32				// 设定的副本数量
	ReplicaCount	uint32				// 应用服务的当前副本数量
	CreateTime   	string      		// 服务创建时间
	StartTime    	string      		// 服务启动时间
}

// 副本信息
type Replica struct {
	Id				string				// 副本ID
	CreateTime		string				// 副本创建时间
	NodeId			string				// 所在节点
	State			uint32				// 副本的状态
	Ctn				CTN					// 容器的副本信息
	CtnStats    	CTN_STATS 			// 服务的CPU使用率、内存使用量和内存使用率
}

// 服务静态配置结构体
type ServiceCfg struct {
	Name 			string 				// 服务名称
	Image       	string 				// 服务的镜像
	Scale 			uint32				// 服务规模
	Cmd				string				// cmd入口指令
	CmdPars			string				// cmd入口指令参数
	EntryPoint		string				// 入口指令
	EntryPointPars 	string				// 入口参数
	Policy			Policy      		// 服务的调度策略
}

// 服务调度策略结构体
type Policy struct {
	Type        	int          		// 策略类型：  0:指定节点；1：根据资源使用情况动态分配；2：为每个节点分配一个副本
	AssignNodes 	[]string     		// 策略类型0： 指定的节点列表
	Algorithm   	int          		// 策略类型1： 算法类型
	RcWeight 		RcWeight     		// 资源权重
	RcEvaUsage  	RcUsage      		// 记录服务预估资源使用量
}


// 资源权重结构体
type RcWeight struct {
	NodeWeight     	float64 			// 硬盘权重
	CpuWeight      	float64 			// CPU权重
	MemWeight      	float64 			// 内存权重
	HardDiskWeight 	float64 			// 硬盘权重
}

// 预估的服务资源使用量
type RcUsage struct {
	CpuUsage      	float64 			// CPU使用量
	MemUsage      	float64 			// 内存使用量
	HardDiskUsage 	float64 			// 硬盘使用量
}