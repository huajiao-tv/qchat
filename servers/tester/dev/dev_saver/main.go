package main

import (
	"flag"
	"fmt"
	"runtime"
	"strings"
	"time"

	"io"
	"os"
	"sort"

	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/logic"
)

var (
	testcase       string
	saveraddr      string
	receiver       string
	sender         string
	msgid          int64
	count          int
	chatChannel    string
	expireInterval int
	outbox         int
	domain         string
	group          string
	owner          string
	members        string
)

const (
	DefaultChannel = "default"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.StringVar(&saveraddr, "host", "127.0.0.1:6520", "saver addr")
	flag.StringVar(&testcase, "tc", "retrieve", "test case,like store, retrive")
	flag.StringVar(&receiver, "receiver", "jid_1000", "receiver")
	flag.StringVar(&sender, "sender", "jid_2000", "sender")
	flag.Int64Var(&msgid, "msgid", 0, "message inbox id")
	flag.IntVar(&count, "count", 0, "messages count")
	flag.StringVar(&chatChannel, "ch", DefaultChannel, "chat channel")
	flag.IntVar(&expireInterval, "ei", 604800, "message expire interval")
	flag.IntVar(&outbox, "outbox", 0, "indicates whether store message to outbox, default 0 is not")
	flag.StringVar(&domain, "d", "dev", "savers domain")
	flag.StringVar(&group, "gid", "", "group id")
	flag.StringVar(&owner, "owner", "", "owner id")
	flag.StringVar(&members, "members", "", "group members")
	flag.Parse()
}

func testUnread() {
	fmt.Println("starting retrieving unread test")
	logic.DynamicConf().LocalSaverRpc = saveraddr

	// default to retrieve all channels' messages
	var chnnels []string

	if chatChannel != DefaultChannel {
		chnnels = []string{chatChannel}
	} else {
		chnnels = []string{
			saver.ChatChannelIM,
			saver.ChatChannelNotify,
			saver.ChatChannelPublic,
		}
	}

	if resp, err := saver.RetrieveUnreadCount(2080, []string{receiver}, chnnels); err != nil {
		fmt.Println("testRetrieve failed, error: ", err)
	} else {
		for _, r := range resp {
			for channel, id := range r.LatestID {
				fmt.Println("Channel", channel, "last read ID:", r.LastReadID[channel], "Latest ID:", id)
			}
		}
	}
}

func testRetrieve() {
	fmt.Println("starting retrieving message test")
	logic.DynamicConf().LocalSaverRpc = saveraddr

	// default to retrieve all channels' messages
	var chnnels map[string]*saver.RetrieveChannel

	if chatChannel != DefaultChannel {
		chnnels = map[string]*saver.RetrieveChannel{
			chatChannel: &saver.RetrieveChannel{Channel: chatChannel, StartMsgId: msgid, MaxCount: count}}
	} else {
		chnnels = map[string]*saver.RetrieveChannel{
			saver.ChatChannelNotify: &saver.RetrieveChannel{Channel: saver.ChatChannelNotify, StartMsgId: msgid, MaxCount: count},
			saver.ChatChannelPublic: &saver.RetrieveChannel{Channel: saver.ChatChannelPublic, StartMsgId: msgid, MaxCount: count},
			saver.ChatChannelIM:     &saver.RetrieveChannel{Channel: saver.ChatChannelPublic, StartMsgId: msgid, MaxCount: count}}
	}

	fields := []string{"To", "From", "TraceSN", "Content", "Type", "ExpireInterval", "MsgID",
		"CreationTime", "ExpireTime", "Box"}
	request := &saver.RetrieveMessagesRequest{Appid: 2080, Owner: receiver, ChatChannels: chnnels,
		TraceSN: fmt.Sprint(time.Now().Unix())}
	fmt.Println("Retrieve message request:", request)
	if resp, err := saver.RetrieveChatMessages(request); err != nil {
		fmt.Println("testRetrieve failed, error: ", err)
	} else {
		//fmt.Println("Retrieve message response:", resp)

		fmt.Println("")
		fmt.Println("Response from saver:")

		for channel, _ := range chnnels {
			fmt.Println("Channel", channel, "latest ID: ", resp.LatestID[channel])
			if lastR, ok := resp.LastReadID[channel]; ok {
				fmt.Println("Channel", channel, "last read ID: ", lastR)
			}
			fmt.Println("")
			if messages, ok := resp.Inbox[channel]; ok && len(messages) > 0 {
				fmt.Println("Channel", channel, "inbox messages:")
				for i, message := range messages {
					fmt.Println("index:", i, ",", message.ToString(fields...))
				}
			} else {
				fmt.Println("Channel", channel, "doesn't have inbox messages in response")
			}
			fmt.Println("")
			if messages, ok := resp.Outbox[channel]; ok && len(messages) > 0 {
				fmt.Println("Channel", channel, "outbox messages:")
				for i, message := range messages {
					fmt.Println("index:", i, ",", message.ToString(fields...))
				}
				fmt.Println("")
			}
			fmt.Println("")
		}

	}
}

func testStore() {
	fmt.Println("starting storeing message test")
	logic.DynamicConf().LocalSaverRpc = saveraddr
	traceId := time.Now().UnixNano()

	// set default channel as im channel
	if chatChannel == DefaultChannel {
		chatChannel = "im"
	}

	fields := []string{"To", "From", "TraceSN", "Content", "Type", "ExpireInterval", "MsgID",
		"CreationTime", "ExpireTime", "Box"}

	msgFmt := "{\"userid\":%v,\"type\":1,\"text\":\"%v \u5173\u6ce8\u4e86\u4f60\"," +
		"\"time\":%v,\"expire\":86400,\"extends\":{\"uid\":\"%v\",\"userid\":\"%v\",\"nickname\":\"%v\"," +
		"\"avatar\":\"http:\\/\\/image.huajiao.com\\/d2bd9628df60a6893539a75b6d0db0a9-100_100.jpg\"," +
		"\"exp\":450077296,\"level\":84,\"verified\":false," +
		"\"verifiedinfo\":{\"credentials\":\"\",\"type\":0,\"realname\":\"%v\",\"status\":0," +
		"\"error\":\"\",\"official\":false},\"creatime\":\"%v\"},\"traceid\":\"%v\"}"

	request := &saver.StoreMessagesRequest{Appid: logic.APPID_HUAJIAO,
		Messages:    make(map[string]*saver.ChatMessage, 1),
		TraceSN:     fmt.Sprint(traceId),
		ChatChannel: chatChannel}
	request.Messages[receiver] = &saver.ChatMessage{
		Content: fmt.Sprintf(msgFmt, receiver, sender,
			time.Now().Unix(), sender, sender, sender, sender,
			time.Now().Format("2006-01-02 15:04:05"), time.Now().UnixNano()),
		Type: 1, To: receiver,
		From: sender, TraceSN: traceId, ExpireInterval: expireInterval, StoreOutbox: uint8(outbox)}
	if resp, err := saver.StoreChatMessages(request); err != nil {
		fmt.Println("Test store", chatChannel, "message failed, error:", err.Error())
	} else {
		for rec, message := range resp.Inbox {
			fmt.Println("inbox message, receiver:", rec, ", ", message.ToString(fields...))
		}
		fmt.Println("")
		for rec, message := range resp.Outbox {
			fmt.Println("outbox message, receiver:", rec, ", ", message.ToString(fields...))
		}
	}

	if count > 1 {
		count--
		traceId = time.Now().UnixNano()
		request := &saver.StoreMessagesRequest{Appid: logic.APPID_HUAJIAO,
			Messages:    make(map[string]*saver.ChatMessage, 1),
			TraceSN:     fmt.Sprint(traceId),
			ChatChannel: chatChannel}

		failed := 0
		for i := 0; i < count; i++ {
			request.Messages[receiver] = &saver.ChatMessage{
				Content: fmt.Sprintf(msgFmt, receiver, sender,
					time.Now().Unix(), sender, sender, sender, sender,
					time.Now().Format("2006-01-02 15:04:05"), time.Now().UnixNano()),
				Type: 1, To: receiver, From: sender, TraceSN: traceId, ExpireInterval: expireInterval,
				StoreOutbox: uint8(outbox)}
			if _, err := saver.StoreChatMessages(request); err != nil {
				fmt.Println("Test store", chatChannel, "message failed, error:", err.Error())
				failed++
			}

			if (i+1)%10000 == 0 {
				fmt.Println("has stored", i+1, chatChannel, "messages...")
			}
		}

		fmt.Println("Test store ", chatChannel, " message finished. successful:", count-failed, "failed", failed)
	}
}

func testRecall() {
	fmt.Println("starting recall message test")
	logic.DynamicConf().LocalSaverRpc = saveraddr

	request := &saver.RecallMessagesRequest{Appid: logic.APPID_HUAJIAO, ChatChannel: chatChannel, Sender: sender,
		Receiver: receiver, InboxId: uint64(msgid), TraceSN: fmt.Sprint(time.Now().UnixNano())}

	fmt.Println("Recall message request:", request)

	if resp, err := saver.RecallChatMessages(request); err != nil {
		fmt.Println("testRecall failed, error: ", err)
	} else {
		fmt.Println("Recall message response:", resp)
	}

}

func testGetQps() {
	if stat, err := saver.GetSaverQps(saveraddr); err != nil {
		fmt.Println("testGetQps failed, error: ", err)
	} else {
		fmt.Println(saveraddr, "QPS:", stat.QpsString())
	}
}

func testGetTotalOps() {
	if stat, err := saver.GetSaverTotalOps(saveraddr); err != nil {
		fmt.Println("testGetTotalOps failed, error: ", err)
	} else {
		fmt.Println(saveraddr, "all operation information:", stat.String())
	}
}

func testGetSaverAddrs() {
}

func testCreateGroup() {
	users := strings.Split(members, ",")
	info, err := saver.CreateGroup(group, 2080, owner, users)
	if err != nil {
		fmt.Print("create group", err.Error())
	} else {
		fmt.Println("group created", *info)
	}
}

func testJoinGroup() {
	users := strings.Split(members, ",")
	msgid, err := saver.JoinGroup(group, 2080, users)
	if err != nil {
		fmt.Println("join group", err.Error())
	} else {
		fmt.Println("group joined", msgid)
	}
}

func testQuitGroup() {
	users := strings.Split(members, ",")
	msgid, err := saver.QuitGroup(group, 2080, users)
	if err != nil {
		fmt.Println("quit group", err.Error())
	} else {
		fmt.Println("group quited", msgid)
	}
}

func testDismissGroup() {
	msgid, err := saver.DismissGroup(group, 2080)
	if err != nil {
		fmt.Println("dismiss group", err.Error())
	} else {
		fmt.Println("group dismissed", msgid)
	}
}

// IntSlice attaches the methods of Interface to []int, sorting in increasing order.
type Uint64Slice []uint64

func (p Uint64Slice) Len() int           { return len(p) }
func (p Uint64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func orderFlushMap(w io.Writer, inboxes map[uint64]string, outboxes map[uint64]string) {
	var keys []uint64
	dataMap := make(map[uint64]string, len(inboxes)+len(outboxes))
	for k, v := range inboxes {
		dataMap[k] = v
	}
	for k, v := range outboxes {
		dataMap[k] = v
	}

	for k := range dataMap {
		keys = append(keys, k)
	}
	sort.Sort(Uint64Slice(keys))
	for _, k := range keys {
		fmt.Fprintln(w, dataMap[k])
	}
}

func testDump() {
	fmt.Println("starting dump message...")
	logic.DynamicConf().LocalSaverRpc = saveraddr

	// default to retrieve all channels' messages
	var channels map[string]*saver.RetrieveChannel

	if msgid <= 0 {
		msgid = 1
	}

	needMore := false
	if count <= 0 {
		count = 1000
		needMore = true
	}

	if chatChannel == DefaultChannel {
		chatChannel = saver.ChatChannelIM
	}
	users := strings.Split(members, ",")
	dumpFile := fmt.Sprintf("./%s_%s_.dat", strings.Join(users, "_"), chatChannel)

	f, err := os.Create(dumpFile)

	if err != nil {
		fmt.Println("failed to open file", dumpFile, "with error:", err)
		return
	}

	channels = map[string]*saver.RetrieveChannel{
		chatChannel: &saver.RetrieveChannel{Channel: chatChannel, StartMsgId: msgid, MaxCount: count}}

	saveBar := 8 * count
	size := 10 * count
	inboxes := make(map[uint64]string, size)
	outboxes := make(map[uint64]string, size)
	for _, user := range users {
		pull := true

		fmt.Println("start to dump the ", chatChannel, "message of", user, "...")
		fmt.Fprintln(f, "user:", user, chatChannel, "messages:")
		for pull {
			request := &saver.RetrieveMessagesRequest{Appid: logic.APPID_HUAJIAO, Owner: user, ChatChannels: channels,
				TraceSN: fmt.Sprint(time.Now().Unix())}
			fmt.Println("Retrieve message request:", request)
			if resp, err := saver.RetrieveChatMessages(request); err != nil {
				fmt.Println("Retrieve message failed, error: ", err)
				pull = false
			} else {
				latest, _ := resp.LatestID[chatChannel]
				lastRead, _ := resp.LastReadID[chatChannel]

				maxReadID := uint64(0)
				// Inbox message
				if messages, ok := resp.Inbox[chatChannel]; ok && len(messages) > 0 {
					for _, message := range messages {
						if message.MsgId > maxReadID {
							maxReadID = message.MsgId
						}
						inboxes[message.MsgId] = fmt.Sprintf(
							"{\"msgid\":\"%v\", \"box\":\"outbox\", \"sender\":\"%v\", \"to\":\"%v\", \"content\":%v}",
							message.MsgId, message.From, message.To, message.Content)
					}
				}
				// Outbox message
				if messages, ok := resp.Outbox[chatChannel]; ok && len(messages) > 0 {
					for _, message := range messages {
						if message.MsgId > maxReadID {
							maxReadID = message.MsgId
						}
						outboxes[message.MsgId] = fmt.Sprintf(
							"{\"msgid\":\"%v\", \"box\":\"outbox\", \"sender\":\"%v\", \"to\":\"%v\", \"content\":%v}",
							message.MsgId, message.From, message.To, message.Content)
					}
				}

				// 写缓冲数据
				if len(inboxes) > saveBar || len(outboxes) > saveBar {
					orderFlushMap(f, inboxes, outboxes)

					if len(inboxes) > 0 {
						inboxes = make(map[uint64]string, size)
					}
					if len(outboxes) > 0 {
						outboxes = make(map[uint64]string, size)
					}
				}

				// 检查是不是还要进行下一次读取
				if needMore && maxReadID > 0 && maxReadID < latest {
					msgid = int64(maxReadID + 1)
					if msgid > int64(lastRead) {
						msgid = int64(lastRead + 1)
						count = int(latest - lastRead + 100)
					}
					channels[chatChannel].MaxCount = count
					channels[chatChannel].StartMsgId = msgid
				} else {
					pull = false
				}
			}
		}

		if len(inboxes) > 0 || len(outboxes) > 0 {
			orderFlushMap(f, inboxes, outboxes)

			if len(inboxes) > 0 {
				inboxes = make(map[uint64]string, size)
			}
			if len(outboxes) > 0 {
				outboxes = make(map[uint64]string, size)
			}
		}

		fmt.Println("dump the ", chatChannel, "message of", user, "done.")
	}
}

func main() {
	switch testcase {
	case "store":
		testStore()
	case "retrieve":
		testRetrieve()
	case "unread":
		testUnread()
	case "recall":
		testRecall()
	case "qps":
		testGetQps()
	case "allop":
		testGetTotalOps()
	case "creategroup":
		testCreateGroup()
	case "joingroup":
		testJoinGroup()
	case "quitgroup":
		testQuitGroup()
	case "dismissgroup":
		testDismissGroup()
	case "dump":
		testDump()
	default:
		fmt.Println("not support test case", testcase)
	}

}
