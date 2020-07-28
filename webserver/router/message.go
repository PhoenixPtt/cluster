// "message.go" file is create by Huxd 2020.07.13

package router

import (
	header "clusterHeader"
	"fmt"
	"time"
	"webserver/router/errcode"

	cluster "clusterServer/clusterServer"
)

// 请求信息 结构体
type requestInf struct {
	typeFlag string
	opertype string
	pars     []header.OperPar
	body     interface{}	// 当存在body时，填充本项内容
}

// 信息 对象结构体
type Message struct {
	msgType  string
	content  interface{}
	errorMsg errcode.Error
}

// 从集群服务端获取指定的数据
func getMessage(reqinfo requestInf) (bSuccess bool, msg *Message) {
	// 根据requestInf中的typeFlag，生成对应的结构体
	var data interface{}
	switch reqinfo.typeFlag {
	case header.FLAG_CLST: // 集群相关
		data = header.CLST{
			Oper: header.Oper{
				Type: reqinfo.opertype,
			},
		}
	case header.FLAG_NODE: // 节点相关

	case header.FLAG_SERV: // 应用服务相关

	case header.FLAG_IMAG: // 镜像相关
		
	default:
		msg = &Message{
			msgType:  reqinfo.typeFlag,
			content:  fmt.Sprintf("Message type is:%v - the time is %v", reqinfo.typeFlag, time.Now()),
			errorMsg: errcode.ErrorCodeUnknown.WithMessage("webserver internal error!"),
		}

		return false, msg
	}

	// 同集群服务端进行交互
	// 创建承载集群服务端返回的通道对象，设定一个缓冲区
	var temRespChan chan interface{} = make(chan interface{}, 1)

	// 调用集群服务端的指定接口函数
	cluster.ResponseURL(reqinfo.typeFlag, data, temRespChan)
	// 等待读取通道对象内的数据
	temResp := <-temRespChan

	//fmt.Println("getMessage => cluster.ResponseURL :", temResp)

	///////////////////////////////////////////////////////////////////////////////////////////////
	// 根据返回值的类型，进行后续处理
	switch resp := temResp.(type) {
	case header.CLST: // 集群相关
		msg = &Message{
			msgType: reqinfo.typeFlag,
			content: resp,
		}
	case header.NODE: // 节点相关
		msg = &Message{
			msgType: reqinfo.typeFlag,
			content: resp,
		}
	case string:
		msg = &Message{
			msgType: reqinfo.typeFlag,
			content: resp + fmt.Sprintf("  Message type is:%v - the time is %v", reqinfo.typeFlag, time.Now()),
		}
	case map[string]interface{}:

	case []interface{}:

	default:
		// Create a simple Message to send to clients, including the current time.
		msg = &Message{
			msgType: reqinfo.typeFlag,
			content: fmt.Sprintf("Message type is:%v - the time is %v", reqinfo.typeFlag, time.Now()),
		}
	}

	return true, msg
}
