package method

import (
	header "clusterHeader"
	"clusterServer/ctn"
	"fmt"
)

const (
	CTN_SIZE             = 50
	NO_SPECIFIED_SERVICE = "NO_SPECIFIED_SERVICE"
)

const (
	ALGO1 = iota
)

//各项资源权重
type WEIGHT struct {
	Node_Weight     float64 //硬盘权重
	CPU_Weight      float64 //CPU权重
	Mem_Weight      float64 //内存权重
	Harddisk_Weight float64 //硬盘权重
}

//预估的服务资源使用量
type RC_USAGE struct {
	CPU_Usage      float64 //CPU使用量
	Mem_Usage      float64 //内存使用量
	Harddisk_Usage float64 //硬盘使用量
}

//服务调度策略
type SchedulePOLICY struct {
	PolicyType int      //策略类型：     0:指定节点；1：根据资源使用情况动态分配
	NodeName   []string //策略类型0：  罗列节点名称列表
	Algorithm  int      //策略类型1： 算法类型
	//如果算法类型是ALGO1:
	WEIGHT   //记录权重信息
	RC_USAGE //记录服务预估资源使用量
}

type SERVICE struct {
	ServiceName    string         //服务名称
	Image          string         //服务的镜像
	Scale          int            //服务的规模
	Stats          string         //服务的状态
	SchedulePOLICY                //服务的调度策略
}

//创建服务
func (service *SERVICE) Create(AgentAddrs []string) {
	//////////////////////////////待实现//////////////////////////////
	//确定服务的副本在那几个agent上启动,每个agent启动的副本数量
	var agentAddrNumMap map[string]int = make(map[string]int, 50)
	for _, addr := range AgentAddrs {
		fmt.Printf("请输入IP地址为%s的agent上部署的副本数量:\n", addr)
		var num int
		fmt.Scanln(&num)
		agentAddrNumMap[addr] = num
		for i := 0; i < num; i++ {
			var cctn header.CTN
			cctn.AgentAddr = addr
			cctn.Image = service.Image
			cctn.ServiceName = service.ServiceName
			cctn.PrepareData(ctn.CREATE)
			ctn.AddCtn(cctn)
		}
	}
}

//启动服务
func (service *SERVICE) Start() {
	if service.ServiceName == "" {
		fmt.Sprintln("该服务无服务名！")
	}

	ctns := ctn.GetCtnsInService(service.ServiceName)

	if len(ctns) == 0 {
		fmt.Sprintln("该服务不存在符合条件的容器，无法启动服务！")
	}
	for _, cctn := range ctns {
		cctn.PrepareData(ctn.START)
	}
}

//停止服务
func (service *SERVICE) Stop() {
	if service.ServiceName == "" {
		fmt.Sprintln("该服务无服务名！")
	}

	ctns := ctn.GetCtnsInService(service.ServiceName)
	if len(ctns) == 0 {
		fmt.Sprintln("该服务不存在符合条件的容器，无法停止服务！")
	}
	for _, cctn := range ctns {
		cctn.PrepareData(ctn.STOP)
	}
}

//删除服务
func (service *SERVICE) Remove() {
	if service.ServiceName == "" {
		fmt.Sprintln("该服务无服务名！")
	}

	ctns := ctn.GetCtnsInService(service.ServiceName)
	if len(ctns) == 0 {
		fmt.Sprintln("该服务不存在符合条件的容器，无法删除服务！")
	}
	for _, cctn := range ctns {
		cctn.PrepareData(ctn.REMOVE)
	}
}

// //删除服务
// func (service *SERVICE) RemoveCtn(ctnId string) {
// 	if service.ServiceName == "" {
// 		fmt.Sprintln("该服务无服务名！")
// 	}

// 	ctnLen := len(service.CtnMap)
// 	if ctnLen == 0 {
// 		fmt.Sprintln("该服务不存在符合条件的容器，无法启动服务！")
// 	}

// 	_, ok := service.CtnMap[ctnId]
// 	if ok {
// 		service.CtnMap[ctnId].SRemove()
// 		delete(service.CtnMap, ctnId)
// 	}
// }

//调度的内容

