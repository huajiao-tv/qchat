// provides message benchmark test related functions
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/logic"
)

type TestType int

const (
	StoreMsg TestType = iota
	RetrieveMsg
	StoreAndRetrieveMsg
)

func benchmarkBase(t TestType, tf func(index int, counter, respTime chan int64)) {
	beginntf = make(chan bool)
	wg = new(sync.WaitGroup)
	wg.Add(gonum)

	logic.DynamicConf().LocalSaverRpc = saveraddr

	total := int64(0)
	totalResponseTime := int64(0)
	counter := make(chan int64, gonum*100)
	respTime := make(chan int64, gonum*100)

	// start counter thread
	go func() {
		for {
			select {
			case count, ok := <-counter:
				if ok {
					total += count
				} else {
					fmt.Println("counter thread exit for tests are completed.")
					return
				}
			case respT, ok := <-respTime:
				if ok {
					totalResponseTime += respT
				}
			}
		}
	}()

	// start statistics thread
	exit := make(chan bool)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		executedPrev, responseTimePrev := int64(0), int64(0)
		used := 0
		for {
			select {
			case <-ticker.C:
				executed := total
				tps := executed - executedPrev
				respT := totalResponseTime
				averRespT := float64(respT - responseTimePrev)
				if tps > 0 {
					averRespT = averRespT / float64(tps) / float64(time.Millisecond)
				} else {
					averRespT = 0
				}
				used++
				fmt.Printf("Executed %v requests yet, TPS: %v, average response time: %.3f ms, used time(seconds): %v\n",
					executed, tps, averRespT, used)
				executedPrev = executed
				responseTimePrev = respT
			case <-exit:
				fmt.Println("statistics thread exit for tests are completed.")
				return
			}
		}
	}()

	for i := 0; i < gonum; i++ {
		go tf(i, counter, respTime)
	}

	begin := time.Now()
	tc := ""
	switch t {
	case StoreMsg:
		tc = "store message"
	case RetrieveMsg:
		tc = "retrieve message"
	case StoreAndRetrieveMsg:
		tc = "store and retrieve message"
	default:
		tc = "unknown "
	}

	fmt.Printf("Start %s benchmarking test, Channel:%s, Go routine num:%v, Request count each routine:%v, User id is in [%v, %v), Begin time:%v\n",
		tc, channel, gonum, runTimes, userStartId, userStartId+userCount, begin)

	close(beginntf)

	wg.Wait()
	end := time.Now()

	time.Sleep(time.Second)
	// notify counter thread to close
	close(counter)
	// notify statistics thread to close
	close(exit)

	time.Sleep(time.Second)

	used := end.Sub(begin).Seconds()
	if used > 0 {
		tps := int64(float64(total) / used)
		fmt.Println("Total request:", total, "Average response time(ms):",
			fmt.Sprintf("%.3f", float64(totalResponseTime)/float64(total)/float64(time.Millisecond)),
			"Used time(seconds):", fmt.Sprintf("%.3f", used), "Computed TPS:", tps)
	} else {
		fmt.Println("run time finished in one second, cannot compute the tps")
	}
}

func benchmarkStoreMsg() {
	benchmarkBase(StoreMsg, storeMsgTest)
}

func benchmarkRetrieveMsg() {
	benchmarkBase(RetrieveMsg, retrieveMsgTest)
}

func benchmarkStoreAndRetrieveMsg() {
	benchmarkBase(StoreAndRetrieveMsg, storeAndRetrieveMsgTest)
}

func storeMsg(content string) {
	sender := fmt.Sprint(userStartId + rand.Int63n(userCount))
	receiver := fmt.Sprint(userStartId + rand.Int63n(userCount))

	traceId := time.Now().Unix()
	request := &saver.StoreMessagesRequest{Appid: logic.APPID_HUAJIAO,
		Messages:    make(map[string]*saver.ChatMessage, 1),
		TraceSN:     fmt.Sprint(traceId),
		ChatChannel: channel}
	request.Messages[receiver] = &saver.ChatMessage{Content: content, Type: 1, To: receiver,
		From: sender, TraceSN: traceId, ExpireInterval: expireInterval, StoreOutbox: uint8(storeOutbox)}
	if _, err := saver.StoreChatMessages(request); err != nil {
		fmt.Println(time.Now().String(), "Test store", channel, "message failed, error:", err.Error())
	}
}

func storeMsgTest(index int, counter, respTime chan int64) {
	defer wg.Done() // notify main when this is finish
	content := GenRandStringRunes(512 + rand.Intn(512))
	<-beginntf // wait start signal

	for i := 0; i < runTimes; i++ {
		start := time.Now()
		storeMsg(content)
		respTime <- int64(time.Since(start))
		counter <- 1
	}
	fmt.Println("#", index, "worker has performed", runTimes, "store message requests, exiting...")
}

func retrieveMsg(owner string, start int64) {
	chnnels := map[string]*saver.RetrieveChannel{
		channel: &saver.RetrieveChannel{Channel: channel, StartMsgId: start}}
	request := &saver.RetrieveMessagesRequest{Appid: logic.APPID_HUAJIAO, Owner: owner, ChatChannels: chnnels,
		TraceSN: fmt.Sprint(time.Now().Unix())}
	if _, err := saver.RetrieveChatMessages(request); err != nil {
		fmt.Println(time.Now().String(), "Test retrieve", channel, "message failed, error:", err.Error())
	}
}

func retrieveMsgTest(index int, counter, respTime chan int64) {
	defer wg.Done() // notify main when this is finish
	<-beginntf      // wait start signal

	for i := 0; i < runTimes; i++ {
		receiver := fmt.Sprint(userStartId + rand.Int63n(userCount))
		start := time.Now()
		retrieveMsg(receiver, int64(rand.Intn(10)))
		respTime <- int64(time.Since(start))
		counter <- 1
	}
	fmt.Println("#", index, "worker has performed", runTimes, "retrieve message requests, exiting...")
}

func storeAndRetrieveMsg(content string) {
	sender := fmt.Sprint(userStartId + rand.Int63n(userCount))
	receiver := fmt.Sprint(userStartId + rand.Int63n(userCount))

	traceId := time.Now().Unix()
	request := &saver.StoreMessagesRequest{Appid: logic.APPID_HUAJIAO,
		Messages:    make(map[string]*saver.ChatMessage, 1),
		TraceSN:     fmt.Sprint(traceId),
		ChatChannel: channel}
	request.Messages[receiver] = &saver.ChatMessage{Content: content, Type: 1, To: receiver,
		From: sender, TraceSN: traceId, ExpireInterval: expireInterval, StoreOutbox: uint8(storeOutbox)}
	if resp, err := saver.StoreChatMessages(request); err != nil {
		fmt.Println(time.Now().String(), "Test store", channel, "message failed, error:", err.Error())
	} else {
		if message, ok := resp.Inbox[receiver]; ok {
			retrieveMsg(receiver, int64(message.MsgId))
		} else {
			fmt.Println(time.Now().String(), "Channel", channel, "doesn't have inbox messages in response")
		}
	}
}

func storeAndRetrieveMsgTest(index int, counter, respTime chan int64) {
	// notify main when this is finish
	defer wg.Done()

	content := GenRandStringRunes(512 + rand.Intn(512))

	// wait start signal
	<-beginntf

	for i := 0; i < runTimes; i++ {
		start := time.Now()
		storeAndRetrieveMsg(content)
		respTime <- int64(time.Since(start))
		counter <- 2
	}

	fmt.Println("#", index, "worker has performed", runTimes, "store and retrieve message requests, exiting...")
}
