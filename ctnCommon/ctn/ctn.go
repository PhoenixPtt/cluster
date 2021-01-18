package ctn

import (
	"io"

	"github.com/docker/docker/api/types"
)

const (
	//包标识
	FLAG_CTRL  = "CTRL"
	FLAG_CTN   = "INFO"
	FLAG_STATS = "STAT"
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
	CTNEXIT = "EXIT"
)

//容器结构体声明
type CTN struct {
	CtnName   string `json:"ctn_name"`
	ImageName string `json:"image_name"`
	CtnID string `json:"ctn_id"`
	AgentAddr string `json:"agentaddr"`

	//容器状态，容器创建事件，容器状态更新时间
	State         string `json:"state"`
	Dirty         bool   `json:"dirty"`
	Created       int64
	CreatedString string `json:"created_string"`
	Updated       int64
	UpdatedString string `json:"update_string"`

	//操作类型和时间
	OperType     string //记录最近一次的操作
	OperStrategy bool   //是否启动
	OperNum      int
	OperTime     int64
	OperTimeStr  string

	//容器信息和时间
	types.Container

	//容器事件和时间
	CtnAction        string `json:"ctn_action"`
	CtnActionTime    string `json:"ctn_action_time"`
	CtnActionTimeInt int64
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
