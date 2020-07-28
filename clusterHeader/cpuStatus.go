package header

// CPU信息,使用之前必须先调用 IniCpuStatus，否则会报越界
type CpuStatus struct {
	CoreCount       uint32        // 核数
	UsedPercent     Float64Data   // CPU使用率 %
	Temperature     Float64Data   // 温度
	Health          Float64Data   // 健康度
	CoreUsedPercent []Float64Data // cpu每个核心的使用率
}

func (c *CpuStatus) SetUsedPercentData(data float64) {
	c.UsedPercent.Val = data
	judge(&c.UsedPercent)
}

func (c *CpuStatus) SetCoreUsedPercent(data []float64) {

	if len(data) != len(c.CoreUsedPercent) {
		c.CoreCount = uint32(len(data))
		c.CoreUsedPercent = make([]Float64Data, c.CoreCount)
	}

	for i := uint32(0); i < c.CoreCount; i++ {
		c.CoreUsedPercent[i].Val = data[i]
		judge(&c.CoreUsedPercent[i])
	}
}

func (c *CpuStatus) SetTemperatureData(data float64) {
	c.Temperature.Val = data
	judge(&c.Temperature)
}

func (c *CpuStatus) SetHealthData(data float64) {
	c.Health.Val = data
	judge(&c.Health)
}


//func (c *CpuStatus) Data() (data []byte) {
//	bytesBuffer := bytes.NewBuffer([]byte{})
//	var size int32 = 0
//	binary.Write(bytesBuffer, binary.BigEndian, &size)
//	binary.Write(bytesBuffer, binary.BigEndian, &c.CoreCount)
//	binary.Write(bytesBuffer, binary.BigEndian, &c.UsedPercent.Val)
//	binary.Write(bytesBuffer, binary.BigEndian, &c.UsedPercent.IsWarning)
//	for i := uint32(0); i < c.CoreCount; i++ {
//		binary.Write(bytesBuffer, binary.BigEndian, &c.CoreUsedPercent[i].Val)
//		binary.Write(bytesBuffer, binary.BigEndian, &c.CoreUsedPercent[i].IsWarning)
//	}
//	binary.Write(bytesBuffer, binary.BigEndian, &c.Temperature.Val)
//	binary.Write(bytesBuffer, binary.BigEndian, &c.Temperature.IsWarning)
//	binary.Write(bytesBuffer, binary.BigEndian, &c.Health.Val)
//	binary.Write(bytesBuffer, binary.BigEndian, &c.Health.IsWarning)
//
//	data = bytesBuffer.Bytes()
//
//	size = int32(len(data))
//	data[0] = byte((size >> 24) & 0xFF)
//	data[1] = byte((size >> 16) & 0xFF)
//	data[2] = byte((size >> 8) & 0xFF)
//	data[3] = byte(size & 0xFF)
//
//	return
//}
//
//func (c *CpuStatus) SetData(data []byte) (dataLen int32) {
//	bytesBuffer := bytes.NewReader(data)
//	binary.Read(bytesBuffer, binary.BigEndian, &dataLen)
//	binary.Read(bytesBuffer, binary.BigEndian, &c.CoreCount)
//	c.CoreUsedPercent = make([]Float64Data, c.CoreCount)
//	binary.Read(bytesBuffer, binary.BigEndian, &c.UsedPercent.Val)
//	binary.Read(bytesBuffer, binary.BigEndian, &c.UsedPercent.IsWarning)
//	for i := uint32(0); i < c.CoreCount; i++ {
//		binary.Read(bytesBuffer, binary.BigEndian, &c.CoreUsedPercent[i].Val)
//		binary.Read(bytesBuffer, binary.BigEndian, &c.CoreUsedPercent[i].IsWarning)
//	}
//	binary.Read(bytesBuffer, binary.BigEndian, &c.Temperature.Val)
//	binary.Read(bytesBuffer, binary.BigEndian, &c.Temperature.IsWarning)
//	binary.Read(bytesBuffer, binary.BigEndian, &c.Health.Val)
//	binary.Read(bytesBuffer, binary.BigEndian, &c.Health.IsWarning)
//	return
//}