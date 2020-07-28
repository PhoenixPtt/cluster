package header

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
)

type CpuMemStruct struct {
	Id          string
	Read        string
	Preread     string
	CpuStats    cpu
	PreCpuStats cpu
	MemoryStats mem
}

type cpu struct {
	Usage    cpuUsage
	SysUsage float64
}

type cpuUsage struct {
	Total  float64
	PerCpu []float64
}

type mem struct {
	Usage    float64
	MaxUsage float64
	Status   memStats
	Limite   float64
}

type memStats struct {
	ActiveAnon float64
}

type CTNO struct {
	AgentAddr   string
	ServiceName string
	OperType    string
	OperIndex   int
	OperError   error
	ImageName   string
	ImageTag    string
	CtnId       string
	CtnInfoFlag int
	CtnInfo     []types.Container
	//Cpu2MemChan CpuMemStruct
	CtnEvtMsg events.Message
}
