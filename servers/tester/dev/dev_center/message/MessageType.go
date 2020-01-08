package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/huajiao-tv/qchat/utility/cryption"
	"github.com/yvasiyarov/php_session_decoder/php_serialize"
)

var (
	help       string
	server     string
	userid     string
	msgtype    string
	content    string
	httpClient *http.Client
	centerPort string
	appid      string
	room       string
	priority   string
)

type message struct {
	Roomid    string `json:"roomid"`
	MsgType   int    `json:"type"`
	Content   string `json:"text"`
	TimeStamp int64  `json:"time"`
	Expire    int    `json:"expire"`
	Extends   extend `json:"extends"`
	Md5       string `json:"traceid"`
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
}

type giftMessage struct {
	Roomid    string  `json:"roomid"`
	MsgType   int     `json:"type"`
	Content   string  `json:"text"`
	TimeStamp int64   `json:"time"`
	Expire    int     `json:"expire"`
	Extends   extends `json:"extends"`
	Md5       string  `json:"traceid"`
	Traceid   string  `json:"traceid"`
}

type extends struct {
	Content         string   `json:"content"`
	Creatime        string   `json:"creatime"`
	ReceiverBalance int      `json:"receiver_balance"`
	ReceiverIncome  string   `json:"receiver_income"`
	ReceiverIncomeB string   `json:"receiver_income_b"`
	ReceiverIncomeP string   `json:"receiver_income_p"`
	SenderBalance   int      `json:"sender_balance"`
	Receiver        receiver `json:"receiver"`
	Sender          sender   `json:"sender"`
	GiftInfo        giftInfo `json:"giftinfo"`
	Title           string   `json:"title"`
	Scheme          string   `json:"scheme"`
}

type receiver struct {
	Avatar       string       `json:"avatar"`
	NickName     string       `json:"nickname"`
	Uid          int          `json:"uid"`
	Verified     bool         `json:"verified"`
	Verifiedinfo verifiedinfo `json:"verifiedinfo"`
	Exp          int          `json:"exp"`
	Level        int          `json:level`
	Medal        []medal      `json:"medal"`
}

type sender struct {
	Avatar       string       `json:"avatar"`
	Uid          int          `json:"uid"`
	NickName     string       `json:"nickname"`
	Verified     bool         `json:"verified"`
	Verifiedinfo verifiedinfo `json:"verifiedinfo"`
	Exp          int          `json:"exp"`
	Level        int          `json:level`
	Medal        []medal      `json:"medal"`
}

type medal struct {
	Kind  string `json:"kind"`
	Medal string `json:"medal"`
}

type giftInfo struct {
	GiftId       string       `json:"giftid"`
	GiftName     string       `json:"giftname"`
	Amount       string       `json:"amount"`
	Icon         string       `json:"icon"`
	Pic          string       `json:"pic"`
	Content      string       `json:"content"`
	RelativeInfo relativeInfo `json:"relativeInfo"`
}

type relativeInfo struct {
	RepeatId  string   `json:"repeatId"`
	RepeatNum int      `json:"repeatNum"`
	Property  property `json:"property"`
}

type property struct {
	RepeatGift      int             `json:"repeatGift"`
	PropertyAndroid propertyAndroid `json:"property_android"`
	PropertyIos     propertyIos     `json:"property_ios"`
}

type propertyAndroid struct {
	RepeatGift string `json:"repeatGift"`
	Desctop    string `json:"desctop"`
	Desc       string `json:"desc"`
	Pic        string `json:"pic"`
	Points     int    `json:"points"`
	Gif        string `string:"gif"`
}

type propertyIos struct {
	RepeatGift string `json:"repeatGift"`
	Desctop    string `json:"desctop"`
	Desc       string `json:"desc"`
	Pic        string `json:"pic"`
	Points     int    `json:"points"`
	Gif        string `string:"gif"`
}

func init() {
	flag.StringVar(&help, "help", "", "help command")
	flag.StringVar(&server, "server", "http://qchatcenter.huajiao.com:6677", "")
	flag.StringVar(&userid, "userid", "19895503", "userid")
	flag.StringVar(&msgtype, "msg type", "9", "msg type")
	flag.StringVar(&content, "content", "我是花椒君", "msg info")
	flag.StringVar(&room, "room", "123", "roomid")

	flag.Parse()
	appid = "2080"
	priority = "1"
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
	if help != "" {
		fmt.Println("no help ^ ^")
	}
	fmt.Println(chatroom())
}

func chatroom() string {

	var message string
	uri := fmt.Sprintf("%s/send", server)
	//message = huajiaoChatRoomMessage(room, content)
	message = huajiaoChatRoomGift(room, content)
	values := url.Values{
		"roomid":   []string{room},
		"sender":   []string{userid},
		"traceid":  []string{fmt.Sprint(time.Now().UnixNano())},
		"content":  []string{message},
		"appid":    []string{appid},
		"priority": []string{priority},
	}

	fmt.Println(values)
	if resp, err := httpClient.PostForm(uri, values); err != nil {
		return fmt.Sprintf("send chatroom message err:%s", err.Error())
	} else {
		return fmt.Sprintf("success\n%v", resp)
	}
}

func huajiaoChatRoomGift(room, content string) string {
	avatar_rec := "http://image.huajiao.com/002cf744a2716b961eecac5a6da4c506-100_100.jpg"
	avatar_send := "http//image.huajiao.com/e5c3cfb49137dc38c4ba3c168bd13641-100_100.jpg"
	timestamp := time.Now().UnixNano()
	data := php_serialize.PhpArray{
		"roomid": room,
		"type":   30,
		"text":   "",
		"time":   timestamp,
		"expire": 86400,
		"extends": php_serialize.PhpArray{
			"contents":          "",
			"creatime":          time.Now().Format("2006-01-02 15:04:05"),
			"receiver_balance":  764,
			"receiver_income":   "4546",
			"receiver_income_b": "136162",
			"receiver_income_p": "9639429",
			"sender_balance":    33,
			"receiver": php_serialize.PhpArray{
				"avatar":   avatar_rec,
				"nickname": "隐藏的花儿",
				"uid":      56679008,
				"verified": true,
				"verifiedinfo": php_serialize.PhpArray{
					"credentials": "情不知所起 一往而深",
					"realname":    "隐藏的花儿",
					"type":        1,
				},
				"exp":   1008592,
				"level": 24,
				"medal": php_serialize.PhpSlice{
					php_serialize.PhpArray{
						"kind":  "tuhao",
						"medal": "2",
					},
				},
			},
			"sender": php_serialize.PhpArray{
				"avatar":   avatar_send,
				"nickname": "回忆",
				"uid":      19895503,
				"verified": false,
				"verifiedinfo": php_serialize.PhpArray{
					"credentials": "",
					"realname":    "回忆",
					"type":        1,
				},
				"exp":   2,
				"level": 1,
				"medal": php_serialize.PhpSlice{
					php_serialize.PhpArray{
						"kind":  "tuhao",
						"medal": "2",
					},
				},
			},
			"giftinfo": php_serialize.PhpArray{
				"giftid":   "1087",
				"giftname": "金话筒",
				"amount":   "1",
				"icon":     "http://static.huajiao.com/huajiao/gift/jinhuatong222.png",
				"pic":      "http://static.huajiao.com/huajiao/gift/jinhuatong222.png",
				"content":  "",
				"relativeInfo": php_serialize.PhpArray{
					"repeatId":  "@19895503240037962864318614720543999921679",
					"repeatNum": 5,
					"property": php_serialize.PhpArray{
						"repeatGift": 1,
						"property_android": php_serialize.PhpArray{
							"repeatGift": "1",
							"desctop":    "金话筒",
							"desc":       "1豆",
							"pic":        "http://static.huajiao.com/huajiao/gift/jinhuatong222.png",
							"points":     1,
							"gif":        "http://static.huajiao.com/huajiao/gifteffect/20004_31.zip",
						},
						"property_ios": php_serialize.PhpArray{
							"repeatGift": "1",
							"desctop":    "金话筒",
							"desc":       "1豆",
							"pic":        "http://static.huajiao.com/huajiao/gift/jinhuatong222.png",
							"points":     1,
							"gif":        "http://static.huajiao.com/huajiao/gifteffect/20004_31.zip",
						},
					},
				},
			},
			"title":  "",
			"scheme": "huajiao://huajiao.com/goto/wallet?userid=24003796",
		},
		//"traceid": timestamp,
	}
	value, err := php_serialize.NewSerializer().Encode(data)
	if err != nil {
		fmt.Println("php serialize error:", err)
		return ""
	}
	msg := &giftMessage{
		Roomid:    room,
		MsgType:   30,
		Content:   "",
		TimeStamp: timestamp,
		Expire:    86400,
		Extends: extends{
			//Content:         "",
			//Creatime:        time.Now().Format("2006-01-02 15:04:05"),

			/*这部分会更新主播的实时花椒币,必要时候也可以删
			ReceiverBalance: 764,
			ReceiverIncome:  "4546",
			ReceiverIncomeB: "136162",
			ReceiverIncomeP: "9639429",
			SenderBalance:   33,
			*/

			//这部分是主播信息，也可以删,用来区别主播和普通用户
			Receiver: receiver{
				//Avatar:   avatar_rec,
				//NickName: "隐藏的花",
				Uid: 56679008,
				//Verified: true,
				/*Verifiedinfo: verifiedinfo{
					Credentials: "情不知所起 一往而深",
					RealName:    "隐藏的花儿",
					Type:        1,
				},*/
				//Exp:   1008592,
				//Level: 24,
				/*Medal: []medal{
					{
						Kind:  "tuhao",
						Medal: "2",
					},
				},*/
			},

			//以下是送礼者信息
			Sender: sender{
				Avatar: avatar_send,
				//NickName: "回忆",
				Uid: 19895503,
				//Verified: false,
				Verifiedinfo: verifiedinfo{
					//Credentials: "",
					RealName: "回忆",
					//Type:        1,
				},
				Exp:   2,
				Level: 1,
				//Medal: []medal{
				//	{
				//			Kind:  "tuhao",
				//		Medal: "2",
				//	},
				//},
			},

			//以下是礼物信息
			GiftInfo: giftInfo{
				GiftId:   "1087",
				GiftName: "金话筒",
				Amount:   "1",
				Icon:     "http://static.huajiao.com/huajiao/gift/jinhuatong222.png",
				Pic:      "http://static.huajiao.com/huajiao/gift/jinhuatong222.png",
				Content:  "",
				RelativeInfo: relativeInfo{
					RepeatId:  "@19895503240037962864318614720543999921679",
					RepeatNum: 1,
					Property: property{
						RepeatGift: 1,
						PropertyAndroid: propertyAndroid{
							RepeatGift: "1",
							Desctop:    "金话筒",
							Desc:       "1豆",
							Pic:        "http://static.huajiao.com/huajiao/gift/jinhuatong222.png",
							Points:     1,
							Gif:        "http://static.huajiao.com/huajiao/gifteffect/20004_31.zip",
						},
						PropertyIos: propertyIos{
							RepeatGift: "1",
							Desctop:    "金话筒",
							Desc:       "1豆",
							Pic:        "http://static.huajiao.com/huajiao/gift/jinhuatong222.png",
							Points:     1,
							Gif:        "http://static.huajiao.com/huajiao/gifteffect/20004_31.zip",
						},
					},
				},
			},
			//Title: "",
			//Scheme: "huajiao://huajiao.com/goto/wallet?userid=24003796",
		},
		//Traceid: timestamp,
		Md5: cryption.Md5(value),
	}
	ret, err := json.Marshal(*msg)
	if err != nil {
		fmt.Println("json marshal error:", err)
		return ""
	}
	return string(ret)
}

func huajiaoChatRoomMessage(room, text string) string {
	avatar := "http://image.huajiao.com/5f2908e7b902f70b39cbbbbe37a4c344-100_100.jpg"
	timestamp := time.Now().UnixNano()
	data := php_serialize.PhpArray{
		"roomid": room,
		"type":   42,
		"text":   text,
		"time":   timestamp,
		//"expire": 86400,
		"extends": php_serialize.PhpArray{
			"liveid": room,
			"userid": "nihao",
			//"nickname": "王大头",
			"avatar":   avatar,
			"verified": true,
			"verifiedinfo": php_serialize.PhpArray{
				//"credentials": "你真的是个shab",
				//"type":     0,
				"realname": "niba",
				//"status":   0,
				//"error": "",
				//"official": false,
			},
			"gift": 1,
			//"exp":   1000,
			"level": 120,
		},
	}
	value, err := php_serialize.NewSerializer().Encode(data)
	if err != nil {
		fmt.Println("php serialize error:", err)
		return ""
	}
	msg := &message{
		Roomid:    room,
		MsgType:   42,
		Content:   text,
		TimeStamp: timestamp,
		//Expire:    86400,
		Extends: extend{
			LiveId: room,
			Sender: "nihao",
			//NickName: "王大头",
			Avatar:   avatar,
			Verified: true,
			Verifiedinfo: verifiedinfo{
				//Credentials: "你真的是个shab",
				//Type:     0,
				RealName: "niba",
				//Status:   0,
				//Error: "",
				//Official: false,
			},
			Gift: 1,
			//Exp:   1000,
			Level: 120,
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
