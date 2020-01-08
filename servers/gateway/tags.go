/**
 * 用来保存每一个标签下的所有用户
 */
package main

import (
	"sync"
	"time"

	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/cpool"
)

func TagOperationHandler(tag string, pool *TagPool) func(interface{}) {
	return func(data interface{}) {
		gwr, ok := data.(*router.GwResp)
		if !ok {
			Logger.Error("", "", "", "TagOperationHandler", "Consumer error: type not match", data)
			return
		}
		if pool == nil {
			return
		}
		traceid := ""
		if gwr.XimpBuff != nil {
			traceid = gwr.XimpBuff.TraceId
		}

		count := make(map[logic.ConnectionId]bool)
		fake := 0
		begin := time.Now()
		pool.Lock()
		for k, conn := range pool.Conns {
			if conn == nil || conn.IsClose() {
				// 为了防止无效引用引起内存泄漏
				delete(pool.Conns, k)
				Logger.Warn(k, tag, traceid, "TagOperationHandler", "close connection in tagpool")
				fake += 1
			} else {
				if ok, _ := count[k]; ok {
					continue
				}
				conn.Operate(gwr, true)
				count[k] = true
			}
		}
		pool.Unlock()
		if len(count) != 0 || fake != 0 {
			cost := time.Now().Sub(begin)
			if cost > time.Millisecond {
				Logger.Trace("", tag, traceid, "TagOperationHandler", len(count), fake, cost)
			}
			// 这块检查消费这个tag是否有延迟
			countGatewayOperation(gwr.XimpBuff.TimeStamp, tag, "", traceid, "TagOperationHandler")()
		}
	}
}

type TagPool struct {
	sync.RWMutex
	Conns     map[logic.ConnectionId]IConnection
	CP        *cpool.ConsumerPool
	AliveFlag bool
}

func (tp *TagPool) Add(conn IConnection) {
	tp.Lock()
	tp.Conns[conn.GetId()] = conn
	tp.AliveFlag = true
	tp.Unlock()
}

type TagPools struct {
	Tags map[string]*TagPool
	sync.RWMutex
}

func NewTagPools() *TagPools {
	return &TagPools{
		Tags: make(map[string]*TagPool),
	}
}

func NewTagPool(tag string) (tp *TagPool) {
	tp = &TagPool{
		Conns: make(map[logic.ConnectionId]IConnection),
	}
	tp.CP = cpool.NewConsumerPool(uint(netConf().TagConsumerCount), uint64(netConf().TagConsumerChanLen), TagOperationHandler(tag, tp))
	return
}

func (tp *TagPool) PushOperation(gwr *router.GwResp) bool {
	return tp.CP.Add(gwr)
}

func cleanTagJob(tag string, tp *TagPool, tps *TagPools) {
	t := time.NewTicker(time.Duration(netConf().TagCleanDuration) * time.Second)
	for {
		select {
		case <-t.C:
			tp.Lock()
			if !tp.AliveFlag {
				tp.Unlock()
				tps.DelTag(tag)
				t.Stop()
				tp.CP.Cancel()
				Logger.Trace(tag, "", "", "cleanTagJob", "tps.DelTag", "")
				return
			}
			if len(tp.Conns) == 0 {
				tp.AliveFlag = false
			}
			tp.Unlock()
		}
	}
}
func (tps *TagPools) GetPool(tag string) *TagPool {
	tps.Lock()
	tp, _ := tps.Tags[tag]
	tps.Unlock()
	return tp
}

func (tps *TagPools) GetAndCreatePool(tag string) *TagPool {
	tps.Lock()
	tp, ok := tps.Tags[tag]
	if !ok {
		tp = NewTagPool(tag)
		go cleanTagJob(tag, tp, tps)
		tps.Tags[tag] = tp
	}
	tps.Unlock()
	return tp
}

func (tps *TagPools) Check(tag string, connId logic.ConnectionId) bool {
	tps.RLock()
	tp := tps.Tags[tag]
	tps.RUnlock()
	if tp == nil {
		return false
	}
	tp.RLock()
	_, ok := tp.Conns[connId]
	tp.RUnlock()
	return ok
}

// 提供给admin调用，不做异步删除
func (tps *TagPools) DelTag(tag string) {
	tps.Lock()
	defer tps.Unlock()
	if _, ok := tps.Tags[tag]; ok {
		delete(tps.Tags, tag)
	}
}

func (tps *TagPools) Del(tag string, connId logic.ConnectionId) {
	tps.RLock()
	tp := tps.Tags[tag]
	tps.RUnlock()
	if tp == nil {
		return
	}
	tp.Lock()
	delete(tp.Conns, connId)
	tp.Unlock()

}

func (tps *TagPools) Stat(ignore0 bool) map[string]int {
	stat := map[string]int{}
	tmp := map[string]*TagPool{}
	tps.RLock()
	for k, v := range tps.Tags {
		tmp[k] = v
	}
	tps.RUnlock()
	for k, v := range tmp {
		v.RLock()
		if !ignore0 || len(v.Conns) != 0 {
			stat[k] = len(v.Conns)
		}
		v.RUnlock()
	}
	return stat
}

func (tps *TagPools) GetTags(connId logic.ConnectionId) []string {
	tags := []string{}
	tmp := map[string]*TagPool{}
	tps.RLock()
	for k, v := range tps.Tags {
		tmp[k] = v
	}
	tps.RUnlock()
	for k, v := range tmp {
		v.RLock()
		if _, ok := v.Conns[connId]; ok {
			tags = append(tags, k)
		}
		v.RUnlock()
	}
	return tags
}

func (tp *TagPool) DelCloseConn() []logic.ConnectionId {
	result := []logic.ConnectionId{}
	tp.Lock()
	for k, v := range tp.Conns {
		if v.IsClose() {
			result = append(result, k)
			delete(tp.Conns, k)
		}
	}
	tp.Unlock()
	return result
}

func (tp *TagPool) GetCloseConn() []logic.ConnectionId {
	result := []logic.ConnectionId{}
	tp.RLock()
	for k, v := range tp.Conns {
		if v.IsClose() {
			result = append(result, k)
		}
	}
	tp.RUnlock()
	return result
}

func (tp *TagPool) GetAllConn() []logic.ConnectionId {
	result := []logic.ConnectionId{}
	tp.RLock()
	for k, _ := range tp.Conns {
		result = append(result, k)
	}
	tp.RUnlock()
	return result
}
