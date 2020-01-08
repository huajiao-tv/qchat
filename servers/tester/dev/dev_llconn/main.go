package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"flag"

	"github.com/huajiao-tv/qchat/servers/tester/dev/dev_llconn/llconn"

	"math/rand"
	"strconv"
	"strings"
	"time"
)

func main() {

	rand.Seed(time.Now().Unix())

	streamType := flag.CommandLine.String("st", "tcp", "stream type: tcp, websocket or wss")
	server := flag.CommandLine.String("server", "127.0.0.1:9090", "gateway server address")
	center := flag.CommandLine.String("center", "127.0.0.1:8080", "node to send message, will be recover by -cr-url if -cr-url is set!")

	heartbeat := flag.CommandLine.Int("hb", 60, "heartbeat interval in second")
	sendHeartBeat := flag.CommandLine.Int("shb", 1, "send heart beat or not, if value is 0, don't send")
	pv := flag.CommandLine.Int("pv", 1, "protocol version")
	cv := flag.CommandLine.Int("cv", 100, "client version")
	dk := flag.CommandLine.String("key", "894184791415baf5c113f83eaff360f0", "default key")
	appID := flag.CommandLine.Int("appid", 1080, "application ID")
	dumpStream := flag.CommandLine.Int("ds", 1, "dump network byte stream")
	acc_type := flag.CommandLine.String("act", "jid", "account type: phone, jid, qid etc")

	randomID := strconv.FormatInt(rand.Int63n(20000000)+int64(os.Getpid()), 10)

	uid := flag.CommandLine.String("uid", randomID, "signature to obtain token")
	pwd := flag.CommandLine.String("pwd", *uid, "password of uid; if sig is empty, pwd is the same as uid")
	sig := flag.CommandLine.String("sig", "", "signature to obtain token")
	did := flag.CommandLine.String("did", "", "device id")
	plf := flag.CommandLine.String("plf", "pc", "platform")
	ne := flag.CommandLine.Bool("ne", true, "not encrypt")

	//新加roomid参数，如果不为空则执行完 join 后退出
	roomid := flag.CommandLine.String("roomid", "", "roomid to join, if not empty, only run join then exit")

	chatroomUrl := flag.CommandLine.String("cr-url", "", "node to send chatroom message")

	getgwy := flag.CommandLine.String("get-gwy", "", "get gateway address, params value is cluster, e.g. qchat_online")
	gwyport := flag.CommandLine.String("gwy-port", "80,443", "gateway port")
	getcenter := flag.CommandLine.String("get-center", "", "get center address, params value is cluster, e.g. qchat_online")
	centerport := flag.CommandLine.String("center-port", "6666", "center port")
	getcounter := flag.CommandLine.String("get-counter", "", "get counter address, params value is cluster, e.g. qchat_online")
	counterport := flag.CommandLine.String("counter-port", "7200", "counter port")
	autoLogin := flag.CommandLine.Int("autologin", 1, "auto login after init login")
	checkLoginWhenJoin := flag.CommandLine.Int("clwj", 1, "check login when join chatroom")

	flag.Parse()

	if *chatroomUrl != "" {
		tmp := strings.Split(*chatroomUrl, "/")
		*center = tmp[2]
	}

	if *getgwy != "" {
		gatewayNodes := getNodes(*getgwy, "gateway", *gwyport)
		fmt.Println(gatewayNodes)
		return
	}

	if *getcenter != "" {
		centerNodes := getNodes(*getcenter, "center", *centerport)
		fmt.Println(centerNodes)
		return
	}

	if *getcounter != "" {
		counterNodes := getNodes(*getcounter, "counter", *counterport)
		fmt.Println(counterNodes)
		return
	}

	clientConf := llconn.ClientConf{*streamType, *server, *appID, *cv, *pv, *dk, *heartbeat, *sendHeartBeat,
		*autoLogin, *checkLoginWhenJoin}

	accountConf := llconn.AccountInfo{*acc_type, *uid, *pwd, *sig, *did, *plf, *ne}

	ms, _ := llconn.NewMessageSender(*center)

	runInteractively(clientConf, accountConf, ms, *dumpStream > 0, *roomid)

	// time.Sleep(1e9)
}

func stringToIntArray(input string) []int64 {

	if len(input) == 0 {
		return nil
	}

	ids := make([]int64, 0)

	for _, s := range strings.Split(input, ",") {
		s = strings.Trim(s, "\r\n\t ")
		i, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			continue
		}
		ids = append(ids, i)
	}
	return ids
}

func readInput(readChan chan string) {

	if readChan == nil {
		return
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("reader.ReadString failed: ", err.Error())
			readChan <- "exit"
			return
		} else {
			line = strings.Trim(line, "\r\n\t ")
			readChan <- line
		}
	}
}

func runInteractively(clientConf llconn.ClientConf, accountConf llconn.AccountInfo, ms *llconn.MessageSender, dump_stream bool, roomid string) {

	var usage string = "usage: join <roomid>|query <roomid> <start,count>|quit <roomid>|getmsg <id list separated by comma>|robotjoin <roomid> <userid> |robotquit <roomid> <userid> |chat <roomid> <content>| peer <receiver> <content>|public <content>|groupmsg <groupid> <startid,offset>|groupsync <anything>|exit\n"

	conn, err := llconn.New(clientConf, accountConf)
	if err != nil {
		fmt.Println("hjconn.New failed", err)
		return
	}

	chatroomSrv := llconn.NewChatroom()
	groupSrv := llconn.NewGroup()
	conn.Register(chatroomSrv)
	conn.Register(groupSrv)

	conn.SetDumpStream(dump_stream)
	readChan := make(chan string, 100)
	conn.Start(readChan)

	go readInput(readChan)

	if roomid != "" {
		time.Sleep(time.Millisecond * 500) // sleep 500ms, 等待连上
		if chatroomSrv.Join(roomid) > 0 {
			fmt.Println("join room success", roomid)
		} else {
			fmt.Println("join room failed ", roomid, "failed")
		}
		time.Sleep(time.Millisecond * 100) // 等待 tcp 写成功
		return
	}

	for {
		select {

		case line := <-readChan:
			if line == "" {
				continue
			}

			lines := strings.Split(line, " ")
			argLen := len(lines)
			cmd := lines[0]
			if cmd == "exit" {
				return
			}

			if argLen < 2 && cmd != "groupsync" {
				fmt.Println(usage)
				continue
			}

			var args string
			if len(lines) >= 2 {
				args = strings.Trim(lines[1], "\r\n\t ")
			}

			switch cmd {
			case "join":
				if argLen != 2 {
					fmt.Println(usage)
					continue
				}
				if chatroomSrv.Join(args) > 0 {
					fmt.Println("join room ", args)
				} else {
					fmt.Println("join room failed ", args, "failed")
				}
			case "quit":
				if argLen != 2 {
					fmt.Println(usage)
					continue
				}
				if chatroomSrv.Quit(args) > 0 {
					fmt.Println("quit from room ", args)
				} else {
					fmt.Println("quit from room ", args, "failed")
				}

			case "query":
				if argLen != 3 {
					fmt.Println(usage)
					continue
				}
				ids := stringToIntArray(lines[2])
				if ids == nil || len(ids) != 2 {
					fmt.Println(usage)
					continue
				}

				if chatroomSrv.Query(lines[1], int32(ids[0]), int32(ids[1])) > 0 {
					fmt.Println("query room ", lines[1])
				} else {
					fmt.Println("query room ", lines[1], "failed")
				}
			case "groupmsg":
				if argLen != 3 {
					fmt.Println(usage)
					continue
				}
				ids := stringToIntArray(lines[2])
				if ids == nil || len(ids) == 0 {
					fmt.Println(usage)
					continue
				}

				groupSrv.GetMsg(lines[1], uint64(ids[0]), int32(ids[1]))
				fmt.Println("groupmsg,", lines[1])
			case "groupmsgbatch":
				ids := stringToIntArray(lines[1])
				groupSrv.GetMsgM(ids)
			case "groupsync":
				var groups []string
				if len(lines) >= 2 {
					groups = strings.Split(lines[1], ",")
				}
				groupSrv.Sync(groups)

			case "getmsg":
				if argLen != 3 {
					fmt.Println(usage)
					continue
				}
				ids := stringToIntArray(lines[2])
				if ids == nil || len(ids) == 0 {
					fmt.Println(usage)
					continue
				}

				if groupSrv.GetMessage(lines[1], ids) {
					fmt.Println("GetMessage room ", lines[1])
				} else {
					fmt.Println("GetMessage room ", lines[1], "failed")
				}

			case "robotjoin":
				if argLen != 3 {
					fmt.Println(usage)
					continue
				}
				tid, err2 := ms.SendRobotJoinMessage(lines[1], lines[2])

				if nil != err2 {
					fmt.Println("SendChatroomMessage failed", err2)
				} else {
					fmt.Println("SendChatroomMessage success ", tid)
				}
			case "robotquit":
				if argLen != 3 {
					fmt.Println(usage)
					continue
				}
				tid, err2 := ms.SendRobotQuitMessage(lines[1], lines[2])

				if nil != err2 {
					fmt.Println("SendRobotQuitMessage failed", err2)
				} else {
					fmt.Println("SendRobotQuitMessage success ", tid)
				}
			case "chat":
				if argLen != 3 && argLen != 4 && argLen != 5 {
					fmt.Println(usage)
					continue
				}

				priority := int64(1)
				msgtype := int64(9)
				if argLen >= 4 && argLen <= 5 {
					var err error
					priority, err = strconv.ParseInt(lines[3], 10, 32)
					if err != nil {
						fmt.Println("priority: ", priority, " is not int")
						fmt.Println(usage)
						continue
					}
					msgtype, err = strconv.ParseInt(lines[4], 10, 32)
					if err != nil {
						fmt.Println("msgtype: ", priority, " is not int")
						fmt.Println(usage)
						continue
					}
				}
				tid, err2 := ms.SendChatroomMessage(lines[1], accountConf.ID, []byte(lines[2]), 86400, int(msgtype), int(priority), clientConf.AppID)

				if nil != err2 {
					fmt.Println("SendChatroomMessage failed", err2)
				} else {
					fmt.Println("SendChatroomMessage success ", tid)
				}

			case "peer":

				if argLen != 3 {
					fmt.Println(usage)
					continue
				}
				tid, err2 := ms.SendPeerMessage(lines[1], []byte(lines[2]), 86400, 1, clientConf.AppID)

				if nil != err2 {
					fmt.Println("SendPeerMessage failed", err2)
				} else {
					fmt.Println("SendPeerMessage success ", tid)
				}

			case "im":

				if argLen != 3 {
					fmt.Println(usage)
					continue
				}
				tid, err2 := ms.SendImMessage(lines[1], []byte(lines[2]), 86400, 1, clientConf.AppID)

				if nil != err2 {
					fmt.Println("SendImMessage failed", err2)
				} else {
					fmt.Println("SendImMessage success ", tid)
				}

			case "public":
				if argLen < 2 {
					fmt.Println(usage)
					continue
				}
				tid, err2 := ms.SendPublicMessage([]byte(lines[1]), clientConf.AppID)

				if nil != err2 {
					fmt.Println("SendPublicMessage failed", err2)
				} else {
					fmt.Println("SendPublicMessage success ", tid)
				}
			case "logout":
				chatroomSrv.Logout()

			default:
				fmt.Println(usage)
			}
		}

		time.Sleep(1e6)
	}
}

func getNodes(cluster, service, ports string) string {
	portList := strings.Split(ports, ",")
	getNodesUrl := "http://messenger.zhushou.corp.qihoo.net:8080/api/queryAllMachine?services=" + service

	resp, err := http.Get(getNodesUrl)
	if err != nil {
		return "error"
	}

	resJson := checkResponse(resp)
	if resJson == "error" {
		return resJson
	}

	var valMap map[string]map[string][]map[string]string

	json.Unmarshal([]byte(resJson), &valMap)

	serviceMap := valMap[cluster]
	var res []string

	if cluster == "qchat_online" || cluster == "qchat_ben" {
		for _, service := range serviceMap[service] {
			for _, port := range portList {
				res = append(res, service["machine"]+":"+port)
			}
		}
	} else {
		for _, service := range serviceMap[service] {
			port := "8360"
			if service["front_port"] != "0" {
				port = service["front_port"]
			}
			res = append(res, service["machine"]+":"+port)
		}
	}

	return strings.Join(res, ",")
}

func checkResponse(resp *http.Response) string {
	if nil == resp {
		fmt.Println("empty response")
		return "error"
	}

	b1, err1 := ioutil.ReadAll(resp.Body)
	if nil != err1 {
		fmt.Println("ReadAll failed")
		return "error"
	}

	res := string(b1)

	//fmt.Println("status code", resp.StatusCode, "response", res)
	return res
}
