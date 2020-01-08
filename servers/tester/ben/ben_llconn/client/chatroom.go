package client

import (
	"math/rand"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/pb"
	"github.com/huajiao-tv/qchat/utility/gzipPool"
)

const (
	ChatroomServiceID = 10000006

	PayloadQuery = 101 // 查询聊天室信息
	PayloadJoin  = 102 // 加入聊天室
	PayloadQuit  = 103 // 退出聊天室

	PayloadIncomingMsg = 1000 // 消息通知
	PayloadCompressed  = 1003 // 压缩消息
)

type ChatroomService struct {
	c    *UserConnection // long live connection
	room string          // current chatroom
	p    *gzipPool.DecompressPool
}

func (this *ChatroomService) HandleServiceMessage(data []byte, sn uint64, source int) bool {
	if this.c.verbose {
		this.c.log("verbose", "chatroom", "HandleServiceMessage")
	}
	p := this.decode(data)
	if p == nil {
		return false
	}

	toUserData := p.ToUserData
	if toUserData == nil {
		return false
	}

	result := toUserData.GetResult()
	reason := toUserData.GetReason()

	switch toUserData.GetPayloadtype() {
	case PayloadJoin:
		this.parseJoinResp(sn, result, reason, toUserData.Applyjoinchatroomresp)
	case PayloadQuit:
		this.parseQuitResp(sn, result, toUserData.Quitchatroomresp)
	case PayloadQuery:
		this.parseQueryResp(sn, result, toUserData.Getchatroominforesp)
	case PayloadIncomingMsg:
		this.parseChatroomMessage(sn, result, toUserData.Newmsgnotify)
	case PayloadCompressed:
		this.parseCompressedMessage(sn, result, toUserData.Multinotify)
	}
	return true
}

func (this *ChatroomService) HandleGetMessage(msg *UserMessage) bool {
	if msg == nil {
		this.c.log("error", "ChatroomService", "nil message")
		return false
	}

	if msg.Content != nil && len(msg.Content) > 0 {
		b := this.p.Decompress(msg.Content)
		m := pb.ChatRoomNewMsg{}
		err := proto.Unmarshal(b, &m)
		if err != nil {
			this.c.log("error", "ChatroomService", "unmarshal failed", err.Error())
		} else {
			this.parseChatroomMessage(msg.Sn, 0, &m)
		}
	} else {
		this.c.log("warn", "GetMultiInfo Resp", "ChatroomMessage", "no content", "id", msg.ID)
	}
	return true
}

func (this *ChatroomService) GetMultiInfo(args []string) {
	if len(args) < 2 {
		this.c.log("error", "Join", "invalid args")
		return
	}
	s := strings.Split(args[1], ",")
	ids := make([]int64, 0, len(s))
	for _, id := range s {
		v, _ := strconv.ParseInt(id, 10, 0)
		ids = append(ids, v)
	}
	m := &pb.Message{
		Sn:    proto.Uint64(logic.GetSn()),
		Msgid: proto.Uint32(pb.GET_MULITI_INFOS_REQ),
		Req: &pb.Request{
			GetMultiInfos: &pb.GetMultiInfosReq{
				InfoType:   proto.String(ChatRoom),
				GetInfoIds: ids,
				SParameter: []byte(args[0]),
			},
		},
	}
	this.c.send(m)
}

func (this *ChatroomService) Join(args []string) bool {
	if len(args) < 1 {
		this.c.log("error", "Join", "invalid args")
		return false
	}
	p := &pb.ChatRoomPacket{
		Roomid:   []byte(args[0]),
		Appid:    proto.Uint32(this.c.conf.AppID),
		ClientSn: proto.Uint64(uint64(rand.Int63())),
		ToServerData: &pb.ChatRoomUpToServer{
			Applyjoinchatroomreq: &pb.ApplyJoinChatRoomRequest{
				Roomid: []byte(args[0]),
			},
			Payloadtype: proto.Uint32(PayloadJoin),
		},
	}
	return this.send(p)
}

func (this *ChatroomService) Quit(args []string) bool {
	if len(args) < 1 {
		this.c.log("error", "Quit", "invalid args")
		return false
	}
	p := &pb.ChatRoomPacket{
		Roomid:   []byte(args[0]),
		Appid:    proto.Uint32(this.c.conf.AppID),
		ClientSn: proto.Uint64(uint64(rand.Int63())),
		ToServerData: &pb.ChatRoomUpToServer{
			Quitchatroomreq: &pb.QuitChatRoomRequest{
				Roomid: []byte(args[0]),
			},
			Payloadtype: proto.Uint32(PayloadQuit),
		},
	}
	return this.send(p)
}

func (this *ChatroomService) Query(args []string) bool {
	if len(args) < 3 {
		this.c.log("error", "Query", "invalid args")
		return false
	}
	index, _ := strconv.Atoi(args[1])
	count, _ := strconv.Atoi(args[2])
	p := &pb.ChatRoomPacket{
		Roomid:   []byte(args[0]),
		Appid:    proto.Uint32(this.c.conf.AppID),
		ClientSn: proto.Uint64(uint64(rand.Int63())),
		ToServerData: &pb.ChatRoomUpToServer{
			Getchatroominforeq: &pb.GetChatRoomDetailRequest{
				Roomid: []byte(args[0]),
				Index:  proto.Int32(int32(index)),
				Offset: proto.Int32(int32(count)),
			},
			Payloadtype: proto.Uint32(PayloadQuery),
		},
	}
	return this.send(p)
}

func (this *ChatroomService) send(p *pb.ChatRoomPacket) bool {
	data, err := proto.Marshal(p)
	if err != nil {
		this.c.log("error", "Send ChatroomService", "proto.Marshal", err.Error())
		return false
	}
	return this.c.RequestService(ChatroomServiceID, data)
}

func (this *ChatroomService) decode(data []byte) *pb.ChatRoomPacket {
	if data == nil {
		return nil
	}
	var p pb.ChatRoomPacket
	if err := proto.Unmarshal(data, &p); err != nil {
		this.c.log("error", "unmarshal failed", "chatroom packet", err.Error())
		return nil
	}
	return &p
}

// get members and total count from a chatroom
func parseChatRoom(room *pb.ChatRoom) ([]string, int32) {
	var members []string
	var totalCount int32 = 0

	if room == nil {
		return members, totalCount
	}
	if room.Members != nil && len(room.Members) > 0 {
		members = make([]string, len(room.Members))
		for i, p := range room.Members {
			id := string(p.Userid)
			members[i] = id
		}
	}
	if room.Properties != nil && len(room.Properties) > 0 {
		for _, p := range room.Properties {
			if p.GetKey() == "memcount" {
				i, _ := strconv.Atoi(string(p.GetValue()))
				totalCount = int32(i)
				break
			}
		}
	}
	return members, totalCount
}

func (this *ChatroomService) parseJoinResp(sn uint64, result int32, reason []byte, resp *pb.ApplyJoinChatRoomResponse) {
	if resp == nil || resp.Room == nil || resp.Room.Roomid == nil {
		return
	}
	members, count := parseChatRoom(resp.Room)
	this.c.log("JoinResp", result, resp.Room.Roomid, "count", count, "memebers", strings.Join(members, ","))
}

func (this *ChatroomService) parseQuitResp(sn uint64, result int32, resp *pb.QuitChatRoomResponse) {
	if resp != nil {
		members, count := parseChatRoom(resp.Room)
		this.c.log("QuitResp", result, resp.Room.Roomid, "count", count, "members", strings.Join(members, ","))
	}
}

func (this *ChatroomService) parseQueryResp(sn uint64, result int32, resp *pb.GetChatRoomDetailResponse) {
	if resp == nil || resp.Room == nil || resp.Room.Roomid == nil {
		return
	}
	members, count := parseChatRoom(resp.Room)
	this.c.log("QueryResp", result, resp.Room.Roomid, "count", count, "members", strings.Join(members, ","))
}

func (this *ChatroomService) parseChatroomMessage(sn uint64, result int32, msg *pb.ChatRoomNewMsg) {
	if msg != nil {
		this.c.log("ChatroomMessage", result, msg.Roomid, msg.Sender.Userid, "id", msg.GetMsgid(), "max", msg.GetMaxid(), "count", msg.GetMemcount(), "content", msg.Msgcontent)
	}
}

func (this *ChatroomService) parseCompressedMessage(sn uint64, result int32, msgs []*pb.ChatRoomMNotify) {
	if msgs == nil || len(msgs) == 0 {
		return
	}
	for _, msg := range msgs {
		b := this.p.Decompress(msg.Data)
		m := pb.ChatRoomNewMsg{}
		err := proto.Unmarshal(b, &m)
		if err != nil {
			this.c.log("error", "ChatroomService", "unmarshal failed", err.Error())
		} else {
			this.parseChatroomMessage(sn, 0, &m)
		}
	}
}
