package lru

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"github.com/huajiao-tv/qchat/utility/util"
)

const (
	DEFAULT_CACHE_SLOT_NUM = 10
	DEFAULT_CACHE_CAP      = 1000 // 容量
	DEFAULT_CACHE_EXPIRE   = 3600 // 毫秒
)

type Cache struct {
	sync.RWMutex
	slots []*Slot
}

func NewCache(slotNum int, cap int, expire time.Duration) *Cache {
	if slotNum <= 0 {
		slotNum = DEFAULT_CACHE_SLOT_NUM
	}
	if cap <= 0 {
		cap = DEFAULT_CACHE_CAP
	}
	if expire < 0 {
		expire = DEFAULT_CACHE_EXPIRE * time.Second
	}

	c := &Cache{
		slots: make([]*Slot, 0, slotNum),
	}
	c.Lock()
	for i := 0; i < slotNum; i++ {
		c.slots = append(c.slots, NewSlot(cap, expire))
	}
	c.Unlock()
	return c
}

func (c *Cache) Stat() []map[string]uint64 {
	c.RLock()
	length := len(c.slots)
	result := make([]map[string]uint64, 0, length)
	for _, s := range c.slots {
		result = append(result, map[string]uint64{
			"hit":    s.hit,
			"miss":   s.miss,
			"len":    uint64(s.list.Len()),
			"alllen": uint64(length),
		})
	}
	c.RUnlock()
	return result
}

// 更新配置
func (c *Cache) UpdateConfig(slotNum, cap int, expire time.Duration) {
	if slotNum <= 0 {
		slotNum = DEFAULT_CACHE_SLOT_NUM
	}
	if cap <= 0 {
		cap = DEFAULT_CACHE_CAP
	}
	if expire < 0 {
		expire = DEFAULT_CACHE_EXPIRE * time.Second
	}
	c.Lock()
	length := len(c.slots)
	if length != slotNum {
		if length > slotNum {
			c.slots = c.slots[0:slotNum]
		} else if length < slotNum {
			for i := 0; i < slotNum-length; i++ {
				c.slots = append(c.slots, NewSlot(cap, expire))
			}
		}
		for _, s := range c.slots {
			s.UpdateConfig(cap, expire)
			s.Clean()
		}
	} else {
		for _, s := range c.slots {
			s.UpdateConfig(cap, expire)
		}
	}
	c.Unlock()
}

func (c *Cache) getSlot(key interface{}) *Slot {
	stringKey := fmt.Sprint("%v", key)
	c.RLock()
	s := c.slots[util.Sum(stringKey)%len(c.slots)]
	c.RUnlock()
	return s
}

func (c *Cache) Clean() {
	c.RLock()
	tmp := make([]*Slot, 0, len(c.slots))
	for _, s := range c.slots {
		tmp = append(tmp, s)
	}

	c.RUnlock()
	for _, s := range tmp {
		s.Clean()
	}
}

func (c *Cache) Add(key interface{}, value interface{}) {
	c.getSlot(key).Add(key, value)
}

func (c *Cache) Del(key interface{}) {
	c.getSlot(key).Del(key)
}
func (c *Cache) Get(key interface{}) (value interface{}, exist bool) {
	return c.getSlot(key).Get(key)
}

type Slot struct {
	sync.RWMutex
	capacity int
	list     *list.List
	cache    map[interface{}]*list.Element
	expire   time.Duration

	hit  uint64
	miss uint64
}

type Node struct {
	Key, Value interface{}
	Time       time.Time
}

func NewSlot(cap int, expire time.Duration) *Slot {
	return &Slot{
		capacity: cap,
		list:     list.New(),
		cache:    make(map[interface{}]*list.Element),
		expire:   expire,
	}
}

func (s *Slot) UpdateConfig(cap int, expire time.Duration) {
	s.Lock()
	s.capacity = cap
	s.expire = expire
	s.Unlock()
}

func (s *Slot) Clean() {
	s.Lock()
	s.list = list.New()
	s.cache = make(map[interface{}]*list.Element)
	s.Unlock()
}

func (s *Slot) Add(key interface{}, value interface{}) {
	s.Lock()
	defer s.Unlock()
	node := &Node{key, value, time.Now()}
	if ele, ok := s.cache[key]; ok {
		s.list.MoveToFront(ele)
		ele.Value.(*Node).Value = value
		ele.Value.(*Node).Time = time.Now()
		return
	}

	ele := s.list.PushFront(node)
	s.cache[key] = ele
	if s.list.Len() > s.capacity {
		ele := s.list.Back()
		s.list.Remove(ele)
		delete(s.cache, ele.Value.(*Node).Key)
	}
}

func (s *Slot) RemoveLast() {
	s.Lock()
	defer s.Unlock()
	ele := s.list.Back()
	if ele != nil {
		s.list.Remove(ele)
		delete(s.cache, ele.Value.(*Node).Key)
	}
}

func (s *Slot) Get(key interface{}) (value interface{}, exist bool) {
	s.Lock()
	defer s.Unlock()
	if ele, ok := s.cache[key]; ok {
		if s.expire != 0 && time.Now().Sub(ele.Value.(*Node).Time) > s.expire {
			delete(s.cache, ele.Value.(*Node).Key)
			s.list.Remove(ele)
			s.miss += 1
			return nil, false
		}
		s.list.MoveToFront(ele)
		s.hit += 1
		return ele.Value.(*Node).Value, true
	} else {
		s.miss += 1
		return nil, false
	}
}

func (s *Slot) Del(key interface{}) (exist bool) {
	s.Lock()
	if ele, ok := s.cache[key]; ok {
		delete(s.cache, key)
		s.list.Remove(ele)
		exist = true
	} else {
		exist = false
	}
	s.Unlock()
	return
}

func (s *Slot) Len() int {
	s.RLock()
	defer s.RUnlock()
	return s.list.Len()
}
