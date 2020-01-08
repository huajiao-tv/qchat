package llconn

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

var (
	DefaultMessageMeter = &MessageMeter{wrapMessage: true}
)

type TestMessage struct {
	SentTime int64 // UnixNano
	Content  string
}

func GenRandString(length int64) string {
	var s string = ""
	for i := int64(0); i < length; i++ {
		t := 97 + rand.Intn(122-97)
		s = string(append([]byte(s), byte(t)))
	}
	return s
}

type MessageMeter struct {
	wrapMessage bool
}

func (mm *MessageMeter) EnableWrapper(enable bool) {
	mm.wrapMessage = enable
}

func (mm *MessageMeter) WrapMessage(content string) string {

	if !mm.wrapMessage {
		return content
	}

	tm := TestMessage{SentTime: time.Now().UnixNano(), Content: content}
	data, err := json.Marshal(&tm)

	if err == nil {
		return string(data)
	}

	fmt.Println("WrapMessage json.Marshal failed", err)

	return content
}

func (mm *MessageMeter) Check(content string) bool {

	if len(content) < 15 {
		return false
	}

	if content[0:12] != "{\"SentTime\":" {
		//fmt.Println("not json format")
		return false
	}

	tm := TestMessage{}

	if err := json.Unmarshal([]byte(content), &tm); err != nil {
		fmt.Println("Unmarshal failed,", err)
		return false
	}

	delay := time.Now().UnixNano() - tm.SentTime
	delay /= 1e6

	if delay > 3000 {
		fmt.Printf("Message [%s] comes TOO LATE (%d milliseconds)\n", tm.Content, delay)
	} else {
		fmt.Printf("Message [%s] has been delayed for %d milliseconds\n", tm.Content, delay)
	}

	return true
}
