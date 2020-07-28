package tcpSocketConnection

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	"tcpSocket/tcpSocketConnection/tcpList"
)

//tcp连接信息结构体
type TcpConnection struct {
	RemoteIp    string
	RemotePort  int
	TcpType     uint8
	UserDataFun *UserData

	package_No         uint16
	conn               net.Conn
	receiveDataFun     receiveData
	stateChangeFun     stateChange
	data_cache         []byte
	data_package_map   map[uint16]([][]byte)
	data_package_count uint16
	writeCache         *tcpList.TcpList
	stopped            bool
}

type TcpListener struct {
	Listener net.Listener
	quit     chan struct{}
}

type receiveData func(*TcpConnection, uint16, string, []byte)
type stateChange func(*TcpConnection, uint8)

const (
	TCP_CONNECT_SUCCESS uint8 = 1
	TCP_DISCONNECT      uint8 = 2
)

type UserReceiveData func(string, uint16, string, []byte)
type UserStateChange func(string, uint8)
type UserData struct {
	ReceiveDataFun UserReceiveData
	StateChangeFun UserStateChange
}

var (
	listenerMap = make(map[string]*TcpListener)
)

func Listen(IpAddress string, Port int, userData *UserData, funReceiveData receiveData, funStateChange stateChange) string {

	ipString := IpAddress + ":" + strconv.Itoa(Port)
	lner, err := net.Listen("tcp", ipString)

	if err != nil {
		fmt.Println("listener creat error", err)
		return ""
	}

	fmt.Println(lner.Addr())

	listener := new(TcpListener)
	listener.Listener = lner
	listener.quit = make(chan struct{})
	listenerMap[lner.Addr().String()] = listener

	go listener.listenConn(userData, funReceiveData, funStateChange)
	return lner.Addr().String()

}

func ConnectToHost(tarIp string, tarPort int, srcIp string, srcPort int, userData *UserData, funReceiveData receiveData, funStateChange stateChange) {

	remoteIp := tarIp + ":" + strconv.Itoa(tarPort)
	localIp  := srcIp + ":" + strconv.Itoa(srcPort)

	remoteAddr, remoteErr := net.ResolveTCPAddr("tcp", remoteIp)
	if remoteErr != nil {
		fmt.Println("Resolve TCPAddrRemote error", remoteErr)
		return
	}

	localAddr, localErr := net.ResolveTCPAddr("tcp", localIp)
	if localErr != nil {
		fmt.Println("Resolve TCPAddrRemote error", localErr)
		return
	}

	conn, err := net.DialTCP("tcp4", localAddr, remoteAddr)

	if err != nil {
		fmt.Println("connect server error", err)
		return
	}

	remote := strings.Split(conn.RemoteAddr().String(),":")

	tcpConnect := new(TcpConnection)
	tcpConnect.UserDataFun = userData
	tcpConnect.RemoteIp = remote[0]
	tcpConnect.RemotePort, _ = strconv.Atoi(remote[1])
	tcpConnect.conn = conn
	tcpConnect.receiveDataFun = funReceiveData
	tcpConnect.stateChangeFun = funStateChange
	tcpConnect.writeCache = tcpList.New()
	tcpConnect.data_package_map = make(map[uint16]([][]byte))
	tcpConnect.stopped = false
	if funStateChange != nil {
		funStateChange(tcpConnect, TCP_CONNECT_SUCCESS)
	}

	go tcpConnect.readData()
	go tcpConnect.writeFromCache()
}

func (lner *TcpListener) listenConn(userData *UserData, funReceiveData receiveData, funStateChange stateChange) {
	for {
		conn, err := lner.Listener.Accept()
		if err != nil {
			select {
			default:
			case <-lner.quit:
				fmt.Println("listenClose")
				lner = nil
				return
			}
			fmt.Println("accept error", err)
			continue
		}

		remote := strings.Split(conn.RemoteAddr().String(),":")

		tcpConnect := new(TcpConnection)
		tcpConnect.UserDataFun = userData
		tcpConnect.RemoteIp = remote[0]
		tcpConnect.RemotePort, _ = strconv.Atoi(remote[1])
		tcpConnect.conn = conn
		tcpConnect.receiveDataFun = funReceiveData
		tcpConnect.stateChangeFun = funStateChange
		tcpConnect.writeCache = tcpList.New()
		tcpConnect.data_package_map = make(map[uint16]([][]byte))
		tcpConnect.stopped = false

		if funStateChange != nil {
			funStateChange(tcpConnect, TCP_CONNECT_SUCCESS)
		}
		go tcpConnect.readData()
		go tcpConnect.writeFromCache()
	}
}

func (conn *TcpConnection) WriteData(level uint8, pkgId uint16, flag string, data []byte) {

	data_len := len(data)
	curPackage_no := conn.package_No
	conn.package_No++

	var package_size uint16
	var package_num uint16 = 0
	var package_count uint16 = uint16(math.Ceil(float64(data_len) / 64000.0))
	if package_count == 0 {
		package_count = 1
	}
	var i uint64 = 0
	for ; i < uint64(package_count); i++ {

		var send_data [][]byte = make([][]byte, 2)

		if i == (uint64(package_count) - 1) {
			package_size = uint16(data_len)
			send_data[1] = data[i*64000:]

		} else {
			package_size = 64000
			data_len -= 64000
			send_data[1] = data[i*64000 : (i+1)*64000]
		}

		package_size += 13

		byteBuffer := bytes.NewBuffer([]byte{})
		binary.Write(byteBuffer, binary.BigEndian, &package_size)
		binary.Write(byteBuffer, binary.BigEndian, uint8(12))
		binary.Write(byteBuffer, binary.BigEndian, &pkgId)
		byteBuffer.Write([]byte(flag))
		binary.Write(byteBuffer, binary.BigEndian, &package_count)
		binary.Write(byteBuffer, binary.BigEndian, &package_num)

		binary.Write(byteBuffer, binary.BigEndian, &curPackage_no)
		// binary.Write(byteBuffer, binary.BigEndian, uint16(22))

		send_data[0] = byteBuffer.Bytes()
		conn.writeCache.PushData(bytes.Join(send_data, []byte{}), level)
		package_num++
	}
}

func (conn *TcpConnection) Abort() {
	conn.conn.Close()
}

func StopListen(IpAddress string) {
	if listenerMap[IpAddress] != nil {
		close(listenerMap[IpAddress].quit)
		listenerMap[IpAddress].Listener.Close()
	}

}

func (conn *TcpConnection) readData() {
	//1.conn是否有效
	if conn.conn == nil {
		log.Panic("无效的 socket 连接")
		return
	}

	//2.新建网络数据流存储结构
	buf := make([]byte, 64013)
	// var buf []byte
	//3.循环读取网络数据流
	for {
		//3.1 网络数据流读入 buffer
		cnt, err := conn.conn.Read(buf)
		//3.2 数据读尽、读取错误

		if err != nil {
			conn.stateChangeFun(conn, TCP_DISCONNECT)
			conn.data_package_map = nil
			conn.data_cache = nil
			defer conn.conn.Close()
			break
		} else if cnt == 0 {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		conn.data_cache = append(conn.data_cache, buf[:cnt]...)
		conn.analysis()
	}
}

func (conn *TcpConnection) analysis() {
	bytesBuffer := bytes.NewReader(conn.data_cache)
	data_surplus := bytesBuffer.Size()

	var data_len uint16
	var header_len uint8
	var pkgId uint16
	var flag [4]byte
	var data_count uint16
	var data_num uint16
	var package_no uint16

	for data_surplus > 2 {
		binary.Read(bytesBuffer, binary.BigEndian, &data_len)
		data_surplus -= 2
		if data_surplus < int64(data_len) {
			return
		}
		data := make([]byte, data_len-13)
		binary.Read(bytesBuffer, binary.BigEndian, &header_len)
		binary.Read(bytesBuffer, binary.BigEndian, &pkgId)
		binary.Read(bytesBuffer, binary.BigEndian, &flag)
		binary.Read(bytesBuffer, binary.BigEndian, &data_count)
		binary.Read(bytesBuffer, binary.BigEndian, &data_num)
		binary.Read(bytesBuffer, binary.BigEndian, &package_no)
		binary.Read(bytesBuffer, binary.BigEndian, &data)

		if data_count > 1 {
			conn.data_package_map[package_no] = append(conn.data_package_map[package_no], data)
			if data_num == data_count-1 { //daiceshi
				if conn.receiveDataFun != nil {
					go conn.receiveDataFun(conn, pkgId, string(flag[0:4]), bytes.Join(conn.data_package_map[package_no], []byte{}))
					conn.data_package_map[package_no] = nil
				}
			}
		} else {
			if conn.receiveDataFun != nil {
				go conn.receiveDataFun(conn, pkgId, string(flag[0:4]), data)
			}
		}

		conn.data_cache = conn.data_cache[2+data_len:]
		data_surplus -= int64(data_len)
	}
}

func (conn *TcpConnection) writeFromCache() {
	conn.stopped = false

	for !conn.stopped {
		// data := conn.writeCache.GetData()
		// fmt.Println(len(data))

		_, err := conn.conn.Write(conn.writeCache.GetData())
		if err != nil {
			fmt.Println(err.Error())
			conn.stateChangeFun(conn, TCP_DISCONNECT)
			defer conn.conn.Close()
			conn.data_package_map = nil
			conn.data_cache = nil
			conn.stopped = true
			break
		}
	}
}
