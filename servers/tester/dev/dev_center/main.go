package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"math/rand"

	"github.com/huajiao-tv/qchat/client/center"
	"github.com/huajiao-tv/qchat/utility/cryption"
	"github.com/yvasiyarov/php_session_decoder/php_serialize"
)

const (
	GroupSend   = "group_send"
	GroupChat   = "group_chat"
	PushAll     = "push_all"
	PushUser    = "push_user"
	PrivateChat = "priv_chat"
	HighSend    = "high_send"
	ChatRoom    = "chatroom"
	CRPriv      = "chatroom_priv"
	Broadcast   = "bc"
	Qps         = "qps"
	Ops         = "ops"
	Group       = "group"
	Zan         = "z"

	RoomID   = "rid"
	GroupID  = "groupid"
	Sender   = "sender"
	Receiver = "receiver"
	AppID    = "appid"
	Content  = "msg"
	Huajiao  = "hj"
	GroupInf = "groupinf"
)

var (
	help       string
	command    string
	server     string
	room       string
	groupid    string
	groupids   string
	sender     string
	receiver   string
	appid      string
	content    string
	huajiao    string
	httpClient *http.Client
	centeraddr string
	priority   string
	groupinf   string
	ownerid    string
	usercount  string
	count      string
	sinceid    string
	userid     string
	userids    string
	version    string
	typ        string
	summary    string
	number     int
)

func init() {
	flag.StringVar(&help, "help", "", "help command")
	flag.StringVar(&command, "cmd", "", "dev_center command")
	flag.StringVar(&server, "server", "http://127.0.0.1:6600", "center server address")
	flag.StringVar(&priority, "priority", "1", "message priority")

	flag.StringVar(&room, RoomID, "", "chatroom id")
	flag.StringVar(&groupid, GroupID, "", "group id")
	flag.StringVar(&groupids, "groupids", "", "group list")
	flag.StringVar(&sender, Sender, "admin", "sender id")
	flag.StringVar(&receiver, Receiver, "", "receiver id")
	flag.StringVar(&appid, AppID, "2080", "appid")
	flag.StringVar(&content, Content, "", "message content")
	flag.StringVar(&huajiao, Huajiao, "0", "if use huajiao json format")
	flag.StringVar(&centeraddr, "rpc", "127.0.0.1:6620", "center addr")
	flag.StringVar(&groupinf, GroupInf, "", "group interface")
	flag.StringVar(&ownerid, "ownerid", "", "owner id")
	flag.StringVar(&usercount, "usercount", "", "user count")
	flag.StringVar(&count, "count", "", "count of records returned")
	flag.StringVar(&sinceid, "sinceid", "", "offset of records returned")
	flag.StringVar(&userid, "userid", "", "user id")
	flag.StringVar(&userids, "userids", "", "user list")
	flag.StringVar(&version, "version", "", "version")
	flag.StringVar(&typ, "type", "", "type")
	flag.StringVar(&summary, "summary", "", "summary")
	flag.IntVar(&number, "n", 1, "repeate number")

	flag.Parse()

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

func usage() string {
	return `
dev_center is a tool for center HTTP interface development

Usage:
	dev_center -cmd command [arguments]

The commands are:
	push_all    push notify to all users
	push_user   push notify to a specified user
	priv_chat   send private notify chat message
	chatroom    push chatroom message

Use "dev_center -help [command]" for more information about a command.
	`
}

func command_help(command string) string {
	switch command {
	case PushAll:
		return `
usage: dev_center -cmd push_all -msg <content> [-server "http://127.0.0.1:6600"] [-appid "2080"] [-sender "admin"]

The arguments are:
	msg         message content
	server      server address [default: "http://127.0.0.1:6600"]
	appid       appid [default: "2080"]
	sender      sender id [default: "admin"]

		`
	case PushUser:
		return `
usage: dev_center -cmd push_user -receiver <uids> -msg <content> [-server "http://127.0.0.1:6600"] [-appid "2080"] [-sender "admin"]

The arguments are:
	receiver    receiver uids
	msg         message content
	server      server address [default: "http://127.0.0.1:6600"]
	appid       appid [default: "2080"]
	sender      sender id [default: "admin"]

		`
	case PrivateChat:
		return `
usage: dev_center -cmd priv_chat -receiver <uids> -msg <content> [-server "http://127.0.0.1:6600"] [-appid "2080"] [-sender "admin"]

The arguments are:
	receiver    receiver uids
	msg         message content
	server      server address [default: "http://127.0.0.1:6600"]
	appid       appid [default: "2080"]
	sender      sender id [default: "admin"]

		`
	case HighSend:
		return `
usage: dev_center -cmd high_send -rid <roomid> -msg <content> [-server "http://127.0.0.1:6600"] [-appid "2080"] [-sender "admin"] [-type 0]

The arguments are:
	rid         chat room id
	msg         message content
	server      server address [default: "http://127.0.0.1:6600"]
	appid       appid [default: "2080"]
	sender      sender id [default: "admin"]
	priority    message priority [default: "1"]
	type        message type [default: "0"]

		`
	case ChatRoom:
		return `
usage: dev_center -cmd chatroom -rid <roomid> -msg <content> [-server "http://127.0.0.1:6600"] [-appid "2080"] [-sender "admin"] [-hj "0"]

The arguments are:
	rid         chat room id
	msg         message content
	server      server address [default: "http://127.0.0.1:6600"]
	appid       appid [default: "2080"]
	sender      sender id [default: "admin"]
	hj          if use huajiao json format [default: "0"]
	priority    message priority [default: "1"]

		`
	case Broadcast:
		return `
usage: dev_center -cmd bc -rid <roomid> -msg <content> [-server "http://127.0.0.1:6600"] [-appid "2080"] [-sender "admin"]

The arguments are:
	rid         chat room id
	msg         message content
	server      server address [default: "http://127.0.0.1:6600"]
	appid       appid [default: "2080"]
	sender      sender id [default: "admin"]
	hj          if use huajiao json format [default: "0"]
	priority    message priority [default: "1"]

		`
	case Group:
		return `
usage: dev_center -

The arguments are:
	join
	create
	quit
	userlist
	dismiss
	ismemeber
	info
	joinlist
	update
	send
	messagelist
	joincount
		`
	default:
		return "\nerror command \n" + usage()
	}
}

func main() {

	if help != "" {
		fmt.Println(command_help(help))
		return
	}

	switch command {
	case Group:
		fmt.Println(group())
	case PushAll:
		fmt.Println(pushAll())
	case GroupSend:
		fmt.Println(groupSend())
	case GroupChat:
		fmt.Println(groupChat())
	case PushUser:
		fmt.Println(pushUser())
	case PrivateChat:
		fmt.Println(privateChat())
	case ChatRoom:
		fmt.Println(chatroom())
	case CRPriv:
		fmt.Println(chatroomPriv())
	case HighSend:
		fmt.Println(highSend())
	case Qps:
		getQps()
	case Ops:
		getTotalOps()
	case Zan:
		fmt.Println(zan())
	case Broadcast:
		fmt.Println(broadcast())
	default:
		fmt.Println(usage())
	}
}

func checkArgs(args ...string) error {
	var value string
	for _, arg := range args {
		switch arg {
		case RoomID:
			value = room
		case Sender:
			value = sender
		case Receiver:
			value = receiver
		case AppID:
			value = appid
		case Content:
			value = content
		case GroupID:
			value = groupid
		case GroupInf:
			value = groupinf
		}

		if value == "" {
			return errors.New(fmt.Sprintf("missed arg:%s", arg))
		}
	}
	return nil
}

func group() string {
	if typ == "create" {
		uri := fmt.Sprintf("%s/groupchat/create", server)
		body := url.Values{
			"ownerid": []string{"100"},
			"traceid": []string{fmt.Sprint(time.Now().UnixNano())},
			"notice":  []string{"300000"},
			"userids": []string{"300000"},
			"groupid": []string{"300000"},
		}
		values := url.Values{
			"b": []string{cryption.Base64Encode([]byte(body.Encode()))},
		}
		values.Set("m", cryption.Md5(body.Encode()+"wpms"))

		if resp, err := httpClient.PostForm(uri, values); err != nil {
			return fmt.Sprintf("group err:%s", err.Error())
		} else {
			a, k := ioutil.ReadAll(resp.Body)
			return fmt.Sprintf("group success\n%v, ,,,%v", string(a), k)
		}
	} else if typ == "join" {
		uri := fmt.Sprintf("%s/groupchat/join", server)
		body := url.Values{
			"traceid": []string{fmt.Sprint(time.Now().UnixNano())},
			"userids": []string{"200000"},
			"groupid": []string{"300000"},
		}
		values := url.Values{
			"b": []string{cryption.Base64Encode([]byte(body.Encode()))},
		}
		values.Set("m", cryption.Md5(body.Encode()+"wpms"))
		if resp, err := httpClient.PostForm(uri, values); err != nil {
			return fmt.Sprintf("group err:%s", err.Error())
		} else {
			a, k := ioutil.ReadAll(resp.Body)
			return fmt.Sprintf("group success\n%v, ,,,%v", string(a), k)
		}
	} else if typ == "quit" {
		uri := fmt.Sprintf("%s/groupchat/quit", server)
		body := url.Values{
			"traceid": []string{fmt.Sprint(time.Now().UnixNano())},
			"userids": []string{"200000"},
			"groupid": []string{"300000"},
		}
		values := url.Values{
			"b": []string{cryption.Base64Encode([]byte(body.Encode()))},
		}
		values.Set("m", cryption.Md5(body.Encode()+"wpms"))
		if resp, err := httpClient.PostForm(uri, values); err != nil {
			return fmt.Sprintf("group err:%s", err.Error())
		} else {
			a, k := ioutil.ReadAll(resp.Body)
			return fmt.Sprintf("group success\n%v, ,,,%v", string(a), k)
		}
	} else {
		return "err_req"
	}
}

func pushAll() string {
	if err := checkArgs(); err != nil {
		return fmt.Sprintf("\n%s\n%s", err.Error(), command_help(PushAll))
	}

	uri := fmt.Sprintf("%s/huajiao/all", server)
	body := url.Values{
		"msg":         []string{content},
		"msgtype":     []string{"100"},
		"traceid":     []string{fmt.Sprint(time.Now().UnixNano())},
		"expire_time": []string{"300000"},
	}
	values := url.Values{
		"b": []string{cryption.Base64Encode([]byte(body.Encode()))},
		"m": []string{"0"},
	}

	if resp, err := httpClient.PostForm(uri, values); err != nil {
		return fmt.Sprintf("push all user notify err:%s", err.Error())
	} else {
		return fmt.Sprintf("success\n%v", resp)
	}
}

func groupChat() string {
	if err := checkArgs(GroupInf); err != nil {
		return fmt.Sprintf("\n%s\n%s", err.Error(), "grouChat groupinf")
	}

	uri := fmt.Sprintf("%s/groupchat/%s", server, groupinf)
	body := url.Values{
		"groupid":   []string{groupid},
		"appid":     []string{appid},
		"ownerid":   []string{ownerid},
		"sender":    []string{sender},
		"userids":   []string{userids},
		"version":   []string{version},
		"groupids":  []string{groupids},
		"content":   []string{content},
		"usercount": []string{usercount},
		"userid":    []string{userid},
		"summary":   []string{summary},
		"count":     []string{count},
		"sinceid":   []string{sinceid},
		"msg":       []string{content},
		"type":      []string{typ},
		"traceid":   []string{fmt.Sprint(time.Now().UnixNano())},
	}
	values := url.Values{
		"b": []string{cryption.Base64Encode([]byte(body.Encode()))},
		"m": []string{cryption.Md5(body.Encode() + "wpms")},
	}

	if resp, err := httpClient.PostForm(uri, values); err != nil {
		return fmt.Sprintf("groupchat/%s err:%s", groupinf, err.Error())
	} else {
		s, e := ioutil.ReadAll(resp.Body)
		return fmt.Sprintf("groupchat/%s succ:%v,%v\n", groupinf, string(s), e)
	}
}

func groupSend() string {
	if err := checkArgs(GroupID, Content, Sender); err != nil {
		return fmt.Sprintf("\n%s\n%s", err.Error(), "groupSend sender, msg, groupid")
	}

	uri := fmt.Sprintf("%s/groupchat/send", server)
	body := url.Values{
		"groupid": []string{groupid},
		"appid":   []string{appid},
		"traceid": []string{fmt.Sprint(time.Now().UnixNano())},
		"sender":  []string{sender},
		"type":    []string{typ},
		"summary": []string{summary},
		"content": []string{content},
	}
	values := url.Values{
		"b": []string{cryption.Base64Encode([]byte(body.Encode()))},
		"m": []string{cryption.Md5(body.Encode() + "wpms")},
	}

	if resp, err := httpClient.PostForm(uri, values); err != nil {
		return fmt.Sprintf("send group err:%s", err.Error())
	} else {
		s, e := ioutil.ReadAll(resp.Body)
		return fmt.Sprintf("send group succ:%v,%v\n", string(s), e)
	}
}

func pushUser() string {
	if err := checkArgs(Content, Receiver); err != nil {
		return fmt.Sprintf("\n%s\n%s", err.Error(), command_help(PushUser))
	}

	msgFmt := "{\"userid\":%v,\"type\":1,\"text\":\"%v \u5173\u6ce8\u4e86\u4f60\"," +
		"\"time\":%v,\"expire\":86400,\"extends\":{\"uid\":\"%v\",\"userid\":\"%v\",\"nickname\":\"%v\"," +
		"\"avatar\":\"http:\\/\\/image.huajiao.com\\/d2bd9628df60a6893539a75b6d0db0a9-100_100.jpg\"," +
		"\"exp\":450077296,\"level\":84,\"verified\":false," +
		"\"verifiedinfo\":{\"credentials\":\"\",\"type\":0,\"realname\":\"%v\",\"status\":0," +
		"\"error\":\"\",\"official\":false},\"creatime\":\"%v\"},\"traceid\":\"%v\"}"
	uri := fmt.Sprintf("%s/huajiao/users", server)

	for i := 0; i < number; i++ {
		body := url.Values{
			"msg": []string{fmt.Sprintf(msgFmt, receiver, sender,
				time.Now().Unix(), sender, sender, sender, sender,
				time.Now().Format("2006-01-02 15:04:05"), time.Now().UnixNano())},
			"msgtype":     []string{"100"},
			"traceid":     []string{fmt.Sprint(time.Now().UnixNano())},
			"expire_time": []string{"10800"},
			"sender":      []string{sender},
			"receivers":   []string{receiver},
		}
		values := url.Values{
			"b": []string{cryption.Base64Encode([]byte(body.Encode()))},
			"m": []string{"0"},
		}

		if _, err := httpClient.PostForm(uri, values); err != nil {
			fmt.Println("index:", i, "push user notify err:", err.Error())
		} else {
			fmt.Println("index:", i)
		}

		time.Sleep(time.Duration(100+rand.Intn(900)) * time.Millisecond)
	}

	return "test done\n"
}

func privateChat() string {
	if err := checkArgs(Content, Receiver); err != nil {
		return fmt.Sprintf("\n%s\n%s", err.Error(), command_help(PrivateChat))
	}

	msgFmt := "{\"userid\":\"%v\",\"type\":39,\"text\":\"" +
		"{\\\"type\\\":1,\\\"text\\\":\\\"ha ha ha\\\",\\\"url1\\\":\\\"\\\",\\\"url2\\\":\\\"\\\"}\",\"time\":%v," +
		"\"expire\":10800,\"extends\":{\"messageid\":75756914,\"userid\":\"%v\",\"nickname\":\"%v\"," +
		"\"avatar\":\"http:\\/\\/image.huajiao.com\\/ea39d896d387faf6cab0ae5f4edae972-100_100.jpg\"," +
		"\"verified\":false,\"verifiedinfo\":{\"credentials\":\"\",\"type\":0,\"realname\":\"%v\"," +
		"\"status\":0,\"error\":\"\",\"official\":false}," +
		"\"verify_student\":{\"vs_status\":0,\"option_student\":\"Y\",\"vs_realname\":\"\",\"vs_school\":\"\"}," +
		"\"exp\":3858328,\"level\":30,\"followed\":false,\"isfriends\":false,\"seqid\":\"%v|%v\"}}"

	//msgFmt := "{\"userid\":\"40000419\",\"type\":39,\"text\":\"{\"type\":1,\"text\":\"ha ha ha\",\"url1\":\"\",\"url2\":\"\"}\",\"time\":1473132861,\"expire\":604800,\"extends\":{\"messageid\":1016,\"userid\":\"40000409\",\"nickname\":\"\u5b59\u534e\",\"avatar\":\"http:\\/\\/image.huajiao.com\\/d2bd9628df60a6893539a75b6d0db0a9-100_100.jpg\",\"verified\":false,\"verifiedinfo\":{\"credentials\":\"\",\"type\":0,\"realname\":\"\u5b59\u534e\",\"status\":0,\"error\":\"\",\"official\":false},\"verify_student\":{\"vs_status\":0,\"option_student\":\"Y\",\"vs_realname\":\"\",\"vs_school\":\"\"},\"exp\":450102832,\"level\":84,\"followed\":false,\"isfriends\":false,\"seqid\":\"40000419|1473132861880.701904\"},\"traceid\":\"f7c98ad730a9cefd125141a870c6f3bd\"}\"
	uri := fmt.Sprintf("%s/huajiao/chat", server)

	for i := 0; i < number; i++ {
		body := url.Values{
			"msg": []string{fmt.Sprintf(msgFmt, receiver, time.Now().Unix(), sender,
				sender, sender, sender, time.Now().UnixNano())},
			"msgtype":     []string{"100"},
			"traceid":     []string{fmt.Sprint(time.Now().UnixNano())},
			"expire_time": []string{"10800"},
			"sender":      []string{sender},
			"receivers":   []string{receiver},
		}
		values := url.Values{
			"b": []string{cryption.Base64Encode([]byte(body.Encode()))},
			"m": []string{"0"},
		}

		if _, err := httpClient.PostForm(uri, values); err != nil {
			fmt.Println("index:", i, "user chat notify err:", err.Error())
		} else {
			fmt.Println("index:", i)
		}
		time.Sleep(time.Duration(100+rand.Intn(900)) * time.Millisecond)
	}

	return "test done\n"
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
	Num          int          `json:"num"`
	Total        int          `json:"total"`
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

/*
{
    "roomid": "10544917",
    "type": 9,
    "text": "xxxxxxxx",
    "time": 1461131141,
    "expire": 86400,
    "extends": {
        "liveid": "10544917",
        "userid": "27628584",
        "nickname": "废品厂厂长",
        "avatar": "http://image.huajiao.com/527dc0ebc228e52db82dbf9d2b2e13fc-100_100.jpg",
        "verified": false,
        "verifiedinfo": {
            "credentials": "我只是个拣破烂的",
            "type": 0,
            "realname": "废品厂厂长",
            "status": 0,
            "error": "",
            "official": false
        },
        "gift": 0,
        "exp": 15270,
        "level": 4
    }
}
*/
func huajiaoChatRoomMessage(room, text string) string {
	avatar := "http://image.huajiao.com/5f2908e7b902f70b39cbbbbe37a4c344-100_100.jpg"
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
func chatroomPriv() string {
	if err := checkArgs(RoomID, Content); err != nil {
		return fmt.Sprintf("\n%s\n%s", err.Error(), command_help(ChatRoom))
	}

	var message string
	uri := fmt.Sprintf("%s/chatroom/private/send", server)
	if huajiao == "0" {
		message = content
	} else {
		message = huajiaoChatRoomMessage(room, content)
	}
	values := url.Values{
		"roomid":   []string{room},
		"sender":   []string{sender},
		"receiver": []string{receiver},
		"traceid":  []string{fmt.Sprint(time.Now().UnixNano())},
		"content":  []string{message},
		"appid":    []string{appid},
		"priority": []string{priority},
	}

	if resp, err := httpClient.PostForm(uri, values); err != nil {
		return fmt.Sprintf("send chatroom message err:%s", err.Error())
	} else {
		return fmt.Sprintf("success\n%v", resp)
	}
}

func highSend() string {
	if err := checkArgs(RoomID, Content); err != nil {
		return fmt.Sprintf("\n%s\n%s", err.Error(), command_help(HighSend))
	}

	uri := fmt.Sprintf("%s/chatroom/send/high/priority", server)
	body := url.Values{
		"content":  []string{content},
		"type":     []string{typ},
		"traceid":  []string{fmt.Sprint(time.Now().UnixNano())},
		"roomid":   []string{room},
		"priority": []string{priority},
		"sender":   []string{sender},
	}
	values := url.Values{
		"b": []string{cryption.Base64Encode([]byte(body.Encode()))},
		"m": []string{cryption.Md5(body.Encode() + "wpms")},
	}

	if resp, err := httpClient.PostForm(uri, values); err != nil {
		return fmt.Sprintf("chatroom send high priority", err.Error())
	} else {
		a, k := ioutil.ReadAll(resp.Body)
		return fmt.Sprintf("success\n%v\n%v", string(a), k)
	}
}

func chatroom() string {
	if err := checkArgs(RoomID, Content); err != nil {
		return fmt.Sprintf("\n%s\n%s", err.Error(), command_help(ChatRoom))
	}

	var message string
	uri := fmt.Sprintf("%s/send", server)
	if huajiao == "0" {
		message = content
	} else {
		message = huajiaoChatRoomMessage(room, content)
	}
	values := url.Values{
		"roomid":   []string{room},
		"sender":   []string{sender},
		"traceid":  []string{fmt.Sprint(time.Now().UnixNano())},
		"content":  []string{message},
		"appid":    []string{appid},
		"priority": []string{priority},
	}

	if resp, err := httpClient.PostForm(uri, values); err != nil {
		return fmt.Sprintf("send chatroom message err:%s", err.Error())
	} else {
		return fmt.Sprintf("success\n%v", resp)
	}
}

func broadcast() string {
	if err := checkArgs(); err != nil {
		return fmt.Sprintf("\n%s\n%s", err.Error(), command_help(ChatRoom))
	}

	uri := fmt.Sprintf("%s/online/broadcast", server)
	body := url.Values{
		"roomid":   []string{room},
		"sender":   []string{sender},
		"traceid":  []string{fmt.Sprint(time.Now().UnixNano())},
		"content":  []string{content},
		"appid":    []string{appid},
		"priority": []string{priority},
	}
	values := url.Values{
		"b": []string{cryption.Base64Encode([]byte(body.Encode()))},
		"m": []string{"0"},
	}

	if resp, err := httpClient.PostForm(uri, values); err != nil {
		return fmt.Sprintf("push all user notify err:%s", err.Error())
	} else {
		return fmt.Sprintf("success\n%v", resp)
	}
}

func z() string {
	avatar := "http://image.huajiao.com/5f2908e7b902f70b39cbbbbe37a4c344-100_100.jpg"
	timestamp := time.Now().UnixNano()
	c, _ := strconv.Atoi(count)
	data := php_serialize.PhpArray{
		"roomid": room,
		"type":   8,
		"text":   nil,
		"time":   timestamp,
		"expire": 86400,
		"extends": php_serialize.PhpArray{
			"liveid":   room,
			"num":      c,
			"userid":   sender,
			"total":    c,
			"nickname": "阳光都是我前世的盼望9w",
			"verified": false,
			"verifiedinfo": php_serialize.PhpArray{
				"credentials": "",
				"type":        0,
				"realname":    "阳光都是我前世的盼望9w",
				"status":      0,
				"error":       "",
				"official":    false,
			},
			"exp":   0,
			"level": 1,
		},
	}
	value, err := php_serialize.NewSerializer().Encode(data)
	if err != nil {
		fmt.Println("php serialize error:", err)
		return ""
	}

	// @todo, 点赞数ios和安卓会对num、total字段做不通用途
	msg := &message{
		Roomid:    room,
		MsgType:   8,
		TimeStamp: timestamp,
		Expire:    86400,
		Extends: extend{
			LiveId:   room,
			Num:      c,
			Sender:   sender,
			Total:    c,
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

func zan() string {
	if err := checkArgs(RoomID); err != nil {
		return fmt.Sprintf("\n%s", err.Error())
	}

	uri := fmt.Sprintf("%s/send", server)
	message := z()
	values := url.Values{
		"roomid":   []string{room},
		"sender":   []string{sender},
		"traceid":  []string{fmt.Sprint(time.Now().UnixNano())},
		"content":  []string{message},
		"appid":    []string{appid},
		"priority": []string{priority},
	}

	if resp, err := httpClient.PostForm(uri, values); err != nil {
		return fmt.Sprintf("send chatroom message err:%s", err.Error())
	} else {
		return fmt.Sprintf("success\n%v\n%v", message, resp)
	}
}

func getQps() {
	if stat, err := center.GetCenterQps(centeraddr); err != nil {
		fmt.Println("testGetQps failed, error: ", err)
	} else {
		fmt.Println(centeraddr, "QPS:", stat.QpsString())
	}
}

func getTotalOps() {
	if stat, err := center.GetCenterTotalOps(centeraddr); err != nil {
		fmt.Println("testGetTotalOps failed, error: ", err)
	} else {
		fmt.Println(centeraddr, "all operation information:", stat.String())
	}
}
