package header

type StorageStatus struct {
	UsedPercent Float64Data  		// 使用率
	Used        Float64Data       	// 内存使用率 %
	Total       Float64Data         // 内存总量
	Temperature Float64Data   		// 温度
	Health      Float64Data  		// 温度
}

func (c *StorageStatus) SetUsedData(used, total uint64) {
	c.Used.Val = float64(used)
	judge(&c.Used)

	c.Total.Val = float64(total)
	judge(&c.Total)

	c.UsedPercent.Val = 100 * c.Used.Val / c.Total.Val
	judge(&c.UsedPercent)
}

func (c *StorageStatus) SetTemperatureData(temp float64) {
	c.Temperature.Val = temp
	judge(&c.Temperature)
}

// 根据阈值计算是否超过阈值
func (c *StorageStatus) SetHealthData(health float64) {
	c.Health.Val = health
	judge(&c.Health)
}

//func (c *StorageStatus) Data() (data []byte) {
//	bytesBuffer := bytes.NewBuffer([]byte{})
//	var size int32 = 0
//	binary.Write(bytesBuffer, binary.BigEndian, &size)
//
//	binary.Write(bytesBuffer, binary.BigEndian, &c.UsedPercent.Val)
//	binary.Write(bytesBuffer, binary.BigEndian, &c.UsedPercent.IsWarning)
//	binary.Write(bytesBuffer, binary.BigEndian, &c.Used.Val)
//	binary.Write(bytesBuffer, binary.BigEndian, &c.Used.IsWarning)
//	binary.Write(bytesBuffer, binary.BigEndian, &c.Total.Val)
//	binary.Write(bytesBuffer, binary.BigEndian, &c.Total.IsWarning)
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
//func (c *StorageStatus) SetData(data []byte) (dataLen int32) {
//	bytesBuffer := bytes.NewReader(data)
//	binary.Read(bytesBuffer, binary.BigEndian, &dataLen)
//	binary.Read(bytesBuffer, binary.BigEndian, &c.UsedPercent.Val)
//	binary.Read(bytesBuffer, binary.BigEndian, &c.UsedPercent.IsWarning)
//	binary.Read(bytesBuffer, binary.BigEndian, &c.Used.Val)
//	binary.Read(bytesBuffer, binary.BigEndian, &c.Used.IsWarning)
//	binary.Read(bytesBuffer, binary.BigEndian, &c.Total.Val)
//	binary.Read(bytesBuffer, binary.BigEndian, &c.Total.IsWarning)
//	binary.Read(bytesBuffer, binary.BigEndian, &c.Temperature.Val)
//	binary.Read(bytesBuffer, binary.BigEndian, &c.Temperature.IsWarning)
//	binary.Read(bytesBuffer, binary.BigEndian, &c.Health.Val)
//	binary.Read(bytesBuffer, binary.BigEndian, &c.Health.IsWarning)
//	return
//}
