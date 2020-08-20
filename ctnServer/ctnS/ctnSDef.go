package ctnS

import (
	"ctnCommon/headers"
	"fmt"
)

//新建一个容器对象
func NewCtnS(image string, agentAddr string, configMap map[string]string) (pCtnS *CTNS) {
	pCtnS = &CTNS{}
	pCtnS.CtnName = fmt.Sprintf("CTN_%s",headers.UniqueId())
	pCtnS.AgentAddr = agentAddr
	pCtnS.Image = image
	configMap = make(map[string]string, 100)
	return
}

