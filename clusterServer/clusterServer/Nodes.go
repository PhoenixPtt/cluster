package clusterServer

import (
	header "clusterHeader"
	"github.com/shirou/gopsutil/host"
	"sync"
	"tcpSocket"
	"time"
)

type Node struct {
	header.Node
	mutex sync.RWMutex
}

type Nodes struct {
	nodes map[string]*Node
	mutex sync.RWMutex
}

// 初始化
func (n *Nodes)Init() {
	n.nodes = make(map[string]*Node)
}

// 添加节点
func (n *Nodes)AddNode(h string, state bool) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	node := n.nodes[h]
	if node == nil {
		node = new(Node)
		node.State = state
		node.Handle = h
		n.nodes[h] = node
	} else {
		node.State = state
	}
}

// 移除节点
func (n *Nodes)RemoveNode(h string) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	node := n.nodes[h]
	if node == nil {
		return
	} else {
		tcpSocket.Abort(h)
		delete(n.nodes, h)
	}
}

// 设置节点状态
func (n *Nodes)SetNodeState(h string, state bool) {
	node := n.GetNode(h)
	if node == nil {
		return
	}

	node.mutex.Lock()
	defer node.mutex.Unlock()

	node.State = state
}

func (n *Nodes)GetNodeIds() []string {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	if len(n.nodes) == 0 {
		return []string{}
	}

	ids := make([]string, len(n.nodes))
	i := 0
	for id,_ := range n.nodes {
		ids[i] = id
		i++
	}
	return ids
}

func (n *Nodes)GetNode(h string) *Node {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.nodes[h]
}

func (n *Nodes)GetState(h string) bool {
	node := n.GetNode(h)
	if node == nil {
		return false
	}

	node.mutex.RLock()
	defer node.mutex.RUnlock()

	return node.State
}

func (n *Nodes)SetResStates(h string, res header.ResourceStatus) {
	node := n.GetNode(h)
	if node == nil {
		return
	}

	node.mutex.Lock()
	defer node.mutex.Unlock()

	node.Res = res
}

func (n *Nodes)SetHostInfo(h string, hostInfo host.InfoStat) {
	node := n.GetNode(h)
	if node == nil {
		return
	}

	node.mutex.Lock()
	defer node.mutex.Unlock()

	node.HostInfo = hostInfo
}

func (n *Nodes)AddLabel(h string, label header.NodeLabel) {
	node := n.GetNode(h)
	if node == nil {
		return
	}

	node.mutex.Lock()
	defer node.mutex.Unlock()

	// 遍历标签，查看当前标签是否已经存在，若存在则更新，则不存在则新增
	labelCount := len(node.Labels)
	for i := 0; i<labelCount; i++ {
		if node.Labels[i].Name == label.Name {
			node.Labels[i].Value = label.Value
			return
		}
	}

	node.Labels = append(node.Labels, label)
}

func (n *Nodes)AddLabels(h string, labels []header.NodeLabel) {
	node := n.GetNode(h)
	if node == nil {
		return
	}

	node.mutex.Lock()
	defer node.mutex.Unlock()

	// 遍历标签，查看当前标签是否已经存在，若存在则更新，则不存在则新增
	labelCount := len(node.Labels)
	for j := 0; j<len(labels); j++ {
		labelExist := false

		for i := 0; i<labelCount; i++ {
			if node.Labels[i].Name == labels[j].Name {
				node.Labels[i].Value = labels[j].Value
				labelExist = true
				break
			}
		}

		if labelExist == false {
			node.Labels = append(node.Labels, labels[j])
		}
	}
}

func (n *Nodes)ClearLabels(h string) {
	node := n.GetNode(h)
	if node == nil {
		return
	}

	node.mutex.Lock()
	defer node.mutex.Unlock()

	node.Labels = []header.NodeLabel{}
}

func (n *Nodes)RemoveLabel(h string, label header.NodeLabel) {
	node := n.GetNode(h)
	if node == nil {
		return
	}

	node.mutex.Lock()
	defer node.mutex.Unlock()

	// 遍历标签，查看当前标签是否已经存在，若存在则更新，则不存在则新增
	labelCount := len(node.Labels)
	for i := 0; i<labelCount; i++ {
		if node.Labels[i].Name == label.Name {
			if i == 0 {
				node.Labels = node.Labels[i+1:]
			} else if i == labelCount-1 {
				node.Labels = node.Labels[:i-1]
			} else {
				node.Labels = append(node.Labels[:i-1], node.Labels[i+1:]...)
			}
			return
		}
	}
}

func (n *Nodes)RemoveLabels(h string, labels []header.NodeLabel) {
	node := n.GetNode(h)
	if node == nil {
		return
	}

	node.mutex.Lock()
	defer node.mutex.Unlock()

	// 遍历标签，查看当前标签是否已经存在，若存在则更新，则不存在则新增
	labelCount := len(node.Labels)
	for j := 0; j<len(labels); j++ {
		for i := 0; i<labelCount; i++ {
			if node.Labels[i].Name == labels[j].Name {
				if i == 0 {
					node.Labels = node.Labels[i+1:]
				} else if i == labelCount-1 {
					node.Labels = node.Labels[:i-1]
				} else {
					node.Labels = append(node.Labels[:i-1], node.Labels[i+1:]...)
				}
				break
			}
		}
	}
}

func (n *Nodes)IsOnline(h string) bool{
	node := n.GetNode(h)
	if node == nil {
		return false
	}

	node.mutex.RLock()
	defer node.mutex.RUnlock()

	return node.State
}

func (n *Nodes)Count() int {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return len(n.nodes)
}

func (n* Nodes)ResCount() *header.ResourceStatus {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	res := new(header.ResourceStatus)

	nodeCount := len(n.nodes)
	if nodeCount == 0 {
		return res
	}

	for _,node := range n.nodes {
		node.mutex.RLock()

		cpu := &node.Res.Cpu
		res.Cpu.CoreCount += cpu.CoreCount
		res.Cpu.UsedPercent.Val += cpu.UsedPercent.Val * float64(cpu.CoreCount)
		res.Cpu.Temperature.Val += cpu.Temperature.Val
		res.Cpu.Health.Val += cpu.Health.Val

		mem := &node.Res.Mem
		res.Mem.Used.Val += mem.Used.Val
		res.Mem.Health.Val += mem.Health.Val
		res.Mem.Temperature.Val += mem.Temperature.Val
		res.Mem.Total.Val += mem.Total.Val

		disk := &node.Res.Disk
		res.Disk.Used.Val += disk.Used.Val
		res.Disk.Health.Val += disk.Health.Val
		res.Disk.Temperature.Val += disk.Temperature.Val
		res.Disk.Total.Val += disk.Total.Val

		node.mutex.RUnlock()
	}

	ndw := 1.0/float64(nodeCount)

	res.Cpu.SetUsedPercentData(res.Cpu.UsedPercent.Val / float64(res.Cpu.CoreCount))
	res.Cpu.SetTemperatureData(res.Cpu.Temperature.Val * ndw)
	res.Cpu.SetHealthData(res.Cpu.Health.Val * ndw)

	res.Mem.SetHealthData(res.Mem.Health.Val * ndw)
	res.Mem.SetTemperatureData(res.Mem.Temperature.Val * ndw)
	res.Mem.SetUsedData(uint64(res.Mem.Used.Val), uint64(res.Mem.Total.Val))

	res.Disk.SetHealthData(res.Disk.Health.Val * ndw)
	res.Disk.SetTemperatureData(res.Disk.Temperature.Val * ndw)
	res.Disk.SetUsedData(uint64(res.Disk.Used.Val), uint64(res.Disk.Total.Val))

	res.Time = time.Now().Format("yyyy-MM-dd hh:mm:ss.zzzzzzzzz")
	return res
}