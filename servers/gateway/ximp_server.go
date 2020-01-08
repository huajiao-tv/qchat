package main

import (
	"sync"
	"time"

	"strconv"

	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/network"
)

const (
	MSGTYPE_NORMAL = iota
	MSGTYPE_PRIORITY
	MSGTYPE_DIRECT
)

// Connection类型的派生类型，用于具体业务
type XimpConnection struct {
	*Connection
	conn            network.INetwork
	clientMsgQueue  chan *network.XimpBuffer // 发给client的消息
	priorityQueue   chan *network.XimpBuffer // 发给client的高优先级消息
	quitWriteSignal chan bool
	quitReadSignal  chan bool

	sync.RWMutex                       // 以下成员的保护锁
	property         map[string]string // 连接属性，可以从后端设置，只在read协程里用
	heartBeatTimeout time.Duration     // 心跳间隔
	rkey             []byte            // 加密需要的对称密钥
	isClose          bool              // 用来确认这个连接是否已经关闭，防止无效的引用
	tags             map[string]bool   // 用于保存一份此连接的所有tag信息，用于退出时清空tag
}

func newXimpConnection(c network.INetwork) *XimpConnection {
	return &XimpConnection{
		conn:             c,
		Connection:       newConnection(),
		clientMsgQueue:   make(chan *network.XimpBuffer, staticConf.ClientQueueLen),
		priorityQueue:    make(chan *network.XimpBuffer, staticConf.ClientQueueLen),
		quitWriteSignal:  make(chan bool, 1),
		quitReadSignal:   make(chan bool, 1),
		heartBeatTimeout: staticConf.HeartBeatTimeout + staticConf.HeartBeatBaseTimeout,
		rkey:             []byte{},
		property:         map[string]string{"ConnectionType": c.Type()},
		tags:             make(map[string]bool),
	}
}

func (this *XimpConnection) Serve() {
	this.serveLoop()
	this.logout()
}

func (this *XimpConnection) GetTraceId() string {
	return logic.GetTraceId(netConf().Manager, this.GetId())
}

func (this *XimpConnection) serveLoop() {
	go this.serveRead()
	for {
		select {
		case toClientBuf := <-this.clientMsgQueue:
			select {
			case toClientBuf := <-this.priorityQueue:
				if err := this.sendToClient(*toClientBuf, MSGTYPE_PRIORITY); err != nil {
					Logger.Debug(this.GetSender(), this.GetAppid(), this.GetTraceId(), "serveLoop", "sendToClient(priorityQueue) error", err, toClientBuf.TraceId)
					this.SetCloseReason("sendPrioMsg:" + err.Error())
					this.quitReadSignal <- true
					return
				}
			default:
			}
			if err := this.sendToClient(*toClientBuf, MSGTYPE_NORMAL); err != nil {
				Logger.Debug(this.GetSender(), this.GetAppid(), this.GetTraceId(), "serveLoop", "sendToClient error", err, toClientBuf.TraceId)
				this.SetCloseReason("sendMsg:" + err.Error())
				this.quitReadSignal <- true
				return
			}
		case toClientBuf := <-this.priorityQueue:
			if err := this.sendToClient(*toClientBuf, MSGTYPE_PRIORITY); err != nil {
				Logger.Debug(this.GetSender(), this.GetAppid(), this.GetTraceId(), "serveLoop", "sendToClient(priorityQueue) error", err, toClientBuf.TraceId)
				this.SetCloseReason("sendPrioMsg:" + err.Error())
				this.quitReadSignal <- true
				return
			}
		case <-this.quitWriteSignal:
			this.quitReadSignal <- true
			return
		}
	}
}

func (this *XimpConnection) SetRKey(rkey []byte) {
	this.Lock()
	this.rkey = rkey
	this.Unlock()
}
func (this *XimpConnection) GetRKey() []byte {
	this.RLock()
	defer this.RUnlock()
	return this.rkey
}

// 两个方便的方法获取属性
func (this *XimpConnection) GetSender() string {
	this.RLock()
	defer this.RUnlock()
	return this.property["Sender"]
}

func (this *XimpConnection) GetConnectionType() (typ string) {
	this.RLock()
	typ = this.property["ConnectionType"]
	this.RUnlock()
	return
}
func (this *XimpConnection) GetAppid() string {
	this.RLock()
	defer this.RUnlock()
	return this.property["Appid"]
}

func (this *XimpConnection) SetHeartBeatTimeout(t time.Duration) {
	this.Lock()
	defer this.Unlock()
	this.heartBeatTimeout = t + staticConf.HeartBeatBaseTimeout
}

func (this *XimpConnection) GetHeartBeatTimeout() time.Duration {
	this.RLock()
	defer this.RUnlock()
	return this.heartBeatTimeout
}

func (this *XimpConnection) close() {
	this.Lock()
	defer this.Unlock()
	this.isClose = true
}

func (this *XimpConnection) SetCloseReason(r string) {
	this.Lock()
	if this.property["CloseReason"] == "" {
		this.property["CloseReason"] = r
	}
	this.Unlock()
}

func (this *XimpConnection) IsClose() bool {
	this.RLock()
	defer this.RUnlock()
	return this.isClose
}

func (this *XimpConnection) GetPropCopy() map[string]string {
	this.RLock()
	propCopy := make(map[string]string, len(this.property)+2)
	for k, v := range this.property {
		propCopy[k] = v
	}
	propCopy["Rkey"] = string(this.rkey)
	propCopy["ClientIp"] = this.conn.RemoteIp()
	this.RUnlock()
	return propCopy
}

func (this *XimpConnection) SetProperty(prop map[string]string) {
	this.Lock()
	for k, v := range prop {
		this.property[k] = v
	}
	this.Unlock()
}

func (this *XimpConnection) Close() {
	select {
	case this.quitWriteSignal <- true:
	default:
	}
}

func (this *XimpConnection) serveRead() {
	for {
		ximpBuf := network.NewXimpBuffer()
		if err := ximpBuf.ReadFrom(true, this.conn, this.GetHeartBeatTimeout()); err != nil {
			Logger.Debug(this.GetSender(), this.GetAppid(), this.GetTraceId(), "serveRead", "readfrom error", err)
			this.SetCloseReason("ReadFrom:" + err.Error())
			this.Close()
			return
		}
		monitor.AddRead(1)
		select {
		case <-this.quitReadSignal:
			return
		default:
		}

		this.RoutePackage(ximpBuf)
	}
}

func (this *XimpConnection) SetTags(tags map[string]bool) {
	this.Lock()
	for k, v := range tags {
		if v {
			this.tags[k] = true
		} else {
			delete(this.tags, k)
		}
	}
	this.Unlock()
}

/**
 * async参数：是否需要异步执行一些 耗时/加锁 的操作
 * 由于这个函数可能在锁定tagpool的前提下调用，所以如果async传true，
 * 那么有加锁的操作将异步进行(settag),耗时的操作也将异步(sendtoclient)
 */
func (this *XimpConnection) Operate(response *router.GwResp, async bool) {
	if response == nil || this.IsClose() {
		return
	}
	if response.Property != nil {
		this.SetProperty(response.Property)
	}
	if response.Rkey != nil && len(response.Rkey) > 0 {
		this.SetRKey(response.Rkey)
	}
	if notEncrypt, ok := response.Property["NotEncrypt"]; ok && notEncrypt == "1" {
		this.SetRKey(nil)
	}
	if response.HeartBeatTimeout != 0 {
		this.SetHeartBeatTimeout(response.HeartBeatTimeout)
	}
	if response.Tags != nil {
		settags := func() {
			this.SetTags(response.Tags)
			for k, v := range response.Tags {
				if v {
					p := tagPools.GetAndCreatePool(k)
					p.Add(this)
				} else {
					tagPools.Del(k, this.GetId())
				}
			}
		}
		if async {
			go settags()
		} else {
			settags()
		}
	}
	if response.XimpBuff != nil {
		if len(response.Err) == 0 {
			if response.Priority {
				this.writeClientPriorityBuffer(response.XimpBuff)
			} else {
				this.writeClientBuffer(response.XimpBuff)
			}
		} else {
			sendtoclient := func() {
				if err := this.sendToClient(*response.XimpBuff, MSGTYPE_DIRECT); err != nil {
					Logger.Debug(this.GetSender(), this.GetAppid(), this.GetTraceId(), "Operate", "sendToClient error", err, response.XimpBuff.TraceId)
				}
			}
			if async {
				go sendtoclient()
			} else {
				sendtoclient()
			}
		}
	}
	if response.Actions != nil && len(response.Actions) != 0 {
		for _, a := range response.Actions {
			switch a {
			case router.DisconnectAction:
				this.SetCloseReason("Operate")
				this.Close()
			default:
			}
		}
	}
}

// 发送给router
func (this *XimpConnection) RoutePackage(buf *network.XimpBuffer) error {
	// 一个连接的默认key是由appid和cversion决定
	if buf.HasHeader {
		if rkey := logic.GetDefaultKey(buf.Appid, buf.CVersion); len(rkey) != 0 && len(this.GetRKey()) == 0 {
			this.SetRKey(rkey)
		}
	}

	if rkey := this.GetRKey(); !buf.IsDecrypt && len(rkey) != 0 {
		if err := buf.Decrypt(rkey); err != nil {
			Logger.Error(this.GetSender(), this.GetAppid(), this.GetTraceId(), "RoutePackage", "decrypt ximpbuffer err", err)
			return err
		}
	}
	Logger.Trace(this.GetSender(), this.GetAppid(), this.GetTraceId(), "RoutePackage", "req", buf.String())
	if response, err := router.RoutePackage(buf, this.GetPropCopy(), netConf().Manager, this.GetId()); err != nil {
		Logger.Error(this.GetSender(), this.GetAppid(), this.GetTraceId(), "RoutePackage", "router.RoutePackage error", err)
		return err
	} else {
		Logger.Trace(this.GetSender(), this.GetAppid(), this.GetTraceId(), "RoutePackage", "resp", response.String())
		this.Operate(response, false)
		return nil
	}

}

// 将消息放入写入客户端队列
func (this *XimpConnection) writeClientBuffer(cbuf *network.XimpBuffer) {
	select {
	case this.clientMsgQueue <- cbuf:
	default:
		// 因为由拉取机制，所以当chan满时，直接丢弃
		Logger.Warn(this.GetSender(), this.GetAppid(), this.GetTraceId(), "writeClientBuffer", "client channel is full", cbuf.String(), cbuf.TraceId)
	}
}
func (this *XimpConnection) writeClientPriorityBuffer(cbuf *network.XimpBuffer) {
	select {
	case this.priorityQueue <- cbuf:
	default:
		// 这块队列也可能满，大部分原因是用户的网络不佳，消耗过慢
		Logger.Warn(this.GetSender(), this.GetAppid(), this.GetTraceId(), "writeClientPriorityBuffer", "client channel is full", cbuf.String(), cbuf.TraceId)
	}
}

// 发送给客户端用户
func (this *XimpConnection) sendToClient(buf network.XimpBuffer, msgtype int) error {
	if netConf().IgnoreWebsocket && this.GetConnectionType() == network.WebSocketNetwork {
		return nil
	}
	if buf.IsDecrypt {
		if err := buf.Encrypt(this.GetRKey()); err != nil {
			return err
		}
	}
	Logger.Debug(this.GetSender(), this.GetAppid(), this.GetTraceId(), "sendToClient", msgtype, buf.TraceId, buf.String())

	bs, err := buf.Encode()
	if err != nil {
		return err
	}

	monitor.AddWrite(1)
	/*
		// 去掉写超时来减少cpu的占用
		if err := this.tcpConn.SetWriteDeadline(time.Time{}); err != nil {
			return err
		}
		if _, err := this.tcpConn.Write(bs); err != nil {
			return err
		}
		return nil
	*/

	// 流量统计
	requestStat.AtomicAddThisSecondFlow(uint64(len(bs) * 8))

	if _, err := this.conn.WriteBytes(bs, logic.StaticConf.ExternalWriteTimeout); err != nil {
		return err
	}
	return nil
}

func (this *XimpConnection) logout() {
	// count request response time if need
	if netConf().StatResponseTime {
		appid, err := strconv.ParseUint(this.GetAppid(), 10, 16)
		if err != nil || appid == 0 {
			appid = uint64(logic.DEFAULT_APPID)
		}
		countFunc := countCloseResponseTime(this.GetSender(), this.GetTraceId(), "logout",
			uint16(appid))
		defer countFunc()
	}

	response, err := router.Logout(this.GetPropCopy(), netConf().Manager, this.GetId())
	if err != nil {
		Logger.Error(this.GetSender(), this.GetAppid(), this.GetTraceId(), "logout", "router.Logout error", err)
		// count close connection operation
		requestStat.AtomicAddCloseConns(1)
	} else {
		this.Operate(response, false)
		Logger.Trace(this.GetSender(), this.GetAppid(), this.GetTraceId(), "logout", response.Err, response.Tags)
		// count close connection operation
		requestStat.AtomicAddCloseConnFails(1)
	}
	this.close()
	// 补偿处理tags清除操作
	this.RLock()
	if len(this.tags) == 0 {
		this.RUnlock()
	} else {
		tags := []string{}
		for k, v := range this.tags {
			if v {
				tags = append(tags, k)
			}
		}
		this.RUnlock()
		for _, v := range tags {
			tagPools.Del(v, this.GetId())
		}
	}
}
