package main

import (
	"bytes"
	"encoding/gob"

	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/cpool"
)

var cacheChatroomMessagePool *cpool.ConsumerPool

func initCron() error {
	cacheChatroomMessageConsumerCount := uint(netConf().CacheChatroomMessageConsumerCount)
	cacheChatroomMessageConsumerChanlen := uint64(netConf().CacheChatroomMessageConsumerChanlen)
	if cacheChatroomMessageConsumerCount <= 0 || cacheChatroomMessageConsumerChanlen <= 0 {
		cacheChatroomMessageConsumerCount = 256
		cacheChatroomMessageConsumerChanlen = 100000
	}
	cacheChatroomMessagePool = cpool.NewConsumerPool(cacheChatroomMessageConsumerCount, cacheChatroomMessageConsumerChanlen, CacheChatRoomMessageFn)
	Logger.Debug("", "", "", "initCron", "cacheChatroomMessageConsumerCount", cacheChatroomMessageConsumerCount, "cacheChatroomMessageConsumerChanlen", cacheChatroomMessageConsumerChanlen)
	return nil
}

func addCacheChatRoomMessage(req *logic.ChatRoomMessage) {
	if ok := cacheChatroomMessagePool.Add(req); !ok {
		Logger.Warn(req.Sender, req.Appid, req.RoomID, "addCacheChatRoomMessage", "chan full", "")
	}
}

func CacheChatRoomMessageFn(d interface{}) {
	req, ok := d.(*logic.ChatRoomMessage)
	if !ok {
		Logger.Error("", "", "", "cacheChatRoomMessageFn", "Consumer error:type not match", d)
		return
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(*req); err != nil {
		Logger.Error(req.Sender, req.Appid, req.RoomID, req.MsgID, "CacheChatRoomMessageCron", "gob.NewEncoder error", err.Error())
	} else if err := cacheChatRoomMessage(req.RoomID, req.Appid, req.MsgID, buf.Bytes()); err != nil {
		Logger.Error(req.Sender, req.Appid, req.RoomID, req.MsgID, "CacheChatRoomMessageCron", "cacheChatRoomMessage error", err.Error())
	}
}
