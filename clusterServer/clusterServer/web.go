package clusterServer

import (
	"clusterHeader"
	"errors"
)

// web 响应网络数据,非阻塞
// flag 	消息标识
// token 	身份认证标识、含用户、主机等信息
// data		接收到的数据包，类型根据flag指定
// respChan	回复当前消息的数据通道，需要在填充数据后立即关闭，且只能填充一次，不可读取
func ResponseURL(flag string, token interface{}, data interface{}, respChan chan<- interface{}) (err error) {

	switch flag {
	case header.FLAG_CLST:   	// 集群相关
		return procFlagCLST(token, data, respChan)
	case header.FLAG_NODE:   	// 节点相关
		return  procFlagNODE(token, data, respChan)
	case header.FLAG_SERV:   	// 应用服务相关
		return  procFlagSERV(token, data, respChan)
	case header.FLAG_IMAG:   	// 镜像相关
		return  procFlagIMAG(token, data, respChan)
	default:
		respChan <- header.MESG{}
		err = errors.New("消息标识错误！")
	}

	close(respChan)
	return err
}


//case header.FLAG_CTNS:   	// 容器操作
//case header.FLAG_FILE:   	// 文件
//case header.FLAG_OTHR:		// 其它
//case header.FLAG_CMSG:		// 消息
//case header.FLAG_EVTM:		// 事件






