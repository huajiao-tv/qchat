package main

import (
	"strconv"
	"sync"
	"time"

	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/lru"
)

var onlineCache *OnlineCache

type OnlineCache struct {
	sync.RWMutex
	onlineCache map[uint16]*lru.Cache
}

func (oc *OnlineCache) Stat() map[string][]map[string]uint64 {
	result := make(map[string][]map[string]uint64)
	oc.RLock()
	for appid, c := range oc.onlineCache {
		result[strconv.Itoa(int(appid))] = c.Stat()
	}
	oc.RUnlock()
	return result
}

func (oc *OnlineCache) Get(appid uint16) *lru.Cache {
	oc.RLock()
	c, ok := oc.onlineCache[appid]
	oc.RUnlock()
	if ok {
		return c
	}
	return nil
}
func (oc *OnlineCache) Clean() {
	oc.RLock()
	for _, c := range oc.onlineCache {
		c.Clean()
	}
	oc.RUnlock()
}
func (oc *OnlineCache) UpdateConfig(slotNum, cap int, expire time.Duration) {
	oc.RLock()
	for _, c := range oc.onlineCache {
		c.UpdateConfig(slotNum, cap, expire)
	}
	oc.RUnlock()
}

func initOnlineCache() {
	slotNum := netConf().OnlineCacheSlot
	cap := netConf().OnlineCacheCap
	expire := netConf().OnlineCacheExpire
	if onlineCache == nil {
		onlineCache = &OnlineCache{
			onlineCache: make(map[uint16]*lru.Cache),
		}
	}
	onlineCache.Lock()
	for _, appid := range logic.NetGlobalConf().Appids {
		id := logic.StringToUint16(appid)
		if _, ok := onlineCache.onlineCache[id]; !ok {
			onlineCache.onlineCache[id] = lru.NewCache(slotNum, cap, time.Duration(expire)*time.Second)
		}
	}
	onlineCache.Unlock()
}

func (oc *OnlineCache) CheckOnline(appid uint16, users []string) map[string][]*logic.UserGateway {
	c := oc.Get(appid)
	if c == nil {
		return nil
	}
	result := make(map[string][]*logic.UserGateway, len(users))
	missUser := []*session.UserSession{}
	for _, u := range users {
		value, ok := c.Get(u)
		if !ok {
			missUser = append(missUser, &session.UserSession{
				UserId: u,
				AppId:  appid,
			})
			continue
		}
		ugs, ok := value.([]*logic.UserGateway)
		if ok {
			result[u] = ugs
		}
	}
	if len(missUser) > 0 {
		if sessions, err := saver.QueryUserSession(missUser); err != nil {
			Logger.Error("", "", "", "CheckOnline", "saver.QueryUserSession error", err)
		} else {
			gateways := map[string][]*logic.UserGateway{}
			for _, s := range sessions {
				gateways[s.UserId] = append(gateways[s.UserId], &logic.UserGateway{s.GatewayAddr, s.ConnectionId})
			}
			for _, u := range missUser {
				if gs, ok := gateways[u.UserId]; ok {
					result[u.UserId] = gs
				} else {
					result[u.UserId] = nil
				}
				c.Add(u.UserId, result[u.UserId])
			}
		}
	}
	return result
}

func (oc *OnlineCache) DelBatch(appid uint16, users []string) {
	c := oc.Get(appid)
	if c == nil {
		return
	}
	for _, u := range users {
		c.Del(u)
	}
}

func (oc *OnlineCache) Del(appid uint16, user string) {
	c := oc.Get(appid)
	if c == nil {
		return
	}
	c.Del(user)
}

func (oc *OnlineCache) AddBatch(appid uint16, userMap map[string][]*logic.UserGateway) {
	c := oc.Get(appid)
	if c == nil {
		return
	}
	for k, v := range userMap {
		c.Add(k, v)
	}
}

func (oc *OnlineCache) Add(appid uint16, user string, gateway []*logic.UserGateway) {
	c := oc.Get(appid)
	if c == nil {
		return
	}
	c.Add(user, gateway)
}
