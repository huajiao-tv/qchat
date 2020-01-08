package client

import (
	"fmt"

	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/huajiao-tv/qchat/logic/pb"
)

const (
	GroupServiceID = 10000001
)

type GroupService struct {
	c *UserConnection // long live connection
}

func (gs *GroupService) GetMsg(groupid string, startid uint64, offset int32) bool {
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

	data, err := proto.Marshal(p)

	if err != nil {
		gs.c.log("error", "send getMsg request", "proto.Marshal", err)
		return false
	}

	//fmt.Println("get group msg :", groupid, startid, offset)
	return gs.c.RequestService(GroupServiceID, data)
}

func (gs *GroupService) GetMsgBatch(ids []int64) bool {
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

	if nil == gs.c {
		gs.c.log("error", "GetMsgBatch failed, conn is nil")
		return false
	}

	data, err := proto.Marshal(p)
	if nil != err {
		gs.c.log("error", "GetMsgBatch failed, proto marshal error:", err)
		return false
	}

	gs.c.log("GetMsgBatch:", ids)
	return gs.c.RequestService(GroupServiceID, data)
}

func (gs *GroupService) Sync() bool {
	p := &pb.GroupUpPacket{
		Payload: proto.Uint32(pb.GC_PAYLOAD_SYNC),
		Syncreq: []*pb.GroupSyncReq{},
	}

	data, err := proto.Marshal(p)

	if err != nil {
		gs.c.log("error", "send sync request", "proto.Marshal", err)
		return false
	}

	return gs.c.RequestService(GroupServiceID, data)
}

func (gs *GroupService) HandleServiceMessage(data []byte, sn uint64, source int) bool {

	p := &pb.GroupDownPacket{}
	if err := proto.Unmarshal(data, p); err != nil {
		gs.c.log("error", "decode group message notify", err)
		return false
	}
	if *p.Payload == uint32(pb.GC_PAYLOAD_MSG_REQ_RESP) {
		gs.c.log("receive get_group_msg resp:", p.String())
		return true
	} else if *p.Payload == uint32(pb.GC_PAYLOAD_SYNC) {
		gs.c.log("receive sync resp:", p.String())
		return true
	}

	if p.Newmsgnotify == nil {
		gs.c.log("error", "receive group new message notify but len is zero!")
		return false
	}
	for _, n := range p.Newmsgnotify {
		snStr := fmt.Sprintf("sn:%v", sn)
		groupIdStr := fmt.Sprintf("receive notify:groupid:%v", *n.Groupid)
		msgIdStr := fmt.Sprintf("msgid:%v", *n.Msgid)
		gs.c.log(snStr, groupIdStr, msgIdStr)
		gs.GetMsg(*n.Groupid, *n.Msgid, -4)
	}
	return true
}

func (gs *GroupService) HandleGetMessage(msg *UserMessage) bool {
	return true
}
