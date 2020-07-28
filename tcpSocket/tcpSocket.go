// tcpSocket
package tcpSocket

import (
	"tcpSocket/tcpSocketConnection"
)

//优先级参数，从高到低
const (
	TCP_TPYE_CONTROLLER uint8 = 0
	TCP_TYPE_MONITOR    uint8 = 1
	TCP_TYPE_FILE       uint8 = 2
	TCP_TYPE_LOG        uint8 = 3
)

//connect状态信息
const (
	TCP_CONNECT_SUCCESS uint8 = 1
	TCP_DISCONNECT      uint8 = 2
)

type stringConnectionMap map[string]*ipConnection
type ipConnection struct {
	conn            *tcpSocketConnection.TcpConnection
	portBegin       int
	connectionCount int8
}

var ipMap = make(stringConnectionMap)
var ipTest string

//开启监听，参数为IP地址、起始端口号、接收数据的回调函数、状态改变的回调函数,返回值为监听的ip地址
func Listen(ip string, port int, funReceiveData tcpSocketConnection.UserReceiveData, funStateChange tcpSocketConnection.UserStateChange) string {
	data := new(tcpSocketConnection.UserData)
	data.ReceiveDataFun = funReceiveData
	data.StateChangeFun = funStateChange

	//返回IP地址，并以非阻塞的方式进行监听数据
	return tcpSocketConnection.Listen(ip, port, data, receive, change)
}

//连接到server，参数为IP地址、起始端口号、接收数据的回调函数、状态改变的回调函数
func ConnectToHost(ip string, port int, funReceiveData tcpSocketConnection.UserReceiveData, funStateChange tcpSocketConnection.UserStateChange) {
	data := new(tcpSocketConnection.UserData)
	data.ReceiveDataFun = funReceiveData
	data.StateChangeFun = funStateChange
	tcpSocketConnection.ConnectToHost(ip, port, data, receive, change)
}

//写数据，参数为IP地址、优先级等级、flag、数据体
//ip地址为状态改变时回调函数返回的ip
//flag必须为4个字符，不能多不能少
//优先级必须是定义的优先级其中之一
func WriteData(ip string, level uint8, pkgId uint16, flag string, data []byte) {
	if ipMap[ip] != nil {
		go ipMap[ip].conn.WriteData(level, pkgId, flag, data)
	}
}

//断开连接，参数为ip地址
func Abort(ip string) {
	ipMap[ip].conn.Abort()

}

//断开所有连接
func AbortAll() {
	for key, _ := range ipMap {
		Abort(key)
	}
}

//停止监听，参数为listen返回的ip地址
func StopListen(ip string) {
	tcpSocketConnection.StopListen(ip)
}

func receive(conn *tcpSocketConnection.TcpConnection, pkgId uint16, flag string, data []byte) {
	if conn.UserDataFun.ReceiveDataFun != nil {
		conn.UserDataFun.ReceiveDataFun(conn.RemoteIp, pkgId, flag, data)
	}
}

func change(conn *tcpSocketConnection.TcpConnection, state uint8) {

	if state == tcpSocketConnection.TCP_CONNECT_SUCCESS {
		//连接成功
		//连接不存在，创建连接
		if ipMap[conn.RemoteIp] == nil {
			ipConn := new(ipConnection)
			ipConn.conn = conn
			ipMap[conn.RemoteIp] = ipConn
		}
		//如果设置了回调函数，则调用
		if conn.UserDataFun.StateChangeFun != nil {
			conn.UserDataFun.StateChangeFun(conn.RemoteIp, tcpSocketConnection.TCP_CONNECT_SUCCESS)
		}
	} else if state == tcpSocketConnection.TCP_DISCONNECT {
		//连接断开
		//连接不存在，返回
		if ipMap[conn.RemoteIp] == nil {
			return
		}
		//如果设置了回调函数，则调用
		if conn.UserDataFun.StateChangeFun != nil {
			ipMap[conn.RemoteIp].conn.UserDataFun.StateChangeFun(conn.RemoteIp, tcpSocketConnection.TCP_DISCONNECT)
		}
		//释放连接
		ipMap[conn.RemoteIp] = nil
	}
}
