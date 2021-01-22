package protocol

import (
	"ctnCommon/ctn"
	"github.com/docker/docker/api/types/events"
)

type REQ_ANS struct {
	CtnName  string //容器名称
	CtnOper  string //容器操作
	CtnImage string //容器镜像
	CtnErr   error  //具体错误信息
	CtnState string //容器状态

	CtnID      []string
	CtnLog     []string
	CtnInspect []ctn.CTN_INSPECT

	CtnErrType []string //错误类型
}

type CTN_MSG struct {
	CtnName string
	OperNum int
	CtnErr  error

	CtnMsg string
}

//Server端与Agent端通信结构体
type SA_TRUCK struct {
	//基本信息
	Flag    string
	Index   int    //计数
	DesAddr string //目标地址
	SrcAddr string //源地址

	//Server请求
	//Agent响应
	Req_Ans []REQ_ANS

	//Agent上传的容器状态信息
	CtnInfo []ctn.CTN
	EvtMsg  []events.Message
	ErrMsg  []error

	//CtnList []types.Container
	//CtnStat []ctn.CTN_STATS
	//CtnMsg  []CTN_MSG

	//消息发送时间
	MsgTimeStr string
	MsgTime    int64
}
