package ctnS

import (
	"ctnCommon/headers"
	"fmt"
)

var (
	AGENT_TRY_NUM = "失败尝试次数"
)

//新建一个容器对象
func NewCtnS(image string, agentAddr string, configMap map[string]string) (pCtnS *CTNS) {
	pCtnS = &CTNS{}
	pCtnS.CtnName = fmt.Sprintf("CTN_%s", headers.UniqueId())
	pCtnS.AgentAddr = agentAddr
	pCtnS.Image = image
	var agentTryNumStr interface{}
	var ok bool
	if agentTryNumStr, ok = configMap[AGENT_TRY_NUM]; ok {
		pCtnS.AgentTryNum = agentTryNumStr.(int)
	}

	return
}
