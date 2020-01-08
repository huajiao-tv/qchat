package client

import (
	"errors"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketStream struct {
	conn   *websocket.Conn
	buffer []byte
}

func NewWebSocketStream(addr string, timeout time.Duration) (*WebSocketStream, error) {
	u := url.URL{Scheme: "ws", Host: addr, Path: "/"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return &WebSocketStream{
		conn: c,
	}, nil
}

func (ws *WebSocketStream) Close() {
	(*ws.conn).Close()
}

func (ws *WebSocketStream) Read(data []byte) (n int, err error) {
	if data == nil {
		return 0, errors.New("byte is nil")
	}
	dLen := len(data)
	if dLen == 0 {
		return 0, errors.New("data is empty")
	}

	if ws.buffer == nil {
		_, ws.buffer, err = (*ws.conn).ReadMessage()
		if err != nil {
			return 0, err
		}
	}

	if dLen < len(ws.buffer) {
		copy(data, ws.buffer[0:dLen])
		ws.buffer = ws.buffer[dLen:]
		return dLen, nil
	} else {
		copy(data, ws.buffer)
		n = len(ws.buffer)
		ws.buffer = nil
		return n, nil
	}
}

func (ws *WebSocketStream) Write(data []byte) (int, error) {
	if data == nil {
		return 0, errors.New("byte is nil")
	}
	dLen := len(data)
	if dLen == 0 {
		return 0, errors.New("data is empty")
	}

	err := (*ws.conn).WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		return 0, err
	}
	return len(data), nil
}

func (ws *WebSocketStream) SetReadDeadline(t time.Time) error {
	return (*ws.conn).SetReadDeadline(t)
}

func (ws *WebSocketStream) SetWriteDeadline(t time.Time) error {
	return (*ws.conn).SetWriteDeadline(t)
}
