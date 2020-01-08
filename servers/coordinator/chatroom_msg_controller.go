package main

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/huajiao-tv/qchat/client/coordinator"
	"github.com/huajiao-tv/qchat/client/gateway"
	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/pb"
)

const (
	PriorityNormal = iota
	PriorityChat
)

var chatRoomAdapterPool *ChatRoomAdapterPool

// 所有房间的消息处理适配器
type ChatRoomAdapterPool struct {
	adapters map[string]*ChatRoomAdapter
	sync.RWMutex
}

func NewChatRoomAdapterPool() *ChatRoomAdapterPool {
	return &ChatRoomAdapterPool{
		adapters: make(map[string]*ChatRoomAdapter),
	}
}

// 获取对应聊天室的消息处理适配器
// 如果不存在就创建一个
func (crap *ChatRoomAdapterPool) Get(appid uint16, roomid string) *ChatRoomAdapter {
	crap.Lock()
	cra, ok := crap.adapters[roomid]
	if !ok {
		cra = NewChatRoomAdapter(appid, roomid)
		crap.adapters[roomid] = cra
		go func() {
			cra.Loop()
			crap.Lock()
			delete(crap.adapters, roomid)
			crap.Unlock()
		}()
	}
	crap.Unlock()
	return cra
}

func (crap *ChatRoomAdapterPool) GetDegradedList(appid uint16) []string {
	crap.RLock()
	defer crap.RUnlock()
	rooms := []string{}
	for roomid, cra := range crap.adapters {
		if cra.Degrade {
			rooms = append(rooms, roomid)
		}
	}
	return rooms
}

//长时间没有消息时，适配器会销毁，重新创建的适配器的Degrade值默认为false
type ChatRoomAdapter struct {
	AppId   uint16
	RoomID  string
	Degrade bool
	Detail  *session.ChatRoomDetail

	// 每一个机房一个处理协程
	ServerRooms map[string]*ServerRoomState

	sync.Mutex
	// 以msgtype和priority做key的map
	MsgPools []*logic.ChatRoomMsgRaw

	// 统计连续多少次没有获取消息了
	// 当超过一定数量时，就直接销毁
	EmptyCount    int
	LastPushCount int
	LastLeftCount int
}

func NewChatRoomAdapter(appid uint16, roomid string) *ChatRoomAdapter {
	cra := &ChatRoomAdapter{
		AppId:  appid,
		RoomID: roomid,
	}
	return cra
}

// 设置聊天室降级
func (cra *ChatRoomAdapter) ForceDegrade(d bool) {
	cra.Lock()
	cra.Degrade = d
	cra.Unlock()
}

// 如果消息总数超过一定数量，将会直接丢掉消息
func (cra *ChatRoomAdapter) AddMsg(msg *logic.ChatRoomMsgRaw) {
	cra.Lock()
	if len(cra.MsgPools) < netConf().MaxKeepMsgs*netConf().MsgPushInterval/1000 {
		cra.MsgPools = append(cra.MsgPools, msg)
		cra.Unlock()
	} else {
		cra.Unlock()
		Logger.Warn(msg.RoomID, msg.Appid, msg.TraceId, "ChatRoomAdapter.AddMsg", "max len reached", msg.Sender, msg.MsgType, msg.Priority)
	}
}

// 每个房间消息适配器处理循环
func (cra *ChatRoomAdapter) Loop() {
	lastMPI := netConf().MsgPushInterval
	if lastMPI == 0 {
		lastMPI = 1000
	}
	pushTicker := time.NewTicker(time.Duration(lastMPI) * time.Millisecond)
	defer pushTicker.Stop()
	// 发送定时的计数器
	counter := 0
	rand := time.Now().UnixNano()
	for {
		t := <-pushTicker.C
		counter += 1
		// 如果配置的时间有更新，需要重新生成定时器
		if netConf().MsgPushInterval != 0 && netConf().MsgPushInterval != lastMPI {
			pushTicker.Stop()
			lastMPI = netConf().MsgPushInterval
			pushTicker = time.NewTicker(time.Duration(lastMPI) * time.Millisecond)
		}
		// 更新send和destroy的次数
		countForSend := netConf().MsgSendInterval / lastMPI
		countForDestroy := netConf().AdapterLiveDuration * 1000 / lastMPI

		cra.Lock()
		degrade := cra.Degrade
		msgs := cra.MsgPools
		if len(msgs) == 0 {
			cra.Unlock()
			cra.EmptyCount += 1
			if cra.EmptyCount >= countForDestroy || countForDestroy == 0 {
				// 超过一定时间没有消息的聊天室，将销毁这个adapter
				Logger.Trace(cra.RoomID, cra.AppId, cra.EmptyCount, "cra.Loop", "finish", countForDestroy)
				break
			} else {
				goto send
			}
		} else {
			cra.EmptyCount = 0
		}
		cra.MsgPools = make([]*logic.ChatRoomMsgRaw, 0)
		cra.Unlock()

		cra.LastPushCount = len(msgs)
		// 全局过滤规则,包括普通聊天室的
		msgs = cra.CommonFilter(msgs)
		if len(msgs) == 0 {
			goto send
		}
		cra.LastLeftCount = len(msgs)
		Logger.Trace(cra.RoomID, cra.AppId, cra.LastPushCount, "cra.Loop", cra.LastLeftCount, cra.Detail.GatewayAddrs, t, lastMPI, rand)

		if _, ok := logic.NetGlobalConf().BigRoom[cra.RoomID]; !ok && !degrade && cra.Detail.ConnCount() < netConf().BigRoomMember {
			// 这块没有continue的原因是，可能原来大聊天室里还有剩下的消息没有发送
			cra.SendNormal(msgs)
		} else {
			cra.PushToServerRoom(msgs)
		}

	send:
		// 如果达到push次数，那就开始发送流程
		if counter >= countForSend {
			counter = 0
			cra.SendComplex()
		}
	}
	for _, sr := range cra.ServerRooms {
		sr.Close()
	}
}

// 返回这个coordinator负责的gateway列表
func filterGateways(gateways map[string]int) map[string]int {
	result := make(map[string]int)
	if len(logic.NetGlobalConf().CoordinatorArea) == 0 {
		for gw, c := range gateways {
			if c > 0 {
				result[gw] = c
			}
		}
		return result
	}
	msn := DynamicConf().MyServerNames
	if len(msn) == 0 {
		return result
	}
	for gw, c := range gateways {
		sr := logic.DynamicConf().GatewaySrMap[gw]
		if sr == "" || c == 0 {
			continue
		}
		if msn[sr] {
			result[gw] = c
		}
	}
	return result
}

// 获取一个adapter下的一个机房对象,如果没有，就创建
func (cra *ChatRoomAdapter) GetServerRoom(sr string) *ServerRoomState {
	if cra.ServerRooms == nil {
		cra.ServerRooms = make(map[string]*ServerRoomState)
	}
	if _, ok := cra.ServerRooms[sr]; !ok {
		cra.ServerRooms[sr] = NewServerRoomState(cra.AppId, cra.RoomID, sr)
	}
	return cra.ServerRooms[sr]
}

// 每一个聊天室机房对象，收集各种数据
type ServerRoomState struct {
	LastMsgCount  int
	LastAdjust    time.Time
	GatewayAddrs  map[string]int
	LastPushCount int
	*ServerRoom
}

func NewServerRoomState(appid uint16, room string, sr string) *ServerRoomState {
	return &ServerRoomState{
		LastMsgCount: netConf().DefaultMsgSend * netConf().MsgSendInterval / 1000,
		LastAdjust:   time.Now(),
		ServerRoom:   NewServerRoom(appid, room, sr),
	}
}

func getMinMsgSend(serverRoomName string) int {
	min, _ := netConf().MinMsgSendSr[serverRoomName]
	if min == 0 {
		min = netConf().MinMsgSend
	}

	return min * netConf().MsgSendInterval / 1000
}

func getMaxMsgSend(serverRoomName string) int {
	max, _ := netConf().MaxMsgSendSr[serverRoomName]
	if max == 0 {
		max = netConf().MaxMsgSend
	}

	return max * netConf().MsgSendInterval / 1000
}

func (srs *ServerRoomState) CalCount(roomid string, serverRoomName string) int {
	if _, ok := DynamicConf().StaticMsgSend[roomid]; ok {
		if count, ok := DynamicConf().StaticMsgSend[roomid][serverRoomName]; ok && count != 0 {
			srs.LastMsgCount = count * netConf().MsgSendInterval / 1000
			return srs.LastMsgCount
		}
	}
	if time.Now().Sub(srs.LastAdjust) < time.Duration(netConf().AdjustInterval)*time.Second {
		return srs.LastMsgCount
	}
	overflow := serverRoomFlowService.OverFlow(serverRoomName)
	if overflow == 0 {
		return srs.LastMsgCount
	} else if overflow > 0 {
		// @todo  反馈等级做
		srs.LastMsgCount -= netConf().MsgCountDecreaseStep
		if min := getMinMsgSend(serverRoomName); srs.LastMsgCount < min {
			srs.LastMsgCount = min
		}
	} else {
		srs.LastMsgCount += netConf().MsgCountIncreaseStep
		if max := getMaxMsgSend(serverRoomName); srs.LastMsgCount > max {
			srs.LastMsgCount = max
		}
	}
	srs.LastAdjust = time.Now()
	return srs.LastMsgCount
}
func (srs *ServerRoomState) WebMoreMsgCount(roomid string, serverRoomName string) int {
	if count, _ := netConf().WebMoreMsgCount[serverRoomName]; count > 0 {
		return count * netConf().MsgSendInterval / 1000
	}
	return 0
}

// 针对每一个gateway做策略
func (cra *ChatRoomAdapter) SendComplex() {
	Logger.Trace(cra.RoomID, cra.AppId, "", "SendComplex", len(cra.ServerRooms))
	for srn, srs := range cra.ServerRooms {
		count := srs.CalCount(cra.RoomID, srn)
		msgs := srs.GetMsgs(count)
		if len(msgs) == 0 || len(srs.GatewayAddrs) == 0 {
			continue
		}
		adapterStats.Add(cra.RoomID, srn, cra, srs, len(msgs))
		if netConf().CompressComplex {
			SendMsgBat(msgs, srs.GatewayAddrs, cra.Detail, false)
		} else {
			SendMsgBatWithoutCompress(msgs, srs.GatewayAddrs, cra.Detail, false)
		}
		if count := srs.WebMoreMsgCount(cra.RoomID, srn); count > 0 {
			msgs := srs.GetSpecificMsg(count, netConf().WebMoreMsgType, netConf().WebMoreMsgPriority)
			if len(msgs) == 0 {
				continue
			}
			if netConf().CompressWeb {
				SendMsgBat(msgs, srs.GatewayAddrs, cra.Detail, true)
			} else {
				SendMsgBatWithoutCompress(msgs, srs.GatewayAddrs, cra.Detail, true)
			}
		}
	}
}

var adapterStats *AdapterStats

type AdapterStats struct {
	// key是roomid-servername
	stats map[string]map[string]*coordinator.AdapterStat
	sync.RWMutex
}

func NewAdapterStats() *AdapterStats {
	return &AdapterStats{
		stats: make(map[string]map[string]*coordinator.AdapterStat),
	}
}

func (as *AdapterStats) Add(roomid, sr string, cra *ChatRoomAdapter, srs *ServerRoomState, realGetCount int) {
	MemCount := 0
	for _, c := range srs.GatewayAddrs {
		MemCount += c
	}
	as.Lock()
	if _, ok := as.stats[roomid]; !ok {
		as.stats[roomid] = make(map[string]*coordinator.AdapterStat)
	}
	as.stats[roomid][sr] = &coordinator.AdapterStat{
		RecordTime:     time.Now(),
		PushInterval:   netConf().MsgPushInterval,
		TotalPushCount: cra.LastPushCount,
		RealPushCount:  srs.LastPushCount,
		RealPushPerS:   srs.LastPushCount * 1000 / netConf().MsgPushInterval,
		SendInterval:   netConf().MsgSendInterval,
		GetCount:       srs.LastMsgCount,
		RealGetCount:   realGetCount,
		RealGetPerS:    realGetCount * 1000 / netConf().MsgSendInterval,
		MemCount:       MemCount,
		TotalMemCount:  cra.Detail.ConnCount(),
		LastAdjust:     srs.LastAdjust,
		FlowCount:      serverRoomFlowService.GetFlow(sr),
		OverFlow:       serverRoomFlowService.OverFlow(sr),
	}
	as.Unlock()
}

func (as *AdapterStats) GetByRoomid(roomid string) map[string]*coordinator.AdapterStat {
	result := make(map[string]*coordinator.AdapterStat)
	as.RLock()
	if _, ok := as.stats[roomid]; ok {
		for sr, a := range as.stats[roomid] {
			result[sr] = a
		}
	}
	as.RUnlock()
	return result
}
func (as *AdapterStats) GetAll() map[string]map[string]*coordinator.AdapterStat {
	result := make(map[string]map[string]*coordinator.AdapterStat)
	now := time.Now()
	as.Lock()
	for r, stats := range as.stats {
		t := make(map[string]*coordinator.AdapterStat)
		for sr, a := range stats {
			if now.Sub(a.RecordTime) > time.Duration(netConf().StatFadeTime)*time.Second {
				delete(stats, sr)
			}
			t[sr] = a
		}
		if len(t) > 0 {
			result[r] = t
		} else {
			delete(as.stats, r)
		}
	}
	as.Unlock()
	return result
}

func (cra *ChatRoomAdapter) PushToServerRoom(msgs []*logic.ChatRoomMsgRaw) {
	// 重置gatewway信息
	for _, srs := range cra.ServerRooms {
		srs.GatewayAddrs = nil
	}
	for sr, managers := range logic.NetGlobalConf().GatewayRpcsSr {
		for _, manager := range managers {
			if c, ok := cra.Detail.GatewayAddrs[manager]; ok && c != 0 {
				if cra.GetServerRoom(sr).GatewayAddrs == nil {
					cra.ServerRooms[sr].GatewayAddrs = make(map[string]int)
				}
				cra.ServerRooms[sr].GatewayAddrs[manager] = c
			}
		}
	}
	for _, srs := range cra.ServerRooms {
		if srs.GatewayAddrs != nil {
			srs.AddMsgs(msgs)
			srs.LastPushCount = len(msgs)
		}
	}
}

func checkDiscardMessagesDetailPolicy(msgtype, priority, members int) bool {
	policy := DynamicConf().NormalChatRoomDiscardMessagePolicy

	var useDefault bool
	key := fmt.Sprintf("%v-%v", msgtype, priority)
	if _, ok := policy[key]; !ok {
		// 默认配置
		useDefault = true
	}

	var percentage int
	switch useDefault {
	case true:
		for numofppl, p := range netConf().CrDropMsgsDtDefault {
			if members >= numofppl && p > percentage {
				percentage = p
			}
		}

	case false:
		var percentageKey int
		for numofppl, _ := range policy[key] {
			if members >= numofppl && numofppl > percentageKey {
				percentageKey = numofppl
			}
		}

		percentage = policy[key][percentageKey]
	}

	// 按千分比丢弃
	ri := rand.Intn(1000)
	if ri < percentage {
		return true
	}

	return false
}

// priority 可能取值范围 加入退出消息（0），聊天消息（1），礼物消息（101），红包消息（201）
// ChatroomDegradePolicy 配置项目 DegradeToZero(0)，DegradeChatToZero（1），DegradeGiftToZero（101），DegradeAllToZero（201）
func checkPolicy(room *session.ChatRoomDetail, priority int) bool {
	switch netConf().ChatroomDegradePolicy {
	case PriorityNormal:
		if room.ConnCount() < netConf().ChatroomDegradeMembers {
			return priority > PriorityNormal
		} else if priority > PriorityChat {
			return true
		}
	default:
		if priority > netConf().ChatroomDegradePolicy {
			return true
		}
	}
	return false
}

// 发送普通聊天室的消息
// 如果saver换到多中心之后，如何解决编号递增的问题
// @todo 保存消息到redis.消息编号，优先级等, 下发router
func (cra *ChatRoomAdapter) SendNormal(msgs []*logic.ChatRoomMsgRaw) {
	detail := cra.Detail
	Logger.Trace(cra.RoomID, cra.AppId, len(msgs), "SendNormal", len(detail.GatewayAddrs))
	if len(detail.GatewayAddrs) == 0 {
		return
	}

	compressMsgs := make([]*logic.ChatRoomMsgRaw, 0, len(msgs))
	for _, msg := range msgs {
		// 普通聊天室消息降级，当消息被降级时，直接丢弃
		membersCount := detail.ConnCount()
		if checkDiscardMessagesDetailPolicy(msg.MsgType, msg.Priority, membersCount) {
			Logger.Warn(msg.RoomID, msg.Appid, msg.TraceId, "SendNormal", "discard policy", msg.Sender, msg.MsgType, msg.Priority, msg.MsgId, msg.MaxId, membersCount)
			continue
		}

		var pullLost bool
		msg.MaxId = detail.MaxID
		message := &logic.ChatRoomMessage{
			RoomID:      msg.RoomID,
			Sender:      msg.Sender,
			Appid:       msg.Appid,
			MsgType:     0, //msg.MsgType,
			MsgContent:  []byte(msg.MsgContent),
			RegMemCount: detail.Registered(),
			MemCount:    detail.MemberCount(),
			TimeStamp:   time.Now().UnixNano() / 1e6,
			MaxID:       detail.MaxID,
		}

		// 检查是否多发，如果是，直接不编号发送
		if len(logic.NetGlobalConf().CoordinatorArea) != 0 {
			goto Send
		}

		pullLost = true
		// 检查是否 pullLost，如果否，直接不编号发送
		if len(logic.NetGlobalConf().PullLost) > 0 {
			if pl, ok := logic.NetGlobalConf().PullLost[msg.RoomID]; ok {
				pullLost = pl
			} else if pl, ok := logic.NetGlobalConf().PullLost["default"]; ok {
				pullLost = pl
			}
		}
		if !pullLost {
			goto Send
		}

		// 检查优先级是否高于设定的值，如果不高于，直接不编号发送
		message.Priority = checkPolicy(detail, msg.Priority)
		if !message.Priority {
			goto Send
		}

		// 存储／编号
		if msgid, err := saver.CacheChatRoomMessage(message); err != nil {
			Logger.Error(msg.RoomID, msg.Appid, msg.TraceId, "SendNormal", "saver.CacheChatRoomMessage error", err, msg.Sender, msg.MsgType, msg.Priority)
		} else {
			message.MsgID, message.MaxID = msgid, msgid
			msg.MsgId, msg.MaxId = msgid, msgid
			detail.MaxID = msgid
		}

	Send:
		if !netConf().CompressNormal {
			// 需要过滤只发自己管理的gateway
			if err := router.SendChatRoomNotify(message, detail.GatewayAddrs, msg.TraceId); err != nil {
				Logger.Warn(msg.RoomID, msg.Appid, msg.TraceId, "SendNormal", "router.SendChatRoomNotify error", msg.Sender, msg.MsgType, msg.Priority, msg.MsgId, msg.MaxId, detail.ConnCount(), err.Error())
			} else {
				Logger.Debug(cra.RoomID, cra.AppId, msg.TraceId, "SendNormal", msg.Sender, msg.MsgType, msg.Priority, msg.MsgId, msg.MaxId, detail.ConnCount())
			}
		} else {
			compressMsgs = append(compressMsgs, msg)
		}
	}
	if len(compressMsgs) > 0 {
		SendMsgBat(compressMsgs, detail.GatewayAddrs, detail, false)
	}
}

// 基础的过滤策略，比如过了一定人数之后某一些消息就不发了等
func (cra *ChatRoomAdapter) CommonFilter(msgs []*logic.ChatRoomMsgRaw) []*logic.ChatRoomMsgRaw {
	memberCount := cra.Detail.ConnCount()
	var key int
	var matched bool
	for mc, _ := range DynamicConf().CommonFilter {
		if memberCount >= mc && mc > key {
			matched = true
			key = mc
		}
	}

	if !matched {
		return msgs
	}

	filter := DynamicConf().CommonFilter[key]

	newMsgs := make([]*logic.ChatRoomMsgRaw, 0, len(msgs))
	for _, msg := range msgs {
		for _, rule := range filter {
			switch {
			case rule.Type == "*" && rule.Priority == "*":
				goto NEXT
			case rule.Type == "*" && rule.Priority != "*":
				filterPriority, _ := strconv.Atoi(rule.Priority)
				if msg.Priority < filterPriority {
					goto NEXT
				}
			case rule.Type != "*" && rule.Priority == "*":
				if rule.Type == strconv.Itoa(msg.MsgType) {
					goto NEXT
				}
			case rule.Type != "*" && rule.Priority != "*":
				filterPriority, _ := strconv.Atoi(rule.Priority)
				if rule.Type == strconv.Itoa(msg.MsgType) && msg.Priority < filterPriority {
					goto NEXT
				}
			}
		}
		newMsgs = append(newMsgs, msg)
		continue

	NEXT:
		Logger.Warn(cra.RoomID, cra.AppId, msg.TraceId, "CommonFilter", "discard", msg.Sender, msg.MsgType, msg.Priority, msg.MsgId, msg.MaxId, memberCount)
	}
	return newMsgs
}

// 获取房间的详细信息，处理错误日志记录
func getDetail(appid uint16, roomid string) (*session.ChatRoomDetail, error) {
	detail, err := saver.QueryChatRoomDetail(roomid, appid)
	if err != nil {
		Logger.Error(roomid, appid, "", "cradapter", "saver.QueryChatRoomDetail error", err.Error())
		return nil, err
	} else if detail == nil {
		Logger.Error(roomid, appid, "", "cradapter", "saver.QueryChatRoomDetail error", "room not exist")
		return nil, errors.New("room not exist")
	} else {
		return detail, nil
	}
}

func SendMsgBat(msgs []*logic.ChatRoomMsgRaw, gwys map[string]int, detail *session.ChatRoomDetail, isWeb bool) {
	blockSize := netConf().MsgSendBlockSize
	for i := 0; i < len(msgs); i += blockSize {
		if i+blockSize >= len(msgs) {
			sendMsgBat(msgs[i:len(msgs)], gwys, detail, isWeb)
			break
		} else {
			sendMsgBat(msgs[i:i+blockSize], gwys, detail, isWeb)
		}
	}
}

func SendMsgBatWithoutCompress(msgs []*logic.ChatRoomMsgRaw, gwys map[string]int, detail *session.ChatRoomDetail, isWeb bool) {
	msgNotifys := make([]*logic.ChatRoomMessageNotify, 0, len(msgs))
	for _, m := range msgs {
		msgNotifys = append(msgNotifys, &logic.ChatRoomMessageNotify{
			&logic.ChatRoomMessage{
				RoomID:      m.RoomID,
				Sender:      m.Sender,
				Appid:       m.Appid,
				MsgType:     0,
				MsgContent:  []byte(m.MsgContent),
				RegMemCount: detail.Registered(),
				MemCount:    detail.MemberCount(),
				MsgID:       m.MsgId,
				MaxID:       m.MaxId, //detail.MaxID,
				TimeStamp:   time.Now().UnixNano() / 1e6,
				Priority:    false,
			},
			nil,
			m.TraceId,
			0,
			0,
		})
		Logger.Debug(m.RoomID, m.Appid, m.TraceId, "SendMsgBatWithoutCompress", m.Sender, m.MsgType, m.Priority, m.MsgId, m.MaxId, isWeb, detail.ConnCount())
	}
	if err := router.SendChatRoomNotifyBatch(msgNotifys, gwys, isWeb); err != nil {
		Logger.Error(detail.RoomID, detail.AppID, isWeb, "sendMsgBatWithoutCompress", "SendChatRoomNotifyBatch error", err)
	}
}

// 批量发送消息
func sendMsgBat(msgs []*logic.ChatRoomMsgRaw, gwys map[string]int, detail *session.ChatRoomDetail, isWeb bool) {
	if len(msgs) == 0 {
		return
	}
	notify := make([]*pb.ChatRoomMNotify, 0, len(msgs))
	for _, msg := range msgs {
		m, err := compressMsg(msg, detail)
		if err != nil {
			Logger.Error(detail.RoomID, detail.AppID, isWeb, "sendMsgBat", "compressMsg err", err.Error(), msg.Sender, msg.MsgType, msg.Priority)
		} else {
			notify = append(notify, m)
		}
		Logger.Debug(msg.RoomID, msg.Appid, msg.TraceId, "sendMsgBat", msg.Sender, msg.MsgType, msg.Priority, msg.MsgId, msg.MaxId, isWeb, detail.ConnCount())
	}
	packet := &pb.ChatRoomPacket{
		Roomid: []byte(detail.RoomID),
		Appid:  proto.Uint32(uint32(detail.AppID)),
		ToUserData: &pb.ChatRoomDownToUser{
			Result:      proto.Int32(0),
			Payloadtype: proto.Uint32(pb.CR_PAYLOAD_COMPRESSED),
			Multinotify: notify,
		},
	}
	content, err := proto.Marshal(packet)
	if err != nil {
		Logger.Error(detail.RoomID, detail.AppID, isWeb, "sendMsgBat", "pb marshal", err.Error())
		return
	}
	ximp, err := pb.CreateMsgNotify("chatroom", content, int64(0), "", "", 0)
	if err != nil {
		Logger.Error(detail.RoomID, detail.AppID, isWeb, "sendMsgBat", "pb.CreateMsgNotify", err.Error())
		return
	}
	ximp.TimeStamp = time.Now().UnixNano() / 1e6
	ximp.TraceId = fmt.Sprintf("COMPRESS-%s-%d", msgs[0].TraceId, len(msgs))

	var tag string
	if !isWeb {
		tag = logic.GenerateChatRoomTag(detail.AppID, detail.RoomID)
	} else {
		tag = logic.GenerateWebChatRoomTag(detail.AppID, detail.RoomID)
	}
	gwr := &router.GwResp{
		XimpBuff: ximp,
	}
	for gwy, _ := range gwys {
		// todo: go?
		go func(g string) {
			if err := gateway.DoOperations(g, []string{tag}, "", nil, gwr); err != nil {
				Logger.Error(detail.RoomID, detail.AppID, ximp.TraceId, "sendMsgBat", "DoOperations", err, g)
			}
		}(gwy)
	}
	Logger.Trace(detail.RoomID, detail.AppID, ximp.TraceId, "sendMsgBat", tag, gwys)
}

// 消息压缩
func compressMsg(msg *logic.ChatRoomMsgRaw, detail *session.ChatRoomDetail) (*pb.ChatRoomMNotify, error) {
	m := &logic.ChatRoomMessage{
		RoomID:      msg.RoomID,
		Sender:      msg.Sender,
		Appid:       msg.Appid,
		MsgType:     0,
		MsgContent:  []byte(msg.MsgContent),
		RegMemCount: detail.Registered(),
		MemCount:    detail.MemberCount(),
		MsgID:       msg.MsgId,
		MaxID:       msg.MaxId, //detail.MaxID,
		TimeStamp:   time.Now().UnixNano() / 1e6,
		Priority:    false,
	}
	data, err := pb.CompressChatRoomNewMsg(m)
	if err != nil {
		return nil, err
	}
	return &pb.ChatRoomMNotify{
		Type:        proto.Int32(pb.CR_PAYLOAD_INCOMING_MSG),
		Data:        data,
		Regmemcount: proto.Int32(int32(detail.Registered())),
		Memcount:    proto.Int32(int32(detail.MemberCount())),
	}, nil
}
