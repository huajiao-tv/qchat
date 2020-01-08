package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/huajiao-tv/qchat/logic"
)

var (
	httpClient *http.Client
)

func init() {
	httpClient = &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(network, addr, time.Duration(1000)*time.Millisecond)
				if err != nil {
					return nil, err
				}
				return c, nil
			},
			MaxIdleConnsPerHost: 100,
		},
		Timeout: time.Duration(60000) * time.Millisecond,
	}
}

type PayloadJson struct {
	TimeStamp int64 `json:"timestamp"`
}

func sendToRoom(roomId string) error {
	s := PayloadJson{
		TimeStamp: time.Now().UnixNano(),
	}
	data, err := json.Marshal(s)
	if err != nil {
		return nil
	}
	values := url.Values{
		"roomid":  []string{roomId},
		"sender":  []string{logic.RandString(8)},
		"traceid": []string{fmt.Sprint(time.Now().UnixNano())},
		"content": []string{string(data)},
	}
	uri := "http://127.0.0.1:8360/send"
	/*88if test > 0 {
		uri = "http://127.0.0.1:8360/send"
	}*/
	_, err = httpClient.PostForm(uri, values)
	return err
}

func ungzip(data []byte) []byte {
	if data == nil || len(data) == 0 {
		return nil
	}

	var done bool = true
	bufLen := 512
	buf := make([]byte, bufLen)
	writer := bytes.NewBuffer(nil)
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		fmt.Println("gzip.NewReader failed", err)
		return nil
	}

	for {
		count, err1 := reader.Read(buf)
		if count == bufLen {
			writer.Write(buf)
		} else if count > 0 {
			writer.Write(buf[0:count])
			break
		} else {
			fmt.Println("ungzip failed", err1)
			done = false
			break
		}
	}

	reader.Close()
	if done {
		return writer.Bytes()
	}
	return nil
}
