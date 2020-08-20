package easylog

const (
	LOG_TO_CONSOLE = iota
	LOG_TO_FILE

	LEVEL_DEBUG = iota
	LEVEL_INFO
	LEVEL_WARNING
	LEVAL_ERROR
	LEVAL_FATAL

	LOG_FILE_NAME = "filename"
)

var (
	logConsole *LOG_CONSOLE
	logFile    *LOG_FILE
)

func init() {
	//初始化结构体
	logConsole = &LOG_CONSOLE{}
	logFile = &LOG_FILE{}
}

func config(configMap map[string]string) {
	filename, ok := configMap[LOG_FILE_NAME]
	if !ok {
		return
	}
	logFile.filename = filename
	return
}

func SetLogOutput(output int, config map[string]string) LOG_BEHAVIOR {
	switch output {
	case LOG_TO_CONSOLE:
		return logConsole
	case LOG_TO_FILE:
		return logFile
	}
	return logConsole
}

func GetLevelStr(level int) (levelStr string) {
	switch level {
	case LEVEL_DEBUG:
		return "DEBUG"
	case LEVEL_INFO:
		return "INFO"
	case LEVEL_WARNING:
		return "WARNING"
	case LEVAL_ERROR:
		return "ERROR"
	case LEVAL_FATAL:
		return "FATAL"
	}
	return "DEBUG"
}

//
//func Debug()  {
//
//}
//
//func Info()  {
//
//}
//func Warning()  {
//
//}
//func Error()  {
//
//}
//func Fatal()  {
//
//}
