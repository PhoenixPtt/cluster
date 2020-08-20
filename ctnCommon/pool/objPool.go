package pool

import (
	"errors"
	"fmt"
	"sync"
)

const(
	MAX_OBJ_NUM = 1000
)

type CtnPoolOper interface {
	AddObj(objName string, pObj *interface{})
	RemoveObj(objName string) string
	GetObj(objName string) *interface{}
	GetObjNames() []string
}

type OBJ_POOL struct{
	obj_pool []interface{}			//对象池
	obj_used_status []bool			//对象池使用状态
	obj_map map[string]int			//对象索引表
	mutex sync.Mutex
}

//添加容器
func (objPool *OBJ_POOL) AddObj(objName string, pObj interface{}) {
	if objPool.isCtnExisted(objName) {
		fmt.Printf("%s已存在，不能重复添加!", objName)
	} else {
		objPool.mutex.Lock()
		defer objPool.mutex.Unlock()
		//判断已分配的存储空间中，切片有无未使用的
		unusedIndex:=-1//第一个未使用的存储空间的序号
		for key,usedStatus:=range objPool.obj_used_status{
			if usedStatus==false{
				unusedIndex=key
			}
		}

		if unusedIndex==-1{//切片中目前已经分配的空间已经使用完毕
			ctnLen:=len(objPool.obj_pool)
			objPool.obj_pool = append(objPool.obj_pool, pObj)
			objPool.obj_used_status = append(objPool.obj_used_status, true)
			objPool.obj_map[objName]=ctnLen//更新索引表
		}else{
			objPool.obj_pool[unusedIndex]=pObj//放置在未使用的位置
			objPool.obj_used_status[unusedIndex]=true//设置标志量为已使用
			objPool.obj_map[objName]=unusedIndex//更新索引表
		}
	}
}

//删除容器
func (objPool *OBJ_POOL) RemoveObj(objName string) error {
	if !objPool.isCtnExisted(objName) {
		errStr := fmt.Sprintf("容器%s不存在，无法执行删除操作!", objName)
		err := errors.New(errStr)
		return err
	} else {
		objPool.mutex.Lock()
		defer objPool.mutex.Unlock()
		pos:=objPool.obj_map[objName]//定位到待删除元素在切片中的位置
		objPool.obj_used_status[pos] = false//设置该位置的标志量为未使用
		delete(objPool.obj_map, objName)//删除索引表
	}
	return nil
}

func (objPool *OBJ_POOL) getObjPos(objName string) int {
	objPool.mutex.Lock()
	defer objPool.mutex.Unlock()

	_, ok:=objPool.obj_map[objName]
	if !ok{
		return -1
	}
	return objPool.obj_map[objName]
}

//通过容器ID获取容器结构体
func (objPool *OBJ_POOL)  GetObj(objName string) interface{} {
	pos := objPool.getObjPos(objName)
	if pos==-1{
		return nil
	}

	objPool.mutex.Lock()
	defer objPool.mutex.Unlock()
	return objPool.obj_pool[pos]
}

func (objPool *OBJ_POOL) GetObjNames() (objNames []string) {
	objNames = make([]string, 0, MAX_OBJ_NUM)

	objPool.mutex.Lock()
	defer objPool.mutex.Unlock()
	for objName,_:=range objPool.obj_map{
		objNames = append(objNames, objName)
	}

	return objNames
}

//判断容器是否存在
func (objPool *OBJ_POOL) isCtnExisted(objName string) bool {
	objPool.mutex.Lock()
	defer objPool.mutex.Unlock()

	//遍历容器索引表
	_,ok:=objPool.obj_map[objName]
	if ok{
		return true
	}
	return false
}

func (objPool *OBJ_POOL) Lock()  {
	objPool.mutex.Lock()
}

func (objPool *OBJ_POOL) Unlock()  {
	objPool.mutex.Unlock()
}

var (
	objPool *OBJ_POOL
)

func init()  {
	objPool = NewObjPool()
}

func NewObjPool() (objPool *OBJ_POOL) {
	objPool = &OBJ_POOL{}
	objPool.obj_pool = make([]interface{}, MAX_OBJ_NUM)
	objPool.obj_used_status = make([]bool, MAX_OBJ_NUM)
	objPool.obj_map = make(map[string]int,MAX_OBJ_NUM)
	return
}

//添加容器
func AddObj(objName string, pObj interface{}) {
	objPool.AddObj(objName, pObj)
}

//删除容器
func RemoveObj(objName string) error {
	return objPool.RemoveObj(objName)
}

//通过容器ID获取容器结构体
func GetObj(objName string) interface{} {
	return objPool.GetObj(objName)
}

func GetObjNames() (objNames []string) {
	return objPool.GetObjNames()
}

func Lock()  {
	objPool.Lock()
}

func Unlock()  {
	objPool.Unlock()
}





