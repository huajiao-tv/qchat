package network

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"net"
	"time"
)

type TLSConnection struct {
	*tls.Conn
	Buf *bufio.Reader
}

func ShakeTLS(tcpConn *TcpConnection) (*TLSConnection, error) {
	crt, err := tls.LoadX509KeyPair("/data/qchat/ssl/certs/*.com.pm", "/data/qchat/ssl/private/*.com.key")
	if err != nil {
		return nil, err
	}

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{crt},
	}

	tlsConn := tls.Server(tcpConn, tlsConf)

	return &TLSConnection{
		Conn: tlsConn,
		Buf:  bufio.NewReader(tlsConn),
	}, nil
}

func (c *TLSConnection) Type() string {
	// return TLSNetwork
	return WebSocketNetwork
}

func (c *TLSConnection) Read(b []byte) (int, error) {
	return c.Buf.Read(b)
}

func (c *TLSConnection) ReadStream(b []byte, n int) error {
	if len(b) < n {
		return errors.New("bad stream")
	}

	var err error
	b = b[:n]
	b, err = c.Buf.Peek(n)
	if err != nil {
		return err
	}

	c.Buf.Discard(n)
	return nil
}

func (c *TLSConnection) WriteBytes(b []byte, timeout time.Duration) (int, error) {
	if err := c.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return 0, err
	}

	n, err := c.Conn.Write(b)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (c *TLSConnection) IsHTTPProto() bool {
	proto, _ := c.Buf.Peek(5)
	if bytes.Equal(proto, []byte("GET /")) {
		return true
	}

	proto, _ = c.Buf.Peek(6)
	if bytes.Equal(proto, []byte("POST /")) {
		return true
	}

	return false
}

func (c *TLSConnection) RemoteIp() string {
	remoteAddr := c.RemoteAddr().(*net.TCPAddr)
	return remoteAddr.IP.String()
}
