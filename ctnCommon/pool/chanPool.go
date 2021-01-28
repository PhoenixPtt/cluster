package pool

import (
	"sync"
)

const (
	MAX_CHAN_NUM = 1000
)

type chanPoolOper interface {
	RegPrivateChanInt(keyInt int)
	UnregPrivateChanInt(keyInt int)
	GetPrivateChanInt(keyInt int) chan interface{}

	RegPrivateChanStr(keyStr string)
	UnregPrivateChanStr(keyStr string)
	GetPrivateChanStr(keyStr string) chan interface{}
}

type CHAN_POOL struct {
	private_chan_int_map    map[int]chan interface{}
	private_chan_string_map map[string]chan interface{}

	intMutex_map map[int]*sync.Mutex
	strMutex     sync.Mutex
}

//注册以整型变量为关键字的私有通道
func (pChanPool *CHAN_POOL) RegPrivateChanInt(keyInt int, bufferSize int) {
	var mutex sync.Mutex
	mutex.Lock()
	defer mutex.Unlock()
	_, ok := pChanPool.private_chan_int_map[keyInt]
	if !ok {
		pChan := make(chan interface{}, bufferSize)
		pChanPool.private_chan_int_map[keyInt] = pChan
	}

	pChanPool.intMutex_map[keyInt] = &mutex
}

func (pChanPool *CHAN_POOL) GetPrivateChanInt(keyInt int) (pChan chan interface{}) {
	pMutex := pChanPool.intMutex_map[keyInt]

	pMutex.Lock()
	defer pMutex.Unlock()
	_, ok := pChanPool.private_chan_int_map[keyInt]
	if ok {
		pChan = pChanPool.private_chan_int_map[keyInt]
		return
	}
	return nil
}

//注销以字符串为关键字的私有通道
func (pChanPool *CHAN_POOL) UnregPrivateChanInt(keyInt int) {
	pMutex := pChanPool.intMutex_map[keyInt]

	_, ok := pChanPool.private_chan_int_map[keyInt]
	if ok {
		pMutex.Lock()
		close(pChanPool.private_chan_int_map[keyInt])
		delete(pChanPool.private_chan_int_map, keyInt)
		pMutex.Unlock()
		delete(pChanPool.intMutex_map, keyInt)
	}

}

//向通道追加元素
func (pChanPool *CHAN_POOL) AppendInt(keyInt int, data interface{}) {
	pMutex := pChanPool.intMutex_map[keyInt]

	pMutex.Lock()
	defer pMutex.Unlock()
	_, ok := pChanPool.private_chan_int_map[keyInt]
	if ok {
		select {
		case pChanPool.private_chan_int_map[keyInt] <- data:
		default:
		}
	}
}

//注册以整型变量为关键字的私有通道
func (pChanPool *CHAN_POOL) RegPrivateChanStr(keyStr string, bufferSize int) {
	pChanPool.strMutex.Lock()
	defer pChanPool.strMutex.Unlock()
	_, ok := pChanPool.private_chan_string_map[keyStr]
	if !ok {
		pChan := make(chan interface{}, bufferSize)
		pChanPool.private_chan_string_map[keyStr] = pChan
	}
}

func (pChanPool *CHAN_POOL) GetPrivateChanStr(keyStr string) (pChan chan interface{}) {
	pChanPool.strMutex.Lock()
	defer pChanPool.strMutex.Unlock()
	_, ok := pChanPool.private_chan_string_map[keyStr]
	if ok {
		pChan = pChanPool.private_chan_string_map[keyStr]
		return
	}
	return nil
}

func (pChanPool *CHAN_POOL) UnregPrivateChanStr(keyStr string) {
	pChanPool.strMutex.Lock()
	defer pChanPool.strMutex.Unlock()
	_, ok := pChanPool.private_chan_string_map[keyStr]
	if ok {
		close(pChanPool.private_chan_string_map[keyStr])
		delete(pChanPool.private_chan_string_map, keyStr)
	}
}

//向通道追加元素
func (pChanPool *CHAN_POOL) AppendStr(keyStr string, data interface{}) {
	pChanPool.strMutex.Lock()
	defer pChanPool.strMutex.Unlock()
	_, ok := pChanPool.private_chan_string_map[keyStr]
	if ok {
		select {
		case pChanPool.private_chan_string_map[keyStr] <- data:
		default:
		}
	}
}

var (
	pChanPool *CHAN_POOL
)

func init() {
	pChanPool = NewChanPool()
}

func NewChanPool() (pChanPool *CHAN_POOL) {
	pChanPool = &CHAN_POOL{}
	pChanPool.private_chan_int_map = make(map[int]chan interface{}, MAX_CHAN_NUM)
	pChanPool.private_chan_string_map = make(map[string]chan interface{}, MAX_CHAN_NUM)
	pChanPool.intMutex_map = make(map[int]*sync.Mutex, MAX_CHAN_NUM)
	return
}

//注册以整型变量为关键字的私有通道
func RegPrivateChanInt(keyInt int, bufferSize int) {
	pChanPool.RegPrivateChanInt(keyInt, bufferSize)
}

func GetPrivateChanInt(keyInt int) (pChan chan interface{}) {
	return pChanPool.GetPrivateChanInt(keyInt)
}

//注销以字符串为关键字的私有通道
func UnregPrivateChanInt(keyInt int) {
	pChanPool.UnregPrivateChanInt(keyInt)
}

func AppendInt(keyInt int, data interface{}) {
	pChanPool.AppendInt(keyInt, data)
}

//注册以整型变量为关键字的私有通道
func RegPrivateChanStr(keyStr string, bufferSize int) {
	pChanPool.RegPrivateChanStr(keyStr, bufferSize)
}

func GetPrivateChanStr(keyStr string) (pChan chan interface{}) {
	return pChanPool.GetPrivateChanStr(keyStr)
}

func UnregPrivateChanStr(keyStr string) {
	pChanPool.UnregPrivateChanStr(keyStr)
}

func AppendStr(keyStr string, data interface{}) {
	pChanPool.AppendStr(keyStr, data)
}
