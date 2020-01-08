package structure

import "time"

// proxy与router的消息格式
type Message struct {
	ProxyAddr    string
	ConnectionId uint64
	Uuid         string
	Appid        string
	Header       []byte
	Data         []byte
	Sid          uint32 // 作为header的sid 或者header中sid的ack
}

func NewMessage(proxyAddr string, id uint64, header, data []byte, uuid, appid string) *Message {
	return &Message{
		ProxyAddr:    proxyAddr,
		ConnectionId: id,
		Header:       header,
		Data:         data,
		Uuid:         uuid,
		Appid:        appid,
	}
}

// router给proxy返回的消息格式
type RegisterResult struct {
	ConnectionId uint64
	Success      bool
	RKey         []byte
	Appid        string
	UUid         string
	ReturnHeader []byte
}

func NewRegisterResult(id uint64, succ bool, RKey []byte, appid, uuid string, header []byte) *RegisterResult {
	return &RegisterResult{
		ConnectionId: id,
		Success:      succ,
		RKey:         RKey,
		Appid:        appid,
		UUid:         uuid,
		ReturnHeader: header,
	}
}

// session rpc通信消息格式
type Device struct {
	Appid             string //appid
	Uuid              string //设备唯一标识
	ProxyAddr         string //连接的proxy
	ConnectionId      uint64 //连接ID
	ConnectTime       uint64 //连接时间
	HeartbeatInterval uint32 //下次心跳间隔
	Data              []byte //用户附加信息

	Ok bool // rpc通信返回结果标记位
}

// session 中存储设备详细信息
type DeviceInfo struct {
	ProxyAddr         string
	ConnectionId      uint64
	ConnectTime       uint64
	LastHeartbeat     time.Time //上次心跳时间
	HeartbeatInterval uint32

	Ok bool //rpc通信返回结果标记位
}

//router给app发送ack
type MessageAck struct {
	Uuid string
	Sid  uint32
	Data []byte
}
