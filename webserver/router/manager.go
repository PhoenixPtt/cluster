// "manager.go" file is create by Huxd 2020.07.13

package router

import (
	header "clusterHeader"
	"log"
	"strings"
	"sync"
	"time"
)

// 信息管理者 对象结构体，包括管理的消息类型、客户列表(可扩展为添加/关闭/删除)、添加客户、删除客户、消息message集合、暂停标识
type Manager struct {
	msgType    string
	user       header.UserInformation // 请求用户的信息
	clients    map[chan Message]bool
	newClients chan chan Message
	delClients chan chan Message
	messages   chan *Message
	pauseFlag  chan bool
}

// 管理者对象列表操作互斥对象
var listOperaMutex sync.Mutex

// 管理者对象列表，存储管理者对象指针，这里必须使用make进行初始化，否则直接赋值时将出现问题
var managerList map[string]*Manager = make(map[string]*Manager)

// 创建新的信息管理者
func NewManager(maType string, user header.UserInformation) (*Manager, chan Message) {
	// 管理者对象列表操作锁定
	listOperaMutex.Lock()
	defer listOperaMutex.Unlock()

	// 查找是否已经创建对应类型的信息管理者，如果已经创建则直接返回
	manager, bOK := managerList[maType]
	if bOK {
		return manager, manager.addClient()
	}

	// 创建Manager实例
	manager = &Manager{
		msgType:    maType,
		user:       user,
		clients:    make(map[chan Message]bool),
		newClients: make(chan (chan Message)),
		delClients: make(chan (chan Message)),
		messages:   make(chan *Message, 10),
		pauseFlag:  make(chan bool, 1),
	}

	// 将该类型的信息管理者对象添加到信息管理者map中
	managerList[maType] = manager

	// 此时必然需要启动该类型的信息管理者协程
	go manager.start()
	// 此时必然是首次创建client，所以在这里添加client
	msg := manager.addClient()
	// 开始进行maType类型数据的获取
	go manager.getMessage()

	// 返回该信息管理者对象
	return manager, msg
}

// 运行信息管理者对象功能
func (m *Manager) start() {
	for {
		select {
		// 添加客户时，赋值并设定标志为true
		case cli := <-m.newClients:
			log.Printf("Manager(%v): add new client!", m.msgType)
			// 在客户map对象中添加新的客户对象，并设定标识为true
			m.clients[cli] = true

			// 设定暂停标志为false，在正常情况下false不会触发堵塞，(暂时不暂停而是停止故而屏蔽本行代码)
			//m.pauseFlag <- false	// <清零结束使用> 屏蔽

		// 删除客户时，从客户列表中移除该客户，并关闭客户消息通道对象
		case cli := <-m.delClients:
			log.Printf("Manager(%v): delete client!", m.msgType)
			// 删除客户，关闭客户消息通道
			delete(m.clients, cli)
			close(cli)

			// 如果客户列表为空，则指示对应的获取消息协程休眠
			if len(m.clients) == 0 {
				// 准备结束并删除管理者列表中的该消息对应的管理者对象
				listOperaMutex.Lock()
				defer listOperaMutex.Unlock()

				log.Printf("Manager(%v): delete from list", m.msgType)
				delete(managerList, m.msgType)

				// 休眠对应的getMessage函数，可以在重新启用之前一直堵塞
				m.pauseFlag <- true
				return // <清零结束使用> 由于此时需要结束，故而跳出循环结束本协程
			}
		// 获取新消息时，通过遍历，将信息发给所有的客户
		case message := <-m.messages:
			log.Printf("Manager(%v) flush message, client count is %v", m.msgType, len(m.clients))
			for cli := range m.clients {
				cli <- *message
			}
		}
	}
}

// 向信息管理者对象中添加客户
func (m *Manager) addClient() chan Message {
	// Create a new channel, over which the Manager can send this client messages.
	newClientChan := make(chan Message)

	// Add this client to the map of those that should receive updates.
	m.newClients <- newClientChan

	// return chan Message object
	return newClientChan
}

// 向信息管理者对象中删除客户
func (m *Manager) delClient(cli chan Message) {
	// Remove this client from the map of attached clients
	m.delClients <- cli
}

// 根据msgType生成requestInf对象
func (m *Manager) getRequestInformation() (req requestInf) {
	// 解析msgType的内容，正常情况下msgType为/XXXXX/YYYY这种格式
	data := strings.Split(m.msgType, "/")

	// 解析后的字符串切片小于3个，则直接返回，因为解析错误
	if len(data) < 3 {
		return
	}

	// 根据msgType的内容，填充req对象
	req.typeFlag = data[1]
	req.opertype = data[2]

	req.user = m.user	// 赋值用户信息

	return req
}

// 获取待发送给客户的消息
func (m *Manager) getMessage() {
	// 调试使用i变量实现计数功能 <正式版本，或需删除>
	var i int = 0

	// 获取请求结构体对象
	req := m.getRequestInformation()

	// 循环获取待发送的数据内容
	for {
		// 计数功能变量自加1 <正式版本，或需删除>
		i++

		// 暂停操作处理
		select {
		case pause := <-m.pauseFlag:
			// 如果暂停标识为true，则进行处理
			if pause {
				// 关闭两个通道对象，并结束协程
				close(m.messages)
				close(m.pauseFlag)
				return

				// 下面是暂停的操作代码
				//log.Printf("Manager(%v): continue getMessage is pause", m.msgType)
				// 循环读取暂停标识，如果标识为false则跳出循环
				//for  {
				//	pause = <-m.pauseFlag
				//	if !pause {
				//		log.Printf("Manager(%v): restart to continue getMessage", m.msgType)
				//		break
				//	}
				//}
			}
		// 默认不进行任何处理，则可以正常进行信息获取
		default:
		}

		// 从集群服务端获取指定类型的信息
		bOK, msg := getMessage(req)

		// 将msg返回的信息，写入到消息数组中
		m.messages <- msg

		// 如果正常获取信息，则显示信息内容
		if bOK {
			// Print a nice log message
			log.Printf("%v msgType is %v, getMessage: %v", i, m.msgType, msg.content)
		} else { // 如果获取的数据出现错误，是否需要在这里处理？
			// Print a error log message
			log.Printf("%v msgType is %v, getMessage is error: %v", i, m.msgType, msg.errorMsg.Error())
		}

		//  临时使用，等待2s，正式版本由集群服务端控制数据间隔
		time.Sleep(2e9)
	}
}
