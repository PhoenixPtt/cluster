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

	DIRTY_POSITION_DOCKER                               = "副本（容器）失效原因：docker服务器"
	DIRTY_POSITION_CTN_EXIST_IN_SERVER_BUT_NOT_IN_AGENT = "副本（容器）失效原因：server端与agent端连接正常但数据不一致，通常是server端有容器信息，而agent端没有改容器信息" //此类型还需要具体判断失效原因
	DIRTY_POSITION_RPL_RUN_OR_REMOVE_ERR                = "副本（容器）失效原因：副本执行RUN操作或者REMOVE操作失败"
	DIRTY_POSITION_ERR_BEFORE_RPL_OPER                  = "副本（容器）失效原因：副本操作在前合法性检测失败"
	DIRTY_POSITION_RPL_OPER_TIMEOUT                     = "副本（容器）失效原因：server端副本操作超时"
	DIRTY_POSITION_REMOVED                              = "正常删除"
	DIRTY_POSTION_IMAGE_RUN_ERR                         = "副本（容器）失效原因：：镜像运行失败"
	DIRTY_POSTION_SERVER_LOST_CONNICTION                = "副本（容器）失效原因：server端与agent端连接断开"

	CTN_NOT_EXIST_ON_AGENT = "not exist on agent"
	CTN_UNKNOWN_CTN_STATUS = "容器状态未知"
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
