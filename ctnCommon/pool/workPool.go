package pool

const(
	WORK_CHAN_NUM = 10000
)

//工作池
//可以向工作池申请一个通道，用来发送数据
//从Agent收来的数据也可以先放入工作池，慢慢处理
type SendObjFunc func (ip string, level uint8, pkgId uint16, flag string, data []byte)
//通知主线程发送网络数据
func CallbackSendCtn(ip string, level uint8, pkgId uint16, flag string, data []byte, sendObjFunc SendObjFunc)  {
	sendObjFunc(ip, level, pkgId, flag, data)
}

type WORK_POOL struct {
	SendTruck chan interface{}
	RecvTruck chan interface{}
	pSendObjFunc SendObjFunc
}

type WorkPoolOper interface {
	Config(sendObjFunc SendObjFunc)
	GetSendChan() chan interface{}
	GetRecvChan() chan interface{}
	Send()
	Recv()
}

func (pWorkPool *WORK_POOL) Config(sendObjFunc SendObjFunc)  {
	pWorkPool.pSendObjFunc = sendObjFunc
}

//获取网口发送通道
func (workPool *WORK_POOL) GetSendChan() chan interface{} {
	return workPool.SendTruck
}

//获取网口接收通道
func (workPool *WORK_POOL) GetRecvChan() chan interface{} {
	return workPool.RecvTruck
}

//获取回调函数
func (workPool *WORK_POOL) GetSendFunc() SendObjFunc {
	return workPool.pSendObjFunc
}

func NewWorkPool() (pWorkPool *WORK_POOL) {
	pWorkPool = &WORK_POOL{}
	pWorkPool.SendTruck = make(chan interface{},WORK_CHAN_NUM)
	pWorkPool.RecvTruck = make(chan interface{},WORK_CHAN_NUM)
	return pWorkPool
}






