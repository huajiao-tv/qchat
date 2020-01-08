package exqueue

import (
	"fmt"
	"sync"
	"time"

	"github.com/huajiao-tv/qchat/logic"
)

const (
	DEFAULT_EX_QUEUE_LEN      = 1000 // 队列长度
	DEFAULT_EX_QUEUE_DURATION = 100  // 空间等待的毫秒
	DEFAULT_EX_QUEUE_MERGE    = 10   // 多少个取出来一起处理
	DEFAULT_EX_QUEUE_SLOT     = 10   // 需要开多少个同样的队列，减少锁粒度
)

// 由多个slot(exqueueinter）组成的数组
type ExQueue struct {
	sync.RWMutex
	eqis   []*ExQueueInter
	length int // 每一个slot的chan长度
	fn     func(map[interface{}][]interface{})
}

// 真正的队列类
type ExQueueInter struct {
	sync.RWMutex
	queue    chan interface{}              // 用来保存进来的key
	values   map[interface{}][]interface{} // 用来保存值
	merge    int                           // 合并数量
	quit     chan bool                     // 退出标志
	duration time.Duration                 // 时间间隔
	ops      uint64                        // 操作次数，用来做debug
	add      uint64                        // 操作次数，用来做debug
}

// 关闭这个队列
func (eqi *ExQueueInter) Quit() {
	select {
	case <-eqi.quit:
	default:
		close(eqi.quit)
	}
}

func (eq *ExQueue) UpdateConfig(l int, d time.Duration, slot int, merge int) {
	if slot == 0 {
		slot = DEFAULT_EX_QUEUE_SLOT
	}
	if l == 0 {
		l = DEFAULT_EX_QUEUE_LEN
	}
	if d == 0 {
		d = time.Duration(DEFAULT_EX_QUEUE_DURATION) * time.Millisecond
	}
	if merge == 0 {
		merge = DEFAULT_EX_QUEUE_MERGE
	}
	eq.Lock()
	if eq.length != l {
		// 删除所有slot重新创建
		for _, v := range eq.eqis {
			v.Quit()
		}
		eq.eqis = make([]*ExQueueInter, 0, slot)
	}
	slotLength := len(eq.eqis)
	if slot != slotLength {
		if slot < slotLength {
			for _, v := range eq.eqis[slot:] {
				v.Quit()
			}
			eq.eqis = eq.eqis[0:slot]
		} else {
			for i := 0; i < slot-slotLength; i++ {
				eq.eqis = append(eq.eqis, NewExQueueInter(l, d, merge, eq.fn))
			}
		}
	}
	for _, v := range eq.eqis {
		v.UpdateConfig(d, merge)
	}
	eq.length = l
	eq.Unlock()
}

func (eqi *ExQueueInter) UpdateConfig(d time.Duration, merge int) {
	eqi.Lock()
	eqi.duration = d
	eqi.merge = merge
	eqi.Unlock()
}

func (eq *ExQueue) Stat() []map[string]uint64 {
	result := []map[string]uint64{}
	eq.RLock()
	length := len(eq.eqis)
	for _, s := range eq.eqis {
		s.RLock()
		result = append(result, map[string]uint64{
			"merge":     uint64(s.merge),
			"duration":  uint64(s.duration.Nanoseconds() / 1e6),
			"queuelen":  uint64(len(s.queue)),
			"queuecap":  uint64(cap(s.queue)),
			"valuelen":  uint64(len(s.values)),
			"ops":       s.ops,
			"add":       s.add,
			"alllength": uint64(length),
		})
		s.RUnlock()
	}
	eq.RUnlock()
	return result
}

/**
 * l表示队列的长度，如果队列满了，那么新的add将返回失败
 * d表示时间间隔，如果在一次获取中无法获取到merge的数量时，就sleep这个时间
 * slot 内部开多个队列来降低锁的粒度
 * merge 尽量取到这个数量的数据后一并处理，该参数针对于 slot 设置，是单个 slot 的 merge 数量
 * fn 处理完成后的回调函数，参数是map[key][]value
 */
func NewExQueue(l int, d time.Duration, slot int, merge int, fn func(map[interface{}][]interface{})) *ExQueue {
	if slot == 0 {
		slot = DEFAULT_EX_QUEUE_SLOT
	}
	if l == 0 {
		l = DEFAULT_EX_QUEUE_LEN
	}
	if d == 0 {
		d = time.Duration(DEFAULT_EX_QUEUE_DURATION) * time.Millisecond
	}
	if merge == 0 {
		merge = DEFAULT_EX_QUEUE_MERGE
	}
	eq := &ExQueue{
		length: l,
		fn:     fn,
		eqis:   make([]*ExQueueInter, 0, slot),
	}
	for i := 0; i < slot; i++ {
		eq.eqis = append(eq.eqis, NewExQueueInter(l, d, merge, fn))
	}
	return eq

}
func NewExQueueInter(l int, d time.Duration, merge int, fn func(map[interface{}][]interface{})) *ExQueueInter {
	eqi := &ExQueueInter{
		queue:    make(chan interface{}, l),
		values:   make(map[interface{}][]interface{}),
		merge:    merge,
		quit:     make(chan bool),
		duration: d,
	}
	go func() {
		localMerge := merge
		for {
			keys := []interface{}{}
			needSleep := false
			select {
			case key := <-eqi.queue:
				keys = append(keys, key)
			case <-eqi.quit:
				// 如果已经关闭这个slot，等这里面所有的通知都处理完之后再退出
				select {
				case key := <-eqi.queue:
					keys = append(keys, key)
				default:
					return
				}
			}
		mergeLabel:
			for i := 0; i < localMerge-1; i++ {
				select {
				case key := <-eqi.queue:
					keys = append(keys, key)
				default:
					needSleep = true
					break mergeLabel
				}
			}
			eqi.Lock()
			values := make(map[interface{}][]interface{}, len(keys))
			for _, key := range keys {
				if v, ok := eqi.values[key]; ok {
					if len(v) != 0 {
						values[key] = v
					}
					delete(eqi.values, key)
				}
			}
			localMerge = eqi.merge
			localD := eqi.duration
			eqi.ops += uint64(len(values))
			eqi.Unlock()
			fn(values)
			if needSleep {
				time.Sleep(localD)
			}
		}
	}()
	return eqi
}

// 添加元素
func (eq *ExQueue) Add(key interface{}, value ...interface{}) (ok bool) {
	eq.RLock()
	stringKey := fmt.Sprint("%v", key)
	eqi := eq.eqis[logic.Sum(stringKey)%len(eq.eqis)]
	eq.RUnlock()
	select {
	case <-eqi.quit:
		return false
	default:
	}
	eqi.Lock()
	defer eqi.Unlock()

	values, ok := eqi.values[key]
	if !ok {
		select {
		case eqi.queue <- key:
		default:
			// 如果队列满的时候最好的做法是通知不发送了
			return false
		}
		values = value
		eqi.values[key] = values
	} else {
		eqi.values[key] = append(eqi.values[key], value...)
	}
	eqi.add += 1
	return true
}
