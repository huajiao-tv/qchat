package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	CounterProduct        = "live"
	CounterPartner        = "internal"
	CounterPlatform       = "server"
	CounterInnerSecretKey = "eac63e66d8c4a6f0303f00bc76d0217c"
	CounterWatchesType    = "watches"
	CounterSharesType     = "shares"
	CounterGuid           = "c1ef726ebc91979c351253267bfeab40"
	CounterUrl            = "http://counter.huajiao.com/"
	CounterIncrease       = "counter/increase"
	CenterAddr            = "http://qchatcenter.huajiao.com:6677/chatroom/send"
)

var httpClient *http.Client
var praiseMsg = `{"roomid":$$roomid$$,"type":8,"text":"","time":1473681599,"expire":86400,"extends":{"liveid":"$$roomid$$","num":$$num$$,"userid":"69247689","total":$$total$$,"nickname":"\u82b1\u6912\u7528\u623709101146","verified":false,"verifiedinfo":{"credentials":"","type":0,"realname":"\u82b1\u6912\u7528\u623709101146","status":0,"error":"","official":false},"verify_student":{},"exp":0,"level":1},"traceid":"$$traceid$$"}`
var totalCount int
var PerCount string
var roomid string

type CounterHttpResponse struct {
	Errno   int         `json:"errno"`
	Errmsg  string      `json:"errmsg"`
	Consume float64     `json:"consume"`
	Time    int64       `json:"time"`
	Md5     string      `json:"md5"`
	Data    interface{} `json:"data"`
}

func ChangeCounterNums(roomid string, count int, countType string) {
	paramRand := strconv.FormatInt(rand.Int63(), 10)
	paramTime := strconv.FormatInt(time.Now().UnixNano(), 10)
	params := []string{"internal", paramRand, paramTime}
	paramSign := md5.Sum([]byte(strings.Join(params, "_") + CounterInnerSecretKey))
	paramSignStr := fmt.Sprintf("%x", paramSign)

	values := url.Values{
		"partner":  []string{CounterPartner},
		"rand":     []string{paramRand},
		"time":     []string{paramTime},
		"product":  []string{CounterProduct},
		"type":     []string{countType},
		"relateid": []string{roomid},
		"guid":     []string{strings.ToLower(paramSignStr)},
		"number":   []string{strconv.Itoa(count)},
		"platform": []string{CounterPlatform},
	}

	resp, err := http.PostForm(CounterUrl+CounterIncrease, values)
	if err != nil {
		fmt.Println("change counter nums fail", err, roomid)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ChangeCounterNums ioutil fail", err.Error, roomid)
	}
	counterHttpResponse := &CounterHttpResponse{}

	json.Unmarshal(body, counterHttpResponse)

	fmt.Println("ChangeCounterNums result", string(body), roomid, count, counterType)
}

func send(priority, typ string) bool {
	PerCountInt, err := strconv.Atoi(PerCount)
	if err != nil {
		return false
	}
	totalCount += PerCountInt
	content := strings.Replace(praiseMsg, "$$num$$", PerCount, -1)
	content = strings.Replace(content, "$$total$$", strconv.Itoa(totalCount), -1)
	content = strings.Replace(content, "$$roomid$$", roomid, -1)
	fmt.Println(content)
	values := make(map[string][]string, 6)
	values["roomid"] = []string{roomid}
	values["sender"] = []string{"admin"}
	values["traceid"] = []string{"1234567890"}
	values["content"] = []string{content}
	values["appid"] = []string{"2080"}
	values["priority"] = []string{priority}
	values["type"] = []string{typ}
	url := CenterAddr

	fmt.Println("bengin change counter")
	ChangeCounterNums(roomid, PerCountInt, "praises")

	if _, err := httpClient.PostForm(url, values); err != nil {
		fmt.Println("SendChatRoomMsg err is ", err)
		return false
	}
	return true
}

func init() {
	flag.StringVar(&roomid, "rid", "", "roomid")
	flag.StringVar(&PerCount, "num", "", "roomid")
	flag.Parse()
	totalCount = 0
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

func main() {
	fmt.Println("start")
	for {
		select {
		case <-time.After(time.Second * time.Duration(60)):
			fmt.Println("exec")
			go send("1", "8")
		case <-time.After(time.Second * time.Duration(7200)):
			return
		}
	}
}
