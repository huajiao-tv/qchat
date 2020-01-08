package main

import (
	"runtime"

	"github.com/huajiao-tv/qchat/utility/msgRedis"
	"github.com/huajiao-tv/qchat/utility/process"
)

var (
	SessionPool *msgRedis.MultiPool // 存储session的redis池子

)

func initRedisPool() error {
	SessionPool = msgRedis.NewMultiPool(
		netConf().SessionAddrs,
		msgRedis.DefaultMaxConnNumber+20,
		msgRedis.DefaultMaxIdleNumber+95,
		msgRedis.DefaultMaxIdleSeconds)

	return nil
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := initSetting(); err != nil {
		panic(err)
	}
	if err := initRedisPool(); err != nil {
		panic(err)
	}
	if err := initMongoSession(); err != nil {
		panic(err)
	}
	if err := initCron(); err != nil {
		panic(err)
	}
	//if err := initMysql(); err != nil {
	//	Logger.Error("Init mysql err", err)
	//}
}

func main() {
	process.SavePid(staticConf.PidFile)
	go GorpcServer()
	go AdminServer()

	select {}
}
