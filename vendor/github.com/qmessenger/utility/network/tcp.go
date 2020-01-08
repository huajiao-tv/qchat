package network

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"net"
	"strconv"
	"time"
)

// 监听一个 tcp 地址
func TcpListen(address string) (*TcpListener, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}
	return newTcpListener(tcpListener), nil
}

// tcp 监听器
type TcpListener struct {
	*net.TCPListener
}

func newTcpListener(tcpListener *net.TCPListener) *TcpListener {
	return &TcpListener{TCPListener: tcpListener}
}

func (this *TcpListener) Accept() (*TcpConnection, error) {
	tcpConn, err := this.TCPListener.AcceptTCP()
	if err != nil {
		return nil, err
	}
	return newTcpConnection(tcpConn), nil
}

// 连接一个 tcp 地址
func TcpConnect(localAddress string, remoteAddress string, timeout time.Duration) (*TcpConnection, error) {
	// @todo: 本 api 调用不绑定 localAddress，可能会影响上层应用程序
	conn, err := net.DialTimeout("tcp", remoteAddress, timeout)
	if err != nil {
		return nil, err
	}
	tcpConn := conn.(*net.TCPConn)
	return newTcpConnection(tcpConn), nil
}

// tcp 连接
type TcpConnection struct {
	*net.TCPConn
	protoBuffer []byte
}

// 创建 tcp 连接对象
func newTcpConnection(tcpConn *net.TCPConn) *TcpConnection {
	return &TcpConnection{TCPConn: tcpConn}
}

func (this *TcpConnection) Type() string {
	return TcpNetwork
}

// 获取连接ip
func (this *TcpConnection) RemoteIp() string {
	remoteAddr := this.RemoteAddr().(*net.TCPAddr)
	return remoteAddr.IP.String()
}

// 获取连接port
func (this *TcpConnection) RemotePort() string {
	remoteAddr := this.RemoteAddr().(*net.TCPAddr)
	return strconv.Itoa(remoteAddr.Port)
}

func (this *TcpConnection) ReadStream(stream []byte, count int) error {
	if len(stream) < count {
		return errors.New("bad stream")
	}
	stream = stream[:count]
	left := count
	for left > 0 {
		n, err := this.Read(stream)
		if n > 0 {
			left -= n
			if left > 0 {
				stream = stream[n:]
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

const protoBufferLen = 8

func (this *TcpConnection) ReadProtoBuffer() error {
	this.protoBuffer = make([]byte, protoBufferLen)
	left := protoBufferLen
	for left > 0 {
		n, err := this.TCPConn.Read(this.protoBuffer)
		if n > 0 {
			left -= n
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *TcpConnection) IsXimpProto() bool {
	if this.protoBuffer == nil {
		return false
	}
	if this.protoBuffer[0] == 'q' && this.protoBuffer[1] == 'h' {
		return true
	}
	return false
}

func (this *TcpConnection) IsHttpProto() bool {
	if this.protoBuffer == nil {
		return false
	}
	if this.protoBuffer[0] == 'G' && this.protoBuffer[1] == 'E' && this.protoBuffer[2] == 'T' && this.protoBuffer[3] == ' ' && this.protoBuffer[4] == '/' {
		return true
	}
	if this.protoBuffer[0] == 'P' && this.protoBuffer[1] == 'O' && this.protoBuffer[2] == 'S' && this.protoBuffer[3] == 'T' && this.protoBuffer[4] == ' ' && this.protoBuffer[5] == '/' {
		return true
	}
	return false
}

func (this *TcpConnection) IsCrossDomain() bool {
	if this.protoBuffer == nil {
		return false
	}
	// 判断是否为“<policy-file-request/>”，只判断了前8个字节
	if this.protoBuffer[0] == '<' && this.protoBuffer[1] == 'p' && this.protoBuffer[2] == 'o' && this.protoBuffer[3] == 'l' &&
		this.protoBuffer[4] == 'i' && this.protoBuffer[5] == 'c' && this.protoBuffer[6] == 'y' && this.protoBuffer[7] == '-' {
		return true
	}
	return false
}

// Miop协议版本3
func (this *TcpConnection) IsMiopV3() bool {
	if this.protoBuffer == nil {
		return false
	}
	miopVersion := binary.BigEndian.Uint16(this.protoBuffer[0:2])
	if miopVersion == 3 {
		return true
	}
	return false
}

// Miop协议版本5
func (this *TcpConnection) IsMiopV5() bool {
	if this.protoBuffer == nil {
		return false
	}
	miopVersion := binary.BigEndian.Uint16(this.protoBuffer[0:2])
	if miopVersion == 5 {
		return true
	}
	return false
}

// Miop协议版本6
func (this *TcpConnection) IsMiopV6() bool {
	if this.protoBuffer == nil {
		return false
	}
	miopVersion := binary.BigEndian.Uint16(this.protoBuffer[0:2])
	if miopVersion == 6 {
		return true
	}
	return false
}

// TLS 协议
func (this *TcpConnection) IsTLSProto() bool {
	if this.protoBuffer == nil {
		return false
	}

	return bytes.Equal(this.protoBuffer[:3], []byte{22, 3, 1}) ||
		bytes.Equal(this.protoBuffer[:3], []byte{22, 3, 2}) ||
		bytes.Equal(this.protoBuffer[:3], []byte{22, 3, 3})
}

func (this *TcpConnection) Read(p []byte) (n int, err error) {
	if this.protoBuffer == nil {
		n, err = this.TCPConn.Read(p)
	} else {
		pLen := len(p)
		if pLen < len(this.protoBuffer) {
			copy(p, this.protoBuffer[0:pLen])
			this.protoBuffer = this.protoBuffer[pLen:]
			n = pLen
		} else {
			copy(p, this.protoBuffer)
			n = len(this.protoBuffer)
			this.protoBuffer = nil
		}
	}
	return
}

func (this *TcpConnection) WriteBytes(p []byte, timeout time.Duration) (int, error) {
	if err := this.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return 0, err
	}
	n, err := this.TCPConn.Write(p)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (this *TcpConnection) ReadRpcRequest(timeout time.Duration) (*RpcRequest, error) {
	if err := this.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}
	decoder := gob.NewDecoder(this)
	request := NewRpcRequest()
	if err := decoder.Decode(request); err != nil {
		return nil, err
	}
	return request, nil
}

func (this *TcpConnection) WriteRpcResponse(response *RpcResponse, timeout time.Duration) error {
	if err := this.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	encoder := gob.NewEncoder(this)
	if err := encoder.Encode(response); err != nil {
		return err
	}
	return nil
}

func (this *TcpConnection) Rpc(request *RpcRequest, readTimeout time.Duration, writeTimeout time.Duration) (*RpcResponse, error) {
	if err := this.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return nil, err
	}
	encoder := gob.NewEncoder(this)
	if err := encoder.Encode(request); err != nil {
		return nil, err
	}
	if err := this.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		return nil, err
	}
	decoder := gob.NewDecoder(this)
	response := NewRpcResponse()
	if err := decoder.Decode(response); err != nil {
		return nil, err
	}
	return response, nil
}

// 注册 rpc 类型
func RegisterRpcTypeForValue(value interface{}) {
	gob.Register(value)
}

type RpcRequest struct {
	Func string
	Args map[string]interface{}
}

func NewRpcRequest() *RpcRequest {
	return &RpcRequest{Args: make(map[string]interface{})}
}

type RpcResponse struct {
	Result bool
	Reason string
	Code   int
	Args   map[string]interface{}
	Data   interface{}
}

func NewRpcResponse() *RpcResponse {
	return &RpcResponse{Args: make(map[string]interface{})}
}
