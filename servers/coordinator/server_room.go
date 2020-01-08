package main

import (
	"container/list"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/huajiao-tv/qchat/logic"
)

//全局配置项目

const (
	ConsumerCount   = 10
	ConsumerChanLen = 10000
	VipMsgChanLen   = 5000
	TextType        = "text"
	OtherType       = "other"
)

//配置项结构，包含消息缓存秒数，每秒条数，各消息类比例,取消息序列，消息序列长度
type ServerRoomConf struct {
	MaxCacheWhiteListSec int
	MaxCacheSec          int
	MaxCacheNumPerSec    int
	MsgTypeMap           map[string]string
	MsgClusterRatio      map[string]int
	MsgTypeOrder         []string
	OrderLen             int
	WhiteListCluster     string
}

//缓存消息的cache结构，list中存储分为每秒的消息
//list的元素结构为map[string]chan *logic.ChatRoomMsgRaw，其中string为消息类型
//MsgIndex用来记录当前取消息取到了哪种类型
type MessagesCachePool struct {
	sync.RWMutex
	List     *list.List
	MsgIndex int
}

// 每一个大聊天室，每一个机房会有一个这个对象
type ServerRoom struct {
	AppId  uint16
	RoomID string
	Name   string

	close         chan bool
	MessagesCache *MessagesCachePool
	//等待向cache中分类的消息队列
	MsgClassifyConsumer chan *logic.ChatRoomMsgRaw
	VipMsgChan          chan *logic.ChatRoomMsgRaw
}

//新建一个存储一秒消息的cache结构,其中string为消息分组，chan为具体消息
func NewMessagePerSec() *map[string]chan *logic.ChatRoomMsgRaw {
	res := make(map[string]chan *logic.ChatRoomMsgRaw, len(MsgPoliceConf().MsgClusterRatio))
	for msgCluster, ratio := range MsgPoliceConf().MsgClusterRatio {
		res[msgCluster] = make(chan *logic.ChatRoomMsgRaw, (MsgPoliceConf().MaxCacheNumPerSec*ratio)/MsgPoliceConf().OrderLen)
	}
	return &res
}

//新建一个房间消息控制类型,并返回这个结构的指针
func NewServerRoom(appid uint16, room string, sr string) *ServerRoom {
	serverRoom := &ServerRoom{
		AppId:  appid,
		RoomID: room,
		Name:   sr,
		close:  make(chan bool),
		MessagesCache: &MessagesCachePool{
			List:     list.New(),
			MsgIndex: 0,
		},
		MsgClassifyConsumer: make(chan *logic.ChatRoomMsgRaw, ConsumerChanLen),
		VipMsgChan:          make(chan *logic.ChatRoomMsgRaw, VipMsgChanLen),
	}
	msgPerSec := NewMessagePerSec()
	serverRoom.MessagesCache.List.PushFront(msgPerSec)
	//从待分类消息队列中分类消息到cache各chan的消费协程
	for i := 0; i < ConsumerCount; i += 1 {
		go serverRoom.StartCosum()
	}
	//每秒钟更新cache list的协程
	go serverRoom.CacheUpdateLoop()
	return serverRoom
}

//消耗消息派发到cache的队列的协程
func (sr *ServerRoom) StartCosum() {
	for {
		select {
		case <-sr.close:
			return
		case d := <-sr.MsgClassifyConsumer:
			sr.ClassifyMsg2Cache(d)
		}
	}
}

//将消息派发队列中的消息派发的具体操作
func (sr *ServerRoom) ClassifyMsg2Cache(msg *logic.ChatRoomMsgRaw) {
	if msg == nil {
		return
	}
	//如果是超级用户则放入超级用户消息队列
	for _, user := range netConf().SuperUsers {
		if msg.Sender == user {
			select {
			case sr.VipMsgChan <- msg:
				return
			default:
				Logger.Error(msg.RoomID, msg.Appid, sr.Name, "ServerRoom.ClassifyMsg2Cache", "vip chan full", msg.Sender, msg.TraceId, msg.MsgType, msg.Priority)
				continue
			}
		}
	}
	sr.MessagesCache.RLock()
	element := sr.MessagesCache.List.Front()
	sr.MessagesCache.RUnlock()
	if element == nil {
		Logger.Error(msg.RoomID, msg.Appid, sr.Name, "ServerRoom.ClassifyMsg2Cache", "no list", msg.Sender, msg.TraceId, msg.MsgType, msg.Priority)
		return
	}
	currentPerSecCache := element.Value.(*map[string]chan *logic.ChatRoomMsgRaw)
	msgType := strconv.Itoa(msg.MsgType)
	msgPri := strconv.Itoa(msg.Priority)
	//如果配置中存在“消息类型-消息优先级”形式的消息配置，则放入此类型
	//否则放入 ”消息类型- “ 类型
	//否则放入其他类型
	var msgCluster string
	if cluster, ok := MsgPoliceConf().MsgTypeMap[msgType+"-"+msgPri]; ok {
		msgCluster = cluster
	} else if cluster, ok := MsgPoliceConf().MsgTypeMap[msgType+"-*"]; ok {
		msgCluster = cluster
	} else {
		msgCluster = MsgPoliceConf().MsgTypeMap["--"]
	}
	//通道满则丢弃
	select {
	case (*currentPerSecCache)[msgCluster] <- msg:
		Logger.Debug(msg.RoomID, msg.Appid, sr.Name, "ServerRoom.ClassifyMsg2Cache", msgCluster, msg.Sender, msg.TraceId, msg.MsgType, msg.Priority)
		return
	default:
		Logger.Error(msg.RoomID, msg.Appid, sr.Name, "ServerRoom.ClassifyMsg2Cache", msgCluster+" chan full", msg.Sender, msg.TraceId, msg.MsgType, msg.Priority)
		return
	}
}

//每秒更改当前房间消息cache的状态，新建当前秒的消息缓存，删除过期秒的消息缓存
func (sr *ServerRoom) CacheUpdateLoop() {
	for {
		select {
		case <-sr.close:
			return
		case <-time.After(time.Second):
			discard := make([]*map[string]chan *logic.ChatRoomMsgRaw, 0, MsgPoliceConf().MaxCacheWhiteListSec)
			msgPerSec := NewMessagePerSec()
			sr.MessagesCache.Lock()
			sr.MessagesCache.List.PushFront(msgPerSec)
			length := sr.MessagesCache.List.Len()
			//根据最大缓存秒数删除过期的消息缓存
			for i := length; i > MsgPoliceConf().MaxCacheWhiteListSec; i -= 1 {
				delElement := sr.MessagesCache.List.Back()
				v := sr.MessagesCache.List.Remove(delElement)
				if tmp, ok := v.(*map[string]chan *logic.ChatRoomMsgRaw); ok {
					discard = append(discard, tmp)
				} else {
					Logger.Error(sr.RoomID, sr.AppId, sr.Name, "ServerRoom.CacheUpdateLoop", "type err")
				}
			}
			sr.MessagesCache.Unlock()
			for _, m := range discard {
				for msgCluster, msgChan := range *m {
					if len(msgChan) == 0 {
						continue
					}
					Logger.Debug(sr.RoomID, sr.AppId, sr.Name, "ServerRoom.CacheUpdateLoop", "clean cache", msgCluster, len(msgChan))

				LoopChan:
					for {
						select {
						case msg := <-msgChan:
							Logger.Warn(msg.RoomID, msg.Appid, msg.TraceId, "ServerRoom.CacheUpdateLoop", "discard", msgCluster, msg.Sender, msg.MsgType, msg.Priority, msg.MsgId, msg.MaxId)
						default:
							break LoopChan
						}
					}
				}
			}
		}
	}
}

// 将消息加到此机房的信息里
func (sr *ServerRoom) AddMsgs(msgs []*logic.ChatRoomMsgRaw) {
	Logger.Debug(sr.RoomID, sr.AppId, sr.Name, "ServerRoom.AddMsgs", len(msgs), sr.getTrace(msgs))
	//扔到待分配队列里，等待分派
	for _, msg := range msgs {
		select {
		case sr.MsgClassifyConsumer <- msg:
			continue
		default:
			Logger.Error(msg.RoomID, msg.Appid, sr.Name, "ServerRoom.AddMsgs", "chan full", msg.Sender, msg.TraceId, msg.MsgType, msg.Priority)
			break
		}
	}
}

// 按消息配置，从自己的对象里缓存的消息里获取count条消息返回
func (sr *ServerRoom) GetMsgs(count int) []*logic.ChatRoomMsgRaw {
	res := []*logic.ChatRoomMsgRaw{}
	//生成一个各消息类型需要取的数目的map
	msgTypeList := map[string]int{}
	for i := 0; i < count; i += 1 {
		msgTypeList[MsgPoliceConf().MsgTypeOrder[sr.MessagesCache.MsgIndex%MsgPoliceConf().OrderLen]] += 1
		sr.MessagesCache.MsgIndex = (sr.MessagesCache.MsgIndex + 1) % MsgPoliceConf().OrderLen
	}
	//循环取各类消息
	for msgCluster, num := range msgTypeList {
		//批量取本类型消息
		residue := sr.getMsgBatch(msgCluster, num, &res)
		//如果本类消息量不够，则取text消息，再取other消息
		if residue != 0 {
			residue = sr.getMsgBatch(TextType, residue, &res)
			if residue != 0 {
				residue = sr.getMsgBatch(OtherType, residue, &res)
			}
			//此部分是剩下所有消息填补的方式
			//for i := 0; i < residue; i += 1 {
			//	sr.getMsg(0, &res)
			//}
		}
	}
	if len(res) == 0 {
		Logger.Debug(sr.RoomID, sr.AppId, sr.Name, "ServerRoom.GetMsgs", "empty cache", count)
		return nil
	} else {
		Logger.Debug(sr.RoomID, sr.AppId, sr.Name, "ServerRoom.GetMsgs", sr.getTrace(res), len(res), count)
		return res
	}
}

//批量取指定的消息
func (sr ServerRoom) getMsgBatch(msgCluster string, count int, resArr *[]*logic.ChatRoomMsgRaw) int {
	Logger.Debug(sr.RoomID, sr.AppId, sr.Name, "ServerRoom.getMsgBatch", msgCluster, count)
	sr.MessagesCache.RLock()
	stopElement := sr.MessagesCache.List.Front()
	//如果是白名单消息,则先拿超级用户消息，若拿够数量可直接返回，则寻找到最大缓存，否则寻找到指定秒数的缓存
	if msgCluster == MsgPoliceConf().WhiteListCluster {
		count = sr.getVipMsg(count, resArr)
		if count == 0 {
			sr.MessagesCache.RUnlock()
			return count
		}
		stopElement = nil
	} else {
		if sr.MessagesCache.List.Len() < MsgPoliceConf().MaxCacheWhiteListSec-1 {
			stopElement = nil
		} else {
			for i := 0; i < MsgPoliceConf().MaxCacheSec; i += 1 {
				stopElement = stopElement.Next()
			}
		}
	}
	//向之前秒数的缓存寻找本类型消息
	for element := sr.MessagesCache.List.Front(); element != stopElement; element = element.Next() {
		currentCache := element.Value.(*map[string]chan *logic.ChatRoomMsgRaw)
	LOOP:
		for count > 0 {
			select {
			case res := <-(*currentCache)[msgCluster]:
				(*resArr) = append((*resArr), res)
				count -= 1
				//继续取当前消息
			default:
				break LOOP
			}
		}
	}
	//返回剩余需要取的消息数
	sr.MessagesCache.RUnlock()
	return count
}

//获取超级用户消息,返回还剩几个未取成功
func (sr *ServerRoom) getVipMsg(count int, resArr *[]*logic.ChatRoomMsgRaw) int {
	for count > 0 {
		select {
		case res := <-sr.VipMsgChan:
			(*resArr) = append((*resArr), res)
			count -= 1
		default:
			Logger.Debug(sr.RoomID, sr.AppId, sr.Name, "ServerRoom.getVipMsg", "vipchan empty", count)
			return count
		}
	}
	return count
}

// 这个聊天室关闭，不再使用
func (sr *ServerRoom) Close() {
	close(sr.close)
	close(sr.MsgClassifyConsumer)
	close(sr.VipMsgChan)
}

//为websocket多取消息专门开设的函数，从第二秒的缓存开始取，防止取走当前秒消息，存在消息不够的隐患
func (sr *ServerRoom) GetSpecificMsg(count int, msgtype int, priority int) []*logic.ChatRoomMsgRaw {
	var msgCluster string
	msgType := strconv.Itoa(msgtype)
	msgPri := strconv.Itoa(priority)
	if cluster, ok := MsgPoliceConf().MsgTypeMap[msgType+"-"+msgPri]; ok {
		msgCluster = cluster
	} else if cluster, ok := MsgPoliceConf().MsgTypeMap[msgType+"-*"]; ok {
		msgCluster = cluster
	} else {
		msgCluster = MsgPoliceConf().MsgTypeMap["--"]
	}

	resArr := []*logic.ChatRoomMsgRaw{}
	sr.MessagesCache.RLock()
	for element := sr.MessagesCache.List.Front().Next(); element != nil; element = element.Next() {
		currentCache := element.Value.(*map[string]chan *logic.ChatRoomMsgRaw)
	LOOP:
		for count > 0 {
			select {
			case res := <-(*currentCache)[msgCluster]:
				resArr = append(resArr, res)
				count -= 1
				//继续取当前消息
			default:
				break LOOP
			}
		}
	}
	sr.MessagesCache.RUnlock()

	Logger.Debug(sr.RoomID, sr.AppId, sr.Name, "ServerRoom.GetSpecificMsg", sr.getTrace(resArr), len(resArr), count)
	return resArr
}

func (sr *ServerRoom) getTrace(msgs []*logic.ChatRoomMsgRaw) string {
	output := make([]string, 0, len(msgs))
	for _, m := range msgs {
		output = append(output, m.Sender+"-"+m.TraceId)
	}
	return strings.Join(output, ",")
}
