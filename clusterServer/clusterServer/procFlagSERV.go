package clusterServer

import (
	header "clusterHeader"
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
		r.Count = sNames;

	case header.FLAG_SERV_CREATE: // 创建服务
		g_controller.CreateSvcFromFile("config/service1.yaml", controller.YML_FILE)
	case header.FLAG_SERV_CTRL:// 启动服务
		{
			var svcOpers map[int]string
			svcOpers = make(map[int]string)
			svcOpers[2] = controller.SSTART
			svcOpers[3] = controller.SSTOP
			svcOpers[4] = controller.SREMOVE
			svcOpers[5] = controller.SSCALE
			svcOpers[6] = controller.SRESTART
			fmt.Println("请选择服务：")
			sNames := g_controller.GetSvcNames()
			for index, val := range sNames {
				fmt.Printf("%d.%s\n", index, val)
			}
			var sIndex int
			fmt.Scanln(&sIndex)
			if sIndex >= len(sNames) {
				break
			}
			rlt := g_controller.Contains(sNames[sIndex])
			if rlt {
				service := g_controller.GetSvc(sNames[sIndex])
				//fmt.Printf("您选择的服务：%#v\n", service)
				switch index {
				case 2:
					g_controller.StartSvc(service.SvcName)
				case 3:
					g_controller.StopSvc(service.SvcName)
				case 4:
					g_controller.RemoveSvc(service.SvcName)
				case 5:
					g_controller.ScaleSvc(service.SvcName, 8)
				case 6:
					g_controller.StopSvc(service.SvcName)
					g_controller.StartSvc(service.SvcName)
				}
				//g_cluster.cr
				//if index==2||index==5||index==6{//启动或扩容缩容或重启
				//	//如果是扩容缩容服务
				//	fmt.Println("请输入副本规模：")
				//	var sScale int
				//	fmt.Scanln(&sScale)
				//	service.SvcOperChan <- cluster.SVC_OPER{
				//		SOperName: svcOpers[index],//将服务操作放入操作通道
				//		Scale: sScale,
				//	}
				//}else{
				//	service.SvcOperChan <- cluster.SVC_OPER{
				//		SOperName: svcOpers[index],//将服务操作放入操作通道
				//	}
				//}
			} else {
				fmt.Println("服务%s不存在！", sNames[sIndex])
			}
		}
	case header.FLAG_SERV_STATS:

	case header.FLAG_SERV_INFO :

	case header.FLAG_CTNS_CRET:  // 创建容器
		{
			var agentNames []string
			var addrIndex int
			for {
				for {
					agentNames = g_controller.GetNodeList(controller.ACTIVE_NODES)
					for index, agentName := range agentNames {
						fmt.Printf("%d:%s\n", index, agentName)
					}
					fmt.Println("请输入agent端IP地址序号：\n")

					fmt.Scanln(&addrIndex)
					if addrIndex >= 0 && addrIndex < len(agentNames) {
						goto TT
					}
				}
			}
		TT:
			fmt.Println("创建容器:")
			fmt.Println("请输入镜像名称：")
			var imageName string
			fmt.Scanln(&imageName)
			fmt.Printf("%s\n", imageName)
			var config map[string]string
			pCtn = ctnS.NewCtnS(imageName, agentNames[addrIndex], config)
			var ctx context.Context
			var cancel context.CancelFunc
			ctx,cancel=context.WithTimeout(context.TODO(), time.Second*time.Duration(5))
			defer cancel()
			pCtn.Oper(ctx, ctn.CREATE)
		}
	case header.FLAG_CTNS_CTRL: // 启动容器
		{
			var flagMap map[int]string
			flagMap = make(map[int]string)
			flagMap[2] = ctn.START
			flagMap[3] = ctn.STOP
			flagMap[4] = ctn.KILL
			flagMap[5] = ctn.REMOVE
			flagMap[6] = ctn.INSPECT
			flagMap[7] = ctn.GETLOG

			ctnName := chooseCtnName()
			if ctnName == "" {
				break
			}

			pCtn = ctnS.GetCtn(ctnName)
			pCtn.OperNum = 1
			//fmt.Printf("%s：%s  %#v\n",flagMap[index],ctnName, pCtn)
			var ctx context.Context
			var cancel context.CancelFunc
			ctx,cancel=context.WithTimeout(context.TODO(), time.Second * time.Duration(5))
			defer cancel()
			pCtn.Oper(ctx, flagMap[index])
		}
	case header.FLAG_CTNS_STATS: // 获取容器状态


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
