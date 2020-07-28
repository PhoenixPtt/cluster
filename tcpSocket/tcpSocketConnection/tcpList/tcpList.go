// tcpList
//也可以在链表中添加几个虚拟数据，当做优先级的标志
package tcpList

import (
	"fmt"
	"sync"
	// "time"
)

//优先级设置
const (
	TCP_TPYE_CONTROLLER uint8 = 0
	TCP_TYPE_MONITOR    uint8 = 1
	TCP_TYPE_FILE       uint8 = 2
	TCP_TYPE_LOG        uint8 = 3
)

//优先级对应元素
type levelElementMap map[uint8]*Element

//元素结构体
type Element struct {
	prev  *Element //前一个指针
	next  *Element //后一个指针
	value []byte   //数据
}

//链表结构体
type TcpList struct {
	begin    *Element
	end      *Element
	levelMap levelElementMap
	cond     sync.Cond
	len      int
}

//新建结构体
func New() *TcpList {
	list := new(TcpList)
	list.Clear()
	list.levelMap = make(levelElementMap)
	list.cond.L = new(sync.Mutex)

	return list
}

func (list *TcpList) Clear() {
	//如果链表没有首地址，则返回
	if list.begin == nil {
		return
	}

	//将第二个元素的前一元素设置为空
	if firstelem := list.begin.next; firstelem != nil {
		firstelem.prev = nil
	}

	//将第一个元素设置为空
	list.begin.next = nil
	list.begin.prev = nil
	list.begin = nil
	list.len = 0
}

//输出链表
func (list *TcpList) Show() {
	for it := list.begin; it != nil; it = it.next {
		fmt.Println(it.value)
	}
}

//添加元素，参数为值和优先级
func (list *TcpList) PushData(data []byte, level uint8) {

	//生成元素
	eData := &Element{value: data}
	//遍历优先级等级
	curLevel := level

	list.cond.L.Lock()
	defer list.cond.L.Unlock()
	for {
		//如果当前优先级下没有数据
		if list.levelMap[curLevel] == nil {
			if curLevel == 0 {
				//已经是最高优先级，且该优先级下没有数据，则元素为首个元素
				nbak := list.begin
				list.begin = eData
				eData.prev = nil
				eData.next = nbak
				//设置该元素为所属优先级的最后一个元素
				list.levelMap[level] = eData
				break
			} else {
				//如果不是最高优先级，且优先级中没有数据，继续查找更高优先级中有没有数据
				curLevel--
				continue
			}
		} else {
			//如果当前优先级下存在数据，将此数据插入到目前优先级对尾
			nbak := list.levelMap[curLevel].next
			list.levelMap[curLevel].next = eData
			eData.prev = list.levelMap[curLevel]
			eData.next = nbak
			//设置该元素为所属优先级的最后一个元素
			list.levelMap[level] = eData
			break
		}
	}
	//队列中元素加一
	list.len++

	if list.len == 1 {
		list.cond.Signal()
	}
}

//获取优先级最高的元素
func (list *TcpList) GetData() []byte {

	list.cond.L.Lock()
	defer list.cond.L.Unlock()

	//如果列表为空，则返回空
	if list.IsEmpty() {
		list.cond.Wait()
	}

	//获取首个元素，即优先级最高的元素
	eData := list.begin

	//如果有第二个元素，将第二元素的前一个元素设置为空
	if eData.next != nil {
		eData.next.prev = nil
	}
	//将第二个元素设置为首元素
	list.begin = eData.next

	//查询该元素是否为某一优先级的尾部数据

	for level := TCP_TPYE_CONTROLLER; level <= TCP_TYPE_LOG; level++ {
		//如果找到则讲该优先级尾部元素的指针设置为空
		if list.levelMap[level] == eData {
			list.levelMap[level] = nil
		}
	}

	//元素个数减一
	list.len--

	return eData.value

}

//获取List中元素的个数
func (list *TcpList) Size() int {
	return list.len
}

//获取List是否wei空
func (list *TcpList) IsEmpty() bool {
	if list.len == 0 {
		return true
	}
	return false
}
