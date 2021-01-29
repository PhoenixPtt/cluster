package ctn

import (
	"github.com/docker/docker/api/types"
	"io"
)

const (
	//包标识
	FLAG_CTRL = "CTRL"
	FLAG_CTN  = "INFO"
	//FLAG_STATS = "STAT"
	FLAG_EVENT = "EVTM"

	//操作标识
	CREATE  = "CREATE"
	START   = "START"
	RUN     = "RUN"
	STOP    = "STOP"
	KILL    = "KILL"
	REMOVE  = "REMOVE"
	GETLOG  = "GETLOG"
	INSPECT = "INSPECT"
	//CTNEXIT = "EXIT"

	DIRTY_POSITION_DOCKER  = "失效位置：docker服务器"
	DIRTY_POSITION_AGENT   = "失效位置：agent端"
	DIRTY_POSITION_REMOVED = "正常删除"
	DIRTY_POSTION_IMAGE    = "失效位置：镜像"
	DIRTY_POSTION_SERVER   = "失效位置：server端"

	CTN_NOT_EXIST_ON_AGENT = "not exist on agent"
)

//容器结构体声明
type CTN struct {
	CtnName   string `json:"ctn_name"`
	ImageName string `json:"image_name"`
	CtnID     string `json:"ctn_id"`
	AgentAddr string `json:"agentaddr"`

	//容器状态
	//State         string `json:"state"`
	Dirty         bool   `json:"dirty"`
	DirtyPosition string `json:"dirty_position"`

	//操作
	OperType     string //记录最近一次的操作
	OperIndex    int    //操作序号
	AgentTryNum  int    //agent执行容器操作失败后允许的最大尝试次数
	OperStrategy bool   //是否启动

	//应答
	OperErr         string
	CtnLog          string
	CtnInspect      CTN_INSPECT
	types.Container //容器信息和时间
	CTN_STATS       //容器资源使用情况

	Created       int64
	CreatedString string `json:"created_string"`
	Updated       int64
	UpdatedString string `json:"update_string"`

	////容器事件和时间
	//CtnAction        string `json:"ctn_action"`
	//CtnActionTime    string `json:"ctn_action_time"`
	//CtnActionTimeInt int64
	//OperNum      int
	//OperTime     int64
	//OperTimeStr  string
}

//容器操作接口声明
type ctnO interface {
	Create() error
	Start() error
	Run() error
	Stop() error
	Kill() error
	Remove() error
	GetLog() (io.ReadCloser, error)
	Inspect() (CTN_INSPECT, error)
}
