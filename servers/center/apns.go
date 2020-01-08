package main

import (
	"time"

	"github.com/huajiao-tv/qchat/client/apns"
	"github.com/huajiao-tv/qchat/utility/cpool"
)

var apnsPool *cpool.ConsumerPool

func init() {
	apnsPool = cpool.NewConsumerPool(10, 10000, ApnsSendMsg)
}

func ApnsSendMsg(d interface{}) {
	m, ok := d.([]*apns.GroupChatRpcMsg)
	if !ok {
		Logger.Error("", "", "", "ApnsSendMsg", "ConsumerPool error :type not match", d)
		return
	}
	start := time.Now()
	if err := apns.PushMsg(m); err != nil {
		Logger.Error("", "", "", "ApnsSendMsg", "apns.PushMsg error", err)
		return
	}
	Logger.Debug("", "", "", "ApnsSendMsg", time.Now().Sub(start), len(m))

}
