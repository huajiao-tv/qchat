package main

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/logger"
)

// 组件名称
const Component = "logger_benter"
const NodeID = "localhost"

// 包全局变量
var (
	loginfo  string
	loglevel string
	gonum    int
	linenum  int

	beginntf chan bool
	wg       *sync.WaitGroup

	Logger *logger.Logger
	funcs  map[string]interface{}
)

func init() {
	var err error
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.StringVar(&loginfo, "loginfo", "please input your log", "loginfo")
	flag.StringVar(&loglevel, "loglevel", "Error", "loglevel")
	flag.IntVar(&gonum, "gonum", 1, "go routine num")
	flag.IntVar(&linenum, "linenum", 100, "log line num")

	flag.Parse()

	logic.StaticConf.Init()
	filename := filepath.Join(logic.StaticConf.LogDir, fmt.Sprintf("%s-%s", Component, NodeID))
	Logger, err = logger.NewLogger(filename, Component+"|"+NodeID, logic.StaticConf.BackupLogDir)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	Logger.SetLevel(0)
	fmt.Printf("init, logger=%#v\n", Logger)
	//funcs = map[string]interface{}{"Error": Logger.Error, "Warn": Logger.Warn, "Debug": Logger.Debug, "Trace": Logger.Trace}
}

func call(m map[string]interface{}, name string, params ...interface{}) (result []reflect.Value, err error) {
	f := reflect.ValueOf(m[name])
	fmt.Println(f)
	if len(params) != f.Type().NumIn() {
		err = errors.New("The number of params is not adapted.")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	return
}

func testlog(num int) {
	defer wg.Done()
	var logfun func(string)

	<-beginntf

	//fmt.Printf("testlog begin work, logger=%#v\n", Logger)

	switch loglevel {
	case "Debug":
		logfun = func(log string) {
			Logger.Debug(log)
		}
	case "Error":
		logfun = func(log string) {
			Logger.Error(log)
		}
	}

	ticker := time.NewTicker(time.Microsecond)
	defer ticker.Stop()

	for i := 0; i < num; i++ {
		<-ticker.C
		loginfo := fmt.Sprintf("%d: %s\n", i, loginfo)
		//call(funcs, loglevel, loginfo)
		logfun(loginfo)
	}
}

func main() {
	fmt.Printf("main, logger=%#v\n", Logger)

	beginntf = make(chan bool)
	wg = new(sync.WaitGroup)
	wg.Add(gonum)

	for i := 0; i < gonum; i++ {
		go testlog(linenum)
	}

	close(beginntf)

	fmt.Println("main:before wg wait")

	wg.Wait()

	Logger.Debug("main end")

	time.Sleep(time.Second * 2)
}
