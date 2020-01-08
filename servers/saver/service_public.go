package main

import (
	"fmt"
	"sync"
	"time"

	"strconv"

	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/logic"
)

var (
	reloadPublicInterval  = 1
	maxPublicMessageCache = 500
	publicMsgRwLock       sync.RWMutex
	publicMsgCaches       = make(map[uint16]*PublicMsgCache)
)

type PublicMsgCache struct {
	messages map[uint64]*saver.ChatMessage
	maxMsgId uint64
	minMsgId uint64
	latest   uint64
}

func reloadPublicMsgCacheHandler() {
	if netConf().ReloadPublicInterval > 0 {
		reloadPublicInterval = netConf().ReloadPublicInterval
	}

	ticker := time.NewTicker(time.Second * time.Duration(reloadPublicInterval))
	defer func() {
		ticker.Stop()
	}()

	for {
		<-ticker.C // wait a moment

		// reload public message to cache
		// update max public message cache limit if need
		if netConf().MaxPublicCache != maxPublicMessageCache && netConf().MaxPublicCache > 0 {
			maxPublicMessageCache = netConf().MaxPublicCache
		}
		// reload public message to cache
		appids := logic.NetGlobalConf().Appids
		newCaches := make(map[uint16]*PublicMsgCache, len(appids))
		needUpdate := false
		for _, appidStr := range appids {
			appid, err := strconv.ParseUint(appidStr, 10, 16)
			if err != nil || appid == 0 {
				Logger.Error("public", appid, "", "reloadPublicMsgCacheHandler",
					fmt.Sprintf("get wrong appid, appid[%v], error[%v]", appid, err))
				continue
			}

			publicCache, err := RetrievePublicRecords(uint16(appid))
			if err != nil {
				Logger.Error("public", appid, "", "reloadPublicMsgCacheHandler",
					fmt.Sprintf("load public message of appid failed, appid[%v], error[%v]", appid, err))
				continue
			}
			newCaches[uint16(appid)] = publicCache

			if cached, ok := publicMsgCaches[uint16(appid)]; !ok {
				needUpdate = true
			} else if len(publicCache.messages) != len(cached.messages) ||
				publicCache.maxMsgId != cached.maxMsgId ||
				publicCache.minMsgId != cached.minMsgId ||
				publicCache.latest != cached.latest {
				needUpdate = true
			}
		}
		// update public message caches
		if needUpdate || len(newCaches) != len(publicMsgCaches) {
			publicMsgRwLock.Lock()
			publicMsgCaches = newCaches
			publicMsgRwLock.Unlock()
			Logger.Trace("public", "", "", "reloadPublicMsgCacheHandler", "public messages cache has been updated")
		}

		// update tick timer if config is changed
		if netConf().ReloadPublicInterval != reloadPublicInterval && netConf().ReloadPublicInterval > 0 {
			ticker.Stop() // stop old ticker explicitly
			Logger.Debug("public", "", "", "reloadPublicMsgCacheHandler",
				fmt.Sprintf("public message reload interval is changed from %v to %v.",
					reloadPublicInterval, netConf().ReloadPublicInterval))

			reloadPublicInterval = netConf().ReloadPublicInterval
			// make new time ticker
			ticker = time.NewTicker(time.Second * time.Duration(reloadPublicInterval))
		}
	}
}

/*
 * retrieve public messages from cache
 * @param appid is application id
 * @param owner is owner of messages
 * @param channelInfo includes required information to retrieve messages
 * @param traceSn is used to trace procedure
 * @param resp is response will return to rpc caller
 *
 * @return nil if operation is successful; otherwise an error is returned
 */
func RetrievePublicRecordsFromCache(appid uint16, owner string, channelInfo *saver.RetrieveChannel,
	traceSn string, resp *saver.RetrieveMessagesResponse) error {
	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countPublicResponseTime(owner, traceSn, "RetrievePublicRecordsFromCache", appid)
		defer countFunc()
	}

	publicMsgRwLock.RLock()
	caches := publicMsgCaches
	publicMsgRwLock.RUnlock()

	if cache, ok := caches[appid]; ok {
		if channelInfo.MaxCount <= 0 {
			channelInfo.MaxCount = DefaultReturnCount
		}

		maxMsgId := uint64(0)
		messages := make([]*saver.ChatMessage, 0, channelInfo.MaxCount)
		count := len(cache.messages)
		if count > channelInfo.MaxCount {
			count = channelInfo.MaxCount
		}

		if channelInfo.StartMsgId >= 0 {
			idx := uint64(channelInfo.StartMsgId)
			if idx < cache.minMsgId {
				idx = cache.minMsgId
			}
			for i := 0; idx <= cache.maxMsgId && i < count; idx++ {
				if message, ok := cache.messages[idx]; ok { // note that cache.messages type is map[uint64]*saver.ChatMessage
					messages = append(messages, message)
					i++
					maxMsgId = idx
				}
			}
		} else {
			// get specified latest records
			for idx, i := cache.maxMsgId, 0; idx >= cache.minMsgId && i < count; idx-- {
				if message, ok := cache.messages[idx]; ok {
					messages = append(messages, message)
					i++
				}
			}
			maxMsgId = cache.maxMsgId
		}

		resp.Inbox[channelInfo.Channel] = messages
		resp.LatestID[channelInfo.Channel] = cache.latest
		Logger.Trace(owner, appid, traceSn, "RetrievePublicRecordsFromCache", "Retrieved public records",
			fmt.Sprintf("start:%v,len:%v,return:%v,maxid:%v,cacechedlen:%v,cachedmaxid:%v,cachedlatestid:%v",
				channelInfo.StartMsgId, channelInfo.MaxCount, len(messages), maxMsgId,
				len(cache.messages), cache.maxMsgId, cache.latest))
	} else {
		Logger.Debug(owner, appid, traceSn, "RetrievePublicRecordsFromCache",
			fmt.Sprint("there is not any public message of appid", appid))
	}

	return nil
}
