package cluster

import (
	"ctnServer/easylog"
)

var (
	Mylog easylog.LOG_BEHAVIOR
)

func init() {
	var config map[string]string
	config = make(map[string]string)
	config[easylog.LOG_FILE_NAME] = "./log/logFiles/log"
	Mylog = easylog.SetLogOutput(easylog.LOG_TO_CONSOLE, config)
}
