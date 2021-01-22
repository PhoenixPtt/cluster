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
	ctnIdMap   map[string]string          //从容器名称到容器Id的映射
	ctnInfoMap map[string]types.Container //从容器ID到容器信息的映射
	ctnStatMap map[string]ctn.CTN_STATS   //从容器ID到容器资源使用状态的映射
	ctnClstMap map[string]string          //从容器ID到集群的映射
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
		ctnWorkPool: pool.NewWorkPool(),
		ctnObjPool:  pool.NewObjPool(),
		agentAddr:   agentAddr,
		serverMap:   make(map[string]*serverInfor, MAX_SERVER_NUM),
		ctnIdMap:    make(map[string]string, MAX_CTN_NUM),          //从容器名称到容器Id的映射
		ctnInfoMap:  make(map[string]types.Container, MAX_CTN_NUM), //从容器ID到容器信息的映射
		ctnStatMap:  make(map[string]ctn.CTN_STATS, MAX_CTN_NUM),   //从容器ID到容器资源使用状态的映射
		ctnClstMap:  make(map[string]string, MAX_CTN_NUM),          //从容器ID到集群的映射
	}

	//初始化采样率
	G_samplingRate = 1

	//配置发送数据接口
	G_ctnMgr.ctnWorkPool.Config(sendObjFunc)

	//监听容器操作状态变化
	go WatchCtns()

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

	//监听容器信息
	server.ctx, G_ctnMgr.cancel_monitor = context.WithCancel(context.TODO())
	MonitorCtns(server.ctx, serverAddr)
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
		pCtnA *ctn.CTN
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

			reqAns := pSaTruck.Req_Ans[0]
			reqAns.CtnErrType = make([]string, 1)

			//获取对应的容器
			if pObj = G_ctnMgr.ctnObjPool.GetObj(reqAns.CtnName); pObj == nil {
				continue
			}
			pCtnA = pObj.(*ctn.CTN)

			//回复Server端,只有当容器id不存在的
			if pCtnA == nil {
				switch reqAns.CtnOper {
				case ctn.CREATE, ctn.RUN: //创建和运行操作，需在容器池中添加容器对象
					pCtnA = &ctn.CTN{}
					pCtnA.CtnName = reqAns.CtnName
					pCtnA.Image = reqAns.CtnImage

					G_ctnMgr.ctnObjPool.AddObj(pCtnA.CtnName, pCtnA)    //将容器加入容器池
					G_ctnMgr.ctnIdMap[pCtnA.CtnID] = pCtnA.CtnName      //容器ID映射到容器名称
					G_ctnMgr.ctnClstMap[pCtnA.CtnID] = pSaTruck.SrcAddr //记录该容器所属的集群
				default:
					//对于非创建、运行操作，如果在容器池中找不到该容器则，返回错误。
					reqAns.CtnErr = fmt.Errorf("容器%s不存在，无法执行%s操作。", reqAns.CtnName, reqAns.CtnOper)
				}
			} else {
				reqAns.CtnErr = nil
			}
			pSaTruck.Req_Ans[0] = reqAns
			pSaTruck.DesAddr = pSaTruck.SrcAddr   //源地址变目标地址
			pSaTruck.SrcAddr = G_ctnMgr.agentAddr //agent的地址作为源地址
			pSendChan := G_ctnMgr.ctnWorkPool.GetSendChan()
			pSendChan <- pSaTruck

			PutCtnOper(pCtnA, reqAns.CtnOper)
		}
	}
}

func PutCtnOper(pCtn *ctn.CTN, oper string) {
	pCtn.OperType = oper
	switch pCtn.OperType {
	case ctn.CREATE:
		pool.RegPrivateChanStr(OPERATE_WATCH, OPERATE_CHAN_BUFFER)
		go WatchCtnOper()
	}
	pChan := pool.GetPrivateChanStr(OPERATE_WATCH)
	pChan <- pCtn
}

func handleCtnOperSuccess(ctnName string, operType string, response interface{}) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
		ok     bool
		ctnId  string = G_ctnMgr.ctnIdMap[ctnName]
	)

	serverAddr := G_ctnMgr.ctnClstMap[ctnName]
	cli := G_ctnMgr.serverMap[serverAddr].cli
	switch operType {
	case ctn.CREATE, ctn.RUN:
		obj := G_ctnMgr.ctnObjPool.GetObj(ctnName)
		pCtn := obj.(*ctn.CTN)
		pCtn.CtnID = response.(string)
	case ctn.START:
		ctx, cancel = context.WithCancel(context.TODO())
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

		delete(G_ctnMgr.ctnIdMap, ctnId)
		delete(G_ctnMgr.ctnClstMap, ctnId)
		G_ctnMgr.ctnObjPool.RemoveObj(ctnName)
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

func OperateN(ctx context.Context, pCtnA *ctn.CTN, num int) (response interface{}, err error) {
	for i := 0; i < num; i++ {
		response, err = Operate(context.TODO(), pCtnA)
		if err == nil {
			handleCtnOperSuccess(pCtnA.CtnName, pCtnA.OperType, response)
			break
		}
	}
	return
}

//集群工作协程
func WatchCtnOper() {
	var (
		pCtnA *ctn.CTN
		err   error

		ctx      context.Context
		pSaTruck protocol.SA_TRUCK

		response interface{}
	)
	ctx, G_ctnMgr.cancelWatchCtnOper = context.WithCancel(context.Background())
	for {
		select {
		case <-ctx.Done():
			pool.UnregPrivateChanStr(OPERATE_WATCH)
			return
		case obj := <-pool.GetPrivateChanStr(OPERATE_WATCH):
			pCtnA = obj.(*ctn.CTN)
			pSaTruck.DesAddr = G_ctnMgr.ctnClstMap[pCtnA.CtnID]
			pSaTruck.SrcAddr = pCtnA.AgentAddr
			reqAns := pSaTruck.Req_Ans[0]
			reqAns.CtnOper = pCtnA.OperType
			response, err = OperateN(context.TODO(), pCtnA, pCtnA.AgentTryNum) //默认重复执行3次
			switch pCtnA.OperType {
			case ctn.CREATE, ctn.RUN, ctn.START, ctn.STOP, ctn.KILL, ctn.REMOVE:
				if err != nil { //执行N次仍然失败，则上报给server端
					reqAns.CtnErr = err
				}
			case ctn.GETLOG:
				reqAns.CtnLog = make([]string, 1)
				if err == nil {
					reqAns.CtnLog[0] = response.(string)
				}
			case ctn.INSPECT:
				reqAns.CtnInspect = make([]ctn.CTN_INSPECT, 1)
				if err == nil {
					reqAns.CtnInspect[0] = response.(ctn.CTN_INSPECT)
				}
			}
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

		pool.CallbackSendCtn(pSaTruck.DesAddr, 1, 0, pSaTruck.Flag, byteStream, G_ctnMgr.ctnWorkPool.GetSendFunc()) //通知主线程发送数据
	}
}

//获取指定集群的所有容器
func getCtns(clstName string) (ctnIds []string) {
	for ctnId, _ := range G_ctnMgr.ctnClstMap {
		if G_ctnMgr.ctnClstMap[ctnId] == clstName {
			ctnIds = append(ctnIds, ctnId)
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

			//获取属于该集群的所有容器
			ctnIds := getCtns(clstName)

			var pSaTruck protocol.SA_TRUCK
			pool.AddIndex()
			pSaTruck.Flag = ctn.FLAG_CTN
			pSaTruck.Index = pool.GetIndex()
			pSaTruck.SrcAddr = G_ctnMgr.agentAddr
			pSaTruck.DesAddr = clstName
			//更新容器信息
			var ctnInfo []ctn.CTN
			ctnInfo = make([]ctn.CTN, 0, MAX_CTN_NUM)
			for _, ctnId := range ctnIds {
				//获取容器结构体
				ctnName, ok := G_ctnMgr.ctnIdMap[ctnId]
				if !ok {
					continue
				}
				pObj := G_ctnMgr.ctnObjPool.GetObj(ctnName)
				pCtn := pObj.(*ctn.CTN)
				if pCtn == nil {
					continue
				}

				//更新容器信息
				if container, ok := G_ctnMgr.ctnInfoMap[ctnId]; ok {
					//容器在docker中实际存在
					pCtn.Dirty = false
					pCtn.Container = container
				} else {
					//容器在docker中已不存在
					pCtn.Dirty = true
					pCtn.Container = types.Container{} //清空容器信息
					pCtn.CTN_STATS = ctn.CTN_STATS{}   //清空资源状态信息
				}
				//注：此处dirty参数表征了容器中的数据是有效的还是已经失效了。如果dirty=true，表明容器对象中的容器信息与资源使用状态已经失效了。

				//更新资源使用状态信息
				if ctnStat, ok := G_ctnMgr.ctnStatMap[ctnId]; ok {
					pCtn.CTN_STATS = ctnStat
				}

				ctnInfo = append(ctnInfo, *pCtn)
			}
			G_ctnMgr.ctnWorkPool.GetSendChan() <- &pSaTruck
			timer.Reset(interval)
		}
	}
}

//更新容器事件信息
func handleEventMessage(evtMsg events.Message) {
	var (
		//ok      bool
		//ctnName string
		addr string
	)
	//if evtMsg.Type == "container" {
	//	if ctnName, ok = ctnIdMap[evtMsg.ID]; !ok {
	//		return
	//	}
	//	ctnEvtMsgMap[ctnName] = evtMsg
	//
	//	//更新容器对象池中的容器状态
	//}

	//向服务端发送容器事件信息
	var pSaTruck protocol.SA_TRUCK
	pool.AddIndex()
	pSaTruck.Flag = ctn.FLAG_EVENT
	pSaTruck.Index = pool.GetIndex()

	pSaTruck.EvtMsg = make([]events.Message, 0, 1)
	pSaTruck.EvtMsg = append(pSaTruck.EvtMsg, evtMsg)

	if evtMsg.Type == "container" {
		//如果是容器相关事件，则只发给容器所属集群
		if destAddr, ok := G_ctnMgr.ctnClstMap[evtMsg.ID]; ok {
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
