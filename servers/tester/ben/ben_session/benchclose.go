package main

import (
	"fmt"
	//"github.com/johntech-o/gorpc"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
	//"strconv"
	//"bytes"
)

func benchCloseSession(gn, n int) {
	/*
		OpenSession(userId string, appId uint16, isLoginUser bool, sessionKey string,
			connectionId logic.ConnectionId, gatewayAddr string, property map[string]string)
	*/

	logic.DynamicConf().LocalSessionRpc = sessaddr

	for i := 0; i < gn; i++ {
		go func(goid, runnum int) {
			defer wg.Done()
			baseuid := fmt.Sprintf("%s-%d", uid, goid)
			for {
				tuid := fmt.Sprintf("%s-%d", baseuid, runnum)
				resp, err := session.Close(tuid, uint16(appid), "127.0.0.1:6220", logic.ConnectionId(connid), prop)
				if err != nil {
					fmt.Println("Closesession err is ", err)
				} else {
					if verbose != 0 {
						fmt.Printf("Closesession resp is %#v\n", resp)
					}
				}
				if runnum--; runnum == 0 {
					break
				}
			}
		}(i, n)
	}
}
