package header

const (
	FLAG_CLST   		= "CLST"
	FLAG_CLST_CFG 		= "CLST_CFG"
	FLAG_CLST_CTRL 		= "CLST_CTRL"
	FLAG_CLST_NODE 		= "CLST_NODE"
	FLAG_CLST_SERVE 	= "CLST_SERVE"
	FLAG_CLST_STATS 	= "CLST_STATS"
)

// 集群结构体
type CLST struct {
	Oper
	Cluster
}

type Cluster struct {
	Cfg       		ClstCfg 		// 集群配置信息
	State			ClstStats		// 集群状态
	Nodes     		Nodes 			// 集群节点信息
	Services    	Services 		// 集群的应用服务名称信息
	WarningInfo		WarningInfo 	// 告警记录
}

// 集群静态配置结构体
type ClstCfg struct {
	Name      					 string   // 集群名称

	AgentUdpPort                 uint32   // Agent端监听的UDP的端口号
	ServerUdpPort                uint32   // 服务端监听的UDP的端口号
	ServerTcpPortForListenClient uint32   // 服务端监听Agent的TCP端口号
	ServerTcpPortForListenAgent  uint32   // 服务端监听Client的TCP端口号
	ResSampleFeq                 uint32   // 采样时间间隔,单位秒
	TaskMigrateTimeFromAgent     uint32   // 主机故障后，所有服务迁移时间，单位秒
}


// 集群资源监控结构体
type ClstStats struct {
	RunState			bool				// 集群系统的运行状态 true-运行中 false-已停止
	NodeCount			uint32				// 集群系统的当前节点数量
	ExecServiceCount 	uint32				// 集群系统的当前运行的服务数量
	Res					ResourceStatus		// 集群系统的资源信息
}
