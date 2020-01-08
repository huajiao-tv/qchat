package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/huajiao-tv/qchat/logic"
)

func Info(args ...interface{}) {
	fmt.Print("INFO:")
	fmt.Println(args...)
	fmt.Print("\n")
}
func Fail(args ...interface{}) {
	fmt.Print("Fail:")
	fmt.Println(args...)
	fmt.Print("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF\n\n")
}
func ExecNTimes(n int, m int64, f func()) {
	sleepTime := time.Second / time.Duration(n)
	next := time.Now()
	var i int64
	for m == 0 || i < m {
		i++
		next = next.Add(sleepTime)
		go f()
		left := next.Sub(time.Now())
		if left > 0 {
			time.Sleep(left)
		}
	}
}

var (
	Server   string
	N        int
	Max      int
	Room     string
	RoomGen  int
	MsgSize  int
	Succ     int64
	Err      int64
	TestType bool
	ConfFile string
)
var httpClient *http.Client
var ConfList []MsgConf

type MsgConf struct {
	MsgType   int
	Priority  int
	Frequency int
}

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
			MaxIdleConnsPerHost: 1000,
		},
		Timeout: time.Duration(2000) * time.Millisecond,
	}
	flag.StringVar(&Server, "s", "http://127.0.0.1:6600", "center addr")
	flag.StringVar(&ConfFile, "f", "", "配置文件路径")
	flag.IntVar(&Max, "m", 10000, "stop when send such count msgs")
	flag.IntVar(&RoomGen, "R", 0, "与ben_connect的roomid一致")
	flag.IntVar(&MsgSize, "S", 100, "消息长度")
	flag.StringVar(&Room, "r", "", "room id")
	flag.BoolVar(&TestType, "t", false, "test type and priority")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	ConfList = []MsgConf{}
	if ConfFile == "" {
		panic("no ConfFile")
	}
	GetConf(ConfFile)
	fmt.Println(ConfList)
}

func GetConf(filePath string) {
	confFile, err := os.Open(filePath)
	defer confFile.Close()
	if nil == err {
		buff := bufio.NewReader(confFile) //读入缓存
		for {
			line, err := buff.ReadString('\n') //以'\n'为结束符读入一行
			if err != nil || io.EOF == err {
				break
			}
			line = strings.Trim(line, "\n")
			confStr := strings.Split(line, ",")
			msgType, _ := strconv.Atoi(confStr[0])
			priority, _ := strconv.Atoi(confStr[1])
			frequency, err := strconv.Atoi(confStr[2])
			if err != nil {
				panic(err.Error())
			}
			ConfList = append(ConfList, MsgConf{MsgType: msgType, Priority: priority, Frequency: frequency})
		}
	}
}

func genRooms() []string {
	rooms := []string{}
	if RoomGen != 0 {
		c := RoomGen
		if c < 0 {
			c = -c
		}
		for i := 0; i < c; i++ {
			rooms = append(rooms, strconv.Itoa(i))
		}
	} else {
		rooms = strings.Split(Room, ",")
	}
	return rooms
}

func ShowInfo() {
	fmt.Printf("Send:%d/%d\t\n",
		atomic.LoadInt64(&Succ),
		atomic.LoadInt64(&Err))
}

func main() {
	rooms := genRooms()
	if len(rooms) == 0 || rooms[0] == "" {
		panic("room is not defined")
	}
	content := logic.RandString(MsgSize)
	go ExecNTimes(1, 0, ShowInfo)
	for _, v := range ConfList {
		go ExecNTimes(v.Frequency, int64(Max), GenSend(rooms, content, v))
	}
	select {}
}

var cur int32 = 0

func GenSend(rooms []string, content string, v MsgConf) func() {
	return func() {
		var r string
		if RoomGen > 0 {
			r = rooms[int(atomic.LoadInt32(&cur))%len(rooms)]
			atomic.AddInt32(&cur, 1)
		} else {
			r = rooms[rand.Int()%len(rooms)]
		}
		if TestType {
			content = "{\"type\":" + strconv.Itoa(v.MsgType) + ",\"priority\":" + strconv.Itoa(v.Priority) + "}"
		}

		if send(Server, r, content, strconv.Itoa(v.Priority), strconv.Itoa(v.MsgType)) {
			atomic.AddInt64(&Succ, 1)
		} else {
			atomic.AddInt64(&Err, 1)
		}
	}
}

func send(addr, roomid, content, priority, msgType string) bool {
	url := fmt.Sprintf("%s/chatroom/send", addr)
	values := make(map[string][]string, 6)
	values["roomid"] = []string{roomid}
	values["sender"] = []string{"admin"}
	values["traceid"] = []string{"1234567890"}
	values["content"] = []string{content}
	values["appid"] = []string{"2080"}
	values["priority"] = []string{priority}
	values["type"] = []string{msgType}

	if _, err := httpClient.PostForm(url, values); err != nil {
		Fail("SendChatRoomMsg err is ", err)
		return false
	}
	return true
}
