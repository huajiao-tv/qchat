package llconn

import (
	"errors"
	"net"
	"time"
)

type TcpStream struct {
	conn *net.Conn
}

func NewTcpStream(addr string, timeout time.Duration) (*TcpStream, error) {
	conn, err := net.DialTimeout("tcp", addr, connectTimeout)
	if err != nil {
		return nil, err
	}
	return &TcpStream{conn: &conn}, nil
}

func (this *TcpStream) Close() {
	(*this.conn).Close()
}

func (this *TcpStream) Read(data []byte) (int, error) {

	if data == nil {
		return 0, errors.New("byte is nil")
	}

	len := len(data)

	if len == 0 {
		return 0, errors.New("data is empty")
	}

	remain := len
	read := 0
	var err error = nil

	for remain > 0 {
		read, err = (*this.conn).Read(data[len-remain:])
		if err != nil {
			return len - remain, err
		}
		remain -= read
	}
	return len, nil
}

func (this *TcpStream) Write(data []byte) (int, error) {

	if data == nil {
		return 0, errors.New("byte is nil")
	}

	len := len(data)

	if len == 0 {
		return 0, errors.New("data is empty")
	}

	remain := len
	read := 0
	var err error = nil

	for remain > 0 {
		read, err = (*this.conn).Write(data[len-remain:])
		if err != nil {
			return len - remain, err
		}
		remain -= read
	}
	return len, nil
}

func (this *TcpStream) SetReadDeadline(t time.Time) error {
	return (*this.conn).SetReadDeadline(t)
}
func (this *TcpStream) SetWriteDeadline(t time.Time) error {
	return (*this.conn).SetWriteDeadline(t)
}
