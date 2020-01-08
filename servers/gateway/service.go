package main

import (
	"bufio"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/network"
)

const (
	ConnectionPollLen = 1000000
)

const (
	CrossDomainXml = `<cross-domain-policy><allow-access-from domain="*" to-ports="*" /><site-control permitted-cross-domain-policies="all" /></cross-domain-policy>`
)

type IConnection interface {
	Serve()
	SetId(id logic.ConnectionId)
	GetId() logic.ConnectionId
	Operate(gwr *router.GwResp, async bool) // 操作集合
	IsClose() bool                          // 些连接是否已经关闭，为了防止tag上的坏引用
	Close()
}

type ConnectionPool struct {
	sync.RWMutex
	connections map[logic.ConnectionId]IConnection
}

func newConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections: make(map[logic.ConnectionId]IConnection, ConnectionPollLen),
	}
}

func (this *ConnectionPool) AddConnection(connection IConnection) {
	this.Lock()
	defer this.Unlock()
	this.connections[connection.GetId()] = connection
}

func (this *ConnectionPool) DelConnection(connectionId logic.ConnectionId) {
	this.Lock()
	defer this.Unlock()
	delete(this.connections, connectionId)
}

func (this *ConnectionPool) Connection(connectionId logic.ConnectionId) IConnection {
	this.RLock()
	defer this.RUnlock()
	if iConnection, ok := this.connections[connectionId]; ok {
		return iConnection
	}
	return nil
}
func (this *ConnectionPool) GetLen() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.connections)
}

func (this *ConnectionPool) Connections() []IConnection {
	this.RLock()
	defer this.RUnlock()
	conns := make([]IConnection, 0, len(this.connections))
	for _, v := range this.connections {
		conns = append(conns, v)
	}
	return conns
}

// 无锁实现，设置和获取的时候小心是否会同时发生
type Connection struct {
	// IConnection接口的实现
	connectionId logic.ConnectionId
}

func newConnection() *Connection {
	return &Connection{}
}

func (this *Connection) SetId(id logic.ConnectionId) {
	this.connectionId = id
}

func (this *Connection) GetId() logic.ConnectionId {
	return this.connectionId
}

// 如果把stopAcceptLock设置成1，表示停止接收请求
var stopAcceptLock int32 = 0

func IsStopAccept() bool {
	return atomic.LoadInt32(&stopAcceptLock) != 0
}

func StopAccept() {
	atomic.StoreInt32(&stopAcceptLock, 1)
}

func StartAccept() {
	atomic.StoreInt32(&stopAcceptLock, 0)
}

func StopListen() {
	for _, v := range tcpListeners {
		v.Close()
	}
}

var tcpListeners []*network.TcpListener

func FrontServer() {
	if len(netConf().MultiListen) != 0 && netConf().MultiListen[0] != "" {
		for _, v := range netConf().MultiListen {
			ListenAndAccept(v)
		}
	} else if netConf().Listen != "" {
		ListenAndAccept(netConf().Listen)
	} else {
		panic("empty listen ip")
	}
}

func ListenAndAccept(addr string) {
	l, err := network.TcpListen(addr)
	if err != nil {
		panic("invalid listen ip :" + err.Error())
	} else {
		Logger.Trace(netConf().MultiListen, netConf().Listen, addr, "ListenAndAccept", "listen", "ok")
	}
	tcpListeners = append(tcpListeners, l)
	go func() {
		for {
			tcpConnection, err := l.Accept()
			if err != nil {
				Logger.Error("", addr, "", "FrontServer", "Accept Error", err)
				continue
			}
			if IsStopAccept() {
				Logger.Warn("", addr, tcpConnection.RemoteIp(), "FrontServer", "close because of stopping accept", "")
				tcpConnection.Close()
			} else {
				go Dispatch(tcpConnection)
			}
		}
	}()
}

func NewConnection(tcpConnection *network.TcpConnection) (IConnection, error) {
	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countOpenResponseTime("NewConnection")
		defer countFunc()
	}

	if err := tcpConnection.SetReadDeadline(time.Now().Add(staticConf.XimpReadTimeout)); err != nil {
		return nil, err
	}
	if err := tcpConnection.ReadProtoBuffer(); err != nil {
		return nil, err
	}

	var conn IConnection

	if tcpConnection.IsXimpProto() {
		conn = newXimpConnection(tcpConnection)
	} else if tcpConnection.IsHttpProto() {
		req, err := http.ReadRequest(bufio.NewReader(tcpConnection))
		if err != nil {
			return nil, err
		}
		switch req.URL.Path {
		case "/":
			c, err := network.ShakeWebSocket(tcpConnection, req)
			if err != nil {
				return nil, err
			}
			conn = newXimpConnection(c)
		}
	} else if tcpConnection.IsCrossDomain() {
		defer tcpConnection.Close()
		// <policy-file-request/>\0
		buf := make([]byte, 23)
		tcpConnection.ReadStream(buf, len(buf))
		buf = []byte(CrossDomainXml)
		buf = append(buf, byte(0))
		tcpConnection.WriteBytes(buf, logic.StaticConf.ExternalWriteTimeout)
	} else if tcpConnection.IsTLSProto() {
		tlsConn, err := network.ShakeTLS(tcpConnection)
		if err != nil {
			return nil, err
		}
		switch {
		case tlsConn.IsHTTPProto():
			req, err := http.ReadRequest(tlsConn.Buf)
			if err != nil {
				return nil, err
			}
			switch req.URL.Path {
			case "/":
				c, err := network.ShakeWebSocketSecure(tlsConn, req)
				if err != nil {
					return nil, err
				}
				conn = newXimpConnection(c)
			}
		default:
			return nil, errors.New("unsupport protocal")
		}
	}

	if conn != nil {
		conn.SetId(connectionIdGenerator.NextConnectionId())
		// count open connection operation
		requestStat.AtomicAddOpenConns(1)
		return conn, nil
	} else {
		// count open connection operation
		requestStat.AtomicAddOpenConnFails(1)
		return nil, errors.New("unknown protocal")
	}
}

func Dispatch(tcpConnection *network.TcpConnection) {
	conn, err := NewConnection(tcpConnection)
	if err != nil {
		Logger.Warn("", "", tcpConnection.RemoteIp(), "Dispatch", "NewConnection error", err)
	} else {
		connectionPool.AddConnection(conn)
		conn.Serve()
		connectionPool.DelConnection(conn.GetId())
	}

	tcpConnection.Close()
	return
}
