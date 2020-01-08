package main

import (
	"flag"
	"fmt"

	"github.com/huajiao-tv/qchat/client/dispatcher"
)

var (
	addr string
)

func init() {
	flag.StringVar(&addr, "server", "127.0.0.1:8088", "")
	flag.Parse()
}

func main() {
	if resp, err := dispatcher.GetQps(addr); err != nil {
		fmt.Println("failed", err.Error())
	} else {
		fmt.Println("ok", resp)
	}
}
