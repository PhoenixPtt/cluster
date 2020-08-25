package ctnA

import (
	"ctnCommon/ctn"
	"time"
)

func Operate(pCtn *CTNA, operType string) (errType string,err error) {
	switch operType {
	case ctn.CREATE:
		errType,err=pCtn.Create()
	case ctn.START:
		errType,err=pCtn.Start()
	case ctn.RUN:
		errType,err=pCtn.Run()
	case ctn.STOP:
		errType,err=pCtn.Stop()
	case ctn.KILL:
		errType,err=pCtn.Kill()
	case ctn.REMOVE:
		errType,err=pCtn.Remove()
	}
	return
}

func OperateWithStratgy(pCtn *CTNA, operType string) (errType string,err error) {
	if pCtn.OperStrategy{
		//第一阶段：操作一次
		errType,err=Operate(pCtn,operType)
		if err==nil{
			return
		}else{
			//向Server端发送状态消息
		}

		//第二阶段：操作若干次
		for i:=0; i<pCtn.OperNum; i++{
			errType,err=Operate(pCtn,operType)
			if err==nil{
				return
			}else{
				//向Server端发送状态消息
			}
		}

		//第三阶段：删除
		for{//操作不成功，删除容器
			errType,err=Operate(pCtn,ctn.REMOVE)
			if err==nil{
				return
			}
			time.Sleep(time.Second)
		}
	}else{
		errType,err=Operate(pCtn,operType)
		return
	}
}
