package ctnS

import (
	"context"
	"ctnCommon/ctn"
	"ctnCommon/pool"
	"ctnCommon/protocol"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"time"
)

var (
	pCtnPool *pool.OBJ_POOL
	ctnIDMap map[string]*CTNS

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
					if pCtn != nil {
						switch reqAns.CtnOper {
						case ctn.CREATE, ctn.RUN, ctn.START:
							pCtn.OperType = ""
							pCtn.OperErr = ""
							pCtn.CtnID = reqAns.CtnID[0]
						case ctn.STOP, ctn.KILL:
							pCtn.OperType = ""
							pCtn.OperErr = ""
							pCtn.CTN_STATS = ctn.CTN_STATS{} //有容器信息，但容器资源状态信息为空
						case ctn.REMOVE:
							pCtn.CtnID = "" //此时已经没有容器id了
							pCtn.OperType = ""
							pCtn.OperErr = ""
							pCtn.Container = types.Container{} //容器设置为空
							pCtn.CTN_STATS = ctn.CTN_STATS{}   //容器资源状态信息为空
						case ctn.GETLOG:
							pCtn.OperType = ""
							pCtn.OperErr = ""
							pCtn.CtnLog = reqAns.CtnLog[0]
						case ctn.INSPECT:
							pCtn.OperType = ""
							pCtn.OperErr = ""
							pCtn.CtnInspect = reqAns.CtnInspect[0]
						}

						operIndex := -pSaTruck.Index
						pCtn.OperMapMutex.Lock()
						_, ok := pCtn.OperMap[operIndex]
						if ok {
							Mylog.Debug(fmt.Sprintf("\n容器名称：%s\n容器操作：%s\n操作结果：%s\n", pCtn.CtnName, reqAns.CtnOper, reqAns.CtnErr))
							delete(pCtn.OperMap, operIndex)
						}
						pCtn.OperMapMutex.Unlock()
					}
				}
			case ctn.FLAG_CTN: //更新容器信息
				{
					var containerNames []string
					var containerIds []string
					containerNames = make([]string, 0, 1000)
					containerIds = make([]string, 0, 1000)
					for _, container := range pSaTruck.CtnInfos {
						containerNames = append(containerNames, container.CtnName)
						containerIds = append(containerIds, container.CtnID)
					}
					var containerStr string
					if len(containerNames) > 0 {
						for index, containerName := range containerNames {
							ctnName := containerName
							if len(ctnName) > 0 {
								ctnName = containerName[:14]
							}
							ctnId := containerIds[index]
							if len(ctnId) > 0 {
								ctnId = containerIds[index][:10]
							}

							containerStr = fmt.Sprintf("%s\n容器名称：%s	|	容器id：%s", containerStr, ctnName, ctnId)
						}
						Mylog.Debug(fmt.Sprintf("\n******收到agent端容器信息******\n消息时间：%s\n%v\n", pSaTruck.MsgTimeStr, containerStr))
					}
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
							pCtn.OperMapMutex.Lock()
							mapLen := len(pCtn.OperMap)
							Mylog.Debug(fmt.Sprintf("\n\n！！！！！！！！！！！！！！！！注意： server端有但agent端没有\n容器名称：%s\n操作类型：%s\n错误信息：%s\n容器ID：%s\nmaplen的值：%d\n\n", pCtn.CtnName, pCtn.OperType, pCtn.OperErr, pCtn.CtnID, mapLen))
							if mapLen == 0 {
								pCtn.Dirty = true //设置容器信息过时
								pCtn.DirtyPosition = ctn.DIRTY_POSITION_CTN_EXIST_IN_SERVER_BUT_NOT_IN_AGENT
								pChan := pool.GetPrivateChanStr(pCtn.CtnName)
								pCtn.State = ctn.CTN_NOT_EXIST_ON_AGENT

								select {
								case pChan <- pCtn.State:
								default:
								}
							}
							pCtn.OperMapMutex.Unlock()
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
