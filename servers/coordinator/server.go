package main

import (
	"fmt"

	"github.com/huajiao-tv/qchat/client/coordinator"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/johntech-o/gorpc"
)

type GorpcService struct{}

var rpcServer *gorpc.Server

func GorpcServer() {
	if netConf().GorpcListen == "" {
		panic("empty gorpc_listen")
	}
	Logger.Trace("gorpc listen", netConf().GorpcListen)
	rpcServer = gorpc.NewServer(netConf().GorpcListen)
	rpcServer.Register(new(GorpcService))
	rpcServer.Serve()
	panic("invalid gorpc listen" + netConf().GorpcListen)
}

func (this *GorpcService) Helloworld(foo string, resp *int) error {
	// @todo 需要处理退出动作
	fmt.Println("helloworld")
	return nil
}

func (this *GorpcService) ChatRoomMsg(req *logic.ChatRoomMsgRaw, resp *int) error {
	Logger.Trace(req.RoomID, req.Appid, req.TraceId, "ChatRoomMsg", req.Sender, req.MsgType, req.Priority)
	chatRoomAdapterPool.Get(req.Appid, req.RoomID).AddMsg(req)
	return nil
}

func (this *GorpcService) GetAdapterStat(req int, resp *map[string]map[string]*coordinator.AdapterStat) error {
	*resp = adapterStats.GetAll()
	return nil
}

func (this *GorpcService) GetDegradedChatRoomList(req uint16, resp *[]string) error {
	*resp = chatRoomAdapterPool.GetDegradedList(req)
	return nil
}

func (this *GorpcService) DegradeChatRoom(req *coordinator.DegradeRequest, resp *int) error {
	Logger.Trace(req.RoomId, req.AppId, "", "DegradeChatRoom", req.Degrade)
	chatRoomAdapterPool.Get(req.AppId, req.RoomId).ForceDegrade(req.Degrade)
	return nil
}

func (this *GorpcService) LiveNotify(req *coordinator.LiveNotifyRequest, resp *int) error {
	notify := "stop"
	if req.Start {
		notify = "start"
	}
	Logger.Trace(req.RoomId, req.AppId, "", "LiveNotify", notify)
	return nil
}
