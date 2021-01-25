package controller

import (
	header "clusterHeader"
	"ctnCommon/pool"
	"fmt"
	//"github.com/docker/docker/api/types"
)

const UPLOAD = "UPLOAD"

func WaitWebService() (pWebServices *header.SERVICE) {
	select {
	case obj := <-pool.GetPrivateChanStr(UPLOAD): //类型：header.SERVICE
		pWebServices = obj.(*header.SERVICE)
		fmt.Println("2222222222222222222222222222", pWebServices)
	}
	return
}

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
