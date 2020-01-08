package main

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/huajiao-tv/qchat/utility/msgRedis"
)

type RedisStats struct {
	property *stat
	gateway  *stat
	member   *stat

	pool *msgRedis.MultiPool
}

type stat struct {
	sync.RWMutex
	address string

	// clients
	connection int
	blocked    int

	// memory
	memory string

	// stats
	ops int
}

func (s *stat) info(p *msgRedis.MultiPool) error {
	ret, err := p.CallWithTimeout(s.address, 5e8, 5e8).Info()
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(ret)
	s.Lock()
	for {
		line, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			s.Unlock()
			return err
		}
		line = strings.TrimRight(line, "\r\n")

		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) != 2 {
			continue
		}
		switch fields[0] {
		case "connected_clients":
			s.connection, _ = strconv.Atoi(fields[1])
		case "blocked_clients":
			s.blocked, _ = strconv.Atoi(fields[1])
		case "used_memory_human":
			s.memory = fields[1]
		case "instantaneous_ops_per_sec":
			s.ops, err = strconv.Atoi(fields[1])
		}
	}
	s.Unlock()
	return nil
}

func (s *stat) Stat(p *msgRedis.MultiPool) {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			s.info(p)
		}
	}
}

func NewStat(prop string, mem string, gwy string) *RedisStats {
	r := &RedisStats{
		property: &stat{address: prop},
		gateway:  &stat{address: mem},
		member:   &stat{address: gwy},
		pool: msgRedis.NewMultiPool(
			[]string{
				prop, mem, gwy,
			},
			5, 1, 5,
		),
	}
	go r.property.Stat(r.pool)
	go r.gateway.Stat(r.pool)
	go r.member.Stat(r.pool)
	return r
}

func (r *RedisStats) Print() {
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:

			r.property.RLock()
			p := fmt.Sprintf("| property: %d(ops) %d(conn) %d(block) %s(mem)|", r.property.ops, r.property.connection, r.property.blocked, r.property.memory)
			r.property.RUnlock()
			r.gateway.RLock()
			g := fmt.Sprintf("gateways: %d(ops) %d(conn) %d(block) %s(mem)|", r.gateway.ops, r.gateway.connection, r.gateway.blocked, r.gateway.memory)
			r.gateway.RUnlock()
			r.member.RLock()
			m := fmt.Sprintf("members: %d(ops) %d(conn) %d(block) %s(mem)|", r.member.ops, r.member.connection, r.member.blocked, r.member.memory)
			r.member.RUnlock()

			fmt.Println(time.Now().Format("15:04:05.000"), p, g, m)
		}
	}
}
