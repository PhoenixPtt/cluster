package header

import (
	"github.com/shirou/gopsutil/host"
)

const (
	FLAG_NODE   = "NODE"			// 节点设置
	FLAG_NODE_RDAT = "NODE_RDAT" 	// 节点的资源数据
	FLAG_NODE_HOST = "NODE_HOST" 	// 节点主机信息
	FLAG_NODE_REMV = "NODE_REMV"	// 移除节点
)

type NODE struct {
	Oper
	Nodes
}

type Nodes struct {
	Count			uint32				// 节点数量
	Node			[]Node				// 所有的节点信息
}

type Node struct {
	NodeInfo
	Res      		ResourceStatus 		// 资源数据
}

type NodeInfo struct {
	Handle   		string              // 节点句柄
	State    		bool                // 连接状态
	HostInfo 		host.InfoStat       // 主机信息
	Labels   		[]NodeLabel			// 主机标签
}

type NodeLabel struct {
	Name 			string	 			// 标签名称
	Value			string  			// 标签值
}
