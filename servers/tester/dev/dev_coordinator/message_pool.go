package main

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/huajiao-tv/qchat/client/gateway"
	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/pb"
)

type ChatRoomMessagePool struct {
	roomId string
	appId  uint16

	timeout  uint32
	poolSize uint32

	messages []*logic.ChatRoomMessageNotify
	l        sync.Mutex
}

func NewChatRoomMessagePool(room string, appid uint16, interval, size int) *ChatRoomMessagePool {
	pool := &ChatRoomMessagePool{
		roomId:   room,
		appId:    appid,
		timeout:  uint32(interval),
		poolSize: uint32(size),
		messages: make([]*logic.ChatRoomMessageNotify, 0, size),
	}
	go pool.MessageFunc(interval)
	return pool
}

func (p *ChatRoomMessagePool) SetTimeout(interval int) {
	atomic.StoreUint32(&p.timeout, uint32(interval))
}

func (p *ChatRoomMessagePool) SetPoolSize(size int) {
	atomic.StoreUint32(&p.poolSize, uint32(size))
}

func (p *ChatRoomMessagePool) Add(msg *logic.ChatRoomMessageNotify) {
	var cache []*logic.ChatRoomMessageNotify
	size := int(atomic.LoadUint32(&p.poolSize))

	p.l.Lock()
	p.messages = append(p.messages, msg)
	if len(p.messages) >= size {
		cache = p.messages
		p.messages = make([]*logic.ChatRoomMessageNotify, 0, size)
	}
	p.l.Unlock()
	if cache != nil && len(cache) > 0 {
		go SendChatRoomMessageBatch(p.roomId, p.appId, cache)
	}
}

func (p *ChatRoomMessagePool) MessageFunc(interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)

	for {
		<-ticker.C

		size := int(atomic.LoadUint32(&p.timeout))
		p.l.Lock()
		if len(p.messages) == 0 {
			p.l.Unlock()
			continue
		}
		msgs := p.messages
		p.messages = make([]*logic.ChatRoomMessageNotify, 0, size)
		p.l.Unlock()
		go SendChatRoomMessageBatch(p.roomId, p.appId, msgs)

		if new := int(atomic.LoadUint32(&p.timeout)); new != interval {
			interval = new
			ticker.Stop()
			ticker = time.NewTicker(time.Duration(interval) * time.Millisecond)
		}
	}
}

func SendChatRoomMessageBatch(room string, appid uint16, msgs []*logic.ChatRoomMessageNotify) {
	fmt.Println("send chatroom message batch", room, appid, len(msgs))
	notify := make([]*pb.ChatRoomMNotify, 0, len(msgs))
	for _, m := range msgs {
		data, err := pb.CompressChatRoomNewMsg(m.ChatRoomMessage)
		if err != nil {
			fmt.Println("compress error", err)
			return
		}
		msg := &pb.ChatRoomMNotify{
			Type:        proto.Int32(pb.CR_PAYLOAD_INCOMING_MSG),
			Data:        data,
			Regmemcount: proto.Int32(int32(m.RegMemCount)),
			Memcount:    proto.Int32(int32(m.MemCount)),
		}
		notify = append(notify, msg)
	}
	packet := &pb.ChatRoomPacket{
		Roomid: []byte(room),
		Appid:  proto.Uint32(uint32(appid)),
		ToUserData: &pb.ChatRoomDownToUser{
			Result:      proto.Int32(0),
			Payloadtype: proto.Uint32(pb.CR_PAYLOAD_COMPRESSED),
			Multinotify: notify,
		},
	}
	content, err := proto.Marshal(packet)
	if err != nil {
		fmt.Println("marshal pb err", err.Error())
		return
	}
	//@todo
	ximp, err := pb.CreateMsgNotify("chatroom", content, int64(0), "", "", 0)
	if err != nil {
		fmt.Println("create msg notify err", err.Error())
		return
	}
	ximp.TimeStamp = time.Now().UnixNano() / 1e6
	//ximp.TraceId = m.TraceId

	tag := logic.GenerateChatRoomTag(appid, room)
	gwr := &router.GwResp{
		XimpBuff: ximp,
	}
	for _, gwy := range strings.Split(gateways, ",") {
		gateway.DoOperations(gwy, []string{tag}, "", nil, gwr)
	}
}
