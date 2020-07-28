// "message.go" file is create by Huxd 2020.07.13

package router

import (
	"fmt"
	"time"
	"webserver/router/errcode"

	cluster "clusterServer/clusterServer"
)

// 信息 对象结构体
type Message struct {
	msgType  string
	content  interface{}
	errorMsg errcode.Error
}

// 从集群服务端获取指定的数据
func getMessage(msgType string) (bool, *Message) {

	// 同集群服务端进行交互
	// 创建承载集群服务端返回的通道对象，设定一个缓冲区
	var temRespChan chan interface{} = make(chan interface{}, 1)
	// 调用集群服务端的指定接口函数
	cluster.ResponseURL(msgType, []byte("abc"), temRespChan)
	// 等待读取通道对象内的数据
	temResp := <-temRespChan

	// 根据返回值的类型，进行后续处理
	var msg *Message
	switch resp := temResp.(type) {
	case string:
		msg = &Message{
			msgType: msgType,
			content: resp + fmt.Sprintf("  Message type is:%v - the time is %v", msgType, time.Now()),
		}
	case map[string]interface{}:

	case []interface{}:

	default:
		// Create a simple Message to send to clients, including the current time.
		msg = &Message{
			msgType: msgType,
			content: fmt.Sprintf("Message type is:%v - the time is %v", msgType, time.Now()),
		}
	}

	return true, msg
}
