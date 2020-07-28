package header

import (
	"github.com/docker/docker/api/types"
)

// 容器操作
const (
	FLAG_CTNS   	 = "CTNS"
	FLAG_CTNS_CTRL   = "CTNS_CTRL"
	FLAG_CTNS_STATS  = "CTNS_STATS"
	FLAG_CTNS_INFO   = "CTNS_INFO"
)

type CTN struct {
	Container
	Logs        string      //容器日志对应getlog
	CtnInspect  CTN_INSPECT //容器信息对应inspect
	Err         string
	OperFlag    string
	OperIndex   int
	AgentAddr   string
	ServiceName string
}

type Container types.Container

var (
	g_ctn_index int
)

//准备数据
func (c *CTN) PrepareData(operFlag string) int {
	g_ctn_index++
	c.OperFlag = operFlag
	c.OperIndex = g_ctn_index
	c.Err = ""
	return g_ctn_index
}
