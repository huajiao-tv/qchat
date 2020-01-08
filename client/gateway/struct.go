package gateway

import (
	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/logic"
)

type Operations struct {
	ConnectionIds []logic.ConnectionId // 要发送的连接id
	Gwr           *router.GwResp
	Tags          []string // 要发送的tag
	PrefixTag     string   // 发指定前缀的tag
}

type ConnTags struct {
	ConnId logic.ConnectionId
	Tags   []string
}
