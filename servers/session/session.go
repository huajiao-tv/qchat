package main

import (
	"math/rand"
	"runtime"
	"time"

	"github.com/huajiao-tv/qchat/utility/process"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	rand.Seed(time.Now().UnixNano())

	if err := initSetting(); err != nil {
		panic(err)
	}
}

func main() {
	process.SavePid(staticConf.PidFile)

	initOnlineCache()
	go GorpcServer()
	go AdminServer()
	initCron()

	select {}
}
