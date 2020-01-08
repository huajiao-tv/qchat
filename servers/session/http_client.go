package main

import (
	"net"
	"net/http"
	"sync/atomic"
	"time"
	"unsafe"
)

func HttpClient() *HttpClientType {
	return (*HttpClientType)(atomic.LoadPointer(&HttpClientPtr))
}

type HttpClientType struct {
	Client              *http.Client
	dialTimeout         time.Duration
	timeout             time.Duration
	maxIdleConnsPerHost int
}

var HttpClientPtr unsafe.Pointer = unsafe.Pointer(&HttpClientType{})

func newHttpClient(dialTimeout, timeout time.Duration, maxIdleConnsPerHost int) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(network, addr, dialTimeout)
				if err != nil {
					return nil, err
				}
				return c, nil
			},
			MaxIdleConnsPerHost: maxIdleConnsPerHost,
		},
		Timeout: timeout,
	}
}

func UpdateHttpClientConfig() bool {
	dialTimeout := time.Duration(netConf().CallbackConnTimeout) * time.Millisecond
	timeout := time.Duration(netConf().CallbackTimeout) * time.Millisecond
	maxIdleConnsPerHost := netConf().CallbackMaxIdleConns

	hc := HttpClient()

	if hc.Client != nil {
		var needUpdate bool
		switch {
		case hc.dialTimeout != dialTimeout:
			needUpdate = true
		case hc.timeout != timeout:
			needUpdate = true
		case hc.maxIdleConnsPerHost != maxIdleConnsPerHost:
			needUpdate = true
		}

		if !needUpdate {
			return false
		}
	}

	client := newHttpClient(dialTimeout, timeout, maxIdleConnsPerHost)
	newHttpClient := &HttpClientType{
		Client:              client,
		dialTimeout:         dialTimeout,
		timeout:             timeout,
		maxIdleConnsPerHost: maxIdleConnsPerHost,
	}

	atomic.StorePointer(&HttpClientPtr, unsafe.Pointer(newHttpClient))

	return true
}
