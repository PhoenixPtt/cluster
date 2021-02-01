package ctnA

import (
	"context"
	"ctnCommon/ctn"
	"ctnCommon/headers"
	"ctnCommon/pool"
	"ctnCommon/protocol"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"sync"
	"time"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
//获取容器信息
const (
	RUN_CTN = iota
	ALL_CTN
)

//获取容器列表
func CtnList(cli *client.Client, ctx context.Context, flag int) (containers []types.Container, err error) {
	mutex_ls.Lock()
	defer mutex_ls.Unlock()

	switch flag {
	case RUN_CTN:
		//获取运行中的容器列表
		containers, err = cli.ContainerList(ctx, types.ContainerListOptions{})
	case ALL_CTN:
		//获取运行和停止的所有容器列表
		containers, err = cli.ContainerList(ctx, types.ContainerListOptions{
			All: true,
		})
	}

	return
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
//容器管理器接口
type serverInfor struct {
	serverAddr   string
	serverOnline bool
	cli          *client.Client
	ctx          context.Context
	ctx_stat     context.Context
	cancel_stat  context.CancelFunc
}

type CtnMgr struct {
	agentAddr string //agent端的ip地址

	serverMap map[string]*serverInfor

	ctnWorkPool *pool.WORK_POOL //容器工作池
	ctnObjPool  *pool.OBJ_POOL  //容器对象池

	cancel_monitor     context.CancelFunc            //取消监控容器
	cancel_stats_map   map[string]context.CancelFunc //取消监控容器资源信息
	cancelWatchCtnOper context.CancelFunc            //取消监控容器操作

	//ctnEvtMsgMap map[string]events.Message  //与容器相关的事件集合
	ctnIdMap     map[string]string          //从容器名称到容器Id的映射
	ctnInfoMap   map[string]types.Container //从容器ID到容器信息的映射
	ctnStatMap   map[string]ctn.CTN_STATS   //从容器ID到容器资源使用状态的映射
	ctnClstMap   map[string]string          //从容器ID到集群的映射
	clstMutexMap map[string]*sync.Mutex     //从集群名称到互斥量的映射
}

func (pCtnMgr *CtnMgr) UpdateServerOnlineStatus(serverAddr string, bOnline bool) {
	if pServerInfo, ok := pCtnMgr.serverMap[serverAddr]; ok {
		pServerInfo.serverOnline = bOnline
	}
	return
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
var (
	OPERATE_WATCH       = "OperateCtnWatch"
	OPERATE_CHAN_BUFFER = 1000
	MAX_SERVER_NUM      = 100
	MAX_CTN_NUM         = 1000

	//全部变量，容器管理器
	G_ctnMgr       *CtnMgr //容器管理器
	G_samplingRate int     //采样率
)

//初始化容器管理器
func InitCtnMgr(sendObjFunc pool.SendObjFunc, agentAddr string) {
	//初始化容器管理器
	G_ctnMgr = &CtnMgr{
		ctnWorkPool:  pool.NewWorkPool(),
		ctnObjPool:   pool.NewObjPool(),
		agentAddr:    agentAddr,
		serverMap:    make(map[string]*serverInfor, MAX_SERVER_NUM),
		ctnIdMap:     make(map[string]string, MAX_CTN_NUM),          //从容器名称到容器Id的映射
		ctnInfoMap:   make(map[string]types.Container, MAX_CTN_NUM), //从容器ID到容器信息的映射
		ctnStatMap:   make(map[string]ctn.CTN_STATS, MAX_CTN_NUM),   //从容器ID到容器资源使用状态的映射
		ctnClstMap:   make(map[string]string, MAX_CTN_NUM),          //从容器ID到集群的映射
		clstMutexMap: make(map[string]*sync.Mutex, MAX_SERVER_NUM),
	}

	//初始化采样率
	G_samplingRate = 1

	//配置发送数据接口
	G_ctnMgr.ctnWorkPool.Config(sendObjFunc)

	//监听容器操作状态变化
	go WatchCtns()

	pool.RegPrivateChanStr(OPERATE_WATCH, OPERATE_CHAN_BUFFER)
	go WatchCtnOper()

	//反馈容器操作结果或者更新容器状态
	go ResponseCtns()
}

func AddServer(serverAddr string) {
	var (
		err error
	)

	var server serverInfor
	server.serverAddr = serverAddr
	server.serverOnline = false
	//初始化docker客户端
	if server.cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()); err != nil {
		fmt.Println(err.Error())
	}
	G_ctnMgr.serverMap[serverAddr] = &server
	var pMutex *sync.Mutex = &sync.Mutex{}
	G_ctnMgr.clstMutexMap[serverAddr] = pMutex

	//监听容器信息
	server.ctx, G_ctnMgr.cancel_monitor = context.WithCancel(context.TODO())
	go MonitorCtns(server.ctx, serverAddr)
}

//接收server端的数据
func Unload(pSaTruck *protocol.SA_TRUCK) {
	pRecvChan := G_ctnMgr.ctnWorkPool.GetRecvChan()
	pRecvChan <- pSaTruck
}

//监听容器操作
func WatchCtns() {
	var (
		pObj  interface{}
		pCtnA *ctn.CTN = nil
	)

	for {
		select {
		case obj := <-G_ctnMgr.ctnWorkPool.GetRecvChan():
			pSaTruck := obj.(*protocol.SA_TRUCK)
			if pSaTruck.Flag != ctn.FLAG_CTRL { //仅接收控制指令
				continue
			}

			if len(pSaTruck.Req_Ans) < 1 {
				continue
			}

			reqAns := &pSaTruck.Req_Ans[0]

			//获取对应的容器
			pObj = G_ctnMgr.ctnObjPool.GetObj(reqAns.CtnName)

			//回复Server端,只有当容器id不存在的
			if pObj == nil {
				switch reqAns.CtnOper {
				case ctn.CREATE, ctn.RUN: //创建和运行操作，需在容器池中添加容器对象
					pCtnA = &ctn.CTN{}
					pCtnA.CtnName = reqAns.CtnName
					pCtnA.ImageName = reqAns.CtnImage
					pCtnA.AgentTryNum = reqAns.AgentTryNum
					pCtnA.OperType = reqAns.CtnOper
					pCtnA.OperIndex = pSaTruck.Index
					pCtnA.OperErr = "nil"

					clstName := pSaTruck.SrcAddr
					G_ctnMgr.clstMutexMap[clstName].Lock()
					G_ctnMgr.ctnObjPool.AddObj(pCtnA.CtnName, pCtnA) //将容器加入容器池
					G_ctnMgr.ctnIdMap[pCtnA.CtnID] = pCtnA.CtnName   //容器ID映射到容器名称
					G_ctnMgr.ctnClstMap[pCtnA.CtnName] = clstName    //记录该容器所属的集群
					G_ctnMgr.clstMutexMap[clstName].Unlock()

				default:
					//对于非创建、运行操作，如果在容器池中找不到该容器则，返回错误。
					reqAns.CtnErr = fmt.Sprintf("容器%s不存在，无法执行%s操作。", reqAns.CtnName, reqAns.CtnOper)
				}
			} else {
				pCtnA = pObj.(*ctn.CTN)
				pCtnA.OperType = reqAns.CtnOper
				pCtnA.OperIndex = pSaTruck.Index
				pCtnA.OperErr = "nil"
			}
			if pCtnA != nil {
				reqAns.CtnErr = pCtnA.OperErr
			}
			pSaTruck.DesAddr = pSaTruck.SrcAddr   //源地址变目标地址
			pSaTruck.SrcAddr = G_ctnMgr.agentAddr //agent的地址作为源地址
			pSendChan := G_ctnMgr.ctnWorkPool.GetSendChan()
			pSendChan <- pSaTruck

			//执行具体容器操作
			if pCtnA != nil {
				pChan := pool.GetPrivateChanStr(OPERATE_WATCH)
				pChan <- pCtnA
			}
		}
	}
}

func handleCtnOperSuccess(ctnName string, operType string, response interface{}, err error) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
		ok     bool
		ctnId  string = G_ctnMgr.ctnIdMap[ctnName]
	)

	obj := G_ctnMgr.ctnObjPool.GetObj(ctnName)
	pCtn := obj.(*ctn.CTN)
	if err != nil {
		pCtn.OperErr = err.Error()
		return
	}

	pCtn.OperErr = "nil"
	switch operType {
	case ctn.CREATE, ctn.RUN:
		pCtn.CtnID = response.(string)
		pCtn.Created = time.Now().UnixNano()
		pCtn.CreatedString = headers.ToString(time.Now(), headers.TIME_LAYOUT_NANO) //创建时间
	case ctn.START:
		ctx, cancel = context.WithCancel(context.TODO())
		serverAddr := G_ctnMgr.ctnClstMap[ctnName]
		cli := G_ctnMgr.serverMap[serverAddr].cli
		CtnStats(cli, ctx, ctnId)
		if _, ok = G_ctnMgr.cancel_stats_map[ctnId]; !ok {
			G_ctnMgr.cancel_stats_map[ctnId] = cancel
		}
	case ctn.STOP, ctn.KILL:
		if _, ok = G_ctnMgr.cancel_stats_map[ctnId]; ok {
			G_ctnMgr.cancel_stats_map[ctnId]()
			delete(G_ctnMgr.cancel_stats_map, ctnId)
		}
	case ctn.REMOVE:
		//取消监听容器资源使用状态
		if _, ok = G_ctnMgr.cancel_stats_map[ctnId]; ok {
			G_ctnMgr.cancel_stats_map[ctnId]()
			delete(G_ctnMgr.cancel_stats_map, ctnId)
		}

		clstName := G_ctnMgr.ctnClstMap[pCtn.CtnName]
		if clstName != "" {
			//Mylog.Debug(fmt.Sprintf("%s %s %s  %v", ctnName, ctnId, clstName, G_ctnMgr.clstMutexMap[clstName]))
			G_ctnMgr.clstMutexMap[clstName].Lock()
			if _, ok := G_ctnMgr.ctnIdMap[ctnId]; ok {
				delete(G_ctnMgr.ctnIdMap, ctnId)
			}
			if _, ok := G_ctnMgr.ctnClstMap[pCtn.CtnName]; ok {
				delete(G_ctnMgr.ctnClstMap, pCtn.CtnName)
			}
			if _, ok = G_ctnMgr.cancel_stats_map[ctnId]; ok {
				G_ctnMgr.cancel_stats_map[ctnId]()
				delete(G_ctnMgr.cancel_stats_map, ctnId)
			}
			G_ctnMgr.ctnObjPool.RemoveObj(ctnName)
			G_ctnMgr.clstMutexMap[clstName].Unlock()
		}
	case ctn.GETLOG:
		pCtn.CtnLog = response.(string)
	case ctn.INSPECT:
		pCtn.CtnInspect = response.(ctn.CTN_INSPECT)
	}
}

func Operate(ctx context.Context, pCtn *ctn.CTN) (response interface{}, err error) {
	ctnName := pCtn.CtnName
	imageName := pCtn.ImageName
	serverAddr := G_ctnMgr.ctnClstMap[ctnName]
	cli := G_ctnMgr.serverMap[serverAddr].cli

	switch pCtn.OperType {
	case ctn.CREATE:
		response, err = Create(cli, ctx, ctnName, imageName)
	case ctn.START:
		response, err = Start(cli, ctx, ctnName)
	case ctn.RUN:
		response, err = Run(cli, ctx, ctnName, imageName)
	case ctn.STOP:
		response, err = Stop(cli, ctx, ctnName)
	case ctn.KILL:
		response, err = Kill(cli, ctx, ctnName)
	case ctn.REMOVE:
		response, err = Remove(cli, ctx, ctnName)
	case ctn.GETLOG:
		response, err = GetLog(cli, ctx, ctnName)
	case ctn.INSPECT:
		response, err = Inspect(cli, ctx, ctnName)
	}
	return
}

func OperateN(ctx context.Context, pCtnA *ctn.CTN, num int) {
	var (
		response interface{}
		err      error
	)
	for i := 1; i <= num; i++ {
		response, err = Operate(context.TODO(), pCtnA)
		if err == nil {
			handleCtnOperSuccess(pCtnA.CtnName, pCtnA.OperType, response, err)
			break
		} else {
			if i == num {
				return
			}

			switch pCtnA.OperType {
			case ctn.START:
				if response != nil {
					ctnId := response.(string)
					Flush(ctnId)
				}
			}
		}
	}
	return
}

//集群工作协程
func WatchCtnOper() {
	var (
		pCtnA *ctn.CTN
		err   error
		ctx   context.Context
	)

	ctx, G_ctnMgr.cancelWatchCtnOper = context.WithCancel(context.Background())
	for {
		select {
		case <-ctx.Done():
			pool.UnregPrivateChanStr(OPERATE_WATCH)
			return
		case obj := <-pool.GetPrivateChanStr(OPERATE_WATCH):
			pCtnA = obj.(*ctn.CTN)
			oper := pCtnA.OperType

			//先查一查自己容器池中的容器对象是否过期
			pObj_in_pool := pool.GetObj(pCtnA.CtnName)
			if pObj_in_pool != nil {
				ctnId := pCtnA.CtnID
				ctnName := pCtnA.CtnName
				Mylog.Debug(fmt.Sprintf("%s %s %s", oper, ctnId, ctnName))
				pCtnA_in_pool := pObj_in_pool.(*ctn.CTN)
				if pCtnA_in_pool.Dirty { //过期了
					//过期的容器对象，只响应删除操作
					Mylog.Debug(fmt.Sprintf(ctnId))
					if oper == ctn.REMOVE {
						//在容器池中删除容器对象
						clstName := G_ctnMgr.ctnClstMap[ctnName]
						if clstName != "" {
							G_ctnMgr.clstMutexMap[clstName].Lock()
							if _, ok := G_ctnMgr.ctnIdMap[ctnId]; ok {
								delete(G_ctnMgr.ctnIdMap, ctnId)
							}
							if _, ok := G_ctnMgr.ctnClstMap[ctnName]; ok {
								delete(G_ctnMgr.ctnClstMap, ctnName)
							}
							if _, ok := G_ctnMgr.cancel_stats_map[ctnId]; ok {
								G_ctnMgr.cancel_stats_map[ctnId]()
								delete(G_ctnMgr.cancel_stats_map, ctnId)
							}
							G_ctnMgr.ctnObjPool.RemoveObj(ctnName)
							G_ctnMgr.clstMutexMap[clstName].Unlock()
						}
					}
					continue
				}
			}

			OperateN(context.TODO(), pCtnA, pCtnA.AgentTryNum) //默认重复执行3次

			var pSaTruck protocol.SA_TRUCK
			//pSaTruck.Index = index
			pSaTruck.Index = -pCtnA.OperIndex
			pSaTruck.Flag = ctn.FLAG_CTRL
			pSaTruck.DesAddr = G_ctnMgr.ctnClstMap[pCtnA.CtnName]
			pSaTruck.SrcAddr = pCtnA.AgentAddr
			reqAns := protocol.REQ_ANS{}
			reqAns.CtnName = pCtnA.CtnName
			reqAns.CtnOper = pCtnA.OperType
			reqAns.CtnErr = pCtnA.OperErr //操作的错误信息返回给server端
			switch pCtnA.OperType {
			case ctn.CREATE, ctn.RUN:
				reqAns.CtnID = make([]string, 1)
				reqAns.CtnID[0] = pCtnA.CtnID
				reqAns.Created = pCtnA.Created
				reqAns.CreatedString = pCtnA.CreatedString
			case ctn.START, ctn.STOP, ctn.KILL, ctn.REMOVE:
			case ctn.GETLOG:
				reqAns.CtnLog = make([]string, 1)
				if err == nil {
					reqAns.CtnLog[0] = pCtnA.CtnLog
				}
			case ctn.INSPECT:
				reqAns.CtnInspect = make([]ctn.CTN_INSPECT, 1)
				if err == nil {
					reqAns.CtnInspect[0] = pCtnA.CtnInspect
				}
			}
			pSaTruck.Req_Ans = make([]protocol.REQ_ANS, 0, 1)
			pSaTruck.Req_Ans = append(pSaTruck.Req_Ans, reqAns)
			pSaTruck.MsgTime = time.Now().UnixNano()
			pSaTruck.MsgTimeStr = headers.ToString(time.Now(), headers.TIME_LAYOUT_NANO)
			G_ctnMgr.ctnWorkPool.GetSendChan() <- &pSaTruck
		}
	}
}

//向Server端反馈容器操作执行结果或者更新容器状态
func ResponseCtns() {
	for obj := range G_ctnMgr.ctnWorkPool.GetSendChan() {
		pSaTruck := obj.(*protocol.SA_TRUCK)
		byteStream, err := headers.Encode(pSaTruck) //打包
		if err != nil {
			errCode := "CTN：网络数据打包失败！"
			fmt.Println(errCode)
			continue
		}

		//Mylog.Debug(fmt.Sprintf("%s, %v", pSaTruck.DesAddr, pSaTruck))
		//for _, pCtn:=range pSaTruck.CtnInfos{
		//	Mylog.Debug(fmt.Sprintf("%s, %s, %s",pCtn.CtnName, pCtn.CtnID, pCtn.Container.State))
		//}
		pool.CallbackSendCtn(pSaTruck.DesAddr, 1, 0, pSaTruck.Flag, byteStream, G_ctnMgr.ctnWorkPool.GetSendFunc()) //通知主线程发送数据
	}
}

//获取指定集群的所有容器
func getCtnNames(clstName string) (ctnNames []string) {
	for ctnName, _ := range G_ctnMgr.ctnClstMap {
		if G_ctnMgr.ctnClstMap[ctnName] == clstName {
			ctnNames = append(ctnNames, ctnName)
		}
	}
	return
}

//监控指定集群中容器状态
func MonitorCtns(ctx context.Context, clstName string) {
	var (
		timer      *time.Timer
		containers []types.Container
		container  types.Container
		//定时器时间间隔
		interval time.Duration = time.Second * time.Duration(G_samplingRate)
	)

	timer = time.NewTimer(interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			containers, _ = CtnList(G_ctnMgr.serverMap[clstName].cli, context.TODO(), ALL_CTN) //获取容器信息

			G_ctnMgr.ctnInfoMap = make(map[string]types.Container) //清空容器信息数据
			for _, container = range containers {
				G_ctnMgr.ctnInfoMap[container.ID] = container
			}

			G_ctnMgr.clstMutexMap[clstName].Lock()
			//获取属于该集群的所有容器
			ctnNames := getCtnNames(clstName)

			var pSaTruck protocol.SA_TRUCK
			pool.AddIndex()
			pSaTruck.Flag = ctn.FLAG_CTN
			pSaTruck.Index = pool.GetIndex()
			pSaTruck.SrcAddr = G_ctnMgr.agentAddr
			pSaTruck.DesAddr = clstName
			//更新容器信息
			pSaTruck.CtnInfos = make([]ctn.CTN, 0, MAX_CTN_NUM)
			for _, ctnName := range ctnNames {
				//获取容器结构体
				pObj := G_ctnMgr.ctnObjPool.GetObj(ctnName)
				//Mylog.Debug(fmt.Sprintf("%s,%s, %v", clstName, ctnName, pObj))
				if pObj == nil {
					continue
				}
				pCtn := pObj.(*ctn.CTN)

				//更新容器信息
				if container, ok := G_ctnMgr.ctnInfoMap[pCtn.CtnID]; ok {
					//容器在docker中实际存在
					pCtn.Dirty = false
					pCtn.DirtyPosition = ""
					pCtn.Container = container
				} else {
					//容器在docker中已不存在
					pCtn.Dirty = true
					pCtn.DirtyPosition = ctn.DIRTY_POSITION_DOCKER
					pCtn.Container = types.Container{} //清空容器信息
					pCtn.CTN_STATS = ctn.CTN_STATS{}   //清空资源状态信息
				}

				//信息更新时间
				pCtn.Updated = time.Now().UnixNano()
				pCtn.UpdatedString = headers.ToString(time.Now(), headers.TIME_LAYOUT_NANO)

				//注：此处dirty参数表征了容器中的数据是有效的还是已经失效了。如果dirty=true，表明容器对象中的容器信息与资源使用状态已经失效了。

				//更新资源使用状态信息
				if ctnStat, ok := G_ctnMgr.ctnStatMap[pCtn.CtnID]; ok {
					pCtn.CTN_STATS = ctnStat
				}

				pSaTruck.CtnInfos = append(pSaTruck.CtnInfos, *pCtn)
			}
			pSaTruck.MsgTime = time.Now().UnixNano()
			pSaTruck.MsgTimeStr = headers.ToString(time.Now(), headers.TIME_LAYOUT_NANO)
			Mylog.Debug(fmt.Sprintf("%s,%v, %v", pSaTruck.MsgTimeStr, len(pSaTruck.CtnInfos), pSaTruck.CtnInfos))

			G_ctnMgr.clstMutexMap[clstName].Unlock()

			G_ctnMgr.ctnWorkPool.GetSendChan() <- &pSaTruck
			timer.Reset(interval)
		}
	}
}

//更新容器事件信息
func handleEventMessage(evtMsg events.Message) {
	var (
		addr string
	)

	//向服务端发送容器事件信息
	var pSaTruck protocol.SA_TRUCK
	pool.AddIndex()
	pSaTruck.Flag = ctn.FLAG_EVENT
	pSaTruck.Index = pool.GetIndex()

	pSaTruck.EvtMsg = make([]events.Message, 0, 1)
	pSaTruck.EvtMsg = append(pSaTruck.EvtMsg, evtMsg)

	if evtMsg.Type == "container" {
		//如果是容器相关事件，则只发给容器所属集群
		ctnName := G_ctnMgr.ctnIdMap[evtMsg.ID]
		if destAddr, ok := G_ctnMgr.ctnClstMap[ctnName]; ok {
			pSaTruck.DesAddr = destAddr
			G_ctnMgr.ctnWorkPool.GetSendChan() <- &pSaTruck
		}
	} else {
		//其它事件，发给所有server端
		for addr = range G_ctnMgr.serverMap {
			pSaTruck.DesAddr = addr
			G_ctnMgr.ctnWorkPool.GetSendChan() <- &pSaTruck
		}
	}
}

//更新错误事件信息
func handleErrorMessage(errMsg error) {
	//向服务端发送错误事件信息
	var (
		pSaTruck protocol.SA_TRUCK
		addr     string
	)
	pool.AddIndex()
	pSaTruck.Flag = ctn.FLAG_EVENT
	pSaTruck.Index = pool.GetIndex()

	pSaTruck.ErrMsg = make([]error, 0, 1)
	pSaTruck.ErrMsg = append(pSaTruck.ErrMsg, errMsg)

	//发给所有server端
	for addr = range G_ctnMgr.serverMap {
		pSaTruck.DesAddr = addr
		G_ctnMgr.ctnWorkPool.GetSendChan() <- &pSaTruck
	}
}
