package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/huajiao-tv/qchat/utility/cryption"
)

const (
	// 720 bytes
	Message = "1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz" +
		"1234567890abcdefghijklmnopqrstuvwxyz"
)

func bench_push_all(gn, n int) {
	for i := 0; i < gn; i++ {
		go func(goid, runnum int) {
			defer wg.Done()

			uri := fmt.Sprintf("%s/huajiao/all", centeraddr)
			result := &testResult{MinTime: 1e9}
			body := url.Values{
				"msg":         []string{Message},
				"msgtype":     []string{"100"},
				"traceid":     []string{"1234567890"},
				"expire_time": []string{"300000"},
			}
			values := url.Values{
				"b": []string{cryption.Base64Encode([]byte(body.Encode()))},
				"m": []string{"0"},
			}

			for {
				start := time.Now()

				if _, err := httpClient.PostForm(uri, values); err != nil {
					fmt.Println("push all err is ", err)
					result.fail()
				} else {
					result.success(start)
				}

				if runnum--; runnum == 0 {
					break
				}
			}

			fmt.Println(result.String())
		}(i, n)
	}
}

func bench_push_hot(gn, n int) {
	for i := 0; i < gn; i++ {
		go func(goid, runnum int) {
			defer wg.Done()

			uri := fmt.Sprintf("%s/huajiao/hot", centeraddr)
			result := &testResult{MinTime: 1e9}
			body := url.Values{
				"msg":         []string{Message},
				"msgtype":     []string{"100"},
				"traceid":     []string{"1234567890"},
				"expire_time": []string{"300000"},
			}
			values := url.Values{
				"b": []string{cryption.Base64Encode([]byte(body.Encode()))},
				"m": []string{"0"},
			}

			for {
				start := time.Now()

				if _, err := httpClient.PostForm(uri, values); err != nil {
					fmt.Println("push all err is ", err)
					result.fail()
				} else {
					result.success(start)
				}

				if runnum--; runnum == 0 {
					break
				}
			}

			fmt.Println(result.String())
		}(i, n)
	}
}

func bench_push_user(gn, n int) {
	for i := 0; i < gn; i++ {
		go func(goid, runnum int) {
			defer wg.Done()

			uri := fmt.Sprintf("%s/huajiao/users", centeraddr)
			result := &testResult{MinTime: 1e9}

			for {
				body := url.Values{
					"msg":         []string{Message},
					"msgtype":     []string{"100"},
					"traceid":     []string{fmt.Sprint(time.Now().Unix())},
					"expire_time": []string{fmt.Sprint(expireInterval)},
					"sender":      []string{fmt.Sprint(userStartId + rand.Int63n(userCount))},
					"receivers":   []string{fmt.Sprint(userStartId + rand.Int63n(userCount))},
				}
				values := url.Values{
					"b": []string{cryption.Base64Encode([]byte(body.Encode()))},
					"m": []string{"0"},
				}

				start := time.Now()

				if _, err := httpClient.PostForm(uri, values); err != nil {
					fmt.Println("push peer err is ", err)
					result.fail()
				} else {
					result.success(start)
				}

				if runnum--; runnum == 0 {
					break
				}
			}

			fmt.Println(result.String())
		}(i, n)
	}
}

func bench_push_chat(gn, n int) {
	for i := 0; i < gn; i++ {
		go func(goid, runnum int) {
			defer wg.Done()

			uri := fmt.Sprintf("%s/huajiao/chat", centeraddr)
			result := &testResult{MinTime: 1e9}

			for {
				body := url.Values{
					"msg":         []string{Message},
					"msgtype":     []string{"100"},
					"traceid":     []string{fmt.Sprint(time.Now().Unix())},
					"expire_time": []string{fmt.Sprint(expireInterval)},
					"sender":      []string{fmt.Sprint(userStartId + rand.Int63n(userCount))},
					"receivers":   []string{fmt.Sprint(userStartId + rand.Int63n(userCount))},
				}
				values := url.Values{
					"b": []string{cryption.Base64Encode([]byte(body.Encode()))},
					"m": []string{"0"},
				}
				start := time.Now()

				if _, err := httpClient.PostForm(uri, values); err != nil {
					fmt.Println("push im err is ", err)
					result.fail()
				} else {
					result.success(start)
				}

				if runnum--; runnum == 0 {
					break
				}
			}

			fmt.Println(result.String())
		}(i, n)
	}
}

func bench_push_recall(gn, n int) {
	for i := 0; i < gn; i++ {
		go func(goid, runnum int) {
			defer wg.Done()

			uri := fmt.Sprintf("%s/push/recall", centeraddr)
			result := &testResult{MinTime: 1e9}
			body := url.Values{
				"traceid":   []string{"1234567890"},
				"sender":    []string{"12345678"},
				"receivers": []string{"12345678"},
				"inid":      []string{"1"},
				"outid":     []string{"1"},
			}
			values := url.Values{
				"b": []string{cryption.Base64Encode([]byte(body.Encode()))},
				"m": []string{"0"},
			}

			for {
				start := time.Now()

				if _, err := httpClient.PostForm(uri, values); err != nil {
					fmt.Println("push all err is ", err)
					result.fail()
				} else {
					result.success(start)
				}

				if runnum--; runnum == 0 {
					break
				}
			}

			fmt.Println(result.String())
		}(i, n)
	}
}
