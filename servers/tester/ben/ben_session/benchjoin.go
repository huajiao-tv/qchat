package main

import (
	"fmt"
	"strconv"

	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
)

func benchJoinChatRoom(gn, n int) {

	room := "12345678"
	logic.NetGlobalConf().SessionRpcs = []string{sessaddr}

	for i := 0; i < gn; i++ {
		go func(goid, runnum int) {
			defer wg.Done()
			baseuid := fmt.Sprintf("%s-%d", uid, goid)
			for {
				connId := goid*runnum + runnum
				tuid := fmt.Sprintf("%s-%d", baseuid, runnum)

				resp, err := session.JoinChatRoom(tuid, strconv.Itoa(appid), "127.0.0.1:6220", logic.ConnectionId((connId)), "tcp", room, map[string]string{})
				if err != nil {
					fmt.Println("JoinChatRoom err is ", err)
				} else {
					if verbose != 0 {
						fmt.Println("JoinChatRoom ret:", resp.UserChatRoomResponse)
					}
				}

				if runnum--; runnum == 0 {
					break
				}
			}
		}(i, n)
	}
}
