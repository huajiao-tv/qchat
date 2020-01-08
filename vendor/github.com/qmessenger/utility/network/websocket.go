package network

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketConnection struct {
	mu sync.Mutex
	*websocket.Conn
	ReadBuffer []byte
}

func ShakeWebSocket(tcp *TcpConnection, req *http.Request) (*WebSocketConnection, error) {
	conn := &respWriter{
		TcpConnection: tcp,
		HttpResponse:  NewHttpResponse(),
	}
	u := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	u.CheckOrigin = func(r *http.Request) bool {
		// allow all connections by default
		return true
	}
	c, err := u.Upgrade(conn, req, nil)
	if err != nil {
		return nil, err
	}
	return &WebSocketConnection{
		Conn: c,
		mu:   sync.Mutex{},
	}, nil
}

func (w *WebSocketConnection) Type() string {
	return WebSocketNetwork
}

func (w *WebSocketConnection) Read(b []byte) (n int, err error) {
	if w.ReadBuffer == nil {
		_, w.ReadBuffer, err = w.Conn.ReadMessage()
		if err != nil {
			return
		}
	}
	bLen := len(b)
	if bLen < len(w.ReadBuffer) {
		copy(b, w.ReadBuffer[0:bLen])
		w.ReadBuffer = w.ReadBuffer[bLen:]
		n = bLen
	} else {
		copy(b, w.ReadBuffer)
		n = len(w.ReadBuffer)
		w.ReadBuffer = nil
	}
	return
}

func (w *WebSocketConnection) ReadStream(stream []byte, count int) error {
	if len(stream) < count {
		return errors.New("bad stream")
	}
	stream = stream[:count]
	left := count
	for left > 0 {
		n, err := w.Read(stream)
		if err != nil {
			return err
		}
		if n > 0 {
			left -= n
			if left > 0 {
				stream = stream[n:]
			}
		}
	}
	return nil
}

func (w *WebSocketConnection) Write(b []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	err = w.Conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (w *WebSocketConnection) SetDeadline(t time.Time) error {
	return w.Conn.UnderlyingConn().SetDeadline(t)
}

func (w *WebSocketConnection) WriteBytes(b []byte, t time.Duration) (int, error) {
	err := w.Conn.SetWriteDeadline(time.Now().Add(t))
	if err != nil {
		return 0, err
	}
	return w.Write(b)
}

func (w *WebSocketConnection) RemoteIp() string {
	remoteAddr := w.RemoteAddr().(*net.TCPAddr)
	return remoteAddr.IP.String()
}

type respWriter struct {
	*TcpConnection
	*HttpResponse
}

func (r *respWriter) Header() http.Header {
	h := r.HttpResponse.headers
	header := make(map[string][]string, len(h))
	for k, v := range h {
		header[k] = strings.Split(v, ",")
	}
	return http.Header(header)
}

func (r *respWriter) Write(b []byte) (int, error) {
	r.HttpResponse.SetBodyStream(b)
	return r.TcpConnection.Write(r.HttpResponse.Stream())
}

func (r *respWriter) WriteHeader(code int) {
	r.HttpResponse.SetCode(code)
}

func (r *respWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.TcpConnection, bufio.NewReadWriter(bufio.NewReader(r.TcpConnection), bufio.NewWriter(r.TcpConnection)), nil
}
