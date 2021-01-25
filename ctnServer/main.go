package main

import (
	"context"
	"ctnCommon/headers"
	"ctnCommon/protocol"
	"ctnServer/controller"
	"ctnServer/ctnS"
	"fmt"
	"time"

	"ctnCommon/ctn"
	"tcpSocket"

	"sync"
)

const (
	FLAG_CTRL  = "CTRL"
	FLAG_CTN   = "INFO"
	FLAG_STATS = "STAT"
	FLAG_EVENT = "EVTM"
)

var (
	exit         bool
	mMutex       sync.Mutex
	g_controller *controller.CONTROLLER
)

func init() {
	g_controller = controller.NewController(mySendCtn)
	g_controller.Start()
	//go cluster.MsgEvent()
}

func main() {
	tcpSocket.Listen("0.0.0.0", 10000, myReceiveData, myStateChange)

	for {
		fmt.Println("请选择操作类型：")
		fmt.Println("1.服务操作")
		fmt.Println("2.容器操作")
		fmt.Println("3.启动集群")
		fmt.Println("4.停止集群")
		fmt.Printf("请输入操作序号：")
		var sOc int //标识是服务操作还是容易操作
		fmt.Scanln(&sOc)
		var bStopped bool = true
		switch sOc {
		case 1: //服务操作
			{
				for {
					var index int
					fmt.Println("服务操作列表：")
					fmt.Println("0.获取服务列表")
					fmt.Println("1.创建服务")
					fmt.Println("2.启动服务")
					fmt.Println("3.停止服务")
					fmt.Println("4.删除服务")
					fmt.Println("5.扩容缩容服务")
					fmt.Println("6.重启服务")
					fmt.Println("任意键.Exit")
					fmt.Println("请输入操作序号：")
					fmt.Scanln(&index)
					switch index {
					case 0: //获取服务列表
						{
							fmt.Println("服务列表如下：")
							sNames := g_controller.GetSvcNames()
							sLen := len(sNames)
							if sLen == 0 {
								fmt.Println("当前无服务！")
							}
							for index, val := range sNames {
								fmt.Printf("%d.%s\n", index, val)
							}
							break
						}
					case 1, 7: //创建服务
						{
							//创建服务
							g_controller.CreateSvcFromFile("config/service1.yaml", controller.YML_FILE)
						}
					case 2, 3, 4, 5, 6: //启动服务
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
							} else {
								fmt.Println("服务%s不存在！", sNames[sIndex])
							}
						}
					default: //退出服务操作
						{
							bStopped = false
						}

					}

					if !bStopped {
						break
					}
				}
			}
		case 2: //容器操作
			{
				for {
					fmt.Println("容器操作列表：")
					fmt.Println("1.创建容器")
					fmt.Println("2.启动容器")
					fmt.Println("3.停止容器")
					fmt.Println("4.强制停止容器")
					fmt.Println("5.删除容器")
					fmt.Println("6.获取容器详细信息")
					fmt.Println("7.获取容器内日志")
					fmt.Println("8.启动监控")
					fmt.Println("9.停止监控")
					fmt.Println("任意键.退出")
					fmt.Printf("请输入容器操作序号：")

					var index int
					fmt.Scanln(&index)

					pCtn := &ctnS.CTNS{}
					switch index {
					case 1: //创建容器
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
							ctx, cancel = context.WithTimeout(context.TODO(), time.Second*time.Duration(5))
							defer cancel()
							pCtn.Oper(ctx, ctn.CREATE)
						}
					case 2, 3, 4, 5, 6, 7: //启动容器
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
							//pCtn.OperNum = 1
							//fmt.Printf("%s：%s  %#v\n",flagMap[index],ctnName, pCtn)
							var ctx context.Context
							var cancel context.CancelFunc
							ctx, cancel = context.WithTimeout(context.TODO(), time.Second*time.Duration(5))
							defer cancel()
							pCtn.Oper(ctx, flagMap[index])
						}
					default: //退出
						{
							bStopped = false
						}

					}

					if !bStopped {
						break
					}
				}
			}
		case 3: //启动集群
			g_controller.Start()
		case 4: //停止集群
			g_controller.Stop()
		default:
			continue
		}
	}
}

func chooseCtnName() string {
	for {
		sliceAddr := ctnS.GetCtnNames()
		for i, value := range sliceAddr {
			fmt.Printf("%d:%s\n", i, value)
		}
		fmt.Println("\n请输入容器序号：\n")

		var addrIndex int
		fmt.Scanln(&addrIndex)
		if addrIndex < len(sliceAddr[addrIndex]) {
			return sliceAddr[addrIndex]
		} else {
			return ""
		}
	}
}

func myReceiveData(h string, pkgId uint16, i string, s []byte) {
	ReceiveDataFromAgent(h, 1, pkgId, i, s)
}

func ReceiveDataFromAgent(h string, level uint8, pkgId uint16, i string, s []byte) {
	if i == "CTRL" {
		fmt.Println("\n从agent端收取数据", h, pkgId, i)
	}

	pSaTruck := &protocol.SA_TRUCK{}
	err := headers.Decode(s, pSaTruck)
	if err != nil {
		fmt.Errorf(err.Error())
		return
	}
	//controller.Mylog.Debug(fmt.Sprintf("agent端收到的数据%v",pSaTruck))
	////pSaTruck.SrcAddr = h

	ctnS.GetRecvChan() <- pSaTruck
}

func myStateChange(id string, mystring uint8) {
	fmt.Println(id, mystring)
	var status bool
	switch mystring {
	case 1:
		status = true
	case 2:
		status = false
	}

	g_controller.PutNode(id, status)
}

//向Agent端发送容器操作命令
func mySendCtn(ip string, level uint8, pkgId uint16, flag string, data []byte) {
	fmt.Println("\nmySendCtn:", ip)
	tcpSocket.WriteData(ip, level, pkgId, flag, data)
}
