package main

import (
	"flag"
	"fmt"
	"runtime"
	"time"

	//"github.com/johntech-o/gorpc"
	"strconv"

	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
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

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.StringVar(&testcase, "tc", "open", "testcase num")
	flag.IntVar(&connid, "connid", 1023, "connectid")
	flag.StringVar(&sessaddr, "host", "127.0.0.1:6420", "session addr")
	flag.StringVar(&platform, "pf", "PC", "platform")
	flag.StringVar(&deviceid, "did", "zj_pc_hp", "deviceid")
	flag.StringVar(&mobiletype, "mt", "ios", "mobile type")
	flag.StringVar(&uid, "uid", "dev_test_zj", "user id")
	flag.IntVar(&appid, "appid", 2090, "appid")
	flag.IntVar(&num, "num", 1, "test num")

	flag.Parse()

	fmt.Println("init done")
}

func testOpenSession(n int) {
	/*
		OpenSession(userId string, appId uint16, isLoginUser bool, sessionKey string,
			connectionId logic.ConnectionId, gatewayAddr string, property map[string]string)
	*/

	logic.DynamicConf().LocalSessionRpc = sessaddr

	for {
		prop := map[string]string{}
		prop["ServerRam"] = "zj_serverram"
		prop["SenderType"] = "jid"
		prop["Platform"] = platform
		prop["ClientIp"] = "127.0.0.1"
		prop["LoginTime"] = time.Now().String()
		prop["Deviceid"] = deviceid
		prop["MobileType"] = mobiletype
		prop["CVersion"] = "100"

		baseuid := []byte(uid)
		tuid := string(append(baseuid, strconv.Itoa(n)...))

		resp, err := session.Open(tuid, uint16(appid), true, "zj_sesskey", logic.ConnectionId(connid), "127.0.0.1:6220", prop)
		if err != nil {
			fmt.Println("Opensession err is ", err)
		} else {
			for _, v := range resp.OldUserSessions {
				fmt.Printf("Opensession resp oldusersession is %#v\n", v)
			}

			fmt.Printf("Opensession resp.tags is %#v\n", resp.Tags)
		}

		if n--; n == 0 {
			break
		}
	}
}

func testCloseSession() {
	/*
		OpenSession(userId string, appId uint16, isLoginUser bool, sessionKey string,
			connectionId logic.ConnectionId, gatewayAddr string, property map[string]string)
	*/

	logic.DynamicConf().LocalSessionRpc = sessaddr
	prop := map[string]string{}
	prop["Platform"] = platform
	prop["LoginTime"] = time.Now().String()
	prop["Deviceid"] = deviceid

	resp, err := session.Close(uid, uint16(appid), "127.0.0.1:6220", logic.ConnectionId(connid), prop)
	if err != nil {
		fmt.Println("Closesession err is ", err)
	} else {
		fmt.Printf("Closesession resp is %#v\n", resp)
	}
}

func testQuerySession() {
	logic.DynamicConf().LocalSessionRpc = sessaddr
	prop := map[string]string{}
	prop["Platform"] = platform
	prop["LoginTime"] = time.Now().String()
	prop["Deviceid"] = deviceid

	baseuid := []byte(uid)

	req := []*session.UserSession{}

	for i := num; i > 0; i-- {
		user := &session.UserSession{}
		user.UserId = string(append(baseuid, strconv.Itoa(i)...))
		user.AppId = uint16(appid)

		req = append(req, user)
	}

	resp, err := session.Query(req)
	if err != nil {
		fmt.Println("Querysession err is ", err)
	} else {
		for _, v := range resp {
			fmt.Printf("Opensession resp session is %#v\n", v)
		}

		fmt.Printf("Querysession resp len is %v\n", len(resp))
	}
}

func main() {

	fmt.Println("connid = ", connid)

	switch testcase {
	case "open":
		testOpenSession(num)
	case "close":
		testCloseSession()
	case "query":
		testQuerySession()
	}
}
