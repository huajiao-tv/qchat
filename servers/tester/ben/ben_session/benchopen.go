package main

import (
	"fmt"
	//"github.com/johntech-o/gorpc"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
	//"bytes"
)

func benchOpenSession(gn, n int) {
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

				resp, err := session.Open(tuid, uint16(appid), true, "zj_sesskey", logic.ConnectionId(connid), "127.0.0.1:6220", prop)
				if err != nil {
					fmt.Println("Opensession err is ", err)
				} else {
					if verbose != 0 {
						for _, v := range resp.OldUserSessions {
							fmt.Printf("Opensession resp oldusersession is %#v\n", v)
						}

						fmt.Printf("Opensession resp.tags is %#v\n", resp.Tags)
					}
				}

				if runnum--; runnum == 0 {
					break
				}
			}
		}(i, n)
	}
}
