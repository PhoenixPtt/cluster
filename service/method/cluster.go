package method

import (
	"fmt"
	"sync"
)

type CLUSTER struct {
	ServiceMap map[string]*SERVICE //集群内部的所有服务
	AgentMap   map[string]*NODE    //集群的节点信息
}

var (
	svsMutex sync.Mutex
)

const DEFAULT_MAP_SIZE = 100

//判断服务是否在集群中已经存在
func (cluster *CLUSTER) IsExisted(serviceName string) bool {
	_, ok := cluster.ServiceMap[serviceName]
	if ok {
		return true
	}
	return false
}

//添加服务至集群
func (cluster *CLUSTER) AddService(service *SERVICE) string {
	svsMutex.Lock()
	defer svsMutex.Unlock()
	serviceNum := len(cluster.ServiceMap)

	if serviceNum == 0 {
		cluster.ServiceMap = make(map[string]*SERVICE, DEFAULT_MAP_SIZE) //为集群中的服务变量分配内存空间
	}

	//判断服务是否已经存在
	if cluster.IsExisted(service.ServiceName) {
		return fmt.Sprintln("服务%s已存在，不能重复创建！", service.ServiceName)
	}

	//添加服务至集群
	cluster.ServiceMap[service.ServiceName] = service

	return ""
}

//从集群中删除服务
func (cluster *CLUSTER) RemoveService(sName string) string {
	//mutex.Lock()
	//defer mutex.Unlock()
	//判断服务是否已经存在
	if !cluster.IsExisted(sName) {
		return fmt.Sprintf("服务%s不存在，不能删除！", sName)
	}

	delete(cluster.ServiceMap, sName)
	return ""
}

//获取集群中所有的服务
func (cluster *CLUSTER) GetServiceMap() map[string]*SERVICE {
	return cluster.ServiceMap
}

//从集群获取指定的服务
func (cluster *CLUSTER) GetService(sName string) *SERVICE {
	_, ok := cluster.ServiceMap[sName]
	if ok {
		return cluster.ServiceMap[sName]
	}
	return nil
}

//获取agent地址映射表
func (cluster *CLUSTER) GetAgentMap() map[string]*NODE {
	return cluster.AgentMap
}

//判断agent是否存在
func (cluster *CLUSTER) IsAgentExist(addr string) bool {
	_, ok := cluster.AgentMap[addr]
	if ok {
		return true
	}
	return false
}

//添加节点
func (cluster *CLUSTER) AddAgent(agent *NODE) string {
	if cluster.IsAgentExist(agent.Addr) {
		err := fmt.Sprintf("节点%s已存在，不能重复添加!", agent.Addr)
		return err
	} else {
		aMapLen := len(cluster.AgentMap)
		if aMapLen == 0 { //如果映射表为空需要为其分配内存空间
			cluster.AgentMap = make(map[string]*NODE, DEFAULT_MAP_SIZE)
		}
		cluster.AgentMap[agent.Addr] = agent
	}
	return ""
}

//删除节点
func (cluster *CLUSTER) RemoveAgent(addr string) string {
	if !cluster.IsAgentExist(addr) {
		err := fmt.Sprintf("节点%s不存在，不能删除!", addr)
		return err
	} else {
		delete(cluster.AgentMap, addr)
	}
	return ""
}

//更新节点状态
func (cluster *CLUSTER) UpdateAgent(addr string, status int) string {
	if cluster.IsAgentExist(addr) {
		switch status {
		case 1:
			fmt.Printf("客户端连接成功：\nIP地址：%s\n", addr)
		case 2:
			fmt.Printf("客户端连接断开：\nIP地址：%s\n", addr)
		}
	}
	return ""
}
