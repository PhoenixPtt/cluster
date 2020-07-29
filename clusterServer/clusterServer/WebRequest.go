package clusterServer

import (
	"sync"
	"time"
)

type Request struct {
	handle 		uint16
	Response 	chan<- interface{}
	Completed 	chan bool
}

var requests sync.Map

func (r *Request)Init(respChan chan <- interface{}) {
	r.Completed = make(chan bool)
	r.Response = respChan
	r.handle = NewPkgId()
	//fmt.Println(time.Now().Format("2006-01-02 15:04:05.000000"), "request init", r.handle)
	go r.Wait()
}

func (r *Request)Wait() {
	//fmt.Println(time.Now().Format("2006-01-02 15:04:05.000000"), "request Wait", r.handle)
	outTimeChan := time.Tick(time.Second*3)

	select {
	case <- outTimeChan : // 超时则结束
		//fmt.Println(time.Now().Format("2006-01-02 15:04:05.000000"), "request Wait outTimeChan", r.handle)
	// do nothing
	case <- r.Completed : // 若收到回复则结束
		//fmt.Println(time.Now().Format("2006-01-02 15:04:05.000000"), "request Wait Completed", r.handle)
		// do nothing
	}

	//fmt.Println(time.Now().Format("2006-01-02 15:04:05.000000"), "request End", r.handle, r.Response)
	close(r.Response)
	defer requests.Delete(r.handle)
}

func (r *Request)Answer(data interface{}) {
	r.Response <- data
	r.Completed <- true
}

func AnswerRequest(handle uint16, data interface{}) {
	//fmt.Println(time.Now().Format("2006-01-02 15:04:05.000000"), "request AnswerRequest", handle)
	val,ok := requests.Load(handle)
	if !ok {
		return
	}

	r := val.(*Request)
	r.Answer(data)
}

func NewRequest(respChan chan <- interface{}) (requestHandle uint16) {
	//fmt.Println(time.Now().Format("2006-01-02 15:04:05.000000"), "request NewRequest")
	r := new(Request)
	r.Init(respChan)
	requests.Store(r.handle, r)
	requestHandle = r.handle
	return
}