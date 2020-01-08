package client

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/huajiao-tv/qchat/utility/cryption"
)

var (
	httpClient *http.Client
)

func init() {
	httpClient = &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(network, addr, time.Duration(3000)*time.Millisecond)
				if err != nil {
					return nil, err
				}
				return c, nil
			},
			MaxIdleConnsPerHost: 100,
		},
		Timeout: time.Duration(5000) * time.Millisecond,
	}
}

func Http(tag string, cmds []string) {
	var (
		method string
		url    string
		args   string
	)
	http := flag.NewFlagSet("ben_llconn.http", flag.ContinueOnError)
	http.StringVar(&method, "m", "get", "http method")
	http.StringVar(&url, "u", "", "server url")
	http.StringVar(&args, "a", "", "arguments")
	http.Parse(cmds)
	if url == "" {
		Log(tag, "global", "error", "no url", cmds)
		return
	}
	switch strings.ToLower(method) {
	case "get":
		if args != "" {
			url = fmt.Sprintf("%s?%s", url, args)
		}
		if res, err := httpClient.Get(url); err != nil {
			Log(tag, "global", "get", url, "error", err.Error())
		} else {
			Log(tag, "global", "get", url, "response", res.StatusCode)
		}
	case "post":
		res, err := httpClient.Post(url, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(args)))
		if err != nil {
			Log(tag, "global", "post", url, "error", err.Error())
		} else {
			Log(tag, "global", "post", url, "response", res.StatusCode)
		}
	}
}

func sendChatroom(c *UserConnection, args []string) {
	if len(args) < 2 {
		c.log("error", "sendChatroom", args)
		return
	}
	room := args[0]
	message := args[1]
	priority := "0"
	if len(args) > 2 {
		priority = args[2]
	}
	uri := fmt.Sprintf("http://%s/send", c.conf.CenterAddr)
	values := url.Values{
		"appid":    []string{fmt.Sprint(c.conf.AppID)},
		"roomid":   []string{room},
		"sender":   []string{c.account.UserID},
		"traceid":  []string{fmt.Sprint(time.Now().UnixNano())},
		"content":  []string{message},
		"priority": []string{priority},
	}

	if res, err := httpClient.PostForm(uri, values); err != nil {
		c.log("error", "sendChatroom", err.Error())
	} else {
		ioutil.ReadAll(res.Body)
		res.Body.Close()
		c.log("sendChatroom", res.StatusCode)
	}
}

func sendRobotJoinMessage(c *UserConnection, args []string) {
	if len(args) < 2 {
		c.log("error, args length less than 2", "robotjoin", args)
		return
	}

	roomid := args[0]
	robotid := args[1]
	uri := fmt.Sprintf("http://%s/joinroom", c.conf.CenterAddr)

	values := url.Values{
		"rid":   []string{roomid},
		"uid":   []string{robotid},
		"appid": []string{fmt.Sprint(c.conf.AppID)},
	}

	if res, err := httpClient.PostForm(uri, values); err != nil {
		c.log("error", "robotjoin", err.Error())
	} else {
		ioutil.ReadAll(res.Body)
		res.Body.Close()
		c.log("robotjoin", res.StatusCode)
	}
}

func sendRobotQuitMessage(c *UserConnection, args []string) {
	if len(args) < 2 {
		c.log("error, args length less than 2", "robotquit", args)
		return
	}

	roomid := args[0]
	robotid := args[1]
	uri := fmt.Sprintf("http://%s/quitroom", c.conf.CenterAddr)

	values := url.Values{
		"rid":   []string{roomid},
		"uid":   []string{robotid},
		"appid": []string{fmt.Sprint(c.conf.AppID)},
	}

	if res, err := httpClient.PostForm(uri, values); err != nil {
		c.log("error", "robotquit", err.Error())
	} else {
		ioutil.ReadAll(res.Body)
		res.Body.Close()
		c.log("robotquit", res.StatusCode)
	}
}

func sendPeer(c *UserConnection, args []string) {
	if len(args) < 2 {
		c.log("error", "sendPeer", args)
		return
	}
	uri := fmt.Sprintf("http://%s/huajiao/users", c.conf.CenterAddr)
	body := url.Values{
		"appid":       []string{fmt.Sprint(c.conf.AppID)},
		"msg":         []string{args[1]},
		"msgtype":     []string{"100"},
		"traceid":     []string{fmt.Sprint(time.Now().UnixNano())},
		"expire_time": []string{"300000"},
		"sender":      []string{c.account.UserID},
		"receivers":   []string{args[0]},
	}
	values := url.Values{
		"b": []string{cryption.Base64Encode([]byte(body.Encode()))},
		"m": []string{"0"},
	}

	if res, err := httpClient.PostForm(uri, values); err != nil {
		c.log("error", "sendPeer", err.Error())
	} else {
		ioutil.ReadAll(res.Body)
		res.Body.Close()
		c.log("sendPeer", res.StatusCode)
	}
}

func sendIm(c *UserConnection, args []string) {
	if len(args) < 2 {
		c.log("error", "sendIm", args)
		return
	}
	uri := fmt.Sprintf("http://%s/push/chat", c.conf.CenterAddr)
	body := url.Values{
		"appid":       []string{fmt.Sprint(c.conf.AppID)},
		"msg":         []string{args[1]},
		"msgtype":     []string{"100"},
		"traceid":     []string{fmt.Sprint(time.Now().UnixNano())},
		"expire_time": []string{"300000"},
		"sender":      []string{c.account.UserID},
		"receivers":   []string{args[0]},
	}
	values := url.Values{
		"b": []string{cryption.Base64Encode([]byte(body.Encode()))},
		"m": []string{"0"},
	}

	if res, err := httpClient.PostForm(uri, values); err != nil {
		c.log("error", "sendIm", err.Error())
	} else {
		ioutil.ReadAll(res.Body)
		res.Body.Close()
		c.log("sendIm", res.StatusCode)
	}
}

func sendWg(c *UserConnection, args []string) {
	if len(args) < 2 {
		c.log("error", "sendWg", args)
		return
	}

	uri := fmt.Sprintf("http://%s/chatroom/broadcast", c.conf.CenterAddr)
	body := url.Values{
		"appid":    []string{fmt.Sprint(c.conf.AppID)},
		"sender":   []string{c.account.UserID},
		"content":  []string{args[0]},
		"priority": []string{args[1]},
		"traceid":  []string{fmt.Sprint(time.Now().UnixNano())},
	}

	values := url.Values{
		"b": []string{cryption.Base64Encode([]byte(body.Encode()))},
		"m": []string{"0"},
	}

	if res, err := httpClient.PostForm(uri, values); err != nil {
		c.log("error", "sendwg", err.Error())
	} else {
		ioutil.ReadAll(res.Body)
		res.Body.Close()
		c.log("sendwg", res.StatusCode)
	}
}
