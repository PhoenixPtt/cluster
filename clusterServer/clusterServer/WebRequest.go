package clusterServer

import (
	"sync"
	"time"
)

type Request struct {
	handle 		uint16
	Response 	chan interface{}
	Completed 	chan bool
}

var requests sync.Map

func (r *Request)Init() {
	r.Completed = make(chan bool)
	r.handle = NewPkgId()
	go r.Wait()
}

func (r *Request)Wait() {
	outTimeChan := time.Tick(time.Second*30)

	select {
	case <- outTimeChan : // 超时则结束
		// do nothing
	case <- r.Completed : // 若收到回复则结束
		// do nothing
	}

	close(r.Response)
	defer requests.Delete(r.handle)
}

func (r *Request)Answer(data interface{}) {
	r.Response <- data
	r.Completed <- true
	close(r.Response)
}

func AnswerRequest(handle uint16, data interface{}) {
	val,ok := requests.Load(handle)
	if !ok {
		return
	}

	r := val.(Request)
	r.Answer(data)
}

func NewRequest() (requestHandle uint16) {
	r := new(Request)
	r.Init()
	requests.Store(r.handle, r)
	requestHandle = r.handle
	return
}