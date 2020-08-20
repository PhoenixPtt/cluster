package ctnA

import (
	"ctnCommon/headers"
	"ctnCommon/pool"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
)

var (
	pCtnPool *pool.OBJ_POOL
)

func init()  {
	pCtnPool = pool.NewObjPool()
}

//添加容器
func AddCtn(pCtnA *CTNA) {
	pCtnPool.AddObj(pCtnA.CtnName, pCtnA)
}

//删除容器
func RemoveCtn(ctnName string) error {
	return pCtnPool.RemoveObj(ctnName)
}

//通过容器ID获取容器结构体
func GetCtn(ctnName string) (pCtn *CTNA) {
	obj := pCtnPool.GetObj(ctnName)
	if obj == nil{
		return nil
	}
	pCtn = obj.(*CTNA)
	return
}

//获取指定节点上运行的所有容器
func GetCtns() (pCtnAS []*CTNA) {
	ctnNames := GetCtnNames()

	for _,ctnName:=range ctnNames{
		pCtnAS = append(pCtnAS, GetCtn(ctnName))
	}
	return
}

func GetCtnNames() []string {
	return pCtnPool.GetObjNames()
}

func UpdateCtnEvent(events events.Message){
	if events.Type == "container"{
		pCtnA := GetCtnFromID(events.ID)
		if pCtnA != nil{
			if pCtnA.CtnActionTimeInt < events.TimeNano{//比较时间，谁的时间更近，以谁为准
				pCtnPool.Lock()
				defer pCtnPool.Unlock()
				pCtnA.CtnAction = events.Action
				pCtnA.CtnActionTime = headers.ToStringInt(events.TimeNano,headers.TIME_LAYOUT_NANO)
				pCtnA.CtnActionTimeInt = events.TimeNano
			}
		}
	}
}

func UpdateCtnInfo(container types.Container)  {
	pCtnA := GetCtnFromID(container.ID)
	if pCtnA != nil{
		pCtnPool.Lock()
		defer pCtnPool.Unlock()
		pCtnA.Container = container
	}
}


func GetCtnFromID(id string) *CTNA {
	for _,ctnName:=range pCtnPool.GetObjNames(){
		pCtnA := GetCtn(ctnName)
		if pCtnA.ID == id{
			return pCtnA
		}
	}

	return nil
}

