package clusterServer

import (
	"clusterHeader"
	"container/list"
	"sync"
)

const (
	CPUTYPE = "CPU"
	MEMTYPE = "内存"
	FILESYSTEMTYPE = "硬盘"
	APPSERVICETYPE = "应用服务"
)

// 告警信息，
type Warnings struct {
	count         uint32           // 总数量
	maxLevel      uint8            // 最大等级
	countPerLevel map[uint8]*header.WarningCountOfType // 每个级别的数量
	countPerNode  map[string]*header.WarningCountOfType // 每个节点的告警数量
	list          list.List        // 所有的报警列表，只记录最近
	listCount     uint32           // 当前列表条目数量
	hasChanged    bool             // 标识是否有新增告警条目

	info 		header.WarningInfo

	mutex   	sync.RWMutex
}

// 发送给Client的警告信息
func (w *Warnings) WarningInfo() *header.WarningInfo {

	// 如果告警条目发生变化
	if w.hasChanged == true {

		// 更新等级数量
		levelCount := w.maxLevel+1
		if levelCount < header.WARNING_LEVEL_COUNT {
			levelCount = header.WARNING_LEVEL_COUNT
		}

		// 如果告警条目发生变化，则重新更新填充
		w.info.Count = w.count

		// 每个级别的信息填充
		w.info.CountPerLevel = make([]header.WarningCountOfType, levelCount)
		for i:=uint8(0); i<levelCount; i++ {
			w.info.CountPerLevel[i] = *w.countPerLevel[i]
		}

		// 每个节点的信息填充
		w.info.CountPerNode = make([]header.WarningCountOfType, len(w.countPerNode))
		i := 0
		for _,wct := range w.countPerNode {
			w.info.CountPerNode[i] = *wct
			i++
		}

		i = w.list.Len()-1
		w.info.Warning = make([]header.WarningItem, w.list.Len())
		for e := w.list.Front(); e != nil; e = e.Next() {
			w.info.Warning[i] = e.Value.(header.WarningItem)
			i--
		}
	}

	return &w.info
}

func (w *Warnings) Count() uint32 {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return uint32( w.list.Len() )
}

// 添加一条记录
func (w *Warnings) Add(item header.WarningItem) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// 总告警数量+1
	w.count++

	// 缓存一条条目
	w.list.PushBack(item)
	if uint32(w.list.Len()) > header.WARNING_MAX_COUNT {
		w.list.Remove(w.list.Front())
	}

	// 根据告警条目的等级分别累计数量
	if w.countPerLevel == nil {
		w.countPerLevel = make(map[uint8]*header.WarningCountOfType)
	}
	wct := w.countPerLevel[item.Level]
	if wct == nil {
		wct = new(header.WarningCountOfType)
		w.countPerLevel[item.Level] = wct
	}
	wct.All++
	switch item.Type {
	case CPUTYPE:			wct.Cpu++
	case MEMTYPE:			wct.Mem++
	case FILESYSTEMTYPE:	wct.FileSystem++
	case APPSERVICETYPE:	wct.AppService++
	default: 				wct.Other++
	}

	// 根据告警条目的节点分别累计数量
	if w.countPerNode == nil {
		w.countPerNode = make(map[string]*header.WarningCountOfType)
	}
	wct = w.countPerNode[item.NodeId]
	if wct == nil {
		wct = new(header.WarningCountOfType)
		w.countPerNode[item.NodeId] = wct
	}
	wct.All++
	switch item.Type {
	case CPUTYPE:			wct.Cpu++
	case MEMTYPE:			wct.Mem++
	case FILESYSTEMTYPE:	wct.FileSystem++
	case APPSERVICETYPE:	wct.AppService++
	default: 				wct.Other++
	}

	// 更新最大警告级别
	if item.Level > w.maxLevel {
		w.maxLevel = item.Level
	}

	// 标识条目变化
	w.hasChanged = true
}

// 清空所有记录
func (w *Warnings) Clear() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.maxLevel = header.WARNING_LEVEL_COUNT
	w.count = 0
	w.countPerLevel = make(map[uint8]*header.WarningCountOfType)
	w.countPerNode = make(map[string]*header.WarningCountOfType)
	w.list = list.List{}
}
