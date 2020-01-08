package channelpool

import (
	"time"
)

/*
 * ChanPool是为了解决批量处理数据的问题而封装的包
 * dataQueue：		使用者自定义类型channel
 * dataSlice: 		使用者自定义类型slice
 * tickerSeconds：	每隔tickerSeconds时间间隔，触发执行SaveFunc逻辑
 * executeLength：	channel buffer中的数据超过executeLength，触发执行SaveFunc逻辑
 * QueueLength：	channel buffer大小
 *
 * SaveFunc:		必须实现，根据需求自定义
 * LogFunc:			必须实现，处理出错情况的日志记录
 */

type ChanPool struct {
	dataQueue     chan interface{}
	dataSlice     []interface{}
	tickerSeconds time.Duration
	executeLength int

	SaveFunc func(datas []interface{})
	LogFunc  func(err interface{})
}

func NewChanPool(QueueLength int, exeLength int, tickerSeconds time.Duration) *ChanPool {
	return &ChanPool{
		dataQueue:     make(chan interface{}, QueueLength),
		dataSlice:     []interface{}{},
		tickerSeconds: tickerSeconds,
		executeLength: exeLength,
	}
}

func (this *ChanPool) Push(v interface{}) {
	select {
	case this.dataQueue <- v:
	default:
		select {
		case lost := <-this.dataQueue:
			this.LogFunc(lost)
			select {
			case this.dataQueue <- v:
			default:
				this.LogFunc(v)
			}
		default:
		}
	}
}

func (this *ChanPool) PushDiscard(v interface{}) {
	select {
	case this.dataQueue <- v:
	default:
		this.LogFunc(v)
	}
}

func (this *ChanPool) Receive() {
	ticker := time.NewTicker(time.Second * this.tickerSeconds)
	for {
		select {
		case data := <-this.dataQueue:
			this.dataSlice = append(this.dataSlice, data)
			if len(this.dataSlice) >= this.executeLength {
				go this.SaveFunc(this.dataSlice)
				this.dataSlice = []interface{}{}
			}
		case <-ticker.C:
			if len(this.dataSlice) > 0 {
				go this.SaveFunc(this.dataSlice)
				this.dataSlice = []interface{}{}
			}
		}
	}
}
