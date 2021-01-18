package clusterServer

import (
	header "clusterHeader"
	"context"
	"ctnCommon/ctn"
	"ctnServer/controller"
	"ctnServer/ctnS"
	"errors"
	"fmt"
	"time"
)

// web 响应网络数据,非阻塞
// token 	身份认证标识、含用户、主机等信息
// data		接收到的数据包，类型根据flag指定
// respChan	回复当前消息的数据通道，填充数据后按需关闭，且只能填充一次，不可读取
func procFlagSERV(token interface{}, data interface{}, respChan chan<- interface{}) (err error) {
	// 对应的数据结构为SERVICE
	r := data.(header.SERVICE)
	r.Oper.Success = true		// 默认操作成功
	r.Oper.Progress = 100		// 默认操作完成
	r.Oper.Err = ""				// 默认无错误


	// 根据子标识类型，执行相关操作
	switch r.Oper.Type {

	case header.FLAG_SERV_LIST :   // 获取服务列表
		sNames := g_controller.GetSvcNames()
		r.Count = uint32(len(sNames))
		r.Service = make([]header.Service, r.Count)
		for  i:=uint32(0); i<r.Count; i++ {
			r.Service[i].Id = sNames[i];
		}
	case header.FLAG_SERV_STATS: // 服务转台
		r = *(g_controller.WaitWebService())
		r.Oper.Success = true		// 默认操作成功
		r.Oper.Progress = 100		// 默认操作完成
		r.Oper.Err = ""				// 默认无错误
	case header.FLAG_SERV_CTRL: // 服务转台
		pSvcOperaTruck :=  controller.WebService2ServiceOperTruck(&r)
		g_controller.PutService(pSvcOperaTruck)

	//
	//default:						// 其它操作处理
	//	r.Oper.Success = false					// 操作失败
	//	r.Oper.Err = "子标识错误！"				// 操作失败信息
	}

	// 写入数据到通道
	respChan <- r
	close(respChan)

	// 返回错误信息
	return errors.New(r.Oper.Err)
}

func createContainer(imageName string, h string ) {
	fmt.Printf("创建容器 %s\n", imageName)
	var config map[string]string
	pCtn := ctnS.NewCtnS(imageName, h, config)
	var ctx context.Context
	var cancel context.CancelFunc
	ctx,cancel=context.WithTimeout(context.TODO(), time.Second*time.Duration(5))
	defer cancel()
	pCtn.Oper(ctx, ctn.CREATE)
}

func ctrlContainer(ctnName string, opera string) {
	pCtn := ctnS.GetCtn(ctnName)
	pCtn.OperNum = 1
	//fmt.Printf("%s：%s  %#v\n",flagMap[index],ctnName, pCtn)
	var ctx context.Context
	var cancel context.CancelFunc
	ctx,cancel=context.WithTimeout(context.TODO(), time.Second * time.Duration(5))
	defer cancel()
	pCtn.Oper(ctx, opera)
}