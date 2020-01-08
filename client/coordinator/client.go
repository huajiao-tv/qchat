package coordinator

import (
	"github.com/huajiao-tv/qchat/logic"
	"github.com/johntech-o/gorpc"
)

var GorpcClient *gorpc.Client

func init() {
	netOptions := gorpc.NewNetOptions(gorpc.DefaultConnectTimeout, gorpc.DefaultReadTimeout, gorpc.DefaultWriteTimeout)
	GorpcClient = gorpc.NewClient(netOptions)

	//	statNetOption := gorpc.NewNetOptions(1*time.Second, 1*time.Second, 1*time.Second)
}

// 发送消息到coordinator
// 由于可能需要在业务层指定发送往的coordinator,所以加了一个addr参数
// msgid是为了以后做多个中心时使用
func ChatRoomMsg(appid uint16, roomid, sender, content string, msgtype, priority int, msgid uint, traceid string) error {
	req := &logic.ChatRoomMsgRaw{
		appid,
		roomid,
		sender,
		content,
		msgtype,
		priority,
		msgid,
		traceid,
		msgid,
	}
	var resp int
	addrs := logic.GetStatedCoordinatorGorpcs(roomid)
	for _, addr := range addrs {
		// 是否改成异步？ @todo
		if err := GorpcClient.CallWithAddress(addr, "GorpcService", "ChatRoomMsg", req, &resp); err != nil {
			return err
		}
	}
	return nil
}

func GetAdapterStat(addr string) (map[string]map[string]*AdapterStat, error) {
	resp := map[string]map[string]*AdapterStat{}
	if err := GorpcClient.CallWithAddress(addr, "GorpcService", "GetAdapterStat", 1, &resp); err != nil {
		return nil, err
	}
	return resp, nil

}

func DegradeChatRoom(appid uint16, roomid string, degrade bool) error {
	req := &DegradeRequest{
		AppId:   appid,
		RoomId:  roomid,
		Degrade: degrade,
	}
	var resp int
	addrs := logic.GetStatedCoordinatorGorpcs(roomid)
	for _, addr := range addrs {
		if err := GorpcClient.CallWithAddress(addr, "GorpcService", "DegradeChatRoom", req, &resp); err != nil {
			return err
		}
	}
	return nil
}

func GetDegradedChatRoomList(appid uint16) ([]string, error) {
	resp := []string{}
	addrs := logic.NetGlobalConf().CoordinatorRpcs
	for _, addr := range addrs {
		rooms := []string{}
		err := GorpcClient.CallWithAddress(addr, "GorpcService", "GetDegradedChatRoomList", appid, &rooms)
		if err != nil {
			return resp, err
		}
		resp = append(resp, rooms...)
	}
	return resp, nil
}

func LiveNotify(appid uint16, roomid string, start bool) error {
	req := &LiveNotifyRequest{
		AppId:  appid,
		RoomId: roomid,
		Start:  start,
	}
	var resp int
	addrs := logic.GetStatedCoordinatorGorpcs(roomid)
	for _, addr := range addrs {
		if err := GorpcClient.CallWithAddress(addr, "GorpcService", "LiveNotify", req, &resp); err != nil {
			return err
		}
	}
	return nil
}
