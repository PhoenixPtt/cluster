package clusterServer

import (
	"clusterServer/ctn"
	"clusterServer/method"
	"fmt"
	"time"
)

type TaskInfo struct {
	Name   string     `json:"name"`   // 应用服务名称
	Count  uint32     `json:"count"`  // 副本数量
	Config TaskConfig `json:"config"` // 应用服务的配置信息
}

type TaskConfig struct {
	EntryPoint string `json:"entry_point"` // 入口参数
	Image      string `json:"image"`       // 镜像信息
}

// 处理Agent丢失事宜 迁移所有的应用
func ProcessAgentLostEvent(handle string) {
	time.Sleep(time.Second * time.Duration(d.TaskMigrateTimeFromAgent))

	// 如果主机依然是丢失，则立即迁移所有服务
	//index, ok := NodeIds[handle]
	//if !ok {
	//	return
	//}

	// 如果主机网络已经恢复，则立即结束
	//node := &d.Node[index]
	//if node.State == 1 {
	//	return
	//}

	// 获得主机的所有服务

}

func ProcessTaskFlagDataFromClient(ip string, pkgId uint16, data []byte) {
	//
	//switch index {
	//case "LIST": //获取服务列表
	//	ListService()
	//case 1: //创建服务
	//	CreateService()
	//case 2: //启动服务
	//	StartService()
	//case 3: //停止服务
	//	StopService()
	//case 4: //删除服务
	//	RemoveService()
	//}

}

// 列出所有的服务
func ListService() (list string) {
	sMap := cluster.GetServiceMap()
	for index, val := range sMap {
		list = fmt.Sprintf("%s,%p", index, val)
	}
	return
}

// 批量创建服务
func CreateServices(names []string, imageNames []string) []bool {
	// 如果参数错误，直接返回false
	if len(names) == 0 {
		return []bool{false}
	}

	// 依次启动所有的服务
	var ok []bool = make([]bool, len(names))
	for index, name := range names {
		ok[index] = CreateService(name, imageNames[index])
	}

	return ok
}

// 创建服务
func CreateService(serviceName string, imageName string) bool {
	fmt.Println("创建服务:")

	fmt.Println("请输入服务名称:")
	fmt.Scanf("%s", &serviceName)
	// fmt.Println("请输入镜像名称:")
	// fmt.Scanf("%s", &imageName)
	// fmt.Println(serviceName, imageName)
	//imageName = "oldservertest:1.0"
	imageName = "servertest:1.0"

	//检查该服务是否存在
	//如果该服务存在,检查该服务的规模
	//检查属于该服务的全部容器，判断这些容器中属于“已创建”状态的个数
	//如果"已创建"状态的容器个数正好等于服务的规模，则该服务已经被成功创建，则不允许重复创建，直接返回
	//如果"已创建"状态的容器个数小于服务的规模，则该服务有部分容器创建失败或者还没来得及创建，则需新创建的容器的个数：
	//容器规模-“已创建”状态的容器个数
	//如果“已创建”状态的容器个数大于服务的规模，则需要根据一定的策略，删除多余的容器
	//否则，创建全新的服务

	var service method.SERVICE
	service.ServiceName = serviceName
	service.Image = imageName
	//遍历节点信息得到可用的ip地址
	var nodeAddrs []string
	for nodeAddr, _ := range cluster.AgentMap {
		nodeAddrs = append(nodeAddrs, nodeAddr)
	}
	service.Create(nodeAddrs)
	cluster.AddService(&service)
	for _, val := range ctn.GetCtnsInService(service.ServiceName) {
		send(val.AgentAddr, *val)
	}
	return true
}

// 批量启动服务
func StartServices(names []string) []bool {
	// 如果参数错误，直接返回false
	if len(names) == 0 {
		return []bool{false}
	}

	// 依次启动所有的服务
	var ok []bool = make([]bool, len(names))
	for index, name := range names {
		ok[index] = StartService(name)
	}

	return ok
}

// 启动服务
func StartService(name string) bool {
	//检查服务“×××”是否存在
	//如果服务“×××”已存在
	//检查处于“启动状态”的服务数量，如果“启动状态”的服务数量==服务的规模
	//则服务已经被正确启动，无需任何操作，直接退出
	//如果“启动状态”的服务数量<服务的规模
	//启动的容器数量=服务的规模-“启动状态”的服务数量，启动相应数量的容器
	//否则
	//退出


	sMap := cluster.GetServiceMap()
	fmt.Println("请输入要启动的服务：")
	var sNames []string
	for index, _ := range sMap {
		sNames = append(sNames, index)
	}
	for index, val := range sNames {
		fmt.Printf("%d.%s\n", index, val)
	}
	var sIndex int
	fmt.Scanln(&sIndex)
	if sIndex>=len(sNames){
		return  false
	}
	rlt := cluster.IsExisted(sNames[sIndex])
	if rlt {
		service := cluster.GetService(sNames[sIndex])
		service.Start()
		for _, val := range ctn.GetCtnsInService(service.ServiceName) {
			send(val.AgentAddr, *val)
		}
	} else {
		fmt.Println("服务%s不存在！", sNames[sIndex])
	}

	return true
}

// 批量停止服务
func StopServices(names []string) []bool {
	// 如果参数错误，直接返回false
	if len(names) == 0 {
		return []bool{false}
	}

	// 依次启动所有的服务
	var ok []bool = make([]bool, len(names))
	for index, name := range names {
		ok[index] = StopService(name)
	}

	return ok
}

// 停止服务
func StopService(name string) bool {
	//检查服务“×××”是否已经存在，如果已存在
	//遍历服务“×××”的所有容器，如果存在容器的状态为“运行”，则停止这些容器
	//否则，退出
	//否则，服务都不存在，自然是退出喽
	sMap := cluster.GetServiceMap()
	fmt.Println("请输入要停止的服务：")
	var sNames []string
	for index, _ := range sMap {
		sNames = append(sNames, index)
	}
	for index, val := range sNames {
		fmt.Printf("%d.%s\n", index, val)
	}

	var sIndex int
	fmt.Scanln(&sIndex)
	if sIndex>=len(sNames){
		return false
	}
	rlt := cluster.IsExisted(sNames[sIndex])
	if rlt {
		service := cluster.GetService(sNames[sIndex])
		service.Stop()
		for _, val := range ctn.GetCtnsInService(service.ServiceName) {
			send(val.AgentAddr, *val)
		}
	} else {
		fmt.Println("服务%s不存在！", sNames[sIndex])
	}
	return true
}

// 批量删除服务
func RemoveServices(names []string) []bool {
	// 如果参数错误，直接返回false
	if len(names) == 0 {
		return []bool{false}
	}

	// 依次启动所有的服务
	var ok []bool = make([]bool, len(names))
	for index, name := range names {
		ok[index] = RemoveService(name)
	}

	return ok
}

// 删除服务
func RemoveService(name string) bool {
	//检查服务“×××”是否存在，如果已存在
	//遍历服务的所有容器，删除掉
	//删除服务信息
	//否则，服务都不存在，也不存在删除一说咯
	sMap := cluster.GetServiceMap()
	fmt.Println("请输入要删除的服务：")
	sNames := make([]string, 0, 50)
	for index, _ := range sMap {
		sNames = append(sNames, index)
	}
	for index, val := range sNames {
		fmt.Printf("%d.%s\n", index, val)
	}
	var sIndex int
	fmt.Scanln(&sIndex)
	if sIndex>=len(sNames){
		return false
	}
	rlt := cluster.IsExisted(sNames[sIndex])
	if rlt {
		service := cluster.GetService(sNames[sIndex])
		service.Remove()
		for _, val := range ctn.GetCtnsInService(service.ServiceName) {
			send(val.AgentAddr, *val)
		}
	} else {
		fmt.Println("服务%s不存在！", sNames[sIndex])
	}
	return true
}
