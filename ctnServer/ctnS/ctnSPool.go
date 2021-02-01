package ctnS

import (
	"context"
	"ctnCommon/ctn"
	"ctnCommon/pool"
	"ctnCommon/protocol"
	"fmt"
	"github.com/docker/docker/api/types/events"
	"time"
)

var (
	pCtnPool        *pool.OBJ_POOL
	ctnIDMap        map[string]*CTNS
	eventMsgs       []events.Message
	errMsgs         []error
	CTN_INFO_WATCH  = "CTN_INFO_WATCH"
	EVENT_MSG_WATCH = "EVENT_MSG_WATCH"
	ERR_MSG_WATCH   = "ERR_MSG_WATCH"
	MSG_NUM         = 1000
)

func init() {
	pCtnPool = pool.NewObjPool()
	ctnIDMap = make(map[string]*CTNS)
	pool.RegPrivateChanStr(CTN_INFO_WATCH, MSG_NUM)
	pool.RegPrivateChanStr(EVENT_MSG_WATCH, MSG_NUM)
	pool.RegPrivateChanStr(ERR_MSG_WATCH, MSG_NUM)

	go WatchCtnInfo()
}

//添加容器
func AddCtn(pCtnS *CTNS) {
	pCtnPool.AddObj(pCtnS.CtnName, pCtnS)
}

//删除容器
func RemoveCtn(ctnName string) error {
	pCtnS := GetCtn(ctnName)
	if pCtnS != nil {
		delete(ctnIDMap, pCtnS.CtnID)
	}
	return pCtnPool.RemoveObj(ctnName)
}

//通过容器ID获取容器结构体
func GetCtn(ctnName string) (pCtn *CTNS) {
	obj := pCtnPool.GetObj(ctnName)
	if obj == nil {
		return nil
	}
	pCtn = obj.(*CTNS)
	return
}

//获取指定节点上运行的所有容器
func GetCtns() (pCtnSs []*CTNS) {
	ctnNames := GetCtnNames()

	for _, ctnName := range ctnNames {
		pCtnSs = append(pCtnSs, GetCtn(ctnName))
	}
	return
}

func GetCtnNames() []string {
	return pCtnPool.GetObjNames()
}

func WatchCtnInfo() {
	pChan := pool.GetPrivateChanStr(CTN_INFO_WATCH)
	for {
		select {
		case pObj := <-pChan:
			pSaTruck := pObj.(*protocol.SA_TRUCK)
			switch pSaTruck.Flag {
			case ctn.FLAG_CTRL:
				{
					len := len(pSaTruck.Req_Ans)
					if len < 1 {
						//不处理
						return
					}
					reqAns := pSaTruck.Req_Ans[0]
					pCtn := GetCtn(reqAns.CtnName)
					switch reqAns.CtnOper {
					case ctn.CREATE, ctn.RUN, ctn.START, ctn.STOP, ctn.KILL, ctn.REMOVE:
						pCtn.OperType = reqAns.CtnOper
						pCtn.OperErr = reqAns.CtnErr
					case ctn.GETLOG:
						pCtn.OperType = reqAns.CtnOper
						pCtn.OperErr = reqAns.CtnErr
						pCtn.CtnLog = reqAns.CtnLog[0]
					case ctn.INSPECT:
						pCtn.OperType = reqAns.CtnOper
						pCtn.OperErr = reqAns.CtnErr
						pCtn.CtnInspect = reqAns.CtnInspect[0]
					}
				}
			case ctn.FLAG_CTN: //更新容器信息
				{
					Mylog.Debug(fmt.Sprintf("消息时间：%s", pSaTruck.MsgTimeStr))
					var ctnMap map[string]bool
					ctnMap = make(map[string]bool) //这个变量主要是帮助找到server端有但agent端没有的容器对象
					for _, ctnInfo := range pSaTruck.CtnInfos {
						pCtn := GetCtn(ctnInfo.CtnName)
						if pCtn != nil {
							//更新的容器信息
							//1.容器信息是否过期
							//2.容器信息
							//3.容器资源状态信息
							pCtn.Dirty = ctnInfo.Dirty
							pCtn.DirtyPosition = ctnInfo.DirtyPosition
							if !pCtn.Dirty { //容器未过期的情况下更新其它信息
								pCtn.Container = ctnInfo.Container
								pCtn.CTN_STATS = ctnInfo.CTN_STATS

								//更新容器状态信息
								pCtn.State = ctnInfo.Container.State
							}
							pChan := pool.GetPrivateChanStr(pCtn.CtnName)
							select {
							case pChan <- pCtn.State:
							default:
							}

							ctnMap[ctnInfo.CtnName] = true
						} else {
							//与agent端同步，删除agent端有但server端没有的容器
							//这里只要把消息发出去就可以了，仅等待1m
							ctx, cancel := context.WithTimeout(context.TODO(), time.Second*time.Duration(1))
							defer cancel()
							pCtn := &CTNS{}
							pCtn.CTN = ctnInfo
							pCtn.Remove(ctx)
						}
					}

					//server端有但是agent端没有的容器对象,这些容器对象在server端的容器信息已经过期
					sCtnNames := GetCtnNames()
					for _, ctnName := range sCtnNames {
						pCtn := GetCtn(ctnName)

						if _, ok := ctnMap[ctnName]; !ok {
							pCtn.Dirty = true //设置容器信息过时
							pCtn.DirtyPosition = ctn.DIRTY_POSITION_CTN_EXIST_IN_SERVER_BUT_NOT_IN_AGENT
							pChan := pool.GetPrivateChanStr(pCtn.CtnName)
							pCtn.State = ctn.CTN_NOT_EXIST_ON_AGENT

							select {
							case pChan <- pCtn.State:
							default:
							}
						}
					}
				}
			case ctn.FLAG_EVENT: //更新事件
				if len(pSaTruck.EvtMsg) > 0 {
					//一般事件信息
					eventMsg := pSaTruck.EvtMsg[0]
					pChan := pool.GetPrivateChanStr(EVENT_MSG_WATCH)
					select {
					case pChan <- eventMsg:
					default:
					}

				}

				if len(pSaTruck.ErrMsg) > 0 {
					//错误事件信息
					errMsg := pSaTruck.ErrMsg[0]
					pChan := pool.GetPrivateChanStr(ERR_MSG_WATCH)
					select {
					case pChan <- errMsg:
					default:

					}
				}
			}
		}
	}
}

func GetCtnFromID(ctnID string) *CTNS {
	_, ok := ctnIDMap[ctnID]
	if ok {
		return ctnIDMap[ctnID]
	}
	return nil
}

func UpdateCtnID(pCtnS *CTNS, ctnID string) {
	pCtnS.CtnID = ctnID
	_, ok := ctnIDMap[ctnID]
	if !ok {
		ctnIDMap[ctnID] = pCtnS
	}
	return
}
