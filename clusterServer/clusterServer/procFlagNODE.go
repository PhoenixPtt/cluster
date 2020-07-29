package clusterServer

import (
	header "clusterHeader"
	"errors"
)

// web 响应网络数据,非阻塞
// token 	身份认证标识、含用户、主机等信息
// data		接收到的数据包，类型根据flag指定
// respChan	回复当前消息的数据通道，填充数据后关闭，且只能填充一次，不可读取
func procFlagNODE(token interface{}, data interface{}, respChan chan<- interface{}) (err error) {
	// data 接口的数据类型为header.NODE
	r := data.(header.NODE)
	r.Oper.Success = true		// 默认操作成功
	r.Oper.Progress = 100		// 默认操作完成
	r.Oper.Err = ""				// 默认无错误

	switch r.Oper.Type {
	case header.FLAG_NODE:		// 获取所有的节点数据
		// 集群节点
		prepareClstNodesData(&r.Nodes)

	case header.FLAG_NODE_HOST:	// 获取节点信息
		// 集群节点
		prepareClstNodesData(&r.Nodes)
		// 清空资源信息
		for i:=uint32(0); i<r.Nodes.Count; i++ {
			r.Nodes.Node[i].Res = header.ResourceStatus{}
		}

	case header.FLAG_NODE_RDAT: // 获取资源数据
		// 集群节点
		prepareClstNodesData(&r.Nodes)
		// 清空主机信息
		for i:=uint32(0); i<r.Nodes.Count; i++ {
			r.Nodes.Node[i].NodeInfo = header.NodeInfo{}
		}

	case header.FLAG_NODE_REMV: // 移除节点
		r.Oper.Success = false
		r.Oper.Err = "该功能暂不支持"
		// 移除节点，后续完善

	default:
		r.Oper.Success = false
		r.Oper.Err = "子标识错误！"
	}

	// 写入数据到通道
	respChan <- r
	close(respChan)

	// 返回错误信息
	return errors.New(r.Oper.Err)
}
