package clusterServer

import (
	"clusterHeader"
)

// web 响应网络数据,非阻塞
func ResponseURL(flag string, data interface{}, respChan chan interface{}) (err error) {

	switch flag {
	case header.FLAG_CLST:   	// 集群相关
		//_,_ := data.(header.CLST)
		respChan <- header.CLST{}

	case header.FLAG_NODE:   	// 节点相关
		v := data.(header.NODE)
		switch v.Oper.Type {
		case header.FLAG_NODE:
			v.Oper.Success = true
			v.Oper.Progress = 100
			v.Oper.Err = ""
			v.Nodes.Count = uint32(nodes.Count())
			v.Nodes.Node = make([]header.Node, v.Nodes.Count)
			for i,h := range nodes.GetNodeIds() {
				v.Nodes.Node[i] = nodes.GetNode(h).Node
			}
			respChan <- v

		case header.FLAG_NODE_HOST:
			v.Oper.Success = true
			v.Oper.Progress = 100
			v.Oper.Err = ""
			v.Nodes.Count = uint32(nodes.Count())
			v.Nodes.Node = make([]header.Node, v.Nodes.Count)
			for i,h := range nodes.GetNodeIds() {
				v.Nodes.Node[i].NodeInfo = nodes.GetNode(h).Node.NodeInfo
				v.Nodes.Node[i].Res = header.ResourceStatus{}
			}
			respChan <- v

		case header.FLAG_NODE_RDAT:
			v.Oper.Success = true
			v.Oper.Progress = 100
			v.Oper.Err = ""
			v.Nodes.Count = uint32(nodes.Count())
			v.Nodes.Node = make([]header.Node, v.Nodes.Count)
			for i,h := range nodes.GetNodeIds() {
				v.Nodes.Node[i].NodeInfo = header.NodeInfo{}
				v.Nodes.Node[i].Res = nodes.GetNode(h).Node.Res
			}
			respChan <- v

		case header.FLAG_NODE_REMV: // 移除节点
			v.Oper.Success = true
			v.Oper.Progress = 100
			v.Oper.Err = ""
			// 移除节点，后续完善

			respChan <- v

		default:

		}



		respChan <- header.NODE{}
	case header.FLAG_SERV:   	// 应用服务相关
		respChan <- header.SERVICE{}
	//case header.FLAG_IMAG:   	// 镜像操作
	//case header.FLAG_CTNS:   	// 容器操作
	//case header.FLAG_FILE:   	// 文件
	//case header.FLAG_OTHR:		// 其它
	//case header.FLAG_CMSG:		// 消息
	//case header.FLAG_EVTM:		// 事件

	default:
		respChan <- header.MESG{}
	}

	close(respChan)
	return err
}
