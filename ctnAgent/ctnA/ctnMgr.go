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

type CtnMgr struct {
	ctnWorkPool *pool.WORK_POOL//容器工作池
	//对象池应该根据server端域名进行分别映射
	//server端域名对应的ip地址可能有多个或者ip地址可以变更
	//更新策略：如果是新增或者删除操作，一律根据同步操作来
	ctnObjPool *pool.OBJ_POOL//容器对象池
	serverAddrs []string//状态信息上传目的地址

	cancel_monitor context.CancelFunc//取消监控容器
	cancel_stats_map map[string]context.CancelFunc//取消监控容器资源信息
}

var (
	//全部变量，容器管理器
	G_ctnMgr *CtnMgr//容器管理器
	G_samplingRate int//采样率

	cli *client.Client

	ctnEvtMsgMap map[string]events.Message//与容器相关的事件集合
	ctnInfoMap map[string]types.Container//从容器ID到容器信息的映射
	ctnIdMap map[string]string//从容器名称到容器Id的映射
	ctnStatMap map[string]ctn.CTN_STATS//从容器ID到容器资源使用状态的映射
)

//初始化容器管理器
func InitCtnMgr(sendObjFunc pool.SendObjFunc, serverAddrs []string) {
	var(
		err error
		ctx context.Context
	)

	//初始化容器管理器
	G_ctnMgr = &CtnMgr{
		ctnWorkPool: pool.NewWorkPool(),
		ctnObjPool: pool.NewObjPool(),
		serverAddrs: serverAddrs,
	}

	//初始化采样率
	G_samplingRate = 1

	//初始化docker客户端
	if cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation());err != nil{
		fmt.Println(err.Error())
	}

	//监听容器信息
	ctx, G_ctnMgr.cancel_monitor = context.WithCancel(context.TODO())
	MonitorCtns(ctx)

	////配置发送数据接口
	//G_ctnMgr.ctnWorkPool.Config(sendObjFunc)

	////监听容器操作状态变化
	//go WatchCtns()

	////反馈容器操作结果或者更新容器状态
	//go ResponseCtns()

}

func Operate(ctx context.Context, pCtn *ctn.CTN, operType string) (err error) {
	switch operType {
	case ctn.CREATE:
		err=Create(ctx, pCtn.CtnName, pCtn.ImageName)
	case ctn.START:
		err=Start(ctx, pCtn.CtnName)
	case ctn.RUN:
		err=Run(ctx, pCtn.CtnName, pCtn.ImageName)
	case ctn.STOP:
		err=Stop(ctx, pCtn.CtnName)
	case ctn.KILL:
		err=Kill(ctx, pCtn.CtnName)
	case ctn.REMOVE:
		err=Remove(ctx, pCtn.CtnName)
	}
	return
}

//func OperateWithStratgy(pCtn *ctn.CTN, operType string) (err error) {
//	if pCtn.OperStrategy{
//		//第一阶段：操作一次
//		err=Operate(pCtn,operType)
//		if err==nil{
//			return
//		}else{
//			//向Server端发送状态消息
//		}
//
//		//第二阶段：操作若干次
//		for i:=0; i<pCtn.OperNum; i++{
//			err=Operate(pCtn,operType)
//			if err==nil{
//				return
//			}else{
//				//向Server端发送状态消息
//			}
//		}
//
//		//第三阶段：删除
//		for{//操作不成功，删除容器
//			err=Operate(pCtn,ctn.REMOVE)
//			if err==nil{
//				return
//			}
//			time.Sleep(time.Second)
//		}
//	}else{
//		err=Operate(pCtn,operType)
//		return
//	}
//}

//接收server端的数据
func Unload(pSaTruck *protocol.SA_TRUCK)  {
	pRecvChan := G_ctnMgr.ctnWorkPool.GetRecvChan()
	pRecvChan <- pSaTruck
}

//监听容器操作状态变化
func WatchCtns() {
	var(
		pObj interface{}
		pCtnA *ctn.CTN
		errType string
		err error
		log string
		ctnInspect ctn.CTN_INSPECT

		m_ctx context.Context
		m_cancel context.CancelFunc

		ok bool
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
			if pObj = G_ctnMgr.ctnObjPool.GetObj(reqAns.CtnName);pObj==nil{
				continue
			}
			pCtnA = pObj.(*ctn.CTN)

			switch reqAns.CtnOper {
			case ctn.CREATE, ctn.RUN:
				if pCtnA == nil {
					pCtnA = &ctn.CTN{}
					pCtnA.CtnName = reqAns.CtnName
					pCtnA.Image = reqAns.CtnImage
				}
				err = Operate(context.TODO(), pCtnA, reqAns.CtnOper)
				reqAns.CtnState = pCtnA.State
				reqAns.CtnID = make([]string, 1)
				reqAns.CtnID[0] = pCtnA.ID
				reqAns.CtnErrType[0] = errType
				reqAns.CtnErr = err

				if err == nil {
					//将容器加入容器池
					G_ctnMgr.ctnObjPool.AddObj(pCtnA.CtnName, pCtnA)

					//容器ID映射到容器名称
					ctnIdMap[pCtnA.ID] = pCtnA.CtnName
				}
			default:
				if pCtnA == nil {
					reqAns.CtnState = "not exist"
					reqAns.CtnErrType[0] = ""
					reqAns.CtnErr = nil
				} else {
					switch reqAns.CtnOper {
					case ctn.START:
						err = Operate(context.TODO(), pCtnA, reqAns.CtnOper)
						reqAns.CtnState = pCtnA.State
						reqAns.CtnErrType[0] = errType
						reqAns.CtnErr = err

						if err == nil{
							//监听容器资源使用状态
							m_ctx, m_cancel = context.WithCancel(context.TODO())
							CtnStats(m_ctx, pCtnA.ID)
							if _, ok = G_ctnMgr.cancel_stats_map[pCtnA.ID]; !ok{
								G_ctnMgr.cancel_stats_map[pCtnA.ID] = m_cancel
							}
						}
					case ctn.STOP, ctn.KILL:
						err = Operate(context.TODO(), pCtnA, reqAns.CtnOper)
						reqAns.CtnState = pCtnA.State
						reqAns.CtnErrType[0] = errType
						reqAns.CtnErr = err

						//取消监听容器资源使用状态
						if err == nil{
							if _, ok = G_ctnMgr.cancel_stats_map[pCtnA.ID]; ok{
								G_ctnMgr.cancel_stats_map[pCtnA.ID]()
								delete(G_ctnMgr.cancel_stats_map, pCtnA.ID)
							}
						}
					case ctn.REMOVE:
						err = Operate(context.TODO(), pCtnA, reqAns.CtnOper)
						reqAns.CtnState = pCtnA.State
						reqAns.CtnErrType[0] = errType
						reqAns.CtnErr = err

						if err == nil {
							//取消监听容器资源使用状态
							if _, ok = G_ctnMgr.cancel_stats_map[pCtnA.ID]; ok{
								G_ctnMgr.cancel_stats_map[pCtnA.ID]()
								delete(G_ctnMgr.cancel_stats_map, pCtnA.ID)
							}

							delete(ctnIdMap, pCtnA.ID)
							G_ctnMgr.ctnObjPool.RemoveObj(pCtnA.CtnName)
						}
					case ctn.GETLOG:
						log, err = GetLog(context.TODO(), pCtnA.CtnName)
						if err == nil{
							reqAns.CtnLog = make([]string, 1)
							reqAns.CtnLog[0] = log
							reqAns.CtnErr = err
						}
					case ctn.INSPECT:
						ctnInspect, err = Inspect(context.TODO(), pCtnA.CtnName)
						if err == nil{
							reqAns.CtnInspect = make([]ctn.CTN_INSPECT, 1)
							reqAns.CtnInspect[0] = ctnInspect
							reqAns.CtnErr = err
						}
					}
				}
			}

			pSaTruck.Req_Ans[0] = reqAns
			pSendChan := G_ctnMgr.ctnWorkPool.GetSendChan()
			pSendChan <- pSaTruck
		}
	}
}

//向Server端反馈容器操作执行结果或者更新容器状态
func ResponseCtns() {
	var(
		srcAddr string
	)

	for obj := range G_ctnMgr.ctnWorkPool.GetSendChan() {
		pSaTruck := obj.(*protocol.SA_TRUCK)
		byteStream, err := headers.Encode(pSaTruck) //打包
		if err != nil {
			errCode := "CTN：网络数据打包失败！"
			fmt.Println(errCode)
			continue
		}

		for _,srcAddr=range pSaTruck.SrcAddr{
			pool.CallbackSendCtn(srcAddr, 1, 0, pSaTruck.Flag, byteStream, G_ctnMgr.ctnWorkPool.GetSendFunc()) //通知主线程发送数据
		}
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

//容器状态监控
func MonitorCtns(ctx context.Context)  {
	var(
		timer *time.Timer
		containers []types.Container
		container types.Container
		ctnName string
		obj interface{}
		pCtnA *ctn.CTN
		ok bool
		ctnStat ctn.CTN_STATS
		addr string
		//定时器时间间隔
		interval time.Duration = time.Second * time.Duration(G_samplingRate)
	)

	timer=time.NewTimer(interval)
	for{
		select {
		case <-ctx.Done():
			return
		case temp := <-timer.C:
			fmt.Println("时间到",temp)
			////清空
			ctnInfoMap = make(map[string]types.Container)

			//获取容器信息
			containers, _ = CtnList(cli, context.TODO(), ALL_CTN)

			//建立从容器ID到容器的映射
			for _, container = range containers{
				ctnInfoMap[container.ID] = container
			}

			//整理监控信息
			var pSaTruck protocol.SA_TRUCK
			pool.AddIndex()
			pSaTruck.Flag = ctn.FLAG_CTN
			pSaTruck.Index = pool.GetIndex()

			fmt.Println("0000000000000000", G_ctnMgr.ctnObjPool.GetObjNames())

			//更新对象池
			//对象池区分属于哪个server端的对象池，这个函数的参数也分监控属于哪个服务端的容器
			//1.如果对象池中没有，但实时猜到了，则不做处理，因为不属于该服务端
			//2.对象池里有，并且实时也采到了，则做同步
			//3.对象池中有，但是实时并没有采到，说明该容器故障，则从对象池中把该容器删除掉
			//4.agent端对象池与server端对象池同步策略



			for _, ctnName=range G_ctnMgr.ctnObjPool.GetObjNames(){
				obj = G_ctnMgr.ctnObjPool.GetObj(ctnName)
				pCtnA = obj.(*ctn.CTN)
				if pCtnA.ID==""{
					container = types.Container{}
					ctnStat = ctn.CTN_STATS{}
				}else{
					if _,ok = ctnInfoMap[pCtnA.ID]; !ok{
						container = types.Container{}
					}else{
						container = ctnInfoMap[pCtnA.ID]
					}

					if _,ok = G_ctnMgr.cancel_stats_map[pCtnA.ID]; !ok{
						ctnStat = ctn.CTN_STATS{}
					}else{
						ctnStat = ctnStatMap[pCtnA.ID]
					}
				}

				pSaTruck.CtnList = append(pSaTruck.CtnList, container)
				pSaTruck.CtnStat = append(pSaTruck.CtnStat, ctnStat)
			}

			fmt.Println("111111111111111111", pSaTruck)

			for _, addr = range G_ctnMgr.serverAddrs{
				pSaTruck.Addr = addr
				G_ctnMgr.ctnWorkPool.GetSendChan() <- &pSaTruck
			}
			timer.Reset(interval)
		}
	}

}

//更新容器事件信息
func handleEventMessage(evtMsg events.Message){
	var(
		ok bool
		ctnName string
		addr string
	)
	if evtMsg.Type == "container"{
		if ctnName,ok=ctnIdMap[evtMsg.ID];!ok{
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

	pSaTruck.EvtMsg = make([]events.Message,0,1)
	pSaTruck.EvtMsg = append(pSaTruck.EvtMsg, evtMsg)

	for _, addr = range G_ctnMgr.serverAddrs{
		pSaTruck.Addr = addr
		G_ctnMgr.ctnWorkPool.GetSendChan() <- &pSaTruck
	}
}

//更新错误事件信息
func handleErrorMessage(errMsg error)  {
	//向服务端发送错误事件信息
	var(
		pSaTruck protocol.SA_TRUCK
		addr string
	)
	pool.AddIndex()
	pSaTruck.Flag = ctn.FLAG_EVENT
	pSaTruck.Index = pool.GetIndex()

	pSaTruck.ErrMsg = make([]error,0,1)
	pSaTruck.ErrMsg = append(pSaTruck.ErrMsg, errMsg)

	for _, addr = range G_ctnMgr.serverAddrs{
		pSaTruck.Addr = addr
		G_ctnMgr.ctnWorkPool.GetSendChan() <- &pSaTruck
	}
}



