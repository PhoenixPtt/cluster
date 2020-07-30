package clusterServer

import (
	header "clusterHeader"
	"encoding/json"
	"fmt"
	"log"
	"tcpSocket"
)

func onAgentStateChanged(ip string, state uint8) {
	if state == tcpSocket.TCP_CONNECT_SUCCESS {
		nodes.AddNode(ip, true)
		// 网络连接成功后，立即请求Agent主机信息
		v := header.NODE{
			Oper:header.Oper {
				Type: header.FLAG_NODE_HOST,
			},
		}
		writeAgentData(ip, tcpSocket.TCP_TPYE_CONTROLLER, 0, header.FLAG_CLST, header.JsonByteArray(v))
		log.Println("Agent:", ip, "Connected success!")
	} else {
		// 网络连接端口后，仅仅从网络列表中标记为连接失败，并不从列表中移除
		nodes.SetNodeState(ip, false)
		log.Println("Agent:", ip, "DisConnected!")
	}
}

func writeAgentData(ip string, tcpType uint8, pkgId uint16, flag string, data []byte) {
	//fmt.Println(time.Now().Format("2006-01-02 15:04:05.000000"), "writeAgentData", ip, pkgId, flag)
	if len(ip) <= 0 { // 如果不指定ip，则给所有ip发
		nodeIds := nodes.GetNodeIds()
		for _, h := range nodeIds {
			if nodes.IsOnline(h) {
				tcpSocket.WriteData(h, tcpType, pkgId, flag, data)
			}
		}
	} else { // 如果指定了ip则仅回复指定ip
		tcpSocket.WriteData(ip, tcpType, pkgId, flag, data)
	}
}


func onAgentReadData(ip string, pkgId uint16, flag string, data []byte) {
	//fmt.Println(time.Now().Format("2006-01-02 15:04:05.000000"), "onAgentReadData", ip, pkgId, flag)

	switch flag {
	case header.FLAG_CLST:

	case header.FLAG_NODE: // 节点相关
		var v header.NODE
		err := json.Unmarshal(data, &v)
		if err != nil {
			fmt.Println("json转换为NODE时发生错误", err)
		}
		if len(v.Nodes.Node) > 0 {
			nd := &v.Nodes.Node[0]
			switch v.Oper.Type {
			case header.FLAG_NODE_HOST:
				nodes.SetHostInfo(ip, nd.HostInfo)
			case header.FLAG_NODE_RDAT:
				nodes.SetNodeState(ip, true)
				nodes.SetResStates(ip, nd.Res)



			case header.FLAG_NODE_REMV:
				// 移除节点，需要先迁移所有节点上的副本
				// 然后删除副本
				nodes.RemoveNode(ip)
			}
		}

	case header.FLAG_IMAG: // 镜像和仓库相关
		ReceiveDataFromAgent(ip, pkgId, data)

	case header.FLAG_CTNS: // 容器相关

	default:
	}
}
