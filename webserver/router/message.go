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
	typeFlag string				// 根操作类型字符串
	opertype string				// 操作类型字符串
	pars     []header.OperPar	// 参数切片，每个参数就是一个OperPar结构体
	body     interface{}		// 当存在body时，填充本项内容
}

// 信息 对象结构体
type Message struct {
	msgType  string
	content  interface{}
	errorMsg errcode.Error
}

// 从集群服务端获取指定的数据，这里并不区分不同的http命令
func getMessage(reqinfo requestInf) (bSuccess bool, msg *Message) {
	// 创建返回值对象，初始时仅设定消息类型
	msg = &Message{
		msgType:  reqinfo.typeFlag,
	}

	// 根据requestInf中的typeFlag，生成对应的结构体并赋值给data变量
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
		// 根据请求信息的内容，生成ImageData结构体
		data = createImageData(reqinfo)
	default:
		msg.content = fmt.Sprintf("Message type is:%v - the time is %v", reqinfo.typeFlag, time.Now())
		msg.errorMsg = errcode.ErrorCodeUnknown.WithMessage("无有效对应的请求类型")
		return false, msg
	}

	// 同集群服务端进行交互
	// 创建承载集群服务端返回的通道对象，设定一个缓冲区
	var temRespChan chan interface{} = make(chan interface{}, 1)

	// 调用集群服务端的指定接口函数
	//fmt.Println("ResponseURL data is： ", data)
	cluster.ResponseURL(reqinfo.typeFlag, "", data, temRespChan)
	// 等待读取通道对象内的数据
	temResp := <-temRespChan

	//fmt.Println("getMessage => cluster.ResponseURL :", temResp)

	// 分析返回的数据
	bSuccess = analysisResponseData(temResp, msg)

	return bSuccess, msg
}

// 生成镜像操作使用的结构体
func createImageData(req requestInf) header.ImageData {
	// 根据请求信息中的opertype信息，创建ImageData结构体
	data := header.ImageData {
		DealType: req.opertype,
	}

	// 根据操作类型执行具体的操作
	switch data.DealType {
	case header.FLAG_IMAG_LIST:	// 获取所有镜像的列表
	case header.FLAG_IMAG_TGLS:	// 获取指定镜像名称的tag列表
		// 从请求信息结构体中获取参数信息，必然有一个参数，如果个数大于等于1，则取首个参数的值为镜像名称
		if len(req.pars) >= 1 {
			data.ImageName = req.pars[0].Value
		} else {
			data.ImageName = ""
		}
		// 创建镜像、加载镜像、镜像推送
	case header.FLAG_IMAG_BUID, header.FLAG_IMAG_LOAD, header.FLAG_IMAG_DIST:
		if contentData, bok := req.body.(header.ImageData); bok {
			return contentData
		}
	case header.FLAG_IMAG_REMO:	// 镜像删除
		if contentData, bok := req.body.(header.ImageData); bok {
			return contentData
		}
	//case header.FLAG_IMAG_BUID:	// 创建镜像
	//	if contentData, bok := req.body.(header.ImageData); bok {
	//		return contentData
	//	}
	//case header.FLAG_IMAG_LOAD:	// 加载镜像
	//	if contentData, bok := req.body.(header.ImageData); bok {
	//		return contentData
	//	}
	//case header.FLAG_IMAG_DIST:	// 镜像推送
	//	if contentData, bok := req.body.(header.ImageData); bok {
	//		return contentData
	//	}
	//case header.FLAG_IMAG_REMO:	// 镜像删除
	//	if contentData, bok := req.body.(header.ImageData); bok {
	//		return contentData
	//	}
	default:

	}

	return data
}

// 解析从集群服务端获取的信息，设置Message结构体的内容
func analysisResponseData(resp interface{}, msg *Message) bool {
	// 根据返回值的类型，进行后续处理
	switch data := resp.(type) {
	case header.CLST: // 集群相关
		msg.content = data
	case header.NODE: // 节点相关
		msg.content = data
	case header.ImageData:	// 镜像相关
		msg.content = data
	case string:
		msg.content = fmt.Sprintf("  Message type is:%v - the time is %v", msg.msgType, time.Now())
	case map[string]interface{}:

	case []interface{}:

	default:
		msg.content = data
	}

	return true
}
