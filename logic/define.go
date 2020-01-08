package logic

import (
	"crypto/md5"
	"crypto/rc4"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/huajiao-tv/qchat/utility/idgen"
)

const (
	VERF_CODE_SALT = "360tantan@1408$"

	DEFAULT_APPID     = 2080
	APPID_HUAJIAO     = 2080
	APPID_OVERSEA     = 2081
	APPID_HUAJIAO_STR = "2080"
)

var (
	_SVN_         string
	_VERSION_     string
	_AUTHOR_      string
	_COMPILETIME_ string

	ComponentTags map[string]string = map[string]string{
		"svn":          _SVN_,
		"version":      _VERSION_,
		"author":       _AUTHOR_,
		"compile_time": _COMPILETIME_,
	}
)

type ConnectionId uint64

func (self ConnectionId) String() string {
	return strconv.FormatUint(uint64(self), 10)
}

type ConnectionIdGenerator struct {
	nextId ConnectionId
	mutex  sync.Mutex
}

func (this *ConnectionIdGenerator) NextConnectionId() ConnectionId {
	this.mutex.Lock()
	this.nextId++
	id := this.nextId
	this.mutex.Unlock()
	return id
}

func NewConnectionIdGenerator() *ConnectionIdGenerator {
	return &ConnectionIdGenerator{}
}

func Sum(key string) int {
	var hash uint32 = 0
	for i := 0; i < len(key); i++ {
		hash += uint32(key[i])
		hash += (hash << 10)
		hash ^= (hash >> 6)
	}
	hash += (hash << 3)
	hash ^= (hash >> 11)
	hash += (hash << 15)

	return int(hash)
}

func StringToUint16(s string) uint16 {
	if i, err := strconv.ParseUint(s, 10, 16); err != nil {
		return 0
	} else {
		return uint16(i)
	}
}
func GetSn() uint64 {
	return uint64(rand.Int63())
}

func GetRpcIndex(rpc string, rpcs []string) (index int) {
	index = -1
	rpcIP := strings.Split(rpc, ":")[0]
	for i, node := range rpcs {
		ip := strings.Split(node, ":")[0]
		if ip == rpcIP {
			index = i
			break
		}
	}
	return
}

// 可以优化，目前只返回是数字的随机串，可以加上字符
func RandString(length int) string {
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		b[i] = byte(48 + rand.Intn(10))
	}
	return string(b)
}

type MessageNotify struct {
	InfoType          string
	InfoContent       []byte
	InfoId            int64
	QueryAfterSeconds uint32
	ExpireTime        uint64
	Sender            string
}

func NewMessageNotify(infoType string, infoContent []byte, queryAfterSeconds uint32, expireTime uint64, sender string) *MessageNotify {
	return &MessageNotify{
		InfoType:          infoType,
		InfoContent:       infoContent,
		InfoId:            idgen.GenIdInt64(),
		QueryAfterSeconds: queryAfterSeconds,
		ExpireTime:        expireTime,
		Sender:            sender,
	}
}

func GenerateChatRoomBroadcastTag(appid interface{}) string {
	if str, ok := appid.(string); ok {
		return "chatroom:" + str + ":"
	} else {
		return fmt.Sprintf("chatroom:%d:", appid)
	}
}

func GenerateChatRoomTag(appid interface{}, room string) string {
	if str, ok := appid.(string); ok {
		return fmt.Sprintf("chatroom:%s:%s", str, room)
	} else {
		return fmt.Sprintf("chatroom:%d:%s", appid, room)
	}
}

func GenerateWebChatRoomTag(appid interface{}, room string) string {
	if str, ok := appid.(string); ok {
		return fmt.Sprintf("chatroom:web:%s:%s", str, room)
	} else {
		return fmt.Sprintf("chatroom:web:%d:%s", appid, room)
	}
}

// 聊天室消息的原始请求
type ChatRoomMsgRaw struct {
	Appid      uint16
	RoomID     string
	Sender     string
	MsgContent string
	MsgType    int
	Priority   int
	MsgId      uint
	TraceId    string
	MaxId      uint
}

type ChatRoomMessage struct {
	RoomID      string
	Sender      string
	Appid       uint16
	MsgType     int
	MsgContent  []byte
	RegMemCount int
	MemCount    int
	MsgID       uint
	MaxID       uint
	TimeStamp   int64
	Priority    bool
}

type ChatRoomMessageNotify struct {
	*ChatRoomMessage
	GatewayAddrs map[string]int
	TraceId      string
	Delay        time.Duration // 消息延迟多久后发送，由调用方传入
	Interval     time.Duration // 每个 gateway 发送消息间隔
}

/*
 * describes the gateway information which user logged on
 *
 * GatewayAddr: is gateway address
 * ConnId: is connection id on gateway
 */
type UserGateway struct {
	GatewayAddr string
	ConnId      ConnectionId
}

type PrivateChatRoomMessageNotify struct {
	*ChatRoomMessage
	UserGateways []*UserGateway
	TraceId      string
}

func FilterGatewayAddrs(config map[int]int, members int, addrs map[string]int, priority bool) map[string]int {
	gws := make([]string, 0, len(addrs))
	for gw, mem := range addrs {
		if mem > 0 {
			gws = append(gws, gw)
		} else {
			delete(addrs, gw)
		}
	}
	if priority {
		return addrs
	}
	percent := 0
	for ts, p := range config {
		if members > ts && p > percent {
			percent = p
		}
	}
	if percent <= 0 {
		return addrs
	} else if percent >= 100 {
		return map[string]int{}
	}
	start := rand.Intn(len(gws))
	count := len(gws) * percent / 100
	for i := 0; i < count; i++ {
		delete(addrs, gws[(start+i)%len(gws)])
	}
	return addrs
}

func MakeVerfCode(user string) string {
	sum := md5.Sum([]byte(user + VERF_CODE_SALT))
	str := fmt.Sprintf("%x", sum)
	return str[24:32]
}

func MakeSecretRan(pwd, serverRan string) []byte {
	rc, _ := rc4.NewCipher([]byte(pwd))
	var str string = serverRan + RandString(8)
	b := []byte(str)
	rc.XORKeyStream(b, b)
	return b
}

func ShowJson(j interface{}) {
	ds, err := json.Marshal(j)
	if err != nil {
		fmt.Println("json encode error:", err)
	}
	fmt.Println(string(ds))
}

func GetTraceId(gwaddr string, connectionId ConnectionId) string {
	return fmt.Sprintf("%s-%d", gwaddr, connectionId)
}

func HumSize(bit uint64) string {
	if bit > 1000000000 {
		return fmt.Sprintf("%.2fGbit", float64(bit)/1000000000)
	} else if bit > 1000000 {
		return fmt.Sprintf("%.2fMbit", float64(bit)/1000000)
	} else if bit > 1000 {
		return fmt.Sprintf("%.2fKbit", float64(bit)/1000)
	}
	return fmt.Sprintf("%dbit", bit)
}
