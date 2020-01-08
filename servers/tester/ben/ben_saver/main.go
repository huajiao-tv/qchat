package main

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/huajiao-tv/qchat/logic"
)

var (
	wg *sync.WaitGroup

	saveraddr string

	userStartId    int64  // test start id
	userCount      int64  // test user count
	runTimes       int    // run times
	channel        string // test channel
	testCase       string // test case
	expireInterval int    // expire interval
	storeOutbox    int    // indicates whether store outbox when save im message
	baseGroupId    int64
	baseUserId     int
	redisStr       string
	roomId         string

	gonum    int
	beginntf chan bool
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)
	flag.StringVar(&saveraddr, "host", "127.0.0.1:6520", "saver addr")

	flag.IntVar(&gonum, "gonum", runtime.NumCPU(), "go routine num")
	flag.Int64Var(&userStartId, "usid", 860000000, "user start id")
	flag.Int64Var(&userCount, "uc", 1000000, "user count")
	flag.IntVar(&runTimes, "rt", 1000, "run times")
	flag.IntVar(&expireInterval, "ei", 86400, "expire interval")
	flag.IntVar(&storeOutbox, "outbox", 0, "store im message to outbox too")

	flag.StringVar(&channel, "ch", "peer", "message channel, support peer/im/public")
	flag.StringVar(&testCase, "tc", "storeAndRetrieve", "test case, like store, retrieve, sat (store and then retrieve)")
	flag.StringVar(&redisStr, "redis", RedisAddrs, "monitor redis addresses")
	flag.StringVar(&roomId, "rid", "", "room id")

	flag.Parse()

	rand.Seed(time.Now().UnixNano())
}

func main() {
	switch testCase {
	case "store":
		benchmarkStoreMsg()
	case "retrieve":
		benchmarkRetrieveMsg()
	case "storeAndRetrieve":
		benchmarkStoreAndRetrieveMsg()
	case "redis":
		monitorRedis()

	}

}

const (
	chatRoomPropertyHashKey = "chatroom:property:%s:2080:hset"
	chatRoomGateWayHashKey  = "chatroom:gateway:%s:2080:hset"
	chatRoomMembersSetKey   = "chatroom:members:%s:2080:set"

	RedisAddrs = "127.0.0.1:6379:"
)

func monitorRedis() {
	if roomId == "" {
		fmt.Println("no roomid")
		return
	}
	addresses := strings.Split(redisStr, ",")
	gateways := fmt.Sprintf(chatRoomGateWayHashKey, roomId)
	members := fmt.Sprintf(chatRoomMembersSetKey, roomId)
	fmt.Println("room property", addresses[logic.Sum(roomId)%len(addresses)])
	fmt.Println("room members ", addresses[logic.Sum(members)%len(addresses)])
	fmt.Println("room gateway ", addresses[logic.Sum(gateways)%len(addresses)])
	s := NewStat(addresses[logic.Sum(roomId)%len(addresses)],
		addresses[logic.Sum(members)%len(addresses)],
		addresses[logic.Sum(gateways)%len(addresses)])
	go s.Print()
	select {}
}
