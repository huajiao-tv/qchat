package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"
)

var (
	testcase       string
	centeraddr     string
	content        string
	gonum          int
	num            int
	wg             *sync.WaitGroup
	httpClient     *http.Client
	userStartId    int64  // test start id
	userCount      int64  // test user count
	expireInterval int    // expire interval
	roomid         string //room id
	gap            int    // gap size
	interval       int    // message sending interval in seconds
	normalGapSize  int    // normal gap size in overload mode
	sid            int    // start id
	help           string
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	httpClient = &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(network, addr, time.Duration(1000)*time.Millisecond)
				if err != nil {
					return nil, err
				}
				return c, nil
			},
			MaxIdleConnsPerHost: 100,
		},
		Timeout: time.Duration(2000) * time.Millisecond,
	}

	flag.StringVar(&help, "help", "", "help test case")
	flag.StringVar(&testcase, "tc", "help", "test case name")
	flag.StringVar(&centeraddr, "host", "http://127.0.0.1:6600", "center addr")
	flag.StringVar(&content, "content", "abcdefghijklmn", "send chatroom text")
	flag.IntVar(&num, "num", 1, "test num every goroutinue")
	flag.IntVar(&gonum, "gonum", runtime.NumCPU(), "goroutinue number")
	flag.Int64Var(&userStartId, "usid", 860000000, "user start id")
	flag.Int64Var(&userCount, "uc", 1000000, "user count")
	flag.IntVar(&expireInterval, "ei", 86400, "expire interval")
	flag.StringVar(&roomid, "rid", "1234567", "room id")
	flag.IntVar(&gap, "gap", 100, "gap size")
	flag.IntVar(&interval, "i", 1, "message sending interval, unit might be millisecond or second according to test case")
	flag.IntVar(&normalGapSize, "og", 3, "normal gap size in overload mode")
	flag.IntVar(&sid, "sid", 1, "start id")

	flag.Parse()
}

func main() {
	if help != "" {
		fmt.Println(testcase_help(help))
		return
	}

	wg = new(sync.WaitGroup)
	wg.Add(gonum)

	begin := time.Now()

	switch testcase {
	// for chatroom
	case "join":
		bench_join(gonum, num)
	case "quit":
		bench_quit(gonum, num)
	case "send":
		bench_send(gonum, num)
	case "query":
		bench_query(gonum, num)
	case "rand":
		bench_send_random(gonum, num)
	case "gap":
		bench_send_gap(gonum, num)
	case "ol":
		bench_send_gap_overload(gonum, num)

	// for chat/push
	case "push_all":
		bench_push_all(gonum, num)
	case "push_hot":
		bench_push_hot(gonum, num)
	case "push_user":
		bench_push_user(gonum, num)
	case "push_chat":
		bench_push_chat(gonum, num)
	case "push_recall":
		bench_push_recall(gonum, num)

	// default for help
	case "help":
		fallthrough
	default:
		fmt.Println(usage())
		return
	}

	wg.Wait()
	end := time.Now()

	duration := end.Sub(begin)
	sec := duration.Seconds()
	fmt.Println(float64(gonum*num)/sec, sec)
}

func usage() string {
	return `
ben_center is a benchmark tool for center HTTP interface development

Usage:
	ben_center -tc testcase [arguments] [-host "http://127.0.0.1:6600"] [-gonum 24] [-num 1]

The test cases are:
	join        benchmarks join chatroom operation capability of specified center
	quit        benchmarks quit chatroom operation capability of specified center
	send        benchmarks sending chatroom message operation capability of specified center
	query       benchmarks query chatroom members detail information operation capability of specified center
	rand        simulates send messages unordered to specified chatroom via specified center
	gap         simulates message gap of specified chatroom via specified center
	ol          simulates overload caused by big message gap and restore for specified chatroom via specified center
	push_all    benchmarks push message to all users capability of specified center
	push_hot    benchmarks push message to online users capability of specified center
	push_user   benchmarks push notify to specified user capability of specified center
	push_chat   benchmarks send private chat message to specified user capability of specified center

Use "ben_center -help [test case]" for more information about a test case.
	`
}

func testcase_help(tc string) string {
	switch tc {
	case "join":
		return `
This test case test join chatroom operation capability of specified center

Usage:
	ben_center -tc join [-host "http://127.0.0.1:6600"] [-gonum 24] [-num 1]

More information of arguments:
	host        specified center address, should starts with "http://", must include correct port
	              default value is "http://127.0.0.1:6600"
	gonum       the go routine number, default is the number of cpu
	num         test num each go routinue, default is 1
		`

	case "quit":
		return `
This test case test quit chatroom operation capability of specified center

Usage:
	ben_center -tc quit [-host "http://127.0.0.1:6600"] [-gonum 24] [-num 1]

More information of arguments:
	host        specified center address, should starts with "http://", must include correct port
	              default value is "http://127.0.0.1:6600"
	gonum       the go routine number, default is the number of cpu
	num         test num each go routinue, default is 1
		`

	case "send":
		return `
This test case test send chatroom operation capability of specified center

Usage:
	ben_center -tc send [-host "http://127.0.0.1:6600"] [-gonum 24] [-num 1]

More information of arguments:
	host        specified center address, should starts with "http://", must include correct port
	              default value is "http://127.0.0.1:6600"
	gonum       the go routine number, default is the number of cpu
	num         test num each go routinue, default is 1
		`

	case "query":
		return `
This test case test query chatroom operation capability of specified center

Usage:
	ben_center -tc query [-host "http://127.0.0.1:6600"] [-gonum 24] [-num 1]

More information of arguments:
	host        specified center address, should starts with "http://", must include correct port
	              default value is "http://127.0.0.1:6600"
	gonum       the go routine number, default is the number of cpu
	num         test num each go routinue, default is 1
		`

	case "rand":
		return `
This test case simulates send messages unordered to specified chatroom via specified center

Usage:
	ben_center -tc rand [-host "http://127.0.0.1:6600"] [-gonum 24] [-num 1] [-rid "1234567"] [-sid 1] [-i 1]

More information of arguments:
	host        specified center address, should starts with "http://", must include correct port
	              default value is "http://127.0.0.1:6600"
	gonum       the go routine number, default is the number of cpu
	num         test num each go routinue, default is 1
	rid         specified chatroom ID, default is "1234567"
	sid         specified valid start message id, default is 1
	i           message sending interval, default is 1. goroutines with odd index use milliseconds as unit,
	              while even index goroutines use second as unit
		`

	case "gap":
		return `
This test case test simulates message gap of specified chatroom via specified center

Usage:
	ben_center -tc gap [-host "http://127.0.0.1:6600"] [-gonum 24] [-num 1] [-rid "1234567"] [-sid 1] [-i 1] [-gap 100]

More information of arguments:
	host        specified center address, should starts with "http://", must include correct port
	              default value is "http://127.0.0.1:6600"
	gonum       the go routine number, default is the number of cpu
	num         test num each go routinue, default is 1
	rid         specified chatroom ID, default is "1234567"
	sid         specified valid start message id, default is 1
	i           message sending interval, default is 1. unit is second
	gap         indicates range of two message ID, which will used to build gap
		`

	case "ol":
		return `
This test case test simulates overload caused by big message gap and restore for specified chatroom via specified center

Usage:
	ben_center -tc ol [-host "http://127.0.0.1:6600"] [-gonum 24] [-num 1] [-rid "1234567"] [-sid 1] [-i 1] [-gap 100] [-og 3]

More information of arguments:
	host        specified center address, should starts with "http://", must include correct port
	              default value is "http://127.0.0.1:6600"
	gonum       the go routine number, default is the number of cpu
	num         test num each go routinue, default is 1
	rid         specified chatroom ID, default is "1234567"
	sid         specified valid start message id, default is 1
	i           message sending interval, default is 1. unit is second
	gap         indicates range of two message ID, which will used to build gap
	og          normal gap size in overload restore period, default is 3
		`

	case "push_all":
		return `
This test case test push message to all users capability of specified center

Usage:
	ben_center -tc push_all [-host "http://127.0.0.1:6600"] [-gonum 24] [-num 1]

More information of arguments:
	host        specified center address, should starts with "http://", must include correct port
	              default value is "http://127.0.0.1:6600"
	gonum       the go routine number, default is the number of cpu
	num         test num each go routinue, default is 1
		`

	case "push_hot":
		return `
This test case test push message to online users capability of specified center

Usage:
	ben_center -tc push_hot [-host "http://127.0.0.1:6600"] [-gonum 24] [-num 1]

More information of arguments:
	host        specified center address, should starts with "http://", must include correct port
	              default value is "http://127.0.0.1:6600"
	gonum       the go routine number, default is the number of cpu
	num         test num each go routinue, default is 1
		`

	case "push_user":
		return `
This test case test push notify to specified user capability of specified center

Usage:
	ben_center -tc push_user [-host "http://127.0.0.1:6600"] [-gonum 24] [-num 1] [-usid 860000000] [-uc 1000000] [-ei 86400]

More information of arguments:
	host        specified center address, should starts with "http://", must include correct port
	              default value is "http://127.0.0.1:6600"
	gonum       the go routine number, default is the number of cpu
	num         test num each go routinue, default is 1
	usid        the start receiver user id, test case will send message to random user which user id in [usid, usid+uc].
	              default is 860000000
	uc          rand range from user start id, default is 1000000
	ei          the message expire interval, unit is second. default is 86400
		`

	case "push_chat":
		return `
This test case test send private chat message to specified user capability of specified center

Usage:
	ben_center -tc push_chat [-host "http://127.0.0.1:6600"] [-gonum 24] [-num 1] [-usid 860000000] [-uc 1000000] [-ei 86400]

More information of arguments:
	host        specified center address, should starts with "http://", must include correct port
	              default value is "http://127.0.0.1:6600"
	gonum       the go routine number, default is the number of cpu
	num         test num each go routinue, default is 1
	usid        the start receiver user id, test case will send message to random user which user id in [usid, usid+uc].
	              default is 860000000
	uc          rand range from user start id, default is 1000000
	ei          the message expire interval, unit is second. default is 86400
		`

	default:
		return "\nwrong test case \n" + usage()
	}
}
