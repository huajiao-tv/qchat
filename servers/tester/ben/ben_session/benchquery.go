package main

import (
	"fmt"
	//"github.com/johntech-o/gorpc"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
	//"bytes"
)

func benchQuerySession(gn, n int) {
	logic.DynamicConf().LocalSessionRpc = sessaddr

	for i := 0; i < gn; i++ {
		go func(goid, runnum int) {
			defer wg.Done()
			baseuid := fmt.Sprintf("%s-%d", uid, goid)
			for {
				req := []*session.UserSession{}

				user := &session.UserSession{}
				user.UserId = fmt.Sprintf("%s-%d", baseuid, runnum)
				user.AppId = uint16(appid)
				user.Platform = prop["Platform"]
				user.Deviceid = prop["Deviceid"]

				req = append(req, user)

				resp, err := session.Query(req)
				if err != nil {
					fmt.Println("Querysession err is ", err)
				} else {
					if verbose != 0 {
						for _, v := range resp {
							fmt.Printf("Opensession resp session is %#v\n", v)
						}

						fmt.Printf("Querysession resp len is %v\n", len(resp))
					}
				}
				if runnum--; runnum == 0 {
					break
				}
			}
		}(i, n)
	}
}
