package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/huajiao-tv/qchat/servers/tester/ben/ben_llconn/client"
	"github.com/huajiao-tv/qchat/utility/cpool"
)

var (
	verbose bool
	cmdPool *cpool.ConsumerPool
	closing chan string

	l        sync.RWMutex
	commands map[string]chan []string
)

func init() {
	rand.Seed(time.Now().UnixNano())
	closing = make(chan string, 100)
	commands = make(map[string]chan []string, 100)
	cmdPool = cpool.NewConsumerPool(100, 10000, execCommandFn)
	flag.BoolVar(&verbose, "vvv", false, "verbose")
	flag.Parse()
}

func main() {
	input := bufio.NewReader(os.Stdin)

	usage()
	go cleanConn()

	for {
		data, _, _ := input.ReadLine()
		line := strings.Trim(string(data), "\r\n\t ")
		if len(line) == 0 {
			continue
		}

		if !strings.HasPrefix(line, "#") {
			line = "#ST " + line
		}

		cmds := strings.Split(line, " ")
		if len(cmds) <= 1 {
			continue
		}
		if "exit" == cmds[1] && len(cmds) == 2 {
			return
		}
		if ok := cmdPool.Add(cmds); !ok {
			print("channel full")
		}
	}
}

func usage() {
	fmt.Println(`
commands usage:
    #<tag> conn -s <gateway:port> -c <center:port> -u <uid>  connect to server, use 'conn help' for more args
    #<tag> http -u <URL> [-m get|post] [-a args]             send http request

    #<tag> join <roomid>                                     join chatroom
    #<tag> quit <roomid>                                     quit chatroom
    #<tag> query <roomid> <start> <count>                    query chatroom
    #<tag> getmsg <roomid> <id1[,id2...]>                    pull chatroom message
    #<tag> chat <roomid> <content> [priority]                send chatroom message
    #<tag> peer <receiver1,[receiver2...]> <content>         send peer message
    #<tag> im <receiver1,[receiver2...]> <content>           send im message
    #<tag> disconn                                           close the connection of user refered by the tag

	`)
}

func connHelp() {
	fmt.Println(`
conn arguments:
    -s <gateway_addr:port>             server address
    -c <center_addr:port>              center address
    -u userid                          userid
    [-a appid]                         appid, default: 2080
    [-t stream_type]                   stream type: tcp or ws, default: tcp
    [-h heartbeat]                     heartbeat rate, default: 60

	`)
}

func execCommandFn(c interface{}) {
	cmds, ok := c.([]string)
	if !ok {
		return
	}
	if len(cmds) < 2 {
		usage()
		return
	}

	// 全局指令(需tag)
	switch cmds[1] {
	case "conn":
		addAccount(cmds[0], cmds[2:])
		return
	case "http":
		client.Http(cmds[0], cmds[2:])
		return
	}
	// 用户指令
	switch cmds[1] {
	case "join", "robotjoin", "quit", "robotquit", "query", "getmsg", "chat", "peer", "im", "wg",
		"creategroup", "joingroup", "quitgroup", "dismissgroup", "listgroupuser", "getgroupinfo",
		"ingroups", "ismember", "listcreatedgroup", "sendgroupmsg", "joincount", "groupmsg", "groupmsgbatch", "groupsync", "disconn":
		l.RLock()
		ch, ok := commands[cmds[0]]
		l.RUnlock()
		if !ok {
			print("invalid tag or userid", cmds[0])
		} else {
			ch <- cmds[1:]
		}
	default:
		usage()
	}
}

func addAccount(tag string, cmds []string) {
	var (
		appid         int
		autologin     int
		key           string
		connType      string
		server        string
		center        string
		userid        string
		password      string
		sig           string
		deviceid      string
		heartbeat     int
		sendheartbeat int
		clientVer     int
	)
	connFlag := flag.NewFlagSet("conn", flag.ContinueOnError)

	connFlag.StringVar(&server, "s", "", "server address")
	connFlag.StringVar(&center, "c", "", "center address")
	connFlag.StringVar(&userid, "u", "", "userid")
	connFlag.StringVar(&password, "p", "", "password")
	connFlag.StringVar(&sig, "sig", "", "signature to obtain token")
	connFlag.StringVar(&connType, "t", "tcp", "stream type")
	connFlag.StringVar(&key, "k", "2817cf16edcac37d2762ff2520705b30", "key")
	connFlag.StringVar(&deviceid, "did", "", "signature to obtain token")

	connFlag.IntVar(&appid, "a", 2333, "appid")
	connFlag.IntVar(&autologin, "al", 1, "auto login")
	connFlag.IntVar(&heartbeat, "h", 60, "heartbeat rate")
	connFlag.IntVar(&sendheartbeat, "shb", 1, "don't send heart beat when value is 0")
	connFlag.IntVar(&clientVer, "cv", 102, "heartbeat rate")

	connFlag.Parse(cmds)
	fmt.Println(server, userid, heartbeat)
	if server == "" || userid == "" || heartbeat < 1 {
		connHelp()
		return
	}
	conf := &client.ClientConf{
		Tag:           tag,
		ConnType:      connType,
		ServerAddr:    server,
		CenterAddr:    center,
		Heartbeat:     heartbeat,
		SendHeartbeat: sendheartbeat,
		AppID:         uint32(appid),
		ClientVer:     clientVer,
		ProtoVer:      1,
		DefaultKey:    key,
		AutoLogin:     autologin,
	}

	if len(password) == 0 {
		password = userid
	}

	account := &client.AccountInfo{
		UserID:    userid,
		Password:  password, // todo:
		Platform:  "ben",
		Signature: sig,
		DeviceID:  deviceid,
	}
	c := client.NewUserConnection(conf, account, verbose)
	l.Lock()
	commands[tag] = c.CmdChan
	l.Unlock()
	go c.Start(closing)
}

func cleanConn() {
	for {
		tag := <-closing
		l.Lock()
		if ch, ok := commands[tag]; ok {
			close(ch)
		}
		delete(commands, tag)
		l.Unlock()
	}
}

func print(args ...interface{}) {
	client.Log("#0", "global", args...)
}
