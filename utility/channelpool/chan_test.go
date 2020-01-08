package channelpool

import (
	"fmt"
	//	"reflect"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

/*********************Test int*********************************/
var intPool = NewChanPool(50, 5, 3)

func log(s interface{}) {
	fmt.Println(s)
}

func TestReceive(t *testing.T) {
	intPool.SaveFunc = func(datas []interface{}) {
		fmt.Println("Receive:", datas)
	}
	intPool.LogFunc = log
	go intPool.Receive()
}

func TestPush(t *testing.T) {
	for i := 0; i < 10; i++ {
		intPool.Push(i)
		if i%3 == 0 {
			time.Sleep(time.Second * 1)
		}
	}
}

/*********************Test struct*********************************/
type user struct {
	name string
	age  int
}

var userPool = NewChanPool(50, 50, 3)

func TestUserReceive(t *testing.T) {
	userPool.SaveFunc = func(datas []interface{}) {
		fmt.Println("Receive:", datas)
	}
	userPool.LogFunc = log
	go userPool.Receive()
}

func TestUserPush(t *testing.T) {
	for i := 0; i < 10; i++ {
		userPool.Push(user{strconv.Itoa(i) + "a", i})
		if i%3 == 0 {
			time.Sleep(time.Second * 1)
		}
	}
}
