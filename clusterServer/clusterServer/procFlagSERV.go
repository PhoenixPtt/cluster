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
	case header.FLAG_SERV_CREATE: // 创建服务
		if len(r.Oper.Par)>0  {
			g_controller.CreateSvcFromFile(r.Oper.Par[0].Value, controller.YML_FILE)
		} else {
			r.Oper.Success = false		// 默认操作成功
			r.Oper.Err = "服务配置文件错误！无法正确创建服务。"
		}
	case header.FLAG_SERV_CTRL:// 启动服务 停止服务 删除服务 扩缩容 重启服务
		if len(r.Oper.Par)>0  {
			svrName := r.Oper.Par[0].Value;
			if g_controller.Contains(svrName)  {
				service := g_controller.GetSvc(svrName)
				//fmt.Printf("您选择的服务：%#v\n", service)
				switch r.Oper.Par[0].Name {
				case header.FLAG_SERV_START:
					g_controller.StartSvc(service.SvcName)
				case header.FLAG_SERV_STOP:
					g_controller.StopSvc(service.SvcName)
				case header.FLAG_SERV_REMOVE:
					g_controller.RemoveSvc(service.SvcName)
				case header.FLAG_SERV_SCALE:
					g_controller.ScaleSvc(service.SvcName, 8)
				case header.FLAG_SERV_RESTART:
					g_controller.StopSvc(service.SvcName)
					g_controller.StartSvc(service.SvcName)
				}
			} else {
				r.Oper.Success = false		// 默认操作成功
				r.Oper.Err = "服务[" + svrName + "]不存在"
			}
		} else {
			r.Oper.Success = false		// 默认操作成功
			r.Oper.Err = "请指定服务名称。"
		}
	case header.FLAG_SERV_STATS:

	case header.FLAG_SERV_INFO :

	case header.FLAG_CTNS_CRET:  // 创建容器
			if len(r.Oper.Par) == 0 {
				r.Oper.Success = false // 默认操作成功
				r.Oper.Err = "请指定镜像名称。"
			} else {
				imageName := r.Oper.Par[0].Value
				if len(r.Oper.Par) == 1 {
					hs := g_controller.GetNodeList(controller.ALL_NODES)
					for _,h := range(hs) {
						go createContainer(imageName, h)
					}
				} else {
					for i := 1; i < len(r.Oper.Par); i++ {
						go createContainer(imageName, r.Oper.Par[i].Value)
					}
				}
			}

	case header.FLAG_CTNS_CTRL: // 启动容器
		if len(r.Oper.Par) >= 1 {
			switch r.Oper.Par[0].Name {
			case header.FLAG_CTNS_START:
				go ctrlContainer(r.Oper.Par[0].Value, ctn.START)
			case header.FLAG_CTNS_STOP:
				go ctrlContainer(r.Oper.Par[0].Value, ctn.STOP)
			case header.FLAG_CTNS_FCST:  // 强制停止容器
				go ctrlContainer(r.Oper.Par[0].Value, ctn.KILL)
			case header.FLAG_CTNS_REMV:  // 删除容器
				go ctrlContainer(r.Oper.Par[0].Value, ctn.REMOVE)
			case header.FLAG_CTNS_LOG :   // 获取容器日志
				go ctrlContainer(r.Oper.Par[0].Value, ctn.GETLOG)
			case header.FLAG_CTNS_STATS : // 获取容器状态
				go ctrlContainer(r.Oper.Par[0].Value, ctn.INSPECT)
			default:
				r.Oper.Success = false					// 操作失败
				r.Oper.Err = "控制容器参数错误！"				// 操作失败信息
			}
		} else {
			r.Oper.Success = false					// 操作失败
			r.Oper.Err = "控制容器至少需要容器ID和控制类型！"				// 操作失败信息
		}

	default:						// 其它操作处理
		r.Oper.Success = false					// 操作失败
		r.Oper.Err = "子标识错误！"				// 操作失败信息
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