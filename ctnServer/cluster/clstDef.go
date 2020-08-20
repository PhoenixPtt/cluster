package cluster

import (
	"sync"
)

var(
	SVC_NUM = 1000
	NODE_NUM = 1000

	YML_FILE		= "YAML文件"
	JSON_FILE		= "JSON串"

	SCREATE  = "CREATE"
	SSTART   = "START"
	SSTOP    = "STOP"
	SRESTART = "RESTART"
	SREMOVE  = "REMOVE"
	SSCALE	 = "SCALE"

	ALL_NODES = 0
	ACTIVE_NODES = 1
)

type CLUSTER struct {
	ClstName string//集群名称
	ServiceMap map[string]*SERVICE //集群内部的所有服务
	ServiceCfgMap map[string]*SVC_CFG//服务名称与服务配置的映射
	NodeStatusMap map[string]bool//节点状态列表

	exitWatchServicesChan chan int
	exitWatchNodesChan chan int

	Mutex sync.Mutex
}

//服务配置结构体
type SVC_CFG struct {
	Version string `yaml:"version"`
	Description SVC_DESCRIPTION `yaml:"service"`
}

type Placement struct {
	Constraints []string `yaml:"constraints"`
}

type Resources struct {
	Name string `yaml:"name"`
	Cpus string `yaml:"cpus"`
	Memory string `yaml:"memory"`
}

type Deploy struct {
	Mode string `yaml:"mode"`
	Replicas int `yaml:"replicas"`
	Placement Placement `yaml:"placement"`
	Resources []Resources `yaml:"resources"`
}

type SVC_DESCRIPTION struct {
	Name string `yaml:"name"`
	Image string `yaml:"image"`
	Deploy Deploy `yaml:"deploy"`
}

type SERVICE_OPER_TRUCK struct {
	SvcName string			//服务名称
	OperType string			//操作类型
	ConfigFileName string	//配置文件路径
	ConfigFileType string	//配置文件类型
	ScaleNum int			//规模
}

func NewCluster(clusterName string) (cluster *CLUSTER) {
	cluster = &CLUSTER{}
	cluster.ClstName = clusterName
	cluster.ServiceMap = make(map[string]*SERVICE, SVC_NUM) //为集群中的服务变量分配内存空间
	cluster.ServiceCfgMap = make(map[string]*SVC_CFG, SVC_NUM)
	cluster.NodeStatusMap = make(map[string]bool, NODE_NUM)//节点状态列表
	cluster.exitWatchServicesChan = make(chan int, 100)
	cluster.exitWatchNodesChan = make(chan int, 100)
	return
}




