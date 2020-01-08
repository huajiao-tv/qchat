package apns

import (
	"github.com/huajiao-tv/qchat/logic"
	"github.com/johntech-o/gorpc"
)

var GorpcClient *gorpc.Client

func init() {
	netOptions := gorpc.NewNetOptions(gorpc.DefaultConnectTimeout, gorpc.DefaultReadTimeout, gorpc.DefaultWriteTimeout)
	GorpcClient = gorpc.NewClient(netOptions)

}

type GroupChatRpcMsg struct {
	GroupId string
	Uids    []string
	Summary string
}

// 按groupid hash到一台center上，做消息合并
func PushMsg(msg []*GroupChatRpcMsg) error {
	if err := GorpcClient.CallWithAddress(logic.GetRandApnsGorpc(), "GroupChatRpc", "PushMsg", msg, nil); err != nil {
		return err
	}
	return nil
}
