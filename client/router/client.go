package router

import (
	"time"

	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/network"
	"github.com/johntech-o/gorpc"
)

var GorpcClient *gorpc.Client

func init() {
	netOptions := gorpc.NewNetOptions(gorpc.DefaultConnectTimeout, gorpc.DefaultReadTimeout, gorpc.DefaultWriteTimeout)
	GorpcClient = gorpc.NewClient(netOptions)

	statNetOption := gorpc.NewNetOptions(1*time.Second, 1*time.Second, 1*time.Second)
	GorpcClient.SetMethodNetOptinons("GorpcService", "GetRouterQps", statNetOption)
	GorpcClient.SetMethodNetOptinons("GorpcService", "GetRouterTotalOps", statNetOption)
}

// 给gateway调用，将包路由到router
func RoutePackage(buf *network.XimpBuffer, prop map[string]string, addr string, connectionId logic.ConnectionId) (*GwResp, error) {
	gwResp := &GwResp{}
	req := &GwPackage{
		XimpBuff:     buf,
		Property:     prop,
		GatewayAddr:  addr,
		ConnectionId: connectionId,
	}
	if err := GorpcClient.CallWithAddress(logic.GetRouterGorpc(), "GorpcService", "RoutePackage", req, &gwResp); err != nil {
		return gwResp, err
	}
	return gwResp, nil
}

func Logout(prop map[string]string, addr string, connectionId logic.ConnectionId) (*GwResp, error) {
	gwResp := &GwResp{
		Property: make(map[string]string),
	}
	req := &GwPackage{
		Property:     prop,
		GatewayAddr:  addr,
		ConnectionId: connectionId,
	}
	if err := GorpcClient.CallWithAddress(logic.GetRouterGorpc(), "GorpcService", "Logout", req, gwResp); err != nil {
		return nil, err
	}
	return gwResp, nil
}

// 发送reconnect请求
func SendReConnectNotify(ip string, port uint32, moreIps []string, gateways []string, tags []string, userGateways []*logic.UserGateway) error {
	req := &ReConnectNotify{
		ip,
		port,
		moreIps,
		gateways,
		tags,
		userGateways,
	}
	if err := GorpcClient.CallWithAddress(logic.GetRouterGorpc(), "GorpcService", "SendReConnectNotify", req, nil); err != nil {
		return err
	}
	return nil
}

func SendPushTags(tags, gatewayAddrs []string, sender, channel, traceSn string, msgId uint64) error {
	req := &ChatPushTags{
		tags,
		gatewayAddrs,
		sender,
		channel,
		msgId,
		traceSn,
	}
	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetRouterGorpc(), "GorpcService", "SendPushTags", req, &resp); err != nil {
		return err
	}
	return nil
}

func SendMsgNotify(gatewayAddr string, tags []string, connIds []logic.ConnectionId, messageNotify *logic.MessageNotify) error {
	req := &MsgNotify{
		GatewayAddr:   gatewayAddr,
		Tags:          tags,
		ConnIds:       connIds,
		MessageNotify: messageNotify,
	}

	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetRouterGorpc(), "GorpcService", "SendMsgNotify", req, &resp); err != nil {
		return err
	}
	return nil
}

func SendPushNotifyBatch(req []*ChatPushNotify) error {
	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetRouterGorpc(), "GorpcService", "SendPushNotifyBatch", req, &resp); err != nil {
		return err
	}
	return nil
}

func SendChatRoomNotify(message *logic.ChatRoomMessage, gatewayAddrs map[string]int, traceId string) error {
	req := &logic.ChatRoomMessageNotify{
		ChatRoomMessage: message,
		TraceId:         traceId,
		GatewayAddrs:    gatewayAddrs,
	}
	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetRouterGorpc(), "GorpcService", "SendChatRoomNotify", req, &resp); err != nil {
		return err
	}
	return nil
}

func SendChatRoomNotifyWithDelay(message *logic.ChatRoomMessage, gatewayAddrs map[string]int, traceId string, delay, interval time.Duration) error {
	req := &logic.ChatRoomMessageNotify{
		ChatRoomMessage: message,
		TraceId:         traceId,
		GatewayAddrs:    gatewayAddrs,
		Delay:           delay,
		Interval:        interval,
	}
	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetRouterGorpc(), "GorpcService", "SendChatRoomNotify", req, &resp); err != nil {
		return err
	}
	return nil
}

func PrivateSendChatRoomNotify(message *logic.ChatRoomMessage, userGateways []*logic.UserGateway, traceId string) error {
	req := &logic.PrivateChatRoomMessageNotify{
		ChatRoomMessage: message,
		TraceId:         traceId,
		UserGateways:    userGateways,
	}
	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetRouterGorpc(), "GorpcService", "PrivateSendChatRoomNotify", req, &resp); err != nil {
		return err
	}
	return nil
}

func SendChatRoomBroadcast(message *logic.ChatRoomMessage, gatewayAddrs map[string]int, traceId string) error {
	req := &logic.ChatRoomMessageNotify{
		ChatRoomMessage: message,
		TraceId:         traceId,
		GatewayAddrs:    gatewayAddrs,
	}
	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetRouterGorpc(), "GorpcService", "SendChatRoomBroadcast", req, &resp); err != nil {
		return err
	}
	return nil
}

func SendOnlineBroadcast(message *logic.ChatRoomMessage, gatewayAddrs map[string]int, traceId string) error {
	req := &logic.ChatRoomMessageNotify{
		ChatRoomMessage: message,
		TraceId:         traceId,
		GatewayAddrs:    gatewayAddrs,
	}
	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetRouterGorpc(), "GorpcService", "SendOnlineBroadcast", req, &resp); err != nil {
		return err
	}
	return nil
}

func SendChatRoomNotifyBatch(msgs []*logic.ChatRoomMessageNotify, gwys map[string]int, isWeb bool) error {
	req := &ChatRoomNotifyBatchReq{
		msgs,
		gwys,
		isWeb,
	}
	if err := GorpcClient.CallWithAddress(logic.GetRouterGorpc(), "GorpcService", "SendChatRoomNotifyBatch", req, nil); err != nil {
		return err
	}
	return nil
}
