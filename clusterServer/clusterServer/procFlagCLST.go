package clusterServer

import (
	header "clusterHeader"
	"errors"
)

// web 响应网络数据,非阻塞
// token 	身份认证标识、含用户、主机等信息
// data		接收到的数据包，类型根据flag指定
// respChan	回复当前消息的数据通道，填充数据后关闭，且只能填充一次，不可读取
func procFlagCLST(token interface{}, data interface{}, respChan chan<- interface{}) (err error) {

	// CLST 标识对应的数据结构为CLST
	r := data.(header.CLST)
	r.Oper.Success = true		// 默认操作成功
	r.Oper.Progress = 100		// 默认操作完成
	r.Oper.Err = ""				// 默认无错误

	// 根据子标识类型，执行相关操作
	switch r.Oper.Type {
	case header.FLAG_CLST:			// 获取集群所有数据
		r.Cfg = d								// 集群配置
		r.State = *clstStats					// 集群状态
		r.WarningInfo = *warings.WarningInfo()	// 集群告警信息
		prepareClstNodesData(&r.Nodes)			// 集群节点
		prepareClstServData(&r.Services)		// 集群服务
	case header.FLAG_CLST_CFG:		// 获取集群的配置信息
		r.Cfg = d
	case header.FLAG_CLST_CTRL:		// 集群控制和参数设置
		r.Oper.Success = false					// 操作失败
		r.Oper.Err = "该功能暂未实现"				// 操作失败信息
	case header.FLAG_CLST_NODE:		// 获取集群节点信息
		prepareClstNodesData(&r.Nodes)			// 集群节点
	case header.FLAG_CLST_SERVE:	// 获取集群的应用服务信息
		prepareClstServData(&r.Services)		// 集群服务
	case header.FLAG_CLST_STATS:	// 获取集群的状态信息 "
		r.State = *clstStats					// 集群状态
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

// 准备集群节点数据
func prepareClstNodesData(v *header.Nodes) {
	v.Node = nodes.GetNodes()					// 获取当前集群的所有节点信息
	v.Count = uint32(len(v.Node))				// 获取当前节点体的长度
}

// 准备集群应用服务数据
func prepareClstServData(v *header.Services) {
	v.Service = []header.Service{}				// 获取当前集群的所有应用信息
	v.Count = uint32(len(v.Service))			// 获取当前节点体的长度
}


