package msgRedis

import (
	"errors"
	"time"
)

type OnceConn struct {
	*Conn
}

func (mp *MultiPool) Call(address string) *OnceConn {
	oc := &OnceConn{}
	c := mp.PopByAddr(address)
	if c == nil {
		oc.Conn = &Conn{err: errors.New("get a nil conn address=" + address)}
		return oc
	}
	oc.Conn = c
	oc.Conn.isOnce = true
	// Reset timeout to default
	c.readTimeout = time.Duration(ReadTimeout)
	c.writeTimeout = time.Duration(WriteTimeout)
	return oc
}

func (mp *MultiPool) CallWithTimeout(address string, readTimeout, writeTimeout int64) *OnceConn {
	oc := &OnceConn{}
	c := mp.PopByAddr(address)
	if c == nil {
		oc.Conn = &Conn{err: errors.New("get a nil conn address=" + address)}
		return oc
	}
	oc.Conn = c
	oc.Conn.isOnce = true
	// Set timeout
	c.readTimeout = time.Duration(readTimeout)
	c.writeTimeout = time.Duration(writeTimeout)
	return oc
}
