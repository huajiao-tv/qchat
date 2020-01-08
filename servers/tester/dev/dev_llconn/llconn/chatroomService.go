package llconn

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/huajiao-tv/qchat/logic/pb"
)

const (
	ChatroomServiceID = 10000006

	payloadQuery = 101 // 查询聊天室信息
	payloadJoin  = 102 // 加入聊天室
	payloadQuit  = 103 // 退出聊天室

	payloadIncomingMsg   = 1000 // 消息通知
	payloadMemberAdded   = 1001 // 加入通知
	payloadMemberRemoved = 1002 // 退出通知
	payloadCompressed    = 1003 // 压缩协议
)

type ChatroomService struct {
	llc  *LongLiveConn // long live connection
	room string        // current chatroom
}

type Dumper interface {
	dump()
}

// 加入聊天室的响应
type JoinResp struct {
	Sn             uint64
	Result         int32
	Reason         []byte
	PartnerData    []byte
	RoomID         string
	Members        []string
	TotalMemberCnt int32
}

// 聊天室详情
type QueryResp struct {
	RoomID         string
	Sn             uint64
	Result         int32
	TotalMemberCnt int32
	Members        []string
}

// 退出聊天室响应
type QuitResp struct {
	RoomID string
	Sn     uint64
	Result int32
}

// 成员变化通知
type MemberChangedNotifcation struct {
	RoomID         string
	Action         string
	Sn             uint64
	TotalMemberCnt int32
	Members        []string
}

//
//聊天室消息
//
type ChatroomMessage struct {
	Sn                  uint64
	ID                  uint32
	MaxID               uint32
	RoomID              string
	Sender              string
	MsgType             int32
	Content             []byte
	RegisteredMemberCnt int32 // 注册用户人数
	TotalMemberCnt      int32 // 总人数
	Valid               bool  // 是否有效
}

func NewChatroom() *ChatroomService {
	chatroom := ChatroomService{}
	return &chatroom
}

func (this *ChatroomService) HandleServiceMessage(data []byte, sn uint64, source int) bool {

	p := this.decodeChatroomPacket(data)

	if p == nil {
		return false
	}

	toUserData := p.ToUserData
	if toUserData == nil {
		return false
	}

	payload := toUserData.GetPayloadtype()
	result := toUserData.GetResult()
	reason := toUserData.GetReason()

	switch payload {

	case payloadJoin:
		this.parseJoinResp(sn, result, reason, toUserData.Applyjoinchatroomresp).dump()
	case payloadQuit:
		this.parseQuitResp(sn, result, toUserData.Quitchatroomresp).dump()
	case payloadQuery:
		this.parseQueryResp(sn, result, toUserData.Getchatroominforesp).dump()
	case payloadCompressed:
		dumpers := this.parseCompressedNotification(sn, toUserData.Multinotify)
		if dumpers != nil && len(dumpers) > 0 {
			for _, i := range dumpers {
				i.dump()
			}
		}
	case payloadIncomingMsg:

		this.parseChatroomMessage(sn, 0, 0, false, toUserData.Newmsgnotify).dump()

	case payloadMemberAdded:
		this.parseJoinNotification(sn, toUserData.Memberjoinnotify).dump()

	case payloadMemberRemoved:
		this.parseQuitNotification(sn, toUserData.Memberquitnotify).dump()
	}

	return true
}

func (this *ChatroomService) HandleGetMessage(msg *UserMessage) bool {

	if msg == nil {
		fmt.Println("HandleGetMessage msg is nil")
		return false
	}

	if msg.Content != nil && len(msg.Content) > 0 {
		b := this.ungzip(msg.Content)
		m := pb.ChatRoomNewMsg{}
		err := proto.Unmarshal(b, &m)
		if err != nil {
			fmt.Println("HandleGetMessage Unmarshal failed", err)
		} else {
			chatmsg := this.parseChatroomMessage(msg.Sn, 0, 0, false, &m)
			if chatmsg != nil {
				chatmsg.dump()
			}
		}
	} else {
		fmt.Println("Persist ChatroomMessage", msg.ID, "valid", msg.Valid, "content nil")
	}

	return true
}

func (this *ChatroomService) Register(conn *LongLiveConn) (int, error) {
	if nil != conn {
		this.llc = conn
		return ChatroomServiceID, nil
	} else {
		return ChatroomServiceID, errors.New("LongLiveConn is nil")
	}
}

//
// get chatroom message
//
func (this *ChatroomService) GetMessage(roomid string, ids []int64) bool {
	return this.llc.GetMultiInfo("chatroom", ids, []byte(roomid))
}

// join in a chatroom
// return positive sequence number to match coressponding async response, returning zero means failure
func (this *ChatroomService) Join(roomid string) uint64 {

	rid := []byte(roomid)
	payload := uint32(payloadJoin)
	appid := uint32(this.llc.clientConf.AppID)
	sn := uint64(rand.Int63())
	p := &pb.ChatRoomPacket{
		Roomid:   rid,
		Appid:    &appid,
		ClientSn: &sn,
		ToServerData: &pb.ChatRoomUpToServer{
			Applyjoinchatroomreq: &pb.ApplyJoinChatRoomRequest{
				Roomid: rid,
			},
			Payloadtype: &payload,
		},
	}

	return this.sendMessage(p)
}

// quit from a room
// return positive sequence number to match coressponding async response, returning zero means failure
func (this *ChatroomService) Quit(roomid string) uint64 {

	rid := []byte(roomid)
	payload := uint32(payloadQuit)
	appid := uint32(this.llc.clientConf.AppID)
	sn := uint64(rand.Int63())
	p := &pb.ChatRoomPacket{
		Roomid:   rid,
		Appid:    &appid,
		ClientSn: &sn,
		ToServerData: &pb.ChatRoomUpToServer{
			Quitchatroomreq: &pb.QuitChatRoomRequest{
				Roomid: rid,
			},
			Payloadtype: &payload,
		},
	}

	return this.sendMessage(p)
}

//
// query information of a specific room
// memberIndex and count are used to specify the range of members whose information should be retrieved
// return positive sequence number to match coressponding async response, returning zero means failure
//
func (this *ChatroomService) Query(roomid string, memberIndex, count int32) uint64 {

	rid := []byte(roomid)
	payload := uint32(payloadQuery)
	appid := uint32(this.llc.clientConf.AppID)
	sn := uint64(rand.Int63())
	p := &pb.ChatRoomPacket{
		Roomid:   rid,
		Appid:    &appid,
		ClientSn: &sn,
		ToServerData: &pb.ChatRoomUpToServer{
			Getchatroominforeq: &pb.GetChatRoomDetailRequest{
				Roomid: rid,
				Index:  &memberIndex,
				Offset: &count,
			},
			Payloadtype: &payload,
		},
	}

	return this.sendMessage(p)
}

func (this *ChatroomService) sendMessage(p *pb.ChatRoomPacket) uint64 {
	if nil == this.llc {
		fmt.Println("sendMessage: conn is nil")
		return 0
	}

	data, err := proto.Marshal(p)

	if err != nil {
		fmt.Println("sendMessage: proto.Marshal failed", err)
		return 0
	}

	return this.llc.SendServiceMessage(ChatroomServiceID, data)
}

func (this *ChatroomService) decodeChatroomPacket(data []byte) *pb.ChatRoomPacket {
	if data == nil {
		return nil
	}

	var p pb.ChatRoomPacket
	if err := proto.Unmarshal(data, &p); err != nil {
		fmt.Println("decodeChatroomPacket Unmarshal failed", err)
		return nil
	}

	return &p
}

// get members and total count from a chatroom
func getMembersAndTotalCount(room *pb.ChatRoom) ([]string, int32) {

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

func (this *ChatroomService) parseJoinNotification(sn uint64, notify *pb.MemberJoinChatRoomNotify) *MemberChangedNotifcation {
	if notify == nil || notify.Room == nil || notify.Room.Roomid == nil {
		return nil
	}

	members, totalCount := getMembersAndTotalCount(notify.Room)

	return &MemberChangedNotifcation{
		Sn:             sn,
		Action:         "join",
		RoomID:         string(notify.Room.Roomid),
		Members:        members,
		TotalMemberCnt: totalCount,
	}

}

func (this *ChatroomService) parseQuitNotification(sn uint64, notify *pb.MemberQuitChatRoomNotify) *MemberChangedNotifcation {
	if notify == nil || notify.Room == nil || notify.Room.Roomid == nil {
		return nil
	}

	members, totalCount := getMembersAndTotalCount(notify.Room)

	return &MemberChangedNotifcation{
		Sn:             sn,
		Action:         "quit",
		RoomID:         string(notify.Room.Roomid),
		Members:        members,
		TotalMemberCnt: totalCount,
	}
}

func (this *ChatroomService) parseJoinResp(sn uint64, result int32, reason []byte, resp *pb.ApplyJoinChatRoomResponse) *JoinResp {
	if resp == nil || resp.Room == nil || resp.Room.Roomid == nil {
		return nil
	}

	members, totalCount := getMembersAndTotalCount(resp.Room)

	return &JoinResp{
		Sn:             sn,
		RoomID:         string(resp.Room.Roomid),
		Members:        members,
		TotalMemberCnt: totalCount,
		Result:         result,
		Reason:         reason,
		PartnerData:    resp.GetRoom().GetPartnerdata(),
	}
}

func (this *ChatroomService) parseQuitResp(sn uint64, result int32, resp *pb.QuitChatRoomResponse) *QuitResp {
	if resp == nil {
		return nil
	}

	return &QuitResp{
		RoomID: string(resp.Room.Roomid),
		Sn:     sn,
		Result: result,
	}
}

func (this *ChatroomService) parseQueryResp(sn uint64, result int32, resp *pb.GetChatRoomDetailResponse) *QueryResp {
	if resp == nil || resp.Room == nil || resp.Room.Roomid == nil {
		return nil
	}

	members, totalCount := getMembersAndTotalCount(resp.Room)

	roomInfo := QueryResp{
		Sn:             sn,
		Result:         result,
		RoomID:         string(resp.Room.Roomid),
		Members:        members,
		TotalMemberCnt: totalCount,
	}

	return &roomInfo
}

func (this *ChatroomService) parseChatroomMessage(sn uint64, memCnt, regCnt int32, overwriteCnt bool, msg *pb.ChatRoomNewMsg) *ChatroomMessage {

	if msg == nil {
		return nil
	}

	cm := ChatroomMessage{
		Sn:                  sn,
		ID:                  msg.GetMsgid(),
		MaxID:               msg.GetMaxid(),
		RoomID:              string(msg.Roomid),
		Sender:              string(msg.Sender.Userid),
		MsgType:             msg.GetMsgtype(),
		Content:             msg.Msgcontent,
		Valid:               true,
		TotalMemberCnt:      msg.GetMemcount(),
		RegisteredMemberCnt: msg.GetRegmemcount(),
	}

	if overwriteCnt {
		cm.TotalMemberCnt = memCnt
		cm.RegisteredMemberCnt = regCnt
	}

	return &cm
}

func (this *ChatroomService) ungzip(data []byte) []byte {
	if data == nil || len(data) == 0 {
		return nil
	}

	var done bool = true
	bufLen := 512
	buf := make([]byte, bufLen)
	writer := bytes.NewBuffer(nil)
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		fmt.Println("gzip.NewReader failed", err)
		return nil
	}

	for {
		count, err1 := reader.Read(buf)
		if count == bufLen {
			writer.Write(buf)
		} else if count > 0 {
			writer.Write(buf[0:count])
			break
		} else {
			fmt.Println("ungzip failed", err1)
			done = false
			break
		}
	}

	reader.Close()

	if done {
		return writer.Bytes()
	}

	return nil
}

func (this *ChatroomService) parseCompressedNotification(sn uint64, notify []*pb.ChatRoomMNotify) []Dumper {

	if notify == nil || len(notify) == 0 {
		return nil
	}

	dumpers := make([]Dumper, 0)

	for _, n := range notify {
		memCnt := *n.Memcount
		regCnt := *n.Regmemcount
		if n.Data == nil || len(n.Data) == 0 {
			continue
		}

		unzipped := this.ungzip(n.Data)
		if unzipped == nil {
			continue
		}

		switch *n.Type {

		case int32(payloadMemberAdded):
			packet := pb.MemberJoinChatRoomNotify{}
			if err := proto.Unmarshal(unzipped, &packet); err != nil {
				fmt.Println("proto.Unmarshal MemberJoinChatRoomNotify failed")
			} else {
				mcn := this.parseJoinNotification(sn, &packet)
				dumpers = append(dumpers, mcn)
			}
		case int32(payloadMemberRemoved):
			packet := pb.MemberQuitChatRoomNotify{}
			if err := proto.Unmarshal(unzipped, &packet); err != nil {
				fmt.Println("proto.Unmarshal MemberQuitChatRoomNotify failed")
			} else {
				mcn := this.parseQuitNotification(sn, &packet)
				dumpers = append(dumpers, mcn)
			}
		case int32(payloadIncomingMsg):
			packet := pb.ChatRoomNewMsg{}
			if err := proto.Unmarshal(unzipped, &packet); err != nil {
				fmt.Println("proto.Unmarshal NewMessageNotify failed")
			} else {
				mcn := this.parseChatroomMessage(sn, memCnt, regCnt, true, &packet)
				dumpers = append(dumpers, mcn)
			}
		}

	}

	return dumpers
}

func (this *ChatroomMessage) dump() {
	if this == nil {
		fmt.Println("ChatroomMessage: nil")
		return
	}

	DefaultMessageMeter.Check(string(this.Content))

	fmt.Println("ChatroomMessage from", this.Sender, "room", this.RoomID, "id", this.ID, "/", this.MaxID, string(this.Content), "total member cnt", this.TotalMemberCnt)
	fmt.Println()
}

func (this *MemberChangedNotifcation) dump() {

	if this == nil {
		fmt.Println("MemberChanged: nil")
		return
	}

	var members string = ""
	for _, m := range this.Members {
		members += m + ","
	}

	members = strings.Trim(members, ",")

	fmt.Println("MemberChanged", members, this.Action, this.RoomID, "total member cnt", this.TotalMemberCnt)
}

func (this *JoinResp) dump() {

	if this == nil {
		fmt.Println("JoinResp: nil")
		return
	}

	var members string = ""
	for _, m := range this.Members {
		members += m + ","
	}

	members = strings.Trim(members, ",")

	fmt.Println("JoinResp members", members, "room", this.RoomID, "result", this.Result, "total member cnt", this.TotalMemberCnt)
}

func (this *QueryResp) dump() {

	if this == nil {
		fmt.Println("QueryResp: nil")
		return
	}

	var members string = ""
	for _, m := range this.Members {
		members += m + ","
	}

	members = strings.Trim(members, ",")

	fmt.Println("QueryResp members", members, "room", this.RoomID, "result", this.Result, this.RoomID, "total member cnt", this.TotalMemberCnt)
}

func (this *QuitResp) dump() {

	if this == nil {
		fmt.Println("QuitResp: nil")
		return
	}

	fmt.Println("QuitResp sn", this.Sn, "result", this.Result)
}
