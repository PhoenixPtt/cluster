package controller

import (
	header "clusterHeader"
	"ctnCommon/ctn"
	"github.com/docker/docker/api/types"

	//"github.com/docker/docker/api/types"
)

//数据格式适配器
func Oper2ServiceOper(webSvcCfg *header.ServiceCfg) (svcCfg *SVC_CFG) {
	svcCfg = &SVC_CFG{}
	svcCfg.Version = webSvcCfg.Version
	svcCfg.Description.Name = webSvcCfg.Name
	svcCfg.Description.Image = webSvcCfg.Image
	//svcCfg.Description.Cmd = webSvcCfg.Cmd
	//svcCfg.Description.CmdPars = webSvcCfg.CmdPars
	//svcCfg.Description.EntryPoint = webSvcCfg.EntryPoint
	//svcCfg.Description.EntryPointPars = webSvcCfg.EntryPointPars
	svcCfg.Description.Deploy.Mode = webSvcCfg.Deploy.Mode
	svcCfg.Description.Deploy.Replicas = int(webSvcCfg.Deploy.Replicas)
	svcCfg.Description.Deploy.Placement.Constraints = webSvcCfg.Deploy.Placement.Constraints
	//svcCfg.Description.Deploy.RcWeight = &webSvcCfg.Deploy.RcWeight
	//svcCfg.Description.Deploy.Resources. = webSvcCfg.Deploy.Resources
	return
}

func WebService2ServiceOperTruck(pWebSvc *header.SERVICE) (pSvcOperTruck *SERVICE_OPER_TRUCK) {
	pSvcOperTruck = &SERVICE_OPER_TRUCK{}
	pSvcCfg := Oper2ServiceOper(&pWebSvc.Service[0].Cfg)
	pSvcOperTruck.OperType = pWebSvc.Type
	switch pWebSvc.Par[0].Name {
	case "规模":
		//pSvcOperTruck.ScaleNum = pWebSvc.Par[0].Value
	}
	pSvcOperTruck.SvcName = pSvcCfg.Description.Name
	pSvcOperTruck.SvcCfg = *pSvcCfg
	return
}

func ToWebService(pController *CONTROLLER, ctnInfoMap map[string]types.Container, ctnStatMap map[string]ctn.CTN_STATS) (pWebServices *header.SERVICE)  {
	currController:=*pController

	pWebServices = &header.SERVICE{}

	//转服务
	var webSvcs header.Services
	webSvcs.Service = make([]header.Service,0,SVC_NUM)
	for _, pSvc:=range currController.ServiceMap{
		pWebSvc := &header.Service{}
		//服务的基本信息
		pWebSvc.Id = pSvc.SvcName// 服务Id
		//服务状态
		pWebSvc.Scale = uint32(pSvc.SvcScale) // 设定的副本数量
		pWebSvc.ReplicaCount = uint32(len(pSvc.Replicas)) // 应用服务的当前副本数量
		pWebSvc.CreateTime = pSvc.CreateTime // 服务创建时间
		pWebSvc.StartTime = pSvc.StartTime // 服务启动时间
		pWebSvc.NameSpace = pSvc.NameSpace //服务的命名空间
		//服务配置信息,空缺

		//服务的所有副本
		for _,pRpl:=range pSvc.Replicas{
			pWebRpl := &header.Replica{}
			pWebRpl.Id = pRpl.RplName
			pWebRpl.CreateTime = pRpl.CreateTime
			//pCtn:=ctnS.GetCtn(pRpl.CtnName)
			//if pCtn!=nil{
			//	pWebRpl.Ctn = ctnInfoMap[pCtn.ID]
			//	pWebRpl.CtnStats = ctnStatMap[pCtn.ID]
			//}
			pWebSvc.Replica = append(pWebSvc.Replica, *pWebRpl)
		}

		webSvcs.Service = append(webSvcs.Service, *pWebSvc)
	}
	webSvcs.Count = uint32(len(webSvcs.Service))

	pWebServices.Services = webSvcs

	//fmt.Println("888", pWebServices)

	return
}
