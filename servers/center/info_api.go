package main

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/huajiao-tv/qchat/client/gateway"
	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
)

func unionNode(f func(string) interface{}, addrs []string) map[string]interface{} {
	var wg sync.WaitGroup
	var resultLock sync.Mutex
	result := map[string]interface{}{}
	if len(addrs) > 0 {
		for _, g := range addrs {
			wg.Add(1)
			go func(addr string) {
				defer wg.Done()
				t := f(addr)
				resultLock.Lock()
				result[addr] = t
				resultLock.Unlock()
			}(g)
		}
		wg.Wait()
	}
	return result
}

func GatewayConnlenHandler(w http.ResponseWriter, req *http.Request) {
	resp := errorResponse{errorSuccess, "", nil}
	var allLock sync.Mutex
	all := 0
	result := unionNode(func(g string) interface{} {
		if count, err := gateway.ConnLen(g); err != nil {
			Logger.Error("", "", "", "GatewayConnlenHandler", "gateway.ConnLen error", err)
			return interface{}(0)
		} else {
			allLock.Lock()
			all += count
			allLock.Unlock()
			return interface{}(count)
		}
	}, logic.NetGlobalConf().GatewayRpcs)
	result["all"] = interface{}(all)
	resp.Data = result
	w.Write([]byte(resp.Error()))
}

func GatewayTagstatHandler(w http.ResponseWriter, req *http.Request) {
	resp := errorResponse{errorSuccess, "", nil}
	var allLock sync.Mutex
	all := map[string]int{}
	result := unionNode(func(g string) interface{} {
		if stat, err := gateway.TagStat(g); err != nil {
			Logger.Error("", "", "", "GatewayTagstatHandler", "gateway.TagStat error", err)
			return nil
		} else {
			allLock.Lock()
			for k, v := range stat {
				if v == 0 {
					delete(stat, k)
					continue
				}
				all[k] = all[k] + v
			}
			allLock.Unlock()
			return interface{}(stat)
		}
	}, logic.NetGlobalConf().GatewayRpcs)
	result["all"] = interface{}(all)
	resp.Data = result
	w.Write([]byte(resp.Error()))
}

func UserInfoHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	if err := req.ParseForm(); err != nil {
		res.Code, res.Reason = errorInvalidRequest, errorReasonInvalidRequest
	} else if appid, err := strconv.Atoi(req.FormValue("appid")); err != nil {
		res.Code, res.Reason = errorInvalidRequest, err.Error()
	} else {
		if appid == 0 {
			appid = logic.DEFAULT_APPID
		}
		querySession := []*session.UserSession{
			&session.UserSession{
				UserId: req.FormValue("uid"),
				AppId:  uint16(appid),
			},
		}
		resp, err := session.Query(querySession)
		if err != nil {
			res.Code, res.Reason = errorInternalError, err.Error()
			Logger.Error(req.FormValue("uid"), appid, "", "UserInfoHandler", "session.Query error", err)
		} else {
			for _, session := range resp {
				queryUserSessionChatRooms(session)
			}
			res.Data = resp
		}
	}
	w.Write([]byte(res.Error()))
}

func OnlineLenHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	if err := req.ParseForm(); err != nil {
		res.Code, res.Reason = errorInvalidRequest, errorReasonInvalidRequest
	} else if appid, err := strconv.Atoi(req.FormValue("appid")); err != nil {
		res.Code, res.Reason = errorInvalidRequest, err.Error()
	} else {
		if appid == 0 {
			appid = logic.DEFAULT_APPID
		}
		resp, err := saver.GetActiveUserNum(uint16(appid))
		if err != nil {
			res.Code, res.Reason = errorInternalError, err.Error()
			Logger.Error("", appid, "", "OnlineLenHandler", "saver.GetActiveUserNum error", err)
		} else {
			res.Data = resp
		}
	}
	w.Write([]byte(res.Error()))
}

func ChatRoomLenHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	if err := req.ParseForm(); err != nil {
		res.Code, res.Reason = errorInvalidRequest, errorReasonInvalidRequest
	} else if appid, err := strconv.Atoi(req.FormValue("appid")); err != nil {
		res.Code, res.Reason = errorInvalidRequest, err.Error()
	} else {
		if appid == 0 {
			appid = logic.DEFAULT_APPID
		}
		resp, err := saver.GetActiveChatRoomNum(uint16(appid))
		if err != nil {
			res.Code, res.Reason = errorInternalError, err.Error()
			Logger.Error("", appid, "", "ChatRoomLenHandler", "saver.GetActiveChatRoomNum error", err)
		} else {
			res.Data = resp
		}
	}
	w.Write([]byte(res.Error()))
}
