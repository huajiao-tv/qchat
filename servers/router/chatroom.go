package main

import (
	"encoding/binary"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/pb"
	"github.com/huajiao-tv/qchat/utility/network"
)

func dealChatroom(gwp *router.GwPackage, crm *pb.ChatRoomPacket, gwr *router.GwResp) (*pb.ChatRoomPacket, *Error) {
	var crResp *pb.ChatRoomPacket
	appid, _ := strconv.Atoi(gwp.Property["Appid"])
	switch *crm.ToServerData.Payloadtype {
	case pb.CR_PAYLOAD_JOIN:
		if crm.ToServerData.Applyjoinchatroomreq == nil || len(crm.ToServerData.Applyjoinchatroomreq.Roomid) == 0 {
			return nil, NewError(pb.ERR_BAD_PARAM, "field applyjoinchatroomreq|roomid is nil")
		}
		props := getJoinProperties(crm.ToServerData.Applyjoinchatroomreq, gwp.Property["ConnectionType"] == network.WebSocketNetwork)
		props["ClientIp"] = gwp.Property["ClientIp"]
		props["ConnectionType"] = gwp.Property["ConnectionType"]
		props["Platform"] = gwp.Property["Platform"] + "::" + gwp.Property["MobileType"]
		props["Deviceid"] = gwp.Property["Deviceid"]
		resp, err := session.JoinChatRoom(gwp.Property["Sender"], gwp.Property["Appid"], gwp.GatewayAddr, gwp.ConnectionId, gwp.Property["ConnectionType"], string(crm.Roomid), props)
		if err != nil {
			if netConf().IgnoreChatroomErrors {
				Logger.Warn(gwp.Property["Sender"], appid, logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealChatroom.Join", "session.JoinChatRoom("+string(crm.Roomid)+") error ignored", err)
				resp = session.JoinChatRoomSuccessResp
				if props["audienceflag"] == "0" {
					resp.Response = []byte("")
				}
			} else {
				Logger.Error(gwp.Property["Sender"], appid, logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealChatroom.Join", "session.JoinChatRoom("+string(crm.Roomid)+") error", err)
				return nil, NewError(pb.ERR_SESSION_REFUSED, err.Error())
			}
		}
		pullLost := true
		if len(logic.NetGlobalConf().PullLost) > 0 {
			if pl, ok := logic.NetGlobalConf().PullLost[string(crm.Roomid)]; ok {
				pullLost = pl
			} else if pl, ok := logic.NetGlobalConf().PullLost["default"]; ok {
				pullLost = pl
			}
		}

		crResp = &pb.ChatRoomPacket{
			Roomid: crm.Roomid,
			ToUserData: &pb.ChatRoomDownToUser{
				Result:      proto.Int(resp.Code),
				Reason:      []byte(resp.Reason),
				Payloadtype: proto.Uint32(pb.CR_PAYLOAD_JOIN_RESP),
				Applyjoinchatroomresp: &pb.ApplyJoinChatRoomResponse{
					Room: &pb.ChatRoom{
						Roomid: crm.Roomid,
					},
					PullLost: &pullLost,
				},
			},
		}
		if resp.Code == session.Success {
			crResp.ToUserData.Applyjoinchatroomresp.Room = &pb.ChatRoom{
				Roomid: crm.Roomid,
				// Version:     proto.Uint64(uint64(resp.Version)),
				Version:  proto.Uint64(0),
				Roomtype: proto.String("tantan"),
				// Maxmsgid:    proto.Uint64(uint64(resp.MaxID)),
				Maxmsgid:    proto.Uint64(0),
				Partnerdata: resp.Response,
				Properties: []*pb.CRPair{
					&pb.CRPair{
						Key: proto.String("regmemcount"),
						// Value: []byte(strconv.Itoa(resp.Registered())),
						Value: []byte("0"),
					},
					&pb.CRPair{
						Key: proto.String("memcount"),
						// Value: []byte(strconv.Itoa(resp.MemberCount())),
						Value: []byte("0"),
					},
				},
				Members: []*pb.CRUser{
					&pb.CRUser{
						Userid: []byte(gwp.Property["Sender"]),
					},
				},
			}
			gwr.Tags = make(map[string]bool)
			tag := logic.GenerateChatRoomTag(gwp.Property["Appid"], string(crm.Roomid))
			gwr.Tags[tag] = true
			if gwp.Property["ConnectionType"] == network.WebSocketNetwork {
				tag := logic.GenerateWebChatRoomTag(gwp.Property["Appid"], string(crm.Roomid))
				gwr.Tags[tag] = true
			}
		}

	case pb.CR_PAYLOAD_QUIT:
		if crm.ToServerData.Quitchatroomreq == nil || len(crm.ToServerData.Quitchatroomreq.Roomid) == 0 {
			return nil, NewError(pb.ERR_BAD_PARAM, "field quitchatroomreq|roomid is nil")
		}
		props := map[string]string{
			"ClientIp":       gwp.Property["ClientIp"],
			"ConnectionType": gwp.Property["ConnectionType"],
			"Platform":       gwp.Property["Platform"] + "::" + gwp.Property["MobileType"],
			"Deviceid":       gwp.Property["Deviceid"],
		}
		resp, err := session.QuitChatRoom(gwp.Property["Sender"], gwp.Property["Appid"], gwp.GatewayAddr, gwp.ConnectionId, gwp.Property["ConnectionType"], string(crm.Roomid), props)
		if err != nil {
			if netConf().IgnoreChatroomErrors {
				Logger.Error(gwp.Property["Sender"], appid, logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealChatroom.Quit", "session.QuitChatRoom("+string(crm.Roomid)+") error ignored", err)
				resp = session.QuitChatRoomSuccessResp
			} else {
				Logger.Error(gwp.Property["Sender"], appid, logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealChatroom.Quit", "session.QuitChatRoom("+string(crm.Roomid)+") error", err)
				return nil, NewError(pb.ERR_SESSION_REFUSED, err.Error())
			}
		}
		crResp = &pb.ChatRoomPacket{
			Roomid: crm.Roomid,
			ToUserData: &pb.ChatRoomDownToUser{
				Result:      proto.Int(resp.Code),
				Reason:      []byte(resp.Reason),
				Payloadtype: proto.Uint32(pb.CR_PAYLOAD_QUIT_RESP),
				Quitchatroomresp: &pb.QuitChatRoomResponse{
					Room: &pb.ChatRoom{
						Roomid: crm.Roomid,
					},
				},
			},
		}
		gwr.Tags = make(map[string]bool)
		tag := logic.GenerateChatRoomTag(gwp.Property["Appid"], string(crm.Roomid))
		gwr.Tags[tag] = false
		if gwp.Property["ConnectionType"] == network.WebSocketNetwork {
			tag := logic.GenerateWebChatRoomTag(gwp.Property["Appid"], string(crm.Roomid))
			gwr.Tags[tag] = false
		}

	case pb.CR_PAYLOAD_QUERY:
		if crm.ToServerData.Getchatroominforeq == nil || len(crm.ToServerData.Getchatroominforeq.Roomid) == 0 {
			return nil, NewError(pb.ERR_BAD_PARAM, "field getchatroominforeq|roomid is nil")
		}
		resp, err := session.QueryChatRoom(gwp.Property["Sender"], gwp.Property["Appid"], string(crm.Roomid))
		if err != nil {
			Logger.Error(gwp.Property["Sender"], appid, logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealChatroom.Query", "session.QueryChatRoom("+string(crm.Roomid)+") error", err)
			return nil, NewError(pb.ERR_SESSION_REFUSED, err.Error())
		}
		crResp = &pb.ChatRoomPacket{
			Roomid: crm.Roomid,
			ToUserData: &pb.ChatRoomDownToUser{
				Result:      proto.Int(0),
				Payloadtype: proto.Uint32(pb.CR_PAYLOAD_QUERY),
				Getchatroominforesp: &pb.GetChatRoomDetailResponse{
					Room: &pb.ChatRoom{
						Roomid:   crm.Roomid,
						Maxmsgid: proto.Uint64(uint64(resp.MaxID)),
						Properties: []*pb.CRPair{
							&pb.CRPair{
								Key:   proto.String("regmemcount"),
								Value: []byte(strconv.Itoa(resp.Registered())),
							},
							&pb.CRPair{
								Key:   proto.String("memcount"),
								Value: []byte(strconv.Itoa(resp.MemberCount())),
							},
						},
					},
				},
			},
		}

	case pb.CR_PAYLOAD_SUB:
		if crm.ToServerData.Subreq == nil || len(crm.ToServerData.Subreq.Roomid) == 0 {
			return nil, NewError(pb.ERR_BAD_PARAM, "field subreq|roomid is nil")
		}
		gwr.Tags = make(map[string]bool)
		tag := logic.GenerateChatRoomTag(gwp.Property["Appid"], string(crm.ToServerData.Subreq.Roomid))
		gwr.Tags[tag] = crm.ToServerData.Subreq.GetSub()
		Logger.Trace(gwp.Property["Sender"], gwp.Property["Appid"], logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealChatroom.Subscribe", crm.ToServerData.Subreq.Roomid, crm.ToServerData.Subreq.Sub)
		crResp = &pb.ChatRoomPacket{
			Roomid: crm.Roomid,
			ToUserData: &pb.ChatRoomDownToUser{
				Result:      proto.Int(0),
				Payloadtype: proto.Uint32(pb.CR_PAYLOAD_SUB),
				Subresp: &pb.SubscribeResponse{
					Roomid: crm.ToServerData.Subreq.Roomid,
					Sub:    crm.ToServerData.Subreq.Sub,
				},
			},
		}

	default:
		return nil, NewError(pb.ERR_BAD_PARAM, "unknown payload")
	}
	return crResp, nil

}

func genPrivateChatroomNotifyXimp(m *logic.PrivateChatRoomMessageNotify) (*network.XimpBuffer, error) {
	return genChatroomNotifyXimp(&logic.ChatRoomMessageNotify{ChatRoomMessage: m.ChatRoomMessage, TraceId: m.TraceId})
}
func genChatroomNotifyXimp(m *logic.ChatRoomMessageNotify) (*network.XimpBuffer, error) {
	msgid := proto.Uint32(uint32(m.MsgID))
	if m.MsgID == 0 {
		msgid = nil
	}
	packet := &pb.ChatRoomPacket{
		Roomid: []byte(m.RoomID),
		Appid:  proto.Uint32(uint32(m.Appid)),
		ToUserData: &pb.ChatRoomDownToUser{
			Result:      proto.Int32(0),
			Payloadtype: proto.Uint32(pb.CR_PAYLOAD_INCOMING_MSG),
			Newmsgnotify: &pb.ChatRoomNewMsg{
				Roomid: []byte(m.RoomID),
				Sender: &pb.CRUser{
					Userid: []byte(m.Sender),
				},
				Msgtype:     proto.Int(m.MsgType),
				Msgcontent:  m.MsgContent,
				Regmemcount: proto.Int(m.RegMemCount),
				Memcount:    proto.Int(m.MemCount),
				Msgid:       msgid,
				Maxid:       proto.Uint32(uint32(m.MaxID)),
				Timestamp:   proto.Uint64(uint64(m.TimeStamp)),
			},
		},
	}
	content, err := proto.Marshal(packet)
	if err != nil {
		return nil, err
	}
	ximp, err := pb.CreateMsgNotify("chatroom", content, int64(m.MsgID), m.Sender, "", getQueryAfterSeconds())
	if err != nil {
		return nil, err
	}
	ximp.TimeStamp = m.TimeStamp // set timestamp so that gateway can measure chat room message latency
	ximp.TraceId = m.TraceId
	return ximp, nil

}

func sendChatRoomBroadcast(m *logic.ChatRoomMessageNotify) error {
	ximp, err := genChatroomNotifyXimp(m)
	if err != nil {
		return err
	}
	tag := logic.GenerateChatRoomBroadcastTag(m.Appid)
	//Logger.Debug(tag, m.Appid, m.MsgID, "sendChatRoomBroadcast", "ChatRoomPacket", m.Priority, m.TraceId, time.Now().UnixNano()/1e6-m.TimeStamp)
	for gwAddr, _ := range m.GatewayAddrs {
		if err := sendNotify(gwAddr, nil, tag, nil, ximp, m.Priority); err != nil {
			Logger.Error(tag, m.Appid, m.MsgID, "sendChatRoomBroadcast", "sendNotify error", err, "TraceId", m.TraceId)
		}
	}
	return nil
}

func sendOnlineBroadcast(m *logic.ChatRoomMessageNotify) error {
	ximp, err := genChatroomNotifyXimp(m)
	if err != nil {
		return err
	}
	for gwAddr, _ := range m.GatewayAddrs {
		if err := sendNotify(gwAddr, []string{strconv.Itoa(int(m.Appid))}, "", nil, ximp, m.Priority); err != nil {
			Logger.Error(m.Appid, m.Appid, m.MsgID, "sendOnlineBroadcast", "sendNotify error", err, "TraceId", m.TraceId)
		}
	}
	return nil
}

func sendChatRoomNotify(tags []string, m *logic.ChatRoomMessageNotify) error {
	ximp, err := genChatroomNotifyXimp(m)
	if err != nil {
		return err
	}
	if m.Delay > 0 || m.Interval > 0 {
		go func() {
			if m.Delay > 0 {
				time.Sleep(m.Delay)
			}
			doSendChatRoomNotify(tags, ximp, m)
		}()
		return nil
	}
	return doSendChatRoomNotify(tags, ximp, m)
}

// called ONLY by sendChatRoomNotify()
func doSendChatRoomNotify(tags []string, ximp *network.XimpBuffer, m *logic.ChatRoomMessageNotify) error {
	//Logger.Debug(tag, m.Appid, m.MsgID, "sendChatRoomNotify", "ChatRoomPacket", m.Priority, m.TraceId, time.Now().UnixNano()/1e6-m.TimeStamp)
	for gwAddr, members := range m.GatewayAddrs {
		if members < 0 {
			Logger.Error(tags, m.Appid, m.MsgID, "sendChatRoomNotify", "negative gateway members", "TraceId", m.TraceId)
		}
		if err := sendNotify(gwAddr, tags, "", nil, ximp, m.Priority); err != nil {
			Logger.Error(tags, m.Appid, m.MsgID, "sendChatRoomNotify", "sendNotify error", err, "TraceId", m.TraceId)
		}
		if m.Interval > 0 {
			time.Sleep(m.Interval)
		}
	}
	return nil
}

func privateSendChatRoomNotify(m *logic.PrivateChatRoomMessageNotify) error {
	ximp, err := genPrivateChatroomNotifyXimp(m)
	if err != nil {
		return err
	}

	//Logger.Debug(m.Appid, m.MsgID, "privateSendChatRoomNotify", "ChatRoomPacket", m.Priority, m.TraceId, time.Now().UnixNano()/1e6-m.TimeStamp)
	for _, userGateway := range m.UserGateways {
		if err := sendNotify(userGateway.GatewayAddr, []string{}, "", []logic.ConnectionId{userGateway.ConnId}, ximp, m.Priority); err != nil {
			Logger.Error(m.Appid, m.MsgID, "privateSendChatRoomNotify", "sendNotify error", err, "TraceId", m.TraceId)
		}
	}
	return nil
}

func getCachedChatRoomMessage(req *pb.GetMultiInfosReq, appid string) ([]*pb.Info, error) {
	if netConf().MaxPullMessage <= 0 {
		return []*pb.Info{}, nil
	}
	var cutMsgids []int64
	if len(req.GetInfoIds) > netConf().MaxPullMessage {
		cutMsgids = req.GetInfoIds[len(req.GetInfoIds)-netConf().MaxPullMessage:]
	} else {
		cutMsgids = req.GetInfoIds
	}
	msgids := make([]uint, len(cutMsgids))
	for i, id := range cutMsgids {
		msgids[i] = uint(id)
	}
	fetchReq := &saver.FetchChatRoomMessageReq{
		RoomID: string(req.SParameter),
		Appid:  logic.StringToUint16(appid),
		MsgIDs: msgids,
	}
	resp, err := saver.GetCachedChatRoomMessages(fetchReq)
	if err != nil {
		Logger.Error("", appid, fetchReq.RoomID, "router.getCachedChatRoomMessage", "saver.GetCachedChatRoomMessages error", err.Error())
	}

	infos := make([]*pb.Info, 0, len(msgids))
	buf := make([]byte, len(msgids)*20)
	for i, id := range msgids {
		valid, infoID, timestamp := buf[i*20:i*20+4], buf[i*20+4:i*20+12], buf[i*20+12:i*20+20]
		binary.BigEndian.PutUint64(infoID, uint64(id))
		data, err := pb.CompressChatRoomNewMsg(resp[id])
		if err != nil {
			Logger.Error("", appid, fetchReq.RoomID, "router.getCachedChatRoomMessage", "saver.GetCachedChatRoomMessages error", err.Error())
		}
		if data != nil {
			binary.BigEndian.PutUint32(valid, 1)
			binary.BigEndian.PutUint64(timestamp, uint64(resp[id].TimeStamp))
			infos = append(infos, &pb.Info{
				PropertyPairs: []*pb.Pair{
					&pb.Pair{
						Key:   []byte("msg_valid"),
						Value: valid,
					},
					&pb.Pair{
						Key:   []byte("info_id"),
						Value: infoID,
					},
					&pb.Pair{
						Key:   []byte("chat_body"),
						Value: data,
					},
					&pb.Pair{
						Key:   []byte("time_sent"),
						Value: timestamp,
					},
				},
			})
		} else {
			binary.BigEndian.PutUint32(valid, 0)
			infos = append(infos, &pb.Info{
				PropertyPairs: []*pb.Pair{
					&pb.Pair{
						Key:   []byte("msg_valid"),
						Value: valid,
					},
					&pb.Pair{
						Key:   []byte("info_id"),
						Value: infoID,
					},
				},
			})
		}
	}
	return infos, nil
}

func getJoinProperties(req *pb.ApplyJoinChatRoomRequest, webConnection bool) map[string]string {
	flag := "1"
	if webConnection || req.GetNoUserlist() {
		flag = "0"
	}
	if req.GetRoom() == nil || req.GetRoom().GetProperties() == nil {
		return map[string]string{
			"audienceflag": flag,
		}
	}
	properties := make(map[string]string, len(req.GetRoom().GetProperties())+1)
	properties["audienceflag"] = flag
	for _, prop := range req.Room.Properties {
		key := prop.GetKey()
		value := prop.GetValue()
		if key == "" || value == nil {
			Logger.Error("", "", "", "router.getJoinProperties", "invalid key or value", key, value)
			continue
		}
		properties[key] = string(value)
	}
	return properties
}
