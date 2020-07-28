package header

import (
	"bytes"
	"encoding/binary"
)

// 比较关系 无效 大于 小于 等等
type Relation int

// 关系常量
const (
	Invalid Relation = iota
	Not
	Eq
	Bt
	Lt
	Be
	Le
	Bele
	Belt
	Btle
	Btlt
	Nbele
	Nbelt
	Nbtle
	Nbtlt
)

// 阈值
type Throad struct {
	Re  Relation   // 运算关系
	min float64   // 最小值
	max float64   // 最大值
}

type Float64Data struct {
	Val       		float64        	// 当前值
	ThroadVal 		Throad   		// 阈值
	IsWarning 		bool     		// 是否超过阈值
}

type SysResStatus struct {
	Cpu  			CpuStatus      // CPU
	Mem  			StorageStatus  // 内存
	Disk 			StorageStatus  // 硬盘
}

//type ConResStatus struct {
//	ID  			string      	// 容器ID
//	Cpu 			CpuStatus      	// CPU
//	Mem 			StorageStatus  	// 内存
//}

type ResourceStatus struct {
	Handle  		string        	// 当前Agent标识 ip:port
	Time 			string			// 数据采样时间
	SysResStatus   					// 系统资源
}

//func (c *ResourceStatus) Data() (data []byte) {
//	conCount := len(c.Ctn)
//	var d [][]byte = make([][]byte, 5+conCount*2)
//
//	// 系统资源
//	index := 0
//	d[index] = []byte{0, 0, 0, 0}
//	index++
//	d[index] = c.Sys.Cpu.Data()
//	index++
//	d[index] = c.Sys.Mem.Data()
//	index++
//	d[index] = c.Sys.Disk.Data()
//
//	// 容器资源
//	index++
//	d[index] = []byte{byte(conCount / 256), byte(conCount % 256)}
//	for i := 0; i < conCount; i++ {
//		index++
//		d[index] = c.Ctn[i].Cpu.Data()
//		index++
//		d[index] = c.Ctn[i].Mem.Data()
//	}
//
//	data = bytes.Join(d, []byte{})
//
//	size := len(data)
//	data[0] = byte((size >> 24) & 0xFF)
//	data[1] = byte((size >> 16) & 0xFF)
//	data[2] = byte((size >> 8) & 0xFF)
//	data[3] = byte(size & 0xFF)
//
//	return
//}
//
//func (c *ResourceStatus) SetData(data []byte) (dataLen int32) {
//	var index int32 = 4
//	dataLen = bytesToInt32(data[0:4])
//	index += c.Sys.Cpu.SetData(data[index:])
//	index += c.Sys.Mem.SetData(data[index:])
//	index += c.Sys.Disk.SetData(data[index:])
//	conCount := bytesToUInt16(data[index : index+2])
//	index += 2
//	c.Ctn = make([]ConResStatus, conCount)
//	for i := uint16(0); i < conCount; i++ {
//		index += c.Ctn[i].Cpu.SetData(data[index:])
//		index += c.Ctn[i].Mem.SetData(data[index:])
//	}
//	return
//}

// 阈值比较
func judge(data *Float64Data) bool {
	switch data.ThroadVal.Re {
	case Invalid:
		data.IsWarning = false
	case Not:
		data.IsWarning = data.ThroadVal.min != data.Val
	case Eq:
		data.IsWarning = data.ThroadVal.min == data.Val
	case Bt:
		data.IsWarning = data.Val > data.ThroadVal.max
	case Lt:
		data.IsWarning = data.Val < data.ThroadVal.min
	case Be:
		data.IsWarning = data.Val >= data.ThroadVal.max
	case Le:
		data.IsWarning = data.Val <= data.ThroadVal.min
	case Bele:
		data.IsWarning = (data.Val >= data.ThroadVal.min) && (data.Val <= data.ThroadVal.max)
	case Belt:
		data.IsWarning = (data.Val >= data.ThroadVal.min) && (data.Val < data.ThroadVal.max)
	case Btle:
		data.IsWarning = (data.Val > data.ThroadVal.min) && (data.Val <= data.ThroadVal.max)
	case Btlt:
		data.IsWarning = (data.Val > data.ThroadVal.min) && (data.Val < data.ThroadVal.max)
	case Nbele:
		data.IsWarning = (data.Val <= data.ThroadVal.min) || (data.Val >= data.ThroadVal.max)
	case Nbelt:
		data.IsWarning = (data.Val <= data.ThroadVal.min) || (data.Val > data.ThroadVal.max)
	case Nbtle:
		data.IsWarning = (data.Val < data.ThroadVal.min) || (data.Val >= data.ThroadVal.max)
	case Nbtlt:
		data.IsWarning = (data.Val < data.ThroadVal.min) || (data.Val > data.ThroadVal.max)
	}

	return data.IsWarning
}

func bytesToInt32(b []byte) (data int32) {
	bytesBuffer := bytes.NewReader(b)
	binary.Read(bytesBuffer, binary.BigEndian, &data)
	return
}

func bytesToUInt16(b []byte) (data uint16) {
	bytesBuffer := bytes.NewReader(b)
	binary.Read(bytesBuffer, binary.BigEndian, &data)
	return data
}
