package ctn

import (
	"fmt"
	"sync"

	"clusterHeader"
)

const (
	DEFAULT_MAP_SIZE = 1000
)

const (
	CREATE  = "CREATE"
	START   = "START"
	STOP    = "STOP"
	KILL    = "KILL"
	REMOVE  = "REMOVE"
	GETLOG  = "GETLOG"
	INSPECT = "INSPECT"
	CTNEXIT = "EXIT"
)

var(
	Ctn_pool []header.CTN//容器池
	ctn_used_status []bool//容器池使用状态
	Ctn_index_map map[int]CTN_INDEX//容器索引表
	Mutex sync.Mutex
)


type CTN_INDEX struct{//容器索引表结构
	index int//收发序号
	i int//在切片中的序号
	CtnId string//容器ID
	ServiceName string//容器所属服务名
}

func init()  {
	ctn_size:=1000
	Ctn_pool = make([]header.CTN,0,ctn_size)
	ctn_used_status = make([]bool,0,ctn_size)
	Ctn_index_map = make(map[int]CTN_INDEX)
}

//判断容器是否存在
func IsCtnExisted(index int) bool {
	Mutex.Lock()
	defer Mutex.Unlock()

	//遍历容器索引表
	_,ok:=Ctn_index_map[index]
	if ok{
		return true
	}
	return false
}

//添加容器
func AddCtn(cctn header.CTN) string {
	if IsCtnExisted(cctn.OperIndex) {
		err := fmt.Sprintf("%s已存在，不能重复添加!", cctn.OperIndex)
		return err
	} else {
		Mutex.Lock()
		defer Mutex.Unlock()
		//判断已分配的存储空间中，切片有无未使用的
		unusedIndex:=-1//第一个未使用的存储空间的序号
		for key,ctn_index:=range Ctn_index_map{
			pos:=ctn_index.i
			if ctn_used_status[pos]==false{
				unusedIndex=key
			}
		}

		if unusedIndex==-1{//切片中目前已经分配的空间已经使用完毕
			ctnLen:=len(Ctn_index_map)
			Ctn_pool = append(Ctn_pool, cctn)
			ctn_used_status = append(ctn_used_status, true)
			Ctn_index_map[cctn.OperIndex]=CTN_INDEX{//更新索引表
				index: cctn.OperIndex,
				i:ctnLen,
				CtnId: cctn.ID,
				ServiceName: cctn.ServiceName,
			}
		}else{
			Ctn_pool[unusedIndex]=cctn//放置在未使用的位置
			ctn_used_status[unusedIndex]=true//设置标志量为已使用
			Ctn_index_map[cctn.OperIndex]=CTN_INDEX{//更新索引表
				index: cctn.OperIndex,
				i:unusedIndex,
				CtnId: cctn.ID,
				ServiceName: cctn.ServiceName,
			}
		}
	}
	return ""
}

//更新容器
func UpdateCtn(cctn header.CTN)  {
	Mutex.Lock()
	defer Mutex.Unlock()
	if cctn.OperFlag == CREATE {
		//定位在切片中的位置
		pos:=Ctn_index_map[cctn.OperIndex].i
		Ctn_pool[pos].ID=cctn.ID
		Ctn_pool[pos].Err=cctn.Err
		//更新索引表
		var ctnIndex CTN_INDEX
		ctnIndex=Ctn_index_map[cctn.OperIndex]
		ctnIndex.ServiceName=cctn.ServiceName
		ctnIndex.CtnId=cctn.ID
		Ctn_index_map[cctn.OperIndex]=ctnIndex
		return
	}

	//对于除创建以外的其它操作，根据容器ID定位索引表的索引序号
	oldIndex := -1
	for index, val := range Ctn_index_map {
		if val.CtnId == cctn.ID {
			oldIndex = index
		}
	}

	if oldIndex == -1 {
		fmt.Sprintf("容器%s不存在，无法执行更新操作!", cctn.ID)
	}

	//执行更新操作
	//1.更新容器切片
	pos:=Ctn_index_map[oldIndex].i
	switch cctn.OperFlag {
	case START, STOP, KILL, REMOVE:
		Ctn_pool[pos].Err = cctn.Err //更新错误信息
	case GETLOG:
		Ctn_pool[pos].Err = cctn.Err   //更新错误信息
		Ctn_pool[pos].Logs = cctn.Logs //更新容器内日志
	case INSPECT:
		Ctn_pool[pos].Err = cctn.Err               //更新错误信息
		Ctn_pool[pos].CtnInspect = cctn.CtnInspect //更新容器内日志
	}
	//2.更新索引表
	Ctn_index_map[cctn.OperIndex]=Ctn_index_map[oldIndex]//增加新的索引关系
	delete(Ctn_index_map, oldIndex)//删除旧的索引关系
}

//删除容器
func RemoveCtn(index int) string {
	if !IsCtnExisted(index) {
		err := fmt.Sprintf("容器%s不存在，无法执行删除操作!", index)
		return err
	} else {
		Mutex.Lock()
		defer Mutex.Unlock()
		pos:=Ctn_index_map[index].i//定位到待删除元素在切片中的位置
		ctn_used_status[pos] = false//设置该位置的标志量为未使用
		delete(Ctn_index_map, index)//删除索引表
	}
	return ""
}

//通过容器ID获取容器结构体
func GetCtn(ctnId string) *header.CTN {
	Mutex.Lock()
	defer Mutex.Unlock()
	for key, val := range Ctn_index_map {
		if val.CtnId == ctnId {
			//定位容器在切片中的位置
			pos:=Ctn_index_map[key].i
			return &Ctn_pool[pos]//返回容器结构体
		}
	}
	return nil
}

func GetCtnFromIndex(index int)*header.CTN{
	Mutex.Lock()
	defer Mutex.Unlock()

	_, ok:=Ctn_index_map[index]
	if !ok{
		return nil
	}

	pos:=Ctn_index_map[index].i
	return &Ctn_pool[pos]
}

//获取属于某服务的所有容器
func GetCtnsInService(sName string) []*header.CTN {
	Mutex.Lock()
	defer Mutex.Unlock()

	//遍历容器映射表，找出属于该服务的容器
	var ctns []*header.CTN
	for _, val := range Ctn_index_map {
		if val.ServiceName==sName{
			ctns=append(ctns,&Ctn_pool[val.i])
		}
	}
	return ctns
}


