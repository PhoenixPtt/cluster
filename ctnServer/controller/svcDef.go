package controller

const(
	CTN_SIZE = 1000
)

//服务通道结构体
type SVC_OPER struct {
	SOperName string	//操作名称
	Scale int			//如果操作为扩容或者缩绒的话，该参数为服务的规模
}

//服务结构体
type SERVICE struct {
	SvcName string         		`yaml: "svc_name"`		//服务名称
	Image string         		`yaml: "image"`			//服务指定的镜像
	SvcStats string										//服务状态，
	SvcScale int            	`yaml: "svc_scale"`		//服务的预期规模
	SvcHealthDegree float64								//服务健康度

	Replicas []*REPLICA									//服务的副本
	NodeStatusMap map[string]bool 						//节点状态映射表
	//SchedulePOLICY				//服务的调度策略
}

//以结构体作为配置参数创建服务对象
func NewService(pSvcCfg *SVC_CFG) (pSvc *SERVICE) {
	pSvc = &SERVICE{}
	pSvc.SvcName = pSvcCfg.Description.Name
	pSvc.Image = pSvcCfg.Description.Image
	pSvc.SvcScale = pSvcCfg.Description.Deploy.Replicas
	pSvc.Replicas = make([]*REPLICA, 0, CTN_SIZE)
	pSvc.NodeStatusMap = make(map[string]bool, CTN_SIZE)
	pSvc.SvcStats = SVC_DEFAULT
	return
}
//以文件作为配置参数创建服务对象
func NewServiceFromFile(fileName string, fileType string) (pSvc *SERVICE) {
	var svcCfg *SVC_CFG
	switch fileType {
	case YML_FILE:
		svcCfg = YmlFile2Struct(fileName)
	case JSON_FILE:
		svcCfg = JsonFile2Struct(fileName)
	}
	pSvc=NewService(svcCfg)
	return
}

func (pSvc *SERVICE) GetStats() string {
	return pSvc.SvcStats
}








