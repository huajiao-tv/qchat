package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/cryption"
	"github.com/yvasiyarov/php_session_decoder/php_serialize"
)

var (
	pool *ChatRoomMessagePool

	room     string
	gateways string
	count    int
)

var (
	material = []string{
		"红酥手",
		"黄縢酒",
		"满城春色宫墙柳",
		"东风恶",
		"欢情薄",
		"一怀愁绪",
		"几年离索",
		"错",
		"错",
		"错",
		"春如旧",
		"人空瘦",
		"泪痕红浥鲛绡透",
		"桃花落",
		"闲池阁",
		"山盟虽在",
		"锦书难托",
		"莫",
		"莫",
		"莫",
	}
)

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

func init() {
	flag.IntVar(&count, "c", 10, "message count")
	flag.StringVar(&room, "rid", "", "room id")
	flag.StringVar(&gateways, "gwys", "127.0.0.1:6220", "gateway address")
	flag.Parse()
	pool = NewChatRoomMessagePool(room, 2080, 1000, 10)
}

func main() {
	for i := 0; i < count; i++ {
		msg := &logic.ChatRoomMessage{
			RoomID:      room,
			Sender:      "20777228",
			Appid:       2080,
			MsgType:     9,
			MsgContent:  []byte(huajiaoChatRoomMessage(room, material[i%len(material)])),
			RegMemCount: 0,
			MemCount:    0,
			MsgID:       0,
			MaxID:       0,
			TimeStamp:   time.Now().UnixNano() / 1e6,
			Priority:    false,
		}
		pool.Add(&logic.ChatRoomMessageNotify{ChatRoomMessage: msg})
	}

	select {}
}

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
			"userid":   "20777228",
			"nickname": "nickname",
			"avatar":   avatar,
			"verified": false,
			"verifiedinfo": php_serialize.PhpArray{
				"credentials": "",
				"type":        0,
				"realname":    "realname",
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
			Sender:   "20777228",
			NickName: "nickname",
			Avatar:   avatar,
			Verified: false,
			Verifiedinfo: verifiedinfo{
				Credentials: "",
				Type:        0,
				RealName:    "realname",
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
