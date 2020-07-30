package clusterServer

import (
	"fmt"
	"sync"
	"time"
)

type Request struct {
	handle 		uint16
	Response 	chan<- interface{}
	Completed 	chan bool
}

var requests map[uint16]*Request = make(map[uint16]*Request)
var requestsMutex sync.Mutex

// 初始化请求
func (r *Request)Init(respChan chan <- interface{}) {
	r.Completed = make(chan bool)
	r.Response = respChan
	r.handle = NewPkgId()
	//fmt.Println(time.Now().Format("2006-01-02 15:04:05.000000"), "request init", r.handle)
	go r.Wait()
}

// 等待请求的回复
func (r *Request)Wait() {
	outTimeChan := time.Tick(time.Second*30)

	// 等待收到回复或超时
	select {
	case <- outTimeChan : // 超时则结束
		// do nothing
	case <- r.Completed : // 若收到回复则结束
		// do nothing
	}

	requestsMutex.Lock()
	defer requestsMutex.Unlock()

	// 关闭回复通道，删除请求
	close(r.Response)
	delete(requests, r.handle)
}

// 回复请求，设置数据，设置完成
func (r *Request)Answer(data interface{}) {
	r.Response <- data
	r.Completed <- true
}

// 回复请求
func AnswerRequest(handle uint16, data interface{}) {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05.000000"), "request AnswerRequest", handle)
	requestsMutex.Lock()
	defer requestsMutex.Unlock()

	// 获得当前handle对应的请求，如果对应的请求有效，则回复该请求
	r := requests[handle]
	if r != nil {
		r.Answer(data)
	}
}

// 新的网络请求
// respChan 回复请求的数据回复数据通道
// requestHandle 返回值：返回请求的唯一句柄
func NewRequest(respChan chan <- interface{}) (requestHandle uint16) {
	// 创建一个新的请求，并初始化
	r := new(Request)
	r.Init(respChan)
	requestHandle = r.handle
	fmt.Println(time.Now().Format("2006-01-02 15:04:05.000000"), "NewRequest", requestHandle)

	// 将当前新建的互斥保存到map中
	requestsMutex.Lock()
	requests[r.handle] = r
	requestsMutex.Unlock()

	return
}