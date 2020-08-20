package main

import (
	"ctnCommon/headers"
	"ctnCommon/pool"
	"ctnServer/cluster"
	"ctnServer/ctnS"
	"fmt"

	"ctnCommon/ctn"
	"tcpSocket"

	"sync"
)

const (
	FLAG_CTRL  = "CTRL"
	FLAG_CTN   = "INFO"
	FLAG_STATS = "STAT"
	FLAG_EVENT = "EVTM"

	CLUSTER_NAME = "集群管理平台"
	SERVICE_WATCH = "集群管理平台"+"_"+"服务监视"
	NODE_WATCH = "集群管理平台"+"_"+"节点监视"
)

var (
	exit    bool
	mMutex  sync.Mutex
	g_cluster *cluster.CLUSTER
)

func init() {
	g_cluster = cluster.NewCluster(CLUSTER_NAME)
	ctnS.Config(mySendCtn)
	g_cluster.Start(SERVICE_WATCH,NODE_WATCH)
	//go cluster.MsgEvent()
}

func main() {
	tcpSocket.Listen("0.0.0.0", 10000, myReceiveData, myStateChange)

	for {
		fmt.Println("请选择操作类型：")
		fmt.Println("1.服务操作")
		fmt.Println("2.容器操作")
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
							sNames := g_cluster.GetSvcNames()
							sLen := len(sNames)
							if sLen == 0 {
								fmt.Println("当前无服务！")
							}
							for index, val := range sNames {
								fmt.Printf("%d.%s\n", index, val)
							}
							break
						}
					case 1,7: //创建服务
						{
							//创建服务
							g_cluster.CreateSvc("config/service1.yaml", cluster.YML_FILE)
						}
					case 2,3,4,5,6: //启动服务
						{
							var svcOpers map[int]string
							svcOpers=make(map[int]string)
							svcOpers[2]=cluster.SSTART
							svcOpers[3]=cluster.SSTOP
							svcOpers[4]=cluster.SREMOVE
							svcOpers[5]=cluster.SSCALE
							svcOpers[6]=cluster.SRESTART
							fmt.Println("请选择服务：")
							sNames:=g_cluster.GetSvcNames()
							for index, val := range sNames{
								fmt.Printf("%d.%s\n", index, val)
							}
							var sIndex int
							fmt.Scanln(&sIndex)
							if sIndex>=len(sNames){
								break
							}
							rlt := g_cluster.Contains(sNames[sIndex])
							if rlt {
								service := g_cluster.GetSvc(sNames[sIndex])
								fmt.Printf("您选择的服务：%#v\n", service)
								switch index {
								case 2:
									g_cluster.StartSvc(service.SvcName)
								case 3:
									g_cluster.StopSvc(service.SvcName)
								case 4:
									g_cluster.RemoveSvc(service.SvcName)
								case 5:
									g_cluster.ScaleSvc(service.SvcName, 8)
								case 6:
									g_cluster.StopSvc(service.SvcName)
									g_cluster.StartSvc(service.SvcName)
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
									agentNames = g_cluster.GetNodeList(cluster.ACTIVE_NODES)
									for index, agentName:=range agentNames{
											fmt.Printf("%d:%s\n", index, agentName)
									}
									fmt.Println("请输入agent端IP地址序号：\n")

									fmt.Scanln(&addrIndex)
									if addrIndex >= 0 && addrIndex < len(agentNames){
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
							pCtn = ctnS.NewCtnS(imageName, agentNames[addrIndex],config)
							pCtn.Oper(ctn.CREATE)
						}
					case 2,3,4,5,6,7: //启动容器
						{
							var flagMap map[int]string
							flagMap=make(map[int]string)
							flagMap[2]=ctn.START
							flagMap[3]=ctn.STOP
							flagMap[4]=ctn.KILL
							flagMap[5]=ctn.REMOVE
							flagMap[6]=ctn.INSPECT
							flagMap[7]=ctn.GETLOG

							ctnName := chooseCtnName()
							if ctnName == "" {
								break
							}

							pCtn = ctnS.GetCtn(ctnName)
							pCtn.OperNum = 1
							//fmt.Printf("%s：%s  %#v\n",flagMap[index],ctnName, pCtn)
							pCtn.Oper(flagMap[index])
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
		default:
			continue
		}
	}
}

func chooseCtnName() string {
	for {
		sliceAddr:=ctnS.GetCtnNames()
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
	//fmt.Println("\n从agent端收取数据", h, pkgId,i)
	pSaTruck := &ctn.SA_TRUCK{}
	err := headers.Decode(s, pSaTruck)
	if err != nil {
		fmt.Errorf(err.Error())
	}

	ctnS.GetRecvChan() <- pSaTruck
}


func myStateChange(id string, mystring uint8) {
	fmt.Println(id, mystring)
	nodeStatusMap := make(map[string]bool)
	switch mystring {
	case 1:
		nodeStatusMap[id] = true
	case 2:
		nodeStatusMap[id] = false
	}

	pChan:=pool.GetPrivateChanStr(NODE_WATCH)
	pChan<-nodeStatusMap
}

//向Agent端发送容器操作命令
func mySendCtn (ip string, level uint8, pkgId uint16, flag string, data []byte){
	fmt.Println("\nmySendCtn:", ip)
	tcpSocket.WriteData(ip, level, pkgId, flag, data)
}
