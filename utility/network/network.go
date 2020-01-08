package network

import (
	"net"
	"time"
)

const (
	TcpNetwork             = "tcp"
	WebSocketNetwork       = "web"
	TLSNetwork             = "tls"
	WebSocketSecureNetwork = "wss"
)

type INetwork interface {
	net.Conn

	// network type: tcp, websocket or ...
	Type() string

	// read specified length from the network interface
	ReadStream(b []byte, n int) error

	// write bytes to the network interface
	WriteBytes(b []byte, t time.Duration) (int, error)

	// get remote ip
	RemoteIp() string
}
