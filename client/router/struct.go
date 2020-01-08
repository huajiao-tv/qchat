package router

import (
	"fmt"
	"time"

	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/network"
)

const (
	DisconnectAction = iota
)

// gateway往后路由需要发送的数据结构
type GwPackage struct {
	XimpBuff     *network.XimpBuffer // 前端传过来的包
	Property     map[string]string   // 前端保存的一些连接信息
	GatewayAddr  string              // gateway的gorpc ip和端口
	ConnectionId logic.ConnectionId  // 连接id
}

// 发往gateway的数据结构
type GwResp struct {
	XimpBuff         *network.XimpBuffer // 需要发往客户端的包
	Property         map[string]string   // 需要前端保存的一些连接信息
	Priority         bool                // 是否是优先下发内容
	Tags             map[string]bool     // 需要前端保存这个连接的tag信息，如果false表示删除
	Actions          []int               // 需要执行的动作
	Rkey             []byte              // 设置加密的key值
	HeartBeatTimeout time.Duration
	Err              string // 由于错误不能在gorpc里返回，所以只能在这里返回
}

func (gwr *GwResp) String() string {
	result := ""
	if gwr.XimpBuff != nil {
		result += "Ximp:" + gwr.XimpBuff.String()
	}
	if gwr.Priority {
		result += " Prio:1"
	}
	if len(gwr.Property) > 0 {
		result += " Prop:" + fmt.Sprintf("%v", gwr.Property)
	}
	if len(gwr.Tags) > 0 {
		result += " Tags:" + fmt.Sprintf("%v", gwr.Tags)
	}
	if len(gwr.Rkey) > 0 {
		result += " Rkey:" + string(gwr.Rkey)
	}
	if gwr.HeartBeatTimeout != 0 {
		result += " HBT:" + gwr.HeartBeatTimeout.String()
	}
	if gwr.Err != "" {
		result += " Err:" + gwr.Err
	}
	return result
}

// 发送消息到router 需要带的消息
type MsgNotify struct {
	GatewayAddr   string
	Tags          []string
	ConnIds       []logic.ConnectionId
	MessageNotify *logic.MessageNotify
}

/*
 * Center uses this to send push request to router
 *
 * Appid: application, Huajiao is 2080
 * Receiver: is push receiver
 * GateWays: is gateway information which receiver logged on, there might be more than one gateway
 * Sender: is who sends new message
 * Channel: indicates channel which new message is stored, so far only 'notify'
 * MsgId: new message id
 * TraceSN: trace sn
 */
type ChatPushNotify struct {
	Appid       int
	Receiver    string
	GatewayAddr string
	ConnId      logic.ConnectionId
	Sender      string
	Channel     string
	MsgId       uint64
	TraceSN     string
}

type ChatPushTags struct {
	Tags         []string
	GatewayAddrs []string
	Sender       string
	Channel      string
	MsgId        uint64
	TraceSN      string
}

type ReConnectNotify struct {
	Ip           string
	Port         uint32
	MoreIps      []string
	Gateways     []string
	Tags         []string
	UserGateways []*logic.UserGateway
}
type ChatRoomNotifyBatchReq struct {
	Msgs  []*logic.ChatRoomMessageNotify
	Gwys  map[string]int
	IsWeb bool
}
