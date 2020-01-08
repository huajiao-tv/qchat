package client

import (
	"time"

	"github.com/huajiao-tv/qchat/logic/pb"
)

const (
	Tcp       = "tcp"
	WebSocket = "ws"

	ConnTimeout    = 5
	WriteTimeout   = 20
	RequestTimeout = 30
	ClientChanLen  = 100
)

const (
	IM       = "im"
	Peer     = "peer"
	Public   = "public"
	ChatRoom = "chatroom"
)

type ClientConf struct {
	Tag           string
	ConnType      string // Connection Type: "tcp" or "websocket"
	ServerAddr    string // Server address
	CenterAddr    string // Center address
	AppID         uint32 // ApplicationID
	ClientVer     int    // ClientVersion
	ProtoVer      int    // ProtocolVersion
	DefaultKey    string // Default Key
	Heartbeat     int    // Heartbeat interval (unit second)
	SendHeartbeat int    // Send heart beat or not, if value is 0, don't send, otherwise send.
	AutoLogin     int    // autologin when init login, if the value is not 0
}

type AccountInfo struct {
	UserID    string // User ID
	Password  string // Corresponding password
	Signature string // Signature to obtain token
	DeviceID  string // Device ID
	Platform  string // Platform
}

type UserMessage struct {
	ID       uint64
	Content  []byte
	InfoType string
	MsgType  uint32
	Sn       uint64
	Valid    uint32
}

type MessageSlot struct {
	InfoType string
	LatestID int64
}

type ServiceHandler interface {
	HandleServiceMessage(data []byte, sn uint64, source int) bool
	HandleGetMessage(msg *UserMessage) bool
}

type Stream interface {
	Close()
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

type packet struct {
	isPing bool        // 是否是心跳
	msg    *pb.Message // PB数据结构
	data   []byte      // 对应网络包数据， 包含处理过的magic code, length 和 Marshal/Unmarshal后的数据
}
