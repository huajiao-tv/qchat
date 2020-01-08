package main

import (
	"github.com/huajiao-tv/qchat/logic"
)

func benchMix(gn, n int) {
	/*
		OpenSession(userId string, appId uint16, isLoginUser bool, sessionKey string,
			connectionId logic.ConnectionId, gatewayAddr string, property map[string]string)
	*/

	logic.DynamicConf().LocalSessionRpc = sessaddr

}
