package clusterServer

import (
	header "clusterHeader"
	"github.com/shirou/gopsutil/host"
	"sync"
	"tcpSocket"
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

	node.mutex.Lock()
	defer node.mutex.Unlock()

	return node.State
}

func (n *Nodes)Count() int {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return len(n.nodes)
}