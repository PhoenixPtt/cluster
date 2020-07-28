package servers

import "sync"

type server struct {
	handle	string 		// 服务端的句柄
	state	bool 		// 状态 false-未连接 true-已连接
	ctn 	map[string]bool	// 归属于当前集群的容器，key:id value:状态,false-未运行 true-运行
	mutex   sync.RWMutex
}

func (s* server) gen(h string, state bool) {
	s.handle = h
	s.state = state
	s.ctn = make(map[string]bool)
}

// 返回服务器的句柄
func (s *server) Handle() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.handle
}

// 返回服务器的状态
func (s *server) State() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.state
}

// 返回所有服务器的容器ID和状态的map
func (s *server) GetAllCtn() map[string]bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.ctn
}

// 设置状态
func (s* server) SetState(state bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.state != state {
		s.state = state
	}
}

// 添加容器
func (s* server) AddCtn(id string, state bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.ctn[id] = state
}

// 设置容器状态
func (s* server) SetCtnState(id string, state bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _,ok := s.ctn[id]; ok {
		s.ctn[id] = state
	}
}

// 移除容器
func (s* server) RemoveCtn(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.ctn, id)
}


// 返回服务器的所有的容器ID
func (s *server) GetAllCtnId() []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var ids []string
	if len(s.ctn) == 0 {
		return ids
	} else {
		ids = make([]string, len(s.ctn))
	}

	i := 0
	for id,_ := range s.ctn {
		ids[i] = id
		i++
	}

	return ids
}