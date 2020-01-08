package llconn

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/huajiao-tv/qchat/logic/pb"
)

const (
	GroupServiceID = 10000001
)

type GroupService struct {
	llc *LongLiveConn // long live connection
}

func NewGroup() *GroupService {
	group := GroupService{}
	return &group
}

func (gs *GroupService) GetMsgM(ids []int64) {
	p := &pb.GroupUpPacket{
		Payload: proto.Uint32(pb.GC_PAYLOAD_MSG_REQ_RESP),
	}
	for i := 0; i <= len(ids)-3; i += 3 {
		p.Getmsgreq = append(p.Getmsgreq, &pb.GroupMessageReq{
			Groupid: proto.String(strconv.Itoa(int(ids[i]))),
			Startid: proto.Uint64(uint64(ids[i+1])),
			Offset:  proto.Int32(int32(ids[i+2])),
			Traceid: proto.String(strconv.Itoa(i)),
		})
	}
	if nil == gs.llc {
		fmt.Println("getMsg:sendMessage: conn is nil")
		return
	}

	data, err := proto.Marshal(p)

	if err != nil {
		fmt.Println("getMsg:sendMessage: proto.Marshal failed", err)
		return
	}

	fmt.Println("get group msg :", ids)
	gs.llc.SendServiceMessage(GroupServiceID, data)
}

func (gs *GroupService) GetMsg(groupid string, startid uint64, offset int32) {
	p := &pb.GroupUpPacket{
		Payload: proto.Uint32(pb.GC_PAYLOAD_MSG_REQ_RESP),
		Getmsgreq: []*pb.GroupMessageReq{
			&pb.GroupMessageReq{
				Groupid: &groupid,
				Startid: &startid,
				Offset:  &offset,
				Traceid: proto.String(strconv.Itoa(1)),
			},
		},
	}
	if nil == gs.llc {
		fmt.Println("getMsg:sendMessage: conn is nil")
		return
	}

	data, err := proto.Marshal(p)

	if err != nil {
		fmt.Println("getMsg:sendMessage: proto.Marshal failed", err)
		return
	}

	fmt.Println("get group msg :", groupid, startid, offset)
	gs.llc.SendServiceMessage(GroupServiceID, data)
}

func (gs *GroupService) Sync(groups []string) {
	p := &pb.GroupUpPacket{
		Payload: proto.Uint32(pb.GC_PAYLOAD_SYNC),
		Syncreq: []*pb.GroupSyncReq{},
	}
	if len(groups) > 0 {
		for _, g := range groups {
			p.Syncreq = append(p.Syncreq, &pb.GroupSyncReq{Groupid: proto.String(g)})
		}
	}
	if nil == gs.llc {
		fmt.Println("sync: conn is nil")
		return
	}

	data, err := proto.Marshal(p)

	if err != nil {
		fmt.Println("sync: proto.Marshal failed", err)
		return
	}

	fmt.Println("sync....:")
	gs.llc.SendServiceMessage(GroupServiceID, data)
}

func (gs *GroupService) HandleServiceMessage(data []byte, sn uint64, source int) bool {

	p := &pb.GroupDownPacket{}
	if err := proto.Unmarshal(data, p); err != nil {
		fmt.Println("decode group new message notify error:", err)
		return false
	}
	if *p.Payload == uint32(pb.GC_PAYLOAD_MSG_REQ_RESP) {
		fmt.Println("receive get group resp:", p.String())
		return true
	} else if *p.Payload == uint32(pb.GC_PAYLOAD_SYNC) {
		fmt.Println("receive sync :", p.String())
		return true
	}

	if p.Newmsgnotify == nil {
		fmt.Println("receive group new message notify but len is zero!")
		return false
	}
	for _, n := range p.Newmsgnotify {
		fmt.Println("sn:", sn, ",receive notify:groupid:", *n.Groupid, ",msgid:", *n.Msgid)
		gs.GetMsg(*n.Groupid, *n.Msgid, -4)
	}
	return true
}

func (gs *GroupService) Register(conn *LongLiveConn) (int, error) {
	if nil != conn {
		gs.llc = conn
		return GroupServiceID, nil
	} else {
		return GroupServiceID, errors.New("LongLiveConn is nil")
	}
}

func (this *GroupService) HandleGetMessage(msg *UserMessage) bool {
	return true
}

//
// get chatroom message
//
func (this *GroupService) GetMessage(roomid string, ids []int64) bool {
	return true
}
