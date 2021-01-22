package ctnS

import (
	"ctnCommon/ctn"
	"ctnCommon/headers"
	"ctnCommon/pool"
	"ctnCommon/protocol"
	"github.com/docker/docker/api/types/events"
)

var (
	pCtnPool *pool.OBJ_POOL
	ctnIDMap map[string]*CTNS
)

func init() {
	pCtnPool = pool.NewObjPool()
	ctnIDMap = make(map[string]*CTNS)
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

func getRplStateFromCtnState(ctnState string) (rplState string) {
	switch ctnState {
	case CTN_STATUS_RUNNING:
		rplState = CTN_STATUS_RUNNING
	default:
		rplState = CTN_STATUS_NOTRUNNING
	}
	return
}

func UpdateInfo(pSaTruck *protocol.SA_TRUCK) {
	switch pSaTruck.Flag {
	case ctn.FLAG_CTRL:
		{
			len := len(pSaTruck.Req_Ans)
			if len < 1 {
				//不处理
			}
			reqAns := pSaTruck.Req_Ans[0]
			pCtn := GetCtn(reqAns.CtnName)
			switch reqAns.CtnOper {
			case ctn.CREATE, ctn.RUN, ctn.START, ctn.STOP, ctn.KILL, ctn.REMOVE:
				pCtn.OperType = reqAns.CtnOper
				pCtn.OperErr = reqAns.CtnErr.Error()
			case ctn.GETLOG:
				pCtn.OperType = reqAns.CtnOper
				pCtn.OperErr = reqAns.CtnErr.Error()
				pCtn.CtnLog = reqAns.CtnLog[0]
			case ctn.INSPECT:
				pCtn.OperType = reqAns.CtnOper
				pCtn.OperErr = reqAns.CtnErr.Error()
				pCtn.CtnInspect = reqAns.CtnInspect[0]
			}
		}
	case ctn.FLAG_CTN: //更新容器信息
	case ctn.FLAG_EVENT: //更新事件
		////这些信息都要返回给上层
		//if len(pSaTruck.EvtMsg) > 0 {
		//	eventMsg := pSaTruck.EvtMsg[0]
		//	//fmt.Printf("%#v", eventMsg)
		//	UpdateCtnEvent(eventMsg)
		//}
		//
		//if len(pSaTruck.ErrMsg) > 0 {
		//	//更新错误信息
		//}
	}

	//for _, newCtn := range  ctnInfo{
	//	pCtnS := GetCtn(newCtn.CtnName)

	////判断类型
	//switch pSaTruck.Flag {
	//case ctn.CREATE, ctn.RUN, ctn.START, ctn.STOP, ctn.KILL, ctn.REMOVE:
	//	pCtnS.OperType = newCtn.OperType
	//	pCtnS.op
	//	if err != nil { //执行N次仍然失败，则上报给server端
	//		reqAns.CtnErr = err
	//	}
	//case ctn.GETLOG:
	//	reqAns.CtnLog = make([]string, 1)
	//	if err == nil {
	//		reqAns.CtnLog[0] = response.(string)
	//	}
	//case ctn.INSPECT:
	//	reqAns.CtnInspect = make([]ctn.CTN_INSPECT, 1)
	//	if err == nil {
	//		reqAns.CtnInspect[0] = response.(ctn.CTN_INSPECT)
	//	}
	//}
	//}

	//需要更新的状态

	//根据原始数据计算出来的状态更新

	//ctnLen := len(ctnList)
	//if ctnLen == 0 {
	//	return
	//}
	//for _, container := range ctnList {
	//	pCtnS := GetCtnFromID(container.ID)
	//	if pCtnS != nil {
	//		pCtnPool.Lock()
	//		pCtnS.Container = container
	//		oldRplState := getRplStateFromCtnState(pCtnS.State)
	//		newRplState := getRplStateFromCtnState(container.State)
	//
	//		pCtnS.State = container.State
	//
	//		if oldRplState != newRplState {
	//			pCtnS.Updated = time.Now().UnixNano()
	//			pCtnS.UpdatedString = headers.ToStringInt(pCtnS.Updated, headers.TIME_LAYOUT_NANO)
	//			pChan := pool.GetPrivateChanStr(pCtnS.CtnName)
	//			pChan <- newRplState
	//		}
	//		pCtnPool.Unlock()
	//	}
	//}
	//for _, pCtnS := range GetCtns() {
	//	if pCtnS.CtnID != "" {
	//		var bExisted bool = false
	//		for _, container := range ctnList {
	//			if pCtnS.CtnID == container.ID {
	//				bExisted = true
	//			}
	//		}
	//		if !bExisted {
	//			pCtnPool.Lock()
	//			defer pCtnPool.Unlock()
	//			pCtnS.State = CTN_STATUS_NOTRUNNING
	//			pCtnS.Updated = time.Now().UnixNano()
	//			pCtnS.UpdatedString = headers.ToStringInt(pCtnS.Updated, headers.TIME_LAYOUT_NANO)
	//			pChan := pool.GetPrivateChanStr(pCtnS.CtnName)
	//			pChan <- pCtnS.State
	//		}
	//	}
	//}
}

func UpdateCtnEvent(events events.Message) {
	if events.Type == "container" {
		pCtnS := GetCtnFromID(events.ID)
		if pCtnS != nil {
			if pCtnS.CtnActionTimeInt < events.TimeNano { //比较时间，谁的时间更近，以谁为准
				pCtnPool.Lock()
				defer pCtnPool.Unlock()
				pCtnS.CtnAction = events.Action
				pCtnS.CtnActionTime = headers.ToStringInt(events.TimeNano, headers.TIME_LAYOUT_NANO)
				pCtnS.CtnActionTimeInt = events.TimeNano

				oldRplState := getRplStateFromCtnState(pCtnS.State)

				var ctnStatus string
				switch pCtnS.CtnAction {
				case "start":
					ctnStatus = CTN_STATUS_RUNNING
				default:
					ctnStatus = CTN_STATUS_NOTRUNNING
				}
				if oldRplState != ctnStatus {
					if pCtnS.Updated < events.TimeNano {
						pCtnS.State = ctnStatus
						pCtnS.Updated = events.TimeNano
						pCtnS.UpdatedString = headers.ToStringInt(pCtnS.Updated, headers.TIME_LAYOUT_NANO)
						pChan := pool.GetPrivateChanStr(pCtnS.CtnName)
						pChan <- pCtnS.State
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
