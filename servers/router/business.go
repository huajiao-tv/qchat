package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

var httpClient *http.Client

func initHttpClient() {
	httpClient = &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(netw, addr, time.Duration(netConf().CallbackConnTimeout)*time.Millisecond)
				if err != nil {
					return nil, err
				}
				return c, nil
			},
			MaxIdleConnsPerHost: netConf().CallbackMaxIdleConns,
			DisableKeepAlives:   false,
		},
		Timeout: time.Duration(netConf().CallbackTimeout) * time.Millisecond,
	}
}

func checkSig(appid, sender, sig string) ([]byte, error) {
	url, ok := netConf().CallbackUrl[appid]
	if !ok {
		return nil, errors.New("callback_url not found!")
	}
	realurl := fmt.Sprintf(url, sender, sig)
	begin := time.Now()
	resp, err := httpClient.Get(realurl)
	cost := time.Now().Sub(begin).String()

	if err != nil {
		Logger.Error(sender, appid, sig, "checkSig", "client.Get error("+cost+")", err.Error())
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		Logger.Error(sender, appid, sig, "checkSig", "StatusCode is not ok("+cost+")", resp.StatusCode)
		return nil, errors.New("http code is not ok")
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		Logger.Error(sender, appid, sig, "checkSig", "read body error("+cost+")", err)
		return nil, err
	}
	Logger.Debug(sender, appid, sig, "checkSig", "cost", cost)
	var f interface{}
	if err := json.Unmarshal(body, &f); err != nil {
		Logger.Error(sender, appid, sig, "checkSig", "json unmarshal error", err)
		return nil, err
	}
	if m, ok := f.(map[string]interface{}); ok {
		if data, ok := m["data"].(map[string]interface{}); ok {
			if token, ok := data["token"]; ok {
				if s, ok := token.(string); ok {
					return []byte(s), nil
				}
			}
		}
	}
	Logger.Warn(sender, appid, sig, "checkSig", "bad return", string(body))
	return nil, errors.New("bad return" + string(body))
}
