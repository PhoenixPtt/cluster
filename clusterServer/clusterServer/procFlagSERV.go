package clusterServer

import (
	header "clusterHeader"
	"errors"
)

// web 响应网络数据,非阻塞
// token 	身份认证标识、含用户、主机等信息
// data		接收到的数据包，类型根据flag指定
// respChan	回复当前消息的数据通道，填充数据后按需关闭，且只能填充一次，不可读取
func procFlagSERV(token interface{}, data interface{}, respChan chan<- interface{}) (err error) {
	// 对应的数据结构为SERVICE
	r := data.(header.SERVICE)
	r.Oper.Success = true		// 默认操作成功
	r.Oper.Progress = 100		// 默认操作完成
	r.Oper.Err = ""				// 默认无错误

	// 根据子标识类型，执行相关操作
	switch r.Oper.Type {
	default:						// 其它操作处理
		r.Oper.Success = false					// 操作失败
		r.Oper.Err = "子标识错误！"				// 操作失败信息
	}

	// 写入数据到通道
	respChan <- r
	close(respChan)

	// 返回错误信息
	return errors.New(r.Oper.Err)
}
