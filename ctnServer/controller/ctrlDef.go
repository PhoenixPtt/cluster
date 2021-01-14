package controller

import (
	"context"
	"ctnCommon/pool"
	"ctnServer/ctnS"
	"sync"
)

var(
	SVC_NUM = 1000
	NODE_NUM = 1000
	CHAN_BUFFER = 1000

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

	NODE_WATCH = "NODE_WATCH"
	SERVICE_WATCH = "SVC_WATCH"
	SERVICE_STATUS_WATCH = "SVC_STATUS_WATCH"
)

type CONTROLLER struct {
	ServiceMap map[string]*SERVICE //集群内部的所有服务
	ServiceCfgMap map[string]*SVC_CFG//服务名称与服务配置的映射
	NodeStatusMap map[string]bool//节点状态列表

	CancelWatchSvcs map[string]context.CancelFunc
	//svcExitChanMap map[string]chan int
	CancelWatchNodes context.CancelFunc
	CancelWatchSvcStatus context.CancelFunc
	CancelDaq context.CancelFunc

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
	Timeout int `yaml:"timeout"`
	Placement Placement `yaml:"placement"`
	Resources []Resources `yaml:"resources"`
}

type SVC_DESCRIPTION struct {
	Name string `yaml:"name"`
	Image string `yaml:"image"`
	//Cmd            string 	`yaml:"cmd"`			// cmd入口指令
	//CmdPars        string 	`yaml:"cmdpars"`			// cmd入口指令参数
	//EntryPoint     string 	`yaml:"entrypoint"`			// 入口指令
	//EntryPointPars string 	`yaml:"entrypointpars"`			// 入口参数
	Deploy Deploy `yaml:"deploy"`
}

type SERVICE_OPER_TRUCK struct {
	SvcName string			//服务名称
	OperType string			//操作类型
	ConfigFileName string	//配置文件路径
	ConfigFileType string	//配置文件类型
	ScaleNum int			//规模
	SvcCfg SVC_CFG 			//服务配置信息
}

func NewController(sendObjFunc pool.SendObjFunc) (controller *CONTROLLER) {
	controller = &CONTROLLER{}
	controller.ServiceMap = make(map[string]*SERVICE, SVC_NUM) //为集群中的服务变量分配内存空间
	controller.ServiceCfgMap = make(map[string]*SVC_CFG, SVC_NUM)
	controller.NodeStatusMap = make(map[string]bool, NODE_NUM)//节点状态列表
	//controller.svcExitChanMap = make(map[string]chan int,SVC_NUM)
	controller.CancelWatchSvcs = make(map[string]context.CancelFunc, SVC_NUM)
	ctnS.Config(sendObjFunc)
	return
}

func GetServiceWatchKey(svcName string) (watchKey string) {
	watchKey = svcName+"_"+SERVICE_WATCH
	return
}




