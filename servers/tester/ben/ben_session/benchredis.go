package main

import (
	//"errors"
	"errors"
	"fmt"
	"sync"

	"math/rand"
	"strconv"

	"github.com/huajiao-tv/qchat/utility/msgRedis"
)

var (
	SessionPool *msgRedis.MultiPool
	mux         sync.Mutex
)

type Geodata struct {
	Uid      string
	Distance string
	lon      string
	lat      string
}

const redisaddr string = "127.0.0.1:6379:"

func init() {
	SessionPool = msgRedis.NewMultiPool(
		[]string{redisaddr},
		msgRedis.DefaultMaxConnNumber+20,
		msgRedis.DefaultMaxIdleNumber+95,
		msgRedis.DefaultMaxIdleSeconds)
	return
}

func getActiveUserNum(appid uint16) (int, error) {

	var num int
	for i := 0; i < 100; i++ {
		keyname := fmt.Sprintf("userstat:%d:%d:set", appid, i)
		cnum, err := SessionPool.Call(redisaddr).SCARD(keyname)
		if err != nil {
			return num, errors.New(keyname + ":" + err.Error())
		}
		num += int(cnum)
	}

	return num, nil
}

func benchRedis(gn, n int) {
	for i := 0; i < gn; i++ {
		go func(goid, runnum int) {
			defer wg.Done()
			baseuid := fmt.Sprintf("%s-%d", uid, goid)
			for {
				tuid := fmt.Sprintf("%s-%d", baseuid, runnum)

				oc := SessionPool.Call(redisaddr)
				_, err := oc.SADD(tuid, []string{"1"})
				if err != nil {
					fmt.Println("testredis err is ", err)
				} else {
					mux.Lock()
					fmt.Println(oc.GetPoolInfo())
					mux.Unlock()
				}

				if runnum--; runnum == 0 {
					break
				}
			}
		}(i, n)
	}

	num, err := getActiveUserNum(uint16(appid))
	fmt.Println("activenum= ", num, err)
}

func benchInitGeo(n int) {

	for i := 0; i < n; i++ {
		oc := SessionPool.Call(redisaddr)
		lon := rand.Float64()
		lat := rand.Float64()
		name := "zj" + strconv.Itoa(i)
		fmt.Println("geo add ", name, 102.00+lon, 24.00+lat)
		oc.GEOADD("testgeo", 102.00+lon, 24.00+lat, name)
	}
}

func benchGeo(gn, n int, r float64) {

	for i := 0; i < gn; i++ {
		go func(goid, runnum int, radius float64) {
			defer wg.Done()
			for {
				oc := SessionPool.Call(redisaddr)
				v, err := oc.GEORADIUS("testgeo", 102.00, 24.00, radius, 100)
				if err != nil {
					fmt.Println("testredis err is ", err)
				} else {
					mux.Lock()
					for _, item := range v {
						data := item.([]interface{})
						geo := &Geodata{}
						geo.Uid = string(data[0].([]uint8)[:])
						geo.Distance = string(data[1].([]uint8)[:])
						pos := data[2].([]interface{})
						geo.lon = string(pos[0].([]uint8)[:])
						geo.lat = string(pos[1].([]uint8)[:])
						//fmt.Println(geo)
					}
					mux.Unlock()
				}

				if runnum--; runnum == 0 {
					break
				}
			}
		}(i, n, r)
	}
}
