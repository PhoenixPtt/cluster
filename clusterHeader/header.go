package header

const (
	FLAG_FILE   = "FILE"	// 文件
	FLAG_OTHR	= "OTHR"	// 其它
	FLAG_CMSG	= "CMSG"	// 消息
	FLAG_EVTM	= "EVTM"	// 事件


	//FLAG_RFRE   = "RFRE" 	// 资源监控频率
	//FLAG_HOST   = "HOST" 	// 主机信息
	//FLAG_RDAT   = "RDAT" 	// 资源监控数据
	//FLAG_CTRL   = "CTRL"
	//FLAG_CTN    = "INFO"
	//FLAG_STATS  = "STAT"
	//FLAG_EVENT  = "EVTM"
)

const (
	CREATE  = "CREATE"
	START   = "START"
	STOP    = "STOP"
	KILL    = "KILL"
	REMOVE  = "REMOVE"
	GETLOG  = "GETLOG"
	INSPECT = "INSPECT"
	CTNEXIT = "EXIT"
)


type Oper struct {
	Type  		string 		// 操作类型
	Par			[]OperPar	// 操作参数
	Index 		uint32    	// 操作序号
	Progress	uint8		// 操作进度 >=100 - 完成 其它-未完成
	Success		bool		// 操作成功
	Err       	string 		// 操作结果
}

type OperPar struct {
	Name		string
	Value		string
}



