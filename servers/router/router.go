package main

import (
	"math/rand"
	"runtime"
	"time"

	"github.com/huajiao-tv/qchat/utility/process"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := initSetting(); err != nil {
		panic(err)
	}

	if err := initDoOperationsPools(); err != nil {
		panic(err)
	}

	rand.Seed(time.Now().UnixNano())
}

func main() {
	process.SavePid(staticConf.PidFile)
	initHttpClient()
	go GorpcServer()
	go AdminServer()

	select {}
}
