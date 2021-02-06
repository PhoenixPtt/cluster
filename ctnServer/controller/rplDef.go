package controller

import (
	"context"
)

//副本结构体
type REPLICA struct {
	RplName        string //副本名称
	CreateTime     string // 副本创建时间
	SvcName        string //服务名称
	RplTargetStat  string //副本目标状态
	RplImage       string //副本镜像名称
	AgentAddr      string //副本被分配的节点id
	AgentStatus    bool   //节点状态
	CtnName        string //在容器池中的索引号
	CancelWatchCtn context.CancelFunc
	Timeout        int  //超时时间
	AgentTryNum    int  //agent执行容器操作失败后允许的最大尝试次数
	IsSchedulering bool //是否正在调度
}
