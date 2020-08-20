package controller

import (
	"time"
)

//副本结构体
type REPLICA struct {
	RplName string			//副本名称
	SvcName string			//服务名称
	RplTargetStat string	//副本目标状态
	RplImage string			//副本镜像名称
	AgentAddr string		//副本被分配的节点id
	AgentStatus bool		//节点状态
	Dirty bool				//标记副本为脏
	RplStatus string		//副本状态
	CtnName string			//在容器池中的索引号
	LastRplOper []REPLICA_OPER	//最近一次的副本操作
}

type REPLICA_OPER struct {
	LastOperType string			//操作类型
	LastErrType string			//执行结果的错误类型
	LastErr		error			//执行结果的错误
	LastTime	time.Time		//时间
	LastTimeStr string			//时间的字符串类型
}

