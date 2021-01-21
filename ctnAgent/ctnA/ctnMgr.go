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

const (
	RUN_CTN = iota
	ALL_CTN
)

var (
	OPERATE_WATCH       = "OperateCtnWatch"
	OPERATE_CHAN_BUFFER = 1000
)

type CtnMgr struct {
	ctnWorkPool  *pool.WORK_POOL //容器工作池
	ctnObjPool   *pool.OBJ_POOL  //容器对象池
	serverAddrs  []string        //server端ip地址
	serverOnline []bool          //server端在线状态
	agentAddr    string          //agent端的ip地址

	cancel_monitor     context.CancelFunc            //取消监控容器
	cancel_stats_map   map[string]context.CancelFunc //取消监控容器资源信息
	cancelWatchCtnOper context.CancelFunc            //取消监控容器操作
}

func (pCtnMgr *CtnMgr) IsServerExisted(serverAddr string) (bOnline bool) {
	for _, addr := range pCtnMgr.serverAddrs {
		if addr == serverAddr {
			bOnline = true
			break
		}
	}
	bOnline = false
	return
}

func (pCtnMgr *CtnMgr) UpdateServerOnlineStatus(serverAddr string, bOnline bool) {
	for index, addr := range pCtnMgr.serverAddrs {
		if addr == serverAddr {
			pCtnMgr.serverOnline[index] = bOnline
		}
	}
	return
}

var (
	//全部变量，容器管理器
	G_ctnMgr       *CtnMgr //容器管理器
	G_samplingRate int     //采样率

	cli *client.Client

	ctnEvtMsgMap map[string]events.Message  //与容器相关的事件集合
	ctnInfoMap   map[string]types.Container //从容器ID到容器信息的映射
	ctnIdMap     map[string]string          //从容器名称到容器Id的映射
	ctnStatMap   map[string]ctn.CTN_STATS   //从容器ID到容器资源使用状态的映射
	ctnDirtyMap  map[string]bool            //从容器ID到容器过期标志的映射
	ctnClstMap   map[string]string          //从容器ID到集群的映射
)

//初始化容器管理器
func InitCtnMgr(sendObjFunc pool.SendObjFunc, agentAddr string, serverAddrs []string) {
	var (
		err error
		ctx context.Context
	)

	//初始化容器管理器
	G_ctnMgr = &CtnMgr{
		ctnWorkPool: pool.NewWorkPool(),
		ctnObjPool:  pool.NewObjPool(),
		serverAddrs: serverAddrs,
		agentAddr:   agentAddr,
	}

	//初始化采样率
	G_samplingRate = 1

	//初始化docker客户端
	if cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()); err != nil {
		fmt.Println(err.Error())
	}

	//监听容器信息
	ctx, G_ctnMgr.cancel_monitor = context.WithCancel(context.TODO())
	ctx = ctx
	//MonitorCtns(ctx)

	////配置发送数据接口
	//G_ctnMgr.ctnWorkPool.Config(sendObjFunc)

	////监听容器操作状态变化
	//go WatchCtns()

	////反馈容器操作结果或者更新容器状态
	//go ResponseCtns()

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

					//将容器加入容器池
					G_ctnMgr.ctnObjPool.AddObj(pCtnA.CtnName, pCtnA)

					//容器ID映射到容器名称
					ctnIdMap[pCtnA.ID] = pCtnA.CtnName
					//记录该容器所属的集群
					ctnClstMap[pCtnA.ID] = pSaTruck.SrcAddr
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

func handleCtnOperSuccess(ctnName string, operType string) {
	var (
		m_ctx    context.Context
		m_cancel context.CancelFunc
		ok       bool
		ctnId    string = ctnIdMap[ctnName]
	)

	switch operType {
	case ctn.START:
		m_ctx, m_cancel = context.WithCancel(context.TODO())
		CtnStats(m_ctx, ctnId)
		if _, ok = G_ctnMgr.cancel_stats_map[ctnId]; !ok {
			G_ctnMgr.cancel_stats_map[ctnId] = m_cancel
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

		delete(ctnIdMap, ctnId)
		delete(ctnClstMap, ctnId)
		G_ctnMgr.ctnObjPool.RemoveObj(ctnName)
	}
}

func Operate(ctx context.Context, pCtn *ctn.CTN) (err error) {
	switch pCtn.OperType {
	case ctn.CREATE:
		err = Create(ctx, pCtn.CtnName, pCtn.ImageName)
	case ctn.START:
		err = Start(ctx, pCtn.CtnName)
	case ctn.RUN:
		err = Run(ctx, pCtn.CtnName, pCtn.ImageName)
	case ctn.STOP:
		err = Stop(ctx, pCtn.CtnName)
	case ctn.KILL:
		err = Kill(ctx, pCtn.CtnName)
	case ctn.REMOVE:
		err = Remove(ctx, pCtn.CtnName)
	}
	return
}

func OperateN(ctx context.Context, pCtnA *ctn.CTN, num int) (err error) {
	for i := 0; i < num; i++ {
		err = Operate(context.TODO(), pCtnA)
		if err == nil {
			handleCtnOperSuccess(pCtnA.CtnName, pCtnA.OperType)
			break
		}
	}
	return
}

//集群工作协程
func WatchCtnOper() {
	var (
		pCtnA      *ctn.CTN
		err        error
		log        string
		ctnInspect ctn.CTN_INSPECT

		ctx      context.Context
		pSaTruck protocol.SA_TRUCK
	)
	ctx, G_ctnMgr.cancelWatchCtnOper = context.WithCancel(context.Background())
	for {
		select {
		case <-ctx.Done():
			pool.UnregPrivateChanStr(OPERATE_WATCH)
			return
		case obj := <-pool.GetPrivateChanStr(OPERATE_WATCH):
			pCtnA = obj.(*ctn.CTN)
			pSaTruck.DesAddr = ctnClstMap[pCtnA.CtnID]
			pSaTruck.SrcAddr = pCtnA.AgentAddr
			reqAns := pSaTruck.Req_Ans[0]
			reqAns.CtnOper = pCtnA.OperType
			switch pCtnA.OperType {
			case ctn.CREATE, ctn.RUN, ctn.START, ctn.STOP, ctn.KILL, ctn.REMOVE:
				err = OperateN(context.TODO(), pCtnA, 3) //默认重复执行3次
				if err != nil {                          //执行N次仍然失败，则上报给server端
					reqAns.CtnErr = err
				}
			case ctn.GETLOG:
				log, err = GetLog(context.TODO(), pCtnA.CtnName)
				reqAns.CtnLog = make([]string, 1)
				if err == nil {
					reqAns.CtnLog[0] = log
				}
			case ctn.INSPECT:
				ctnInspect, err = Inspect(context.TODO(), pCtnA.CtnName)
				reqAns.CtnInspect = make([]ctn.CTN_INSPECT, 1)
				if err == nil {
					reqAns.CtnInspect[0] = ctnInspect
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

//获取指定集群的所有容器
func getCtns(clstName string) (ctnIds []string) {
	for ctnId, _ := range ctnClstMap {
		if ctnClstMap[ctnId] == clstName {
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
		//ctnName    string
		//obj        interface{}
		//pCtnA      *ctn.CTN
		//ok         bool
		//ctnStat    ctn.CTN_STATS
		//addr       string
		//定时器时间间隔
		interval time.Duration = time.Second * time.Duration(G_samplingRate)
	)

	timer = time.NewTimer(interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			//清空
			ctnInfoMap = make(map[string]types.Container)

			//获取容器信息
			containers, _ = CtnList(cli, context.TODO(), ALL_CTN)

			//建立从容器ID到容器的映射
			for _, container = range containers {
				ctnInfoMap[container.ID] = container
			}

			////获取属于该集群的所有容器
			//ctnIds := getCtns(clstName)

			////更新容器数据
			//for index, _ := range ctnIds {
			//
			//}
			//
			////更新容器池
			//for _, ctnName = range G_ctnMgr.ctnObjPool.GetObjNames() {
			//	obj = G_ctnMgr.ctnObjPool.GetObj(ctnName)
			//	pCtnA = obj.(*ctn.CTN)
			//	if _, ok = ctnInfoMap[pCtnA.ID]; !ok {
			//		container = types.Container{}
			//	} else {
			//		container = ctnInfoMap[pCtnA.ID]
			//	}
			//
			//	if _, ok = G_ctnMgr.cancel_stats_map[pCtnA.ID]; !ok {
			//		ctnStat = ctn.CTN_STATS{}
			//	} else {
			//		ctnStat = ctnStatMap[pCtnA.ID]
			//	}
			//}
			//
			//fmt.Println("111111111111111111", pSaTruck)
			//
			//for _, addr = range G_ctnMgr.serverAddrs {
			//	pSaTruck.Addr = addr
			//	G_ctnMgr.ctnWorkPool.GetSendChan() <- &pSaTruck
			//}

			////整理监控信息
			//var pSaTruck protocol.SA_TRUCK
			//pool.AddIndex()
			//pSaTruck.Flag = ctn.FLAG_CTN
			//pSaTruck.Index = pool.GetIndex()

			//根据容器所属集群不同，分别将容器信息发送给对应的集群
			//for _, ctnName=range G_ctnMgr.ctnObjPool.GetObjNames(){
			//	obj = G_ctnMgr.ctnObjPool.GetObj(ctnName)
			//	pCtnA = obj.(*ctn.CTN)
			//	if pCtnA.ID==""{
			//		container = types.Container{}
			//		ctnStat = ctn.CTN_STATS{}
			//	}else{
			//		if _,ok = ctnInfoMap[pCtnA.ID]; !ok{
			//			container = types.Container{}
			//		}else{
			//			container = ctnInfoMap[pCtnA.ID]
			//		}
			//
			//		if _,ok = G_ctnMgr.cancel_stats_map[pCtnA.ID]; !ok{
			//			ctnStat = ctn.CTN_STATS{}
			//		}else{
			//			ctnStat = ctnStatMap[pCtnA.ID]
			//		}
			//	}
			//
			//	pSaTruck.CtnList = append(pSaTruck.CtnList, container)
			//	pSaTruck.CtnStat = append(pSaTruck.CtnStat, ctnStat)
			//}
			//
			//fmt.Println("111111111111111111", pSaTruck)
			//
			//for _, addr = range G_ctnMgr.serverAddrs{
			//	pSaTruck.Addr = addr
			//	G_ctnMgr.ctnWorkPool.GetSendChan() <- &pSaTruck
			//}
			timer.Reset(interval)
		}
	}

}

//更新容器事件信息
func handleEventMessage(evtMsg events.Message) {
	var (
		ok      bool
		ctnName string
		addr    string
	)
	if evtMsg.Type == "container" {
		if ctnName, ok = ctnIdMap[evtMsg.ID]; !ok {
			return
		}
		ctnEvtMsgMap[ctnName] = evtMsg

		//更新容器对象池中的容器状态
	}

	//向服务端发送容器事件信息
	var pSaTruck protocol.SA_TRUCK
	pool.AddIndex()
	pSaTruck.Flag = ctn.FLAG_EVENT
	pSaTruck.Index = pool.GetIndex()

	pSaTruck.EvtMsg = make([]events.Message, 0, 1)
	pSaTruck.EvtMsg = append(pSaTruck.EvtMsg, evtMsg)

	if evtMsg.Type == "container" {
		//如果是容器相关事件，则只发给容器所属集群
		if destAddr, ok := ctnClstMap[evtMsg.ID]; ok {
			pSaTruck.DesAddr = destAddr
			G_ctnMgr.ctnWorkPool.GetSendChan() <- &pSaTruck
		}
	} else {
		//其它事件，发给所有server端
		for _, addr = range G_ctnMgr.serverAddrs {
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
	for _, addr = range G_ctnMgr.serverAddrs {
		pSaTruck.DesAddr = addr
		G_ctnMgr.ctnWorkPool.GetSendChan() <- &pSaTruck
	}
}
