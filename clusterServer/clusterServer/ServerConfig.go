package clusterServer

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"clusterHeader"
)

const (
	CfgDir      string = "/tmp/cluster/Server/"
	CfgFileName string = CfgDir + "config.cfg"
)

var d header.ClstCfg // 集群数据

func init() {
	d.Name = "CETC15-基于容器的集群管理平台"

	d.AgentUdpPort = 30000
	d.AgentTcpPort = 40000
	d.ServerUdpPort = 30001
	d.ServerTcpPortForListenClient = 30002
	d.ServerTcpPortForListenAgent = 30003

	d.ResSampleFeq = 1
	d.TaskMigrateTimeFromAgent = 10
}

// 加载名称为name的配置，load之前必须先清空集群
func LoadCfg() (ok bool, err error) {
	if header.PathExists(CfgDir) == false {
		os.MkdirAll(CfgDir, os.ModePerm)
	}

	fileBytes, _ := ioutil.ReadFile(CfgFileName)
	json.Unmarshal([]byte(fileBytes), d)

	Save()

	return
}

func Save() {
	bytes, _ := json.Marshal(d)
	ioutil.WriteFile(CfgFileName, bytes, os.ModePerm)
}
