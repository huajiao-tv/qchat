package main

import (
	"flag"
	"fmt"
	"runtime"
	"sync"
	"time"
	//"github.com/johntech-o/gorpc"
	//"github.com/huajiao-tv/qchat/client/session"
	//"github.com/huajiao-tv/qchat/logic"
	//"strconv"
	//"bytes"
)

//var GorpcClient *gorpc.Client
var testcase string
var connid int
var sessaddr string
var platform string
var deviceid string
var mobiletype string
var uid string
var appid int
var num int
var radius float64
var gonum int
var prop map[string]string
var wg *sync.WaitGroup
var verbose int

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cpun := runtime.NumCPU()

	flag.StringVar(&testcase, "tc", "open", "testcase num")
	flag.IntVar(&connid, "connid", 1023, "connectid")
	flag.StringVar(&sessaddr, "host", "127.0.0.1:6420", "session addr")
	flag.StringVar(&platform, "pf", "PC", "platform")
	flag.StringVar(&deviceid, "did", "zj_pc_hp", "deviceid")
	flag.StringVar(&mobiletype, "mt", "ios", "mobile type")
	flag.StringVar(&uid, "uid", "dev_test_zj", "user id")
	flag.IntVar(&appid, "appid", 2090, "appid")
	flag.IntVar(&num, "num", 1, "test num every goroutinue")
	flag.IntVar(&gonum, "gonum", cpun, "goroutinue number")
	flag.IntVar(&verbose, "v", 0, "verbose mode")
	flag.Float64Var(&radius, "r", 10000, "半径")

	flag.Parse()

	initTest()
	fmt.Println("init done")
}

func initTest() {
	switch testcase {
	case "open":
		prop = map[string]string{}
		prop["ServerRam"] = "zj_serverram"
		prop["SenderType"] = "jid"
		prop["Platform"] = platform
		prop["ClientIp"] = "127.0.0.1"
		prop["LoginTime"] = time.Now().String()
		prop["Deviceid"] = deviceid
		prop["MobileType"] = mobiletype
		prop["CVersion"] = "100"
	case "close":
		prop = map[string]string{}
		prop["Platform"] = platform
		prop["LoginTime"] = time.Now().String()
		prop["Deviceid"] = deviceid
	case "query":
		prop := map[string]string{}
		prop["Platform"] = platform
		prop["LoginTime"] = time.Now().String()
		prop["Deviceid"] = deviceid
	}
}

func main() {

	fmt.Println("connid = ", connid)
	wg = new(sync.WaitGroup)
	wg.Add(gonum)

	begin := time.Now()

	switch testcase {
	case "open":
		benchOpenSession(gonum, num)
	case "close":
		benchCloseSession(gonum, num)
	case "query":
		benchQuerySession(gonum, num)
	case "mix":
		benchMix(gonum, num)
	case "redis":
		benchRedis(gonum, num)
	case "geoprepare":
		benchInitGeo(num)
		goto End
	case "geo":
		benchGeo(gonum, num, radius)
	case "join":
		benchJoinChatRoom(gonum, num)
	case "quit":
		benchQuitChatRoom(gonum, num)
	}

	wg.Wait()

End:
	end := time.Now()
	duration := end.Sub(begin)
	sec := duration.Seconds()
	fmt.Println("test result: ", float64(gonum*num)/sec, sec)
}
