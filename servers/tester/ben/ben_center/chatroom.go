package main

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/huajiao-tv/qchat/utility/cryption"
	"github.com/yvasiyarov/php_session_decoder/php_serialize"
)

type testResult struct {
	MinTime        time.Duration
	MaxTime        time.Duration
	TotalTime      time.Duration
	TotalRequests  int64
	FailedRequests int64
}

type verifiedinfo struct {
	Credentials string `json:"credentials"`
	Type        int    `json:"type"`
	RealName    string `json:"realname"`
	Status      int    `json:"status"`
	Error       string `json:"error"`
	Official    bool   `json:"official"`
}

type extend struct {
	LiveId       string       `json:"liveid"`
	Sender       string       `json:"userid"`
	NickName     string       `json:"nickname"`
	Avatar       string       `json:"avatar"`
	Verified     bool         `json:"verified"`
	Verifiedinfo verifiedinfo `json:"verifiedinfo"`
	Gift         int          `json:"gift"`
	Exp          int          `json:"exp"`
	Level        int          `json:"level"`
}

type message struct {
	Roomid    string `json:"roomid"`
	MsgType   int    `json:"type"`
	Content   string `json:"text"`
	TimeStamp int64  `json:"time"`
	Expire    int    `json:"expire"`
	Extends   extend `json:"extends"`
	Md5       string `json:"traceid"`
}

func (this *testResult) success(start time.Time) {
	cost := time.Now().Sub(start)
	this.TotalTime += cost
	this.TotalRequests++
	if cost > this.MaxTime {
		this.MaxTime = cost
	} else if cost < this.MinTime {
		this.MinTime = cost
	}
}

func (this *testResult) fail() {
	this.FailedRequests++
	this.TotalRequests++
}

func (this testResult) String() string {
	if this.TotalRequests != 0 {
		return fmt.Sprintf("Request result: min:%fs, max:%fs, average:%fs, requests:%d(failed:%d)",
			this.MinTime.Seconds(), this.MaxTime.Seconds(), this.TotalTime.Seconds()/float64(this.TotalRequests), this.TotalRequests, this.FailedRequests)
	} else {
		return "Error: not run"
	}
}

func bench_join(gn, n int) {
	room := "1234567"
	for i := 0; i < gn; i++ {
		go func(goid, runnum int) {
			defer wg.Done()

			url := fmt.Sprintf("%s/chatroom/join", centeraddr)
			result := &testResult{MinTime: 1e9}
			values := make(map[string][]string, 2)
			values["rid"] = []string{room}

			for {
				uid := fmt.Sprintf("%s%d%d", "10000", goid, runnum)
				values["uid"] = []string{uid}

				start := time.Now()

				if _, err := httpClient.PostForm(url, values); err != nil {
					fmt.Println("JoinChatRoom err is ", err)
					result.fail()
				} else {
					result.success(start)
				}

				if runnum--; runnum == 0 {
					break
				}
			}

			fmt.Println(result.String())
		}(i, n)
	}
}

func bench_quit(gn, n int) {
	room := "1234567"
	for i := 0; i < gn; i++ {
		go func(goid, runnum int) {
			defer wg.Done()

			url := fmt.Sprintf("%s/chatroom/quit", centeraddr)
			result := &testResult{MinTime: 1e9}
			values := make(map[string][]string, 2)
			values["rid"] = []string{room}

			for {
				uid := fmt.Sprintf("%s%d%d", "10000", goid, runnum)
				values["uid"] = []string{uid}

				start := time.Now()

				if _, err := httpClient.PostForm(url, values); err != nil {
					fmt.Println("QuitChatRoom err is ", err)
					result.fail()
				} else {
					result.success(start)
				}

				if runnum--; runnum == 0 {
					break
				}
			}

			fmt.Println(result.String())
		}(i, n)
	}
}

func bench_send(gn, n int) {
	for i := 0; i < gn; i++ {
		go func(goid, runnum int) {
			defer wg.Done()

			url := fmt.Sprintf("%s/chatroom/send", centeraddr)
			result := &testResult{MinTime: 1e9}
			values := make(map[string][]string, 6)
			values["roomid"] = []string{roomid}
			values["sender"] = []string{"admin"}
			values["traceid"] = []string{"1234567890"}
			values["content"] = []string{content}
			values["appid"] = []string{"2080"}
			values["priority"] = []string{"0"}

			for {
				start := time.Now()

				if _, err := httpClient.PostForm(url, values); err != nil {
					fmt.Println("SendChatRoomMsg err is ", err)
					result.fail()
				} else {
					result.success(start)
				}

				if runnum--; runnum == 0 {
					break
				}
			}

			fmt.Println(result.String())
		}(i, n)
	}
}

func bench_query(gn, n int) {
	room := "1234567"
	for i := 0; i < gn; i++ {
		go func(goid, runnum int) {
			defer wg.Done()

			url := fmt.Sprintf("%s/chatroom/query/member_detail", centeraddr)
			result := &testResult{MinTime: 1e9}
			values := make(map[string][]string, 3)
			values["content"] = []string{room}
			values["sn"] = []string{"12345"}
			values["m"] = []string{"0"}

			for {
				start := time.Now()

				if _, err := httpClient.PostForm(url, values); err != nil {
					fmt.Println("QueryChatRoom err is ", err)
					result.fail()
				} else {
					result.success(start)
				}

				if runnum--; runnum == 0 {
					break
				}
			}

			fmt.Println(result.String())
		}(i, n)
	}
}

func bench_send_random(gn, n int) {
	url := fmt.Sprintf("%s/send", centeraddr)
	values := make(map[string][]string, 7)
	values["roomid"] = []string{roomid}
	values["sender"] = []string{"admin"}
	values["appid"] = []string{"2080"}
	values["priority"] = []string{"101"}
	values["traceid"] = []string{fmt.Sprint(time.Now().UnixNano())}
	values["content"] = []string{huajiaoChatRoomMessage(roomid, fmt.Sprint(sid-1), "admin")}
	values["msgid"] = []string{fmt.Sprint(sid - 1)}
	if _, err := httpClient.PostForm(url, values); err != nil {
		fmt.Println("SendChatRoomMsg with id", fmt.Sprint(sid-1), "err is ", err)
		return
	}

	for i := 0; i < gn; i++ {
		go func(goid, runnum int) {
			defer wg.Done()

			result := &testResult{MinTime: 1e9}
			values := make(map[string][]string, 7)
			values["roomid"] = []string{roomid}
			values["sender"] = []string{"admin"}
			values["appid"] = []string{"2080"}
			values["priority"] = []string{"101"}

			msgIdCount := sid
			for {
				if interval > 0 {
					if goid%2 == 0 {
						time.Sleep(time.Second * time.Duration(interval))
					} else {
						time.Sleep(time.Millisecond * time.Duration(interval))
					}
				}
				start := time.Now()

				msgId := fmt.Sprint(msgIdCount + goid)
				values["traceid"] = []string{fmt.Sprint(start.UnixNano())}
				values["content"] = []string{huajiaoChatRoomMessage(roomid, msgId, "admin")}
				values["msgid"] = []string{msgId}
				if _, err := httpClient.PostForm(url, values); err != nil {
					fmt.Println("SendChatRoomMsg with random id", msgId, "err is ", err)
					result.fail()
				} else {
					result.success(start)
					//fmt.Println("message with id", msgId, "has been sent.")
				}

				if runnum--; runnum == 0 {
					break
				}
				msgIdCount += gn
			}

			fmt.Println(result.String())
		}(i, n)
	}
}

func bench_send_gap(gn, n int) {
	url := fmt.Sprintf("%s/send", centeraddr)
	values := make(map[string][]string, 7)
	values["roomid"] = []string{roomid}
	values["sender"] = []string{"admin"}
	values["appid"] = []string{"2080"}
	values["priority"] = []string{"101"}
	values["traceid"] = []string{fmt.Sprint(time.Now().UnixNano())}
	values["content"] = []string{huajiaoChatRoomMessage(roomid, fmt.Sprint(sid-1), "admin")}
	values["msgid"] = []string{fmt.Sprint(sid - 1)}
	if _, err := httpClient.PostForm(url, values); err != nil {
		fmt.Println("SendChatRoomMsg with id", fmt.Sprint(sid-1), "err is ", err)
		return
	}

	msgIdCount := int32(sid)
	for i := 0; i < gn; i++ {
		go func(goid, runnum int) {
			defer wg.Done()

			result := &testResult{MinTime: 1e9}
			values := make(map[string][]string, 7)
			values["roomid"] = []string{roomid}
			values["sender"] = []string{"admin"}
			values["appid"] = []string{"2080"}
			values["priority"] = []string{"101"}

			for {
				start := time.Now()

				msgId := fmt.Sprint(atomic.AddInt32(&msgIdCount, int32(gap)))
				values["traceid"] = []string{fmt.Sprint(start.UnixNano())}
				values["content"] = []string{huajiaoChatRoomMessage(roomid, msgId, "admin")}
				values["msgid"] = []string{msgId}
				if _, err := httpClient.PostForm(url, values); err != nil {
					fmt.Println("SendChatRoomMsg with id", msgId, "err is", err)
					result.fail()
				} else {
					result.success(start)
					//fmt.Println("message with id", msgId, "has been sent.")
				}

				if runnum--; runnum == 0 {
					break
				}
				if interval > 0 {
					time.Sleep(time.Second * time.Duration(interval))
				}
			}

			fmt.Println(result.String())
		}(i, n)
	}
}

func huajiaoChatRoomMessage(room, text, sender string) string {
	avatar := ""
	timestamp := time.Now().UnixNano()
	data := php_serialize.PhpArray{
		"roomid": room,
		"type":   9,
		"text":   text,
		"time":   timestamp,
		"expire": 86400,
		"extends": php_serialize.PhpArray{
			"liveid":   room,
			"userid":   sender,
			"nickname": sender,
			"avatar":   avatar,
			"verified": false,
			"verifiedinfo": php_serialize.PhpArray{
				"credentials": "",
				"type":        0,
				"realname":    sender,
				"status":      0,
				"error":       "",
				"official":    false,
			},
			"gift":  0,
			"exp":   1000,
			"level": 20,
		},
	}
	value, err := php_serialize.NewSerializer().Encode(data)
	if err != nil {
		fmt.Println("php serialize error:", err)
		return ""
	}
	msg := &message{
		Roomid:    room,
		MsgType:   9,
		Content:   text,
		TimeStamp: timestamp,
		Expire:    86400,
		Extends: extend{
			LiveId:   room,
			Sender:   sender,
			NickName: sender,
			Avatar:   avatar,
			Verified: false,
			Verifiedinfo: verifiedinfo{
				Credentials: "",
				Type:        0,
				RealName:    sender,
				Status:      0,
				Error:       "",
				Official:    false,
			},
			Gift:  0,
			Exp:   1000,
			Level: 20,
		},
		Md5: cryption.Md5(value),
	}
	ret, err := json.Marshal(*msg)
	if err != nil {
		fmt.Println("json marshal error:", err)
		return ""
	}
	return string(ret)
}

func bench_send_gap_overload(gn, n int) {
	msgIdCount := int32(sid)
	for i := 0; i < gn; i++ {
		go func(goid, runnum int) {
			defer wg.Done()

			url := fmt.Sprintf("%s/chatroom/send", centeraddr)
			result := &testResult{MinTime: 1e9}
			values := make(map[string][]string, 7)
			values["roomid"] = []string{roomid}
			values["sender"] = []string{"admin"}
			values["appid"] = []string{"2080"}
			values["priority"] = []string{"101"}

			count := 0

			delta := gap
			for {
				start := time.Now()

				msgId := fmt.Sprint(atomic.AddInt32(&msgIdCount, int32(delta)))
				values["traceid"] = []string{fmt.Sprint(start.UnixNano())}
				values["content"] = []string{huajiaoChatRoomMessage(roomid, msgId, "admin")}
				values["msgid"] = []string{msgId}
				if _, err := httpClient.PostForm(url, values); err != nil {
					fmt.Println("SendChatRoomMsg with random id err is ", err)
					result.fail()
				} else {
					result.success(start)
					//fmt.Println("message with id", msgId, "has been sent.")
				}

				if runnum--; runnum == 0 {
					break
				}

				if count++; count > 30 {
					delta = normalGapSize
				}
				if interval > 0 {
					time.Sleep(time.Second * time.Duration(interval))
				}
			}

			fmt.Println(result.String())
		}(i, n)
	}
}
