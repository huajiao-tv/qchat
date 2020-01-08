package main

import (
	"flag"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/logger"
)

var (
	servers string
	roomId  string
	mode    string
	test    int
)

var (
	DefaultKey       []byte        = []byte("894184791415baf5c113f83eaff360f0")
	AppId            uint16        = 1080
	ConnectTimeout   time.Duration = 2 * time.Second
	WriteTimeout     time.Duration = 2 * time.Second
	ReadTimeout      time.Duration = 2 * time.Second
	HeartBeatTimeout time.Duration = 60 * time.Second
	Component        string        = "test_llconn_delay"

	Logger *logger.Logger
)

func init() {
	rand.Seed(time.Now().UnixNano())
	filename := filepath.Join(logic.StaticConf.LogDir, Component)
	Logger, _ = logger.NewLogger(filename, Component, logic.StaticConf.BackupLogDir)
	Logger.SetLevel(0)

	flag.StringVar(&roomId, "rid", "123", "room id")
	flag.StringVar(&servers, "servers", "127.0.0.1:8080,127.0.0.1:8888", "gateway addr")
	flag.StringVar(&mode, "mode", "recv", "send or recv")
	flag.IntVar(&test, "test", 0, "only for send mode")
	flag.Parse()
}

func main() {

	if mode == "send" {
		for {
			err := sendToRoom(roomId)
			if err != nil {
				Logger.Debug(roomId, "sendToRoom failed", err.Error())
			}
			time.Sleep(time.Duration(rand.Intn(1000)+1000) * time.Millisecond)
		}
	}

	m := make(map[string]*Client)
	for _, server := range strings.Split(servers, ",") {
		uid := logic.RandString(20)
		c := NewClient(uid, server)
		if c == nil {
			panic("start failed,uid:" + uid + ",server:" + server)
			continue
		}
		m[uid] = c
	}

	select {}
}
