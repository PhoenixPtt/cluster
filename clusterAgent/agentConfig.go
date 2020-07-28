package main

import (
	"clusterHeader"
	"encoding/json"
	"io/ioutil"
	"os"
)

const (
	CfgDir      string = "/tmp/cluster/agent/"
	CfgFileName string = CfgDir + "config.cfg"
)

var d header.ClstCfg // 集群数据

func init() {
	d.AgentUdpPort = 30000
	d.ServerUdpPort = 30001
	d.ServerTcpPortForListenAgent = 30003

	d.ResSampleFeq = 1
}

func init()  {
	if header.PathExists(CfgDir) == false {
		os.MkdirAll(CfgDir, os.ModePerm)
	}

	fileBytes, _ := ioutil.ReadFile(CfgFileName)
	json.Unmarshal([]byte(fileBytes), d)

	Save()
}

func Save() {
	bytes, _ := json.Marshal(d)
	ioutil.WriteFile(CfgFileName, bytes, os.ModePerm)
}
