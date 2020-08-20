package easylog

import (
	"fmt"
	"path"
	"runtime"
	"time"
)

type LOG_CONSOLE struct {
	//file os.File
}

func (logConsole *LOG_CONSOLE) Debug(msg string)  {
	logConsole.WriteLog(GetLevelStr(LEVEL_DEBUG),msg)
}

func (logConsole *LOG_CONSOLE) Info(msg string)  {
	logConsole.WriteLog(GetLevelStr(LEVEL_INFO),msg)
}

func (logConsole *LOG_CONSOLE) Warning(msg string)  {
	logConsole.WriteLog(GetLevelStr(LEVEL_WARNING),msg)
}

func (logConsole *LOG_CONSOLE) Error(msg string)  {
	logConsole.WriteLog(GetLevelStr(LEVAL_ERROR),msg)
}

func (logConsole *LOG_CONSOLE) Fatal(msg string)  {
	logConsole.WriteLog(GetLevelStr(LEVAL_FATAL),msg)
}

func (logConsole *LOG_CONSOLE) WriteLog(logLevel string, msg string)  {
	//打印一个事件字符串
	timeStr := time.Now().Format("2006-01-02 15:04:05.000")//记录日志的时间
	_, logFileName, line, ok:= runtime.Caller(2)
	if !ok{
		return
	}
	logStr:=fmt.Sprintf("%s %s [%s:%d] %s",timeStr,logLevel,path.Base(logFileName),line, msg)
	fmt.Println(logStr)
}
