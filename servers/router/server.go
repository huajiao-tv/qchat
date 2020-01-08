package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/huajiao-tv/qchat/client/gateway"
	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/pb"
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

// 处理从前端路由过来的包
func (this *GorpcService) RoutePackage(gwp *router.GwPackage, gwr *router.GwResp) error {
	tgwr, err := dealPackage(gwp)
	if err != nil {
		tgwr.Err = err.Error()
	}
	*gwr = *tgwr
	return nil
}

// 处理退出
func (this *GorpcService) Logout(gwp *router.GwPackage, gwr *router.GwResp) error {
	// 如果根本就没有登录成功，或者已经被logout处理的，就不再处理后续
	if open, ok := gwp.Property["Open"]; !ok || open == "0" {
		return nil
	}
	userId := gwp.Property["Sender"]
	appId := logic.StringToUint16(gwp.Property["Appid"])
	if resp, err := session.Close(userId, appId, gwp.GatewayAddr, gwp.ConnectionId, gwp.Property); err != nil {
		Logger.Error(userId, appId, logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "Logout", "CloseSession error", err)
		return err
	} else {
		if len(resp.Tags) != 0 {
			gwr.Tags = make(map[string]bool, len(resp.Tags))
			for _, v := range resp.Tags {
				gwr.Tags[v] = false
			}
		}
		Logger.Trace(userId, appId, logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "Logout", gwp.Property["ConnectionType"], resp.Tags)
	}
	return nil
}

func genInfoContent(msgid, pushType uint64) []byte {
	return []byte(`{"cmd":"push","data":{"msg_id":` + strconv.FormatUint(msgid, 10) + `,"pushtype":` + strconv.FormatUint(pushType, 10) + `}}`)
}

const (
	SendPublicRandMax = 10000
	SendPublicRandMin = 10
)

// 发送public消息接口
func (this *GorpcService) SendPushTags(tags *router.ChatPushTags, resp *int) error {
	if checkPushDegrade(tags.Channel, "", false) {
		Logger.Warn(tags.Tags, tags.MsgId, tags.TraceSN, "SendPushTags", "push disabled", tags.Channel, tags.GatewayAddrs)
		return nil
	}
	Logger.Trace(tags.Tags, tags.MsgId, tags.TraceSN, "SendPushTags", tags.Channel, tags.GatewayAddrs)
	ximp, err := pb.CreateMsgNotify(tags.Channel, nil, int64(tags.MsgId), tags.Sender, "", getQueryAfterSeconds())
	if err != nil {
		Logger.Error(tags.Tags, tags.MsgId, tags.TraceSN, "SendPushTags", "createMsgNotify error", err)
		return err
	}
	gwr := &router.GwResp{
		XimpBuff: ximp,
	}
	gateways := logic.NetGlobalConf().GatewayRpcs
	if len(tags.GatewayAddrs) != 0 {
		gateways = tags.GatewayAddrs
	}

	sendRandMax := netConf().SendPublicRandMax
	sendRandMin := netConf().SendPublicRandMin
	if sendRandMax > SendPublicRandMax {
		sendRandMax = SendPublicRandMax
	}
	if sendRandMin < SendPublicRandMin {
		sendRandMin = SendPublicRandMin
	}
	if sendRandMin >= sendRandMax {
		sendRandMax = sendRandMin + 1
	}

	for _, addr := range gateways {
		// because we have a lot of online users, in this way, we should send notification slowly
		time.Sleep(time.Millisecond * time.Duration(sendRandMin+rand.Intn(sendRandMax-sendRandMin)))
		go func(g string, o *router.GwResp) {
			if err := gateway.DoOperations(g, tags.Tags, "", nil, o); err != nil {
				Logger.Error(tags.Tags, tags.MsgId, tags.TraceSN, "SendPushTags:"+g, "DoOperationsBatch error", err)
			}
		}(addr, gwr)
	}
	return nil
}

// 发送peer消息接口
func (this *GorpcService) SendPushNotifyBatch(m []*router.ChatPushNotify, resp *int) error {
	reqs := map[string][]*gateway.Operations{}
	for _, n := range m {
		if checkPushDegrade(n.Channel, n.Receiver, false) {
			Logger.Warn(n.Receiver, n.Appid, n.TraceSN, "SendPushNotifyBatch", "push disabled", logic.GetTraceId(n.GatewayAddr, n.ConnId), fmt.Sprintf("%s-%d", n.Channel, n.MsgId))
			continue
		}
		Logger.Trace(n.Receiver, n.Appid, n.TraceSN, "SendPushNotifyBatch", logic.GetTraceId(n.GatewayAddr, n.ConnId), fmt.Sprintf("%s-%d", n.Channel, n.MsgId))
		ximp, err := pb.CreateMsgNotify(n.Channel, genInfoContent(n.MsgId, 100), int64(n.MsgId), n.Sender, n.Receiver, getQueryAfterSeconds())
		if err != nil {
			Logger.Error(n.Receiver, n.Appid, n.TraceSN, "SendPushNotifyBatch", "createMsgNotify error", err)
			continue
		}

		reqs[n.GatewayAddr] = append(reqs[n.GatewayAddr], &gateway.Operations{
			ConnectionIds: []logic.ConnectionId{n.ConnId},
			Gwr: &router.GwResp{
				XimpBuff: ximp,
			},
		})
	}
	for gatewayAddr, ops := range reqs {
		go func(g string, o []*gateway.Operations) {
			if err := gateway.DoOperationsBatch(g, o); err != nil {
				Logger.Error(g, len(o), "", "SendPushNotifyBatch", "DoOperationsBatch error", err)
			}
		}(gatewayAddr, ops)
	}
	return nil
}

// 发送重新登录消息
func (this *GorpcService) SendReConnectNotify(m *router.ReConnectNotify, resp *int) error {
	Logger.Trace(m.Tags, m.Gateways, m.Ip, "router.SendReConnectNotify", m.Port, m.MoreIps, m.UserGateways)
	ximp, err := createReConnectNotify(m.Ip, m.Port, m.MoreIps)
	if err != nil {
		return err
	}
	for _, gwAddr := range m.Gateways {
		if err := sendNotify(gwAddr, m.Tags, "", nil, ximp, true); err != nil {
			Logger.Error(m.Tags, gwAddr, "", "SendReConnectNotify", "sendNotify error", err)
		}
	}
	for _, userGateway := range m.UserGateways {
		if err := sendNotify(userGateway.GatewayAddr, nil, "", []logic.ConnectionId{userGateway.ConnId}, ximp, true); err != nil {
			Logger.Error(userGateway.GatewayAddr, userGateway.ConnId, "", "SendReConnectNotify", "sendNotify error", err)
		}
	}
	return nil
}

// 批量发送聊天室消息
func (this *GorpcService) SendChatRoomNotifyBatch(req *router.ChatRoomNotifyBatchReq, resp *int) error {
	Logger.Trace("", "", "", "router.SendChatRoomNotifyBatch", req.IsWeb, req.Gwys)
	for _, m := range req.Msgs {
		m.GatewayAddrs = req.Gwys
		var tags []string
		if !req.IsWeb {
			tags = []string{logic.GenerateChatRoomTag(m.Appid, m.RoomID)}
		} else {
			tags = []string{logic.GenerateWebChatRoomTag(m.Appid, m.RoomID)}
		}
		if err := sendChatRoomNotify(tags, m); err != nil {
			Logger.Error(m.RoomID, m.Appid, m.MsgID, "router.SendChatRoomNotifyBatch", "sendChatRoomNotify error", err, "TraceId", m.TraceId)
			return err
		}
	}
	return nil
}

// 发送聊天室消息接口
func (this *GorpcService) SendChatRoomNotify(m *logic.ChatRoomMessageNotify, resp *int) error {
	Logger.Trace(m.RoomID, m.Appid, m.TraceId, "router.SendChatRoomNotify", m.MsgID, m.Priority, m.Sender)
	tags := []string{logic.GenerateChatRoomTag(m.Appid, m.RoomID)}
	if err := sendChatRoomNotify(tags, m); err != nil {
		Logger.Error(m.RoomID, m.Appid, m.MsgID, "router.SendChatRoomNotify", "sendChatRoomNotify error", err, "TraceId", m.TraceId)
		return err
	}
	return nil
}

// 发送聊天室消息给聊天室单个用户接口
func (this *GorpcService) PrivateSendChatRoomNotify(m *logic.PrivateChatRoomMessageNotify, resp *int) error {
	Logger.Trace(m.RoomID, m.Appid, m.TraceId, "router.PrivateSendChatRoomNotify", m.MsgID, m.Priority, m.Sender)
	if err := privateSendChatRoomNotify(m); err != nil {
		Logger.Error(m.RoomID, m.Appid, m.MsgID, "router.PrivateSendChatRoomNotify", "privateSendChatRoomNotify error", err, "TraceId", m.TraceId)
		return err
	}
	return nil
}

// 向所有直播间发送世界消息
func (this *GorpcService) SendChatRoomBroadcast(m *logic.ChatRoomMessageNotify, resp *int) error {
	Logger.Trace(m.RoomID, m.Appid, m.TraceId, "router.SendChatRoomBroadcast", m.MsgID, m.Priority, m.Sender)
	if err := sendChatRoomBroadcast(m); err != nil {
		Logger.Error(m.RoomID, m.Appid, m.MsgID, "router.SendChatRoomBroadcast", "sendChatRoomNotify error", err, "TraceId", m.TraceId)
		return err
	}
	return nil
}

// 向所有连接发送世界消息
func (this *GorpcService) SendOnlineBroadcast(m *logic.ChatRoomMessageNotify, resp *int) error {
	Logger.Trace(m.RoomID, m.Appid, m.TraceId, "router.SendOnlineBroadcast", m.MsgID, m.Priority, m.Sender)
	if err := sendOnlineBroadcast(m); err != nil {
		Logger.Error(m.RoomID, m.Appid, m.MsgID, "router.SendOnlineBroadcast", "sendOnlineBroadcast error", err, "TraceId", m.TraceId)
		return err
	}
	return nil
}
