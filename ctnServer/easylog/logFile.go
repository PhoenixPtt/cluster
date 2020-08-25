package easylog

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"time"
)

type LOG_FILE struct {
	filename string
}

func (logFile *LOG_FILE) Debug(msg string, )  {
	logFile.WriteLog(GetLevelStr(LEVEL_DEBUG),"输出到文件")
}

func (logFile *LOG_FILE) Info(msg string)  {
	logFile.WriteLog(GetLevelStr(LEVEL_INFO),"输出到文件")
}

func (logFile *LOG_FILE) Warning(msg string)  {
	logFile.WriteLog(GetLevelStr(LEVEL_WARNING),"输出到文件")
}

func (logFile *LOG_FILE) Error(msg string)  {
	logFile.WriteLog(GetLevelStr(LEVAL_ERROR),"输出到文件")
}

func (logFile *LOG_FILE) Fatal(msg string)  {
	logFile.WriteLog(GetLevelStr(LEVAL_FATAL),"输出到文件")
}

func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func (logFile *LOG_FILE) WriteLog(logLevel string, msg string)  {
	//打印一个事件字符串
	timeStr := time.Now().Format("2006-01-02 15:04:05.000")//记录日志的时间
	_, logFileName, line, ok:= runtime.Caller(2)
	if !ok{
		return
	}
	logStr:=fmt.Sprintf("%s %s [%s:%d] %s\n",timeStr,logLevel,path.Base(logFileName),line, msg)

	var f *os.File
	var err error
	timeStr = time.Now().Format("2006010215")//记录日志的时间
	logFile.filename = fmt.Sprintf("./logFiles/log_%s",timeStr)
	if checkFileIsExist(logFile.filename) { //如果文件存在
		f, err = os.OpenFile(logFile.filename, os.O_APPEND|os.O_WRONLY, 0666) //打开文件
	} else {
		f, err = os.Create(logFile.filename) //创建文件
	}

	check(err)
	_, err = io.WriteString(f, logStr) //写入文件(字符串)
	check(err)
	f.Close()
}
