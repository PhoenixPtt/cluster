package servers

import (
	"sync"

	"clusterAgent/ctn"
)

// 记录当前已连接的所有的server的ip
var clusterServer map[string]*server = make(map[string]*server)
var clusterServerMapMutex sync.RWMutex

func newServer(h string, state bool) *server {
	s := new(server)
	s.gen(h, state)
	return s
}



// 添加一个集群
func Add(h string) {
	clusterServerMapMutex.Lock()
	defer clusterServerMapMutex.Unlock()

	s := clusterServer[h]
	if s == nil { // 如果集群不存在，则新建
		clusterServer[h] = newServer(h, true)
	} else { // 如果集群已经存在，只是更新状态
		s.SetState(true)
	}
}

// 把当前集群的所有容器都停止并删除
func RemoveServer(h string) {
	clusterServerMapMutex.Lock()
	defer clusterServerMapMutex.Unlock()

	s := clusterServer[h]
	if s == nil {
		return
	}

	// 遍历隶属于当前服务器的所有容器，依次remove
	for _,ctnId := range s.GetAllCtnId() {
		cctn := ctn.CTN{}
		cctn.ID = ctnId
		cctn.Remove()
	}

	delete(clusterServer, h)
}

// 给指定服务器新增一个容器
func AddServerCtn(h string, ctnId string, ctnState bool) {
	clusterServerMapMutex.Lock()
	defer clusterServerMapMutex.Unlock()

	s := clusterServer[h]
	if s == nil { // 如果当前服务器不存在，则先添加服务器
		s = newServer(h, true)
		clusterServer[h] = s
	}

	// 添加容器
	s.AddCtn(ctnId, ctnState)
}

// 指定服务器删除一个容器
func RemoveClusterCtn(h string, ctnId string) {
	clusterServerMapMutex.Lock()
	defer clusterServerMapMutex.Unlock()

	s := clusterServer[h]
	if s != nil {
		s.RemoveCtn(ctnId)
	}
}

// 指定服务器的某个容器状态变更
func UpdateClusterCtn(h string, ctnId string, state bool) {
	clusterServerMapMutex.Lock()
	defer clusterServerMapMutex.Unlock()
	s := clusterServer[h]
	if s != nil {
		s.SetCtnState(ctnId, state)
	}
}

func GetServer(h string) *server{
	clusterServerMapMutex.Lock()
	defer clusterServerMapMutex.Unlock()
	s := clusterServer[h]
	return s
}

// 返回所有的服务器句柄
func GetAllServerHandles() []string {
	clusterServerMapMutex.Lock()
	defer clusterServerMapMutex.Unlock()

	if len(clusterServer) == 0 {
		return []string{}
	}

	var hs []string = make([]string, len(clusterServer))

	i := 0
	for h,_ := range clusterServer {
		hs[i] = h
		i++
	}

	return hs
}

// 返回所有在线的服务器句柄
func GetOnlineServerHandles() []string {
	clusterServerMapMutex.Lock()
	defer clusterServerMapMutex.Unlock()

	if len(clusterServer) == 0 {
		return []string{}
	}

	var hs []string

	for h,s := range clusterServer {
		if s.State() == true {
			hs = append(hs, h)
		}
	}

	return hs
}

// 返回所有不在线的服务器句柄
func GetOfflineServerHandles() []string {
	clusterServerMapMutex.Lock()
	defer clusterServerMapMutex.Unlock()

	if len(clusterServer) == 0 {
		return []string{}
	}

	var hs []string

	for h,s := range clusterServer {
		if s.State() == false {
			hs = append(hs, h)
		}
	}

	return hs
}

