package client

import (
	"bytes"
	"crypto/rc4"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/pb"
	"github.com/huajiao-tv/qchat/utility/gzipPool"
)

const (
	StateIdle = iota
	StateRunning
	StateInitiated
	StateLoggedIn
)

var idNameMap = map[int]string{
	pb.CHAT_REQ:           "ChatReq",
	pb.CHAT_RESP:          "ChatResp",
	pb.GET_INFO_REQ:       "GetInfoReq",
	pb.GET_INFO_RESP:      "GetInfoResp",
	pb.LOGIN_REQ:          "LoginReq",
	pb.LOGIN_RESP:         "LoginResp",
	pb.LOGOUT_REQ:         "LogoutReq",
	pb.LOGOUT_RESP:        "LogoutResp",
	pb.NEW_MESSAGE_NOTIFY: "NewMessageNotify",
	pb.RE_LOGIN_NOTIFY:    "ReLoginNotify",
	pb.RE_CONNECT_NOTIFY:  "ReConnectNotify",

	pb.INIT_LOGIN_REQ:  "InitLoginReq",
	pb.INIT_LOGIN_RESP: "InitLoginResp",

	pb.SERVICE_REQ:          "Service_Req",
	pb.SERVICE_RESP:         "Service_Resp",
	pb.GET_MULITI_INFOS_REQ: "GetMultipleInfoReq",
	pb.GET_MULTI_INFOS_RESP: "GetMultipleInfoResp",
}

var (
	PingData = []byte{0, 0, 0, 0}
)

type UserConnection struct {
	conf    *ClientConf
	account *AccountInfo
	conn    Stream

	verbose bool
	state   int32
	key     string

	clientRan string
	serverRan string

	sendChan chan *packet
	readChan chan *packet
	CmdChan  chan []string // export

	services map[int]ServiceHandler
	cache    map[*packet]time.Time
	infoBox  map[string]*MessageSlot
}

func NewUserConnection(c *ClientConf, a *AccountInfo, v bool) *UserConnection {
	conn := &UserConnection{
		conf:     c,
		account:  a,
		verbose:  v,
		key:      c.DefaultKey,
		sendChan: make(chan *packet, ClientChanLen),
		readChan: make(chan *packet, ClientChanLen),
		CmdChan:  make(chan []string, 100),
		cache:    make(map[*packet]time.Time, 100),
	}
	// add services
	conn.services = map[int]ServiceHandler{
		ChatroomServiceID: &ChatroomService{
			c: conn,
			p: gzipPool.NewGzipDecompressPool(10),
		},
		GroupServiceID: &GroupService{
			c: conn,
		},
	}
	conn.infoBox = map[string]*MessageSlot{
		IM: &MessageSlot{
			InfoType: IM,
			LatestID: -1,
		},
		Peer: &MessageSlot{
			InfoType: Peer,
			LatestID: -1,
		},
		Public: &MessageSlot{
			InfoType: Public,
			LatestID: -1,
		},
	}
	if v {
		conn.log("verbose", c, a)
	}
	return conn
}

func (c *UserConnection) Start(ch chan string) {
	if c.verbose {
		c.log("verbose", "Start")
	}
	old := c.getState()
	if old == StateIdle && c.setState(old, StateRunning) {
		go c.mainLoop(ch)
	} else {
		c.log("error", "already running")
	}
}

func (c *UserConnection) RequestService(serviceID int, data []byte) bool {
	p := &pb.Message{
		Sn:    proto.Uint64(logic.GetSn()),
		Msgid: proto.Uint32(pb.SERVICE_REQ),
		Req: &pb.Request{
			ServiceReq: &pb.Service_Req{
				ServiceId: proto.Uint32(uint32(serviceID)),
				Request:   data,
			},
		},
	}
	return c.send(p)
}

func (c *UserConnection) connect() error {
	if c.verbose {
		c.log("verbose", "connect")
	}
	var err error
	if c.conf.ConnType == Tcp {
		c.log("connect", "tcp", c.conf.ServerAddr)
		c.conn, err = NewTcpStream(c.conf.ServerAddr, time.Duration(ConnTimeout)*time.Second)
	} else if c.conf.ConnType == WebSocket {
		c.log("connect", "websocket", c.conf.ServerAddr)
		c.conn, err = NewWebSocketStream(c.conf.ServerAddr, time.Duration(ConnTimeout)*time.Second)
	} else {
		err = errors.New("unknown stream type")
	}
	if err != nil {
		return err
	} else {
		return nil
	}
}

func (c *UserConnection) mainLoop(ch chan string) {
	if c.verbose {
		c.log("verbose", "start", "mainLoop")
	}

	err := c.connect()
	if err != nil {
		c.log("error", "connect failed", err.Error())
		return
	}
	defer c.shutdown()

	c.send(c.initLogin())

	go c.readLoop()
	go c.commandLoop()

	heartbeatTimeout := time.Duration(c.conf.Heartbeat) * time.Second
	heartbeatTimer := time.NewTimer(heartbeatTimeout)
	packetTimer := time.NewTimer(c.getNextTimeout())

ForLoop:
	for {
		select {
		case p := <-c.sendChan:
			if c.doSend(p) {
				heartbeatTimer.Reset(heartbeatTimeout)
			} else {
				break ForLoop
			}
		case p := <-c.readChan:
			if c.handlePacket(p) {
				packetTimer.Reset(c.getNextTimeout())
			} else {
				break ForLoop
			}
		case <-heartbeatTimer.C:
			if c.conf.SendHeartbeat != 0 {
				c.sendChan <- &packet{true, nil, nil}
			}

		case <-packetTimer.C:
			p := c.getNextPacket()
			if p.isPing {
				c.log("error", "pong timeout")
			} else {
				c.log("error", "response timeout")
			}
			break ForLoop
		}
	}

	if c.verbose {
		c.log("verbose", "stop", "mainLoop")
	}
	if ch != nil {
		time.Sleep(1 * time.Second)
		ch <- c.conf.Tag
	}
}

func (c *UserConnection) handlePacket(p *packet) bool {
	if p == nil {
		return false
	}

	c.removeNextPacket(p)
	if p.isPing {
		if c.verbose {
			c.log("verbose", "pong")
		}
		return true
	}

	m := p.msg
	sn := p.msg.GetSn()
	resp := p.msg.GetResp()

	if resp != nil && resp.Error != nil {
		err := *resp.Error
		if pb.ERR_USER_INVALID == *err.Id {
			c.log("error", pb.ERR_USER_INVALID, "ERR_USER_INVALID")
			return false
		}
		c.log("error", *err.Id, err.Description)
		return false
	}

	switch *m.Msgid {
	case uint32(pb.INIT_LOGIN_RESP):
		clientRan := m.Resp.InitLoginResp.ClientRam
		if *clientRan != c.clientRan {
			c.log("error", "client ram mismatch")
			return false
		}
		c.serverRan = *(m.Resp.InitLoginResp.ServerRam)
		c.setState(StateRunning, StateInitiated)
		if c.conf.AutoLogin != 0 && !c.send(c.login()) {
			c.log("error", "send login failed")
		}

	case uint32(pb.LOGIN_RESP):
		c.setState(StateInitiated, StateLoggedIn)
		c.key = *(m.Resp.Login.SessionKey)
		c.log("success", "LoginResp")
		c.send(c.getInfo(c.infoBox[Peer]))
		c.send(c.getInfo(c.infoBox[Public]))
		c.send(c.getInfo(c.infoBox[IM]))

	case uint32(pb.GET_INFO_RESP):
		getInfo := *m.Resp.GetInfo
		count := len(getInfo.Infos)
		infoType := *getInfo.InfoType
		c.log("GetInfoResp", infoType, "count", count, "lastid", *getInfo.LastInfoId)
		for _, info := range getInfo.Infos {
			var msg UserMessage
			for _, pair := range info.PropertyPairs {
				key := string(pair.Key)
				if key == "info_id" {
					msg.ID = binary.BigEndian.Uint64(pair.Value)
				} else if key == "chat_body" {
					msg.Content = pair.Value
				}
			}
			c.log("GetInfo", infoType, "id", msg.ID, "content", msg.Content)
		}
		if box, ok := c.infoBox[infoType]; ok {
			box.LatestID = *getInfo.LastInfoId
		}
		if len(getInfo.Infos) > 0 {
			c.send(c.getInfo(c.infoBox[infoType]))
		}

	case uint32(pb.GET_MULTI_INFOS_RESP):
		getInfo := *m.Resp.GetMultiInfos
		count := len(getInfo.Infos)
		infoType := *getInfo.InfoType

		c.log("GetMultiInfoResp", infoType, "count", count, "lastid", *getInfo.LastInfoId)
		for _, info := range getInfo.Infos {
			msg := UserMessage{
				InfoType: infoType,
				Sn:       m.GetSn(),
				Valid:    1,
			}
			for _, pair := range info.PropertyPairs {
				var key string = string(pair.Key)
				if key == "info_id" {
					msg.ID = binary.BigEndian.Uint64(pair.Value)
				} else if key == "chat_body" {
					msg.Content = pair.Value
				} else if key == "msg_type" {
					msg.MsgType = binary.BigEndian.Uint32(pair.Value)
				} else if key == "msg_valid" {
					msg.Valid = binary.BigEndian.Uint32(pair.Value)
				}
			}
			if service, ok := c.services[ChatroomServiceID]; ok {
				service.HandleGetMessage(&msg)
			}
		}

	case uint32(pb.NEW_MESSAGE_NOTIFY):
		ntf := m.Notify.NewinfoNtf
		infoType := *ntf.InfoType
		c.log("NewMessageNotify", infoType)
		if infoType == "chatroom" {
			if service, ok := c.services[ChatroomServiceID]; ok {
				service.HandleServiceMessage(ntf.InfoContent, sn, 0)
			}
		} else if infoType == "group" {
			service, ok := c.services[GroupServiceID]
			if ok {
				service.HandleServiceMessage(ntf.InfoContent, sn, 0)
			}
		} else {
			if box, ok := c.infoBox[infoType]; ok {
				c.send(c.getInfo(box))
			}
		}

	case uint32(pb.SERVICE_RESP):
		if *resp.ServiceResp.ServiceId == ChatroomServiceID {
			service, ok := c.services[ChatroomServiceID]
			if ok {
				service.HandleServiceMessage(resp.ServiceResp.Response, sn, 1)
			}
		} else if *resp.ServiceResp.ServiceId == GroupServiceID {
			service, ok := c.services[GroupServiceID]
			if ok {
				service.HandleServiceMessage(resp.ServiceResp.Response, sn, 1)
			}
		}

	default:
		c.log("warn", "unknown packet", *m.Msgid, "sn", *m.Sn)
	}
	return true
}

func (c *UserConnection) commandLoop() {
	if c.verbose {
		c.log("verbose", "start", "commandLoop")
	}

	s, ok := c.services[ChatroomServiceID]
	if !ok {
		c.log("error", "no service")
	}
	cr := s.(*ChatroomService)

	s, ok = c.services[GroupServiceID]
	if !ok {
		c.log("error", "no service")
	}
	groupSrv := s.(*GroupService)

	for {
		cmds := <-c.CmdChan
		if cmds == nil {
			break
		}
		if c.verbose {
			c.log("verbose", "commands", cmds)
		}
		switch cmds[0] {
		case "join":
			cr.Join(cmds[1:])
		case "quit":
			cr.Quit(cmds[1:])
		case "query":
			cr.Query(cmds[1:])
		case "getmsg":
			cr.GetMultiInfo(cmds[1:])
		case "chat":
			sendChatroom(c, cmds[1:])
		case "peer":
			sendPeer(c, cmds[1:])
		case "im":
			sendIm(c, cmds[1:])
		case "wg":
			sendWg(c, cmds[1:])
		case "creategroup":
			createGroup(c, cmds[1:])
		case "joingroup":
			joinGroup(c, cmds[1:])
		case "quitgroup":
			quitGroup(c, cmds[1:])
		case "dismissgroup":
			dismissGroup(c, cmds[1:])
		case "listgroupuser":
			listGroupUser(c, cmds[1:])
		case "getgroupinfo":
			getGroupInfo(c, cmds[1:])
		case "ingroups":
			inGroups(c, cmds[1:])
		case "ismember":
			isMember(c, cmds[1:])
		case "listcreatedgroup":
			listCreatedGroup(c, cmds[1:])
		case "sendgroupmsg":
			sendGroupMsg(c, cmds[1:])
		case "joincount":
			joinCount(c, cmds[1:])
		case "groupmsg":
			ids := stringToIntArray(cmds[2])
			if ids == nil || len(ids) != 2 {
				continue
			}

			groupSrv.GetMsg(cmds[1], uint64(ids[0]), int32(ids[1]))
		case "groupmsgbatch":
			ids := stringToIntArray(cmds[1])
			groupSrv.GetMsgBatch(ids)
		case "groupsync":
			groupSrv.Sync()

		case "disconn":
			c.shutdown()
		case "robotjoin":
			sendRobotJoinMessage(c, cmds[1:])
		case "robotquit":
			sendRobotQuitMessage(c, cmds[1:])
			break
		}
	}
	if c.verbose {
		c.log("verbose", "stop", "commandLoop")
	}
}

func (c *UserConnection) readLoop() {
	if c.verbose {
		c.log("verbose", "start", "readLoop")
	}
	for {
		p := c.readPacket()
		if p != nil {
			c.readChan <- p
		} else {
			break
		}
	}

	close(c.readChan)
	c.log("exit", "read goroutine")
	if c.verbose {
		c.log("verbose", "stop", "readLoop")
	}
}

func (c *UserConnection) magicCode() []byte {
	magic := make([]byte, 12)
	magic[0] = 0x71
	magic[1] = 0x68
	magic[2] = (byte)(((c.conf.ProtoVer & 0xF) << 4) | ((c.conf.ClientVer & 0xF00) >> 8))
	magic[3] = (byte)(c.conf.ClientVer & 0xFF)
	magic[4] = (byte)((c.conf.AppID & 0xFF00) >> 8)
	magic[5] = (byte)(c.conf.AppID & 0xFF)
	return magic
}

func (c *UserConnection) ping() bool {
	if c.verbose {
		c.log("verbose", "ping")
	}
	c.conn.SetWriteDeadline(time.Now().Add(time.Duration(WriteTimeout) * time.Second))
	if count, err := c.conn.Write(PingData); count != len(PingData) {
		c.log("error", "ping", err.Error())
		return false
	}
	return true
}

func (c *UserConnection) doSend(p *packet) bool {
	if p == nil {
		c.log("error", "send", "nil packet")
		return false
	}

	c.cache[p] = time.Now()
	if p.isPing {
		return c.ping()
	}
	if p.data == nil || len(p.data) == 0 {
		c.log("error", "send", "empty data")
		return false
	}

	c.conn.SetWriteDeadline(time.Now().Add(time.Duration(WriteTimeout) * time.Second))
	if count, err := c.conn.Write(p.data); count < len(p.data) {
		c.log("error", "send", "write failed", err.Error())
		return false
	} else {
		id := int(*p.msg.Msgid)
		name := idNameMap[id]
		if c.verbose {
			c.log("verbose", "send", name, *p.msg.Sn)
		}
	}
	return true
}

func (c *UserConnection) send(m *pb.Message) bool {
	if m == nil {
		c.log("error", "nil pb message")
		return false
	}
	pbData, err := proto.Marshal(m)
	if err != nil {
		c.log("error", "proto.Marshal", err.Error())
		return false
	}
	rc, err := rc4.NewCipher([]byte(c.key))
	if err != nil {
		c.log("error", "NewCipher", err.Error())
		return false
	}
	rc.XORKeyStream(pbData, pbData)
	totalLen := (uint32)(len(pbData) + 4)
	if c.getState() == StateRunning {
		totalLen += 12
	}

	buf := bytes.NewBuffer(nil)
	// cache magic code
	if c.getState() == StateRunning {
		count, err := buf.Write(c.magicCode())
		if count != 12 || err != nil {
			c.log("error", "write magic code", err.Error())
			return false
		}
	}
	// cache length bytes
	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, totalLen)
	count, err := buf.Write(lenBytes)
	if err != nil || count != 4 {
		c.log("error", "write length", err.Error())
		return false
	}
	// cache PB data
	if count, err = buf.Write(pbData); count != len(pbData) {
		c.log("error", "write pb", err.Error())
		return false
	}

	data := buf.Bytes()
	c.sendChan <- &packet{false, m, data}
	return true
}

func (c *UserConnection) readPacket() *packet {
	if c.conn == nil {
		return nil
	}
	lenBytes := make([]byte, 4)
	count, err := c.conn.Read(lenBytes)
	if count != 4 {
		c.log("error", "read data", err)
		time.Sleep(5 * time.Second)
		return nil
	}
	if c.getState() == StateRunning && c.conf.ProtoVer == 1 {
		if lenBytes[0] != 0x71 || lenBytes[1] != 0x68 {
			c.log("error", "magic code", lenBytes[0], lenBytes[1])
			return nil
		}
		// read 2 more bytes
		lenBytes[0], lenBytes[1] = lenBytes[2], lenBytes[3]
		count, err = c.conn.Read(lenBytes[2:4])
		if count != 2 {
			c.log("error", "read 2", err)
			return nil
		}
	}
	len := int(binary.BigEndian.Uint32(lenBytes))
	if len == 0 {
		return &packet{isPing: true}
	} else if len < 0 || len > 100000000 {
		c.log("error", "data len", lenBytes)
		return nil
	}

	len -= 4
	if c.getState() == StateRunning && c.conf.ProtoVer == 1 {
		len -= 2
	}
	pbData := make([]byte, len)
	count, err = c.conn.Read(pbData)
	if count != len {
		c.log("error", "read pb failed", err, count)
		return nil
	}

	// decipher data
	key := c.key
	if c.getState() == StateInitiated {
		key = c.account.Password
	}
	rc, err := rc4.NewCipher([]byte(key))
	if err != nil {
		c.log("error", "NewCipher failed", err.Error())
		return nil
	}

	dest := make([]byte, len)
	rc.XORKeyStream(dest, pbData)

	m := &pb.Message{}
	err = proto.Unmarshal(dest, m)
	// the case password is incorrect
	if err != nil && c.getState() == StateInitiated {
		rc, err = rc4.NewCipher([]byte(c.key))
		rc.XORKeyStream(pbData, pbData)
		err = proto.Unmarshal(pbData, m)
	}

	if err != nil {
		c.log("error", "decode pb failed", err.Error())
		return nil
	}

	if c.verbose {
		id := int(*m.Msgid)
		name := idNameMap[id]
		c.log("verbose", "packet received", name)
	}

	if *m.Msgid == 200009 {
		c.log("Receive packet:", m)
	}

	if *m.Msgid == 200001 {
		c.log("Receive packet:", m, "time now", uint32(time.Now().Unix()))
	}

	return &packet{false, m, pbData}
}

func (c *UserConnection) shutdown() {
	if c.conn != nil {
		c.conn.Close()
		c.log("closed by client")
	}
}

func (c *UserConnection) initLogin() *pb.Message {
	c.clientRan = logic.RandString(8)
	m := &pb.Message{
		Sender: proto.String(c.account.UserID),
		Sn:     proto.Uint64(logic.GetSn()),
		Msgid:  proto.Uint32(pb.INIT_LOGIN_REQ),
		Req: &pb.Request{
			InitLoginReq: &pb.InitLoginReq{
				ClientRam: proto.String(c.clientRan),
			},
		},
	}
	if len(c.account.Signature) > 0 {
		m.Req.InitLoginReq.Sig = &c.account.Signature
	}
	return m
}

func (c *UserConnection) login() *pb.Message {
	m := &pb.Message{
		Sender:     proto.String(c.account.UserID),
		SenderType: proto.String("jid"),
		Sn:         proto.Uint64(logic.GetSn()),
		Msgid:      proto.Uint32(pb.LOGIN_REQ),
		Req: &pb.Request{
			Login: &pb.LoginReq{
				AppId:      proto.Uint32(uint32(c.conf.AppID)),
				NetType:    proto.Uint32(3),
				MobileType: proto.String("pc"),
				ServerRam:  proto.String(c.serverRan),
				SecretRam:  logic.MakeSecretRan(c.account.Password, c.serverRan),
				HeartFeq:   proto.Uint32(uint32(c.conf.Heartbeat)),
				Deviceid:   proto.String(c.account.DeviceID),
				Platform:   proto.String(c.account.Platform),
			},
		},
	}
	if len(c.account.Signature) == 0 {
		m.Req.Login.VerfCode = proto.String(logic.MakeVerfCode(c.account.UserID))
	}
	return m
}

func (c *UserConnection) getInfo(box *MessageSlot) *pb.Message {
	return &pb.Message{
		Sn:    proto.Uint64(logic.GetSn()),
		Msgid: proto.Uint32(pb.GET_INFO_REQ),
		Req: &pb.Request{
			GetInfo: &pb.GetInfoReq{
				InfoType:      &box.InfoType,
				GetInfoId:     proto.Int64(box.LatestID + 1),
				GetInfoOffset: proto.Int32(binary.MaxVarintLen32),
			},
		},
	}
}

func (c *UserConnection) getNextTimeout() time.Duration {
	least := time.Duration(1) * time.Hour
	for _, t := range c.cache {
		timeout := t.Add(time.Duration(RequestTimeout) * time.Second).Sub(time.Now())
		if timeout < least {
			least = timeout
		}
	}
	if least < 0 {
		least = time.Duration(1) * time.Nanosecond
	}
	return least
}

func (c *UserConnection) getNextPacket() *packet {
	var p *packet
	now := time.Now()
	least := time.Duration(1) * time.Hour
	for pkt, t := range c.cache {
		if diff := t.Sub(now); diff < least {
			least = diff
			p = pkt
		}
	}
	return p
}

func (c *UserConnection) removeNextPacket(p *packet) {
	if p == nil {
		return
	}
	delete(c.cache, p)
	for pkt, _ := range c.cache {
		if p.isPing && pkt.isPing {
			delete(c.cache, pkt)
		} else if pkt.msg != nil && p.msg != nil {
			if pkt.msg.GetSn() == p.msg.GetSn() {
				delete(c.cache, pkt)
			}
		}
	}
}

func (c *UserConnection) getState() int32 {
	return atomic.LoadInt32(&c.state)
}

func (c *UserConnection) setState(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&c.state, old, new)
}

func (c *UserConnection) log(args ...interface{}) {
	Log(c.conf.Tag, c.account.UserID, args...)
}

func Log(tag, id string, args ...interface{}) {
	content := time.Now().Format("2006-01-02 15:04:05.000") + "|" + id
	for _, arg := range args {
		switch arg.(type) {
		case int:
			content = content + "|" + strconv.Itoa(arg.(int))
		case string:
			content = content + "|" + strings.TrimRight(arg.(string), "\n")
		case int64:
			str := strconv.FormatInt(arg.(int64), 10)
			content = content + "|" + str
		case []byte:
			content = content + "|" + string(arg.([]byte))
		default:
			content = content + "|" + fmt.Sprintf("%v", arg)
		}
	}
	fmt.Println("> "+tag, content)
}

func stringToIntArray(input string) []int64 {

	if len(input) == 0 {
		return nil
	}

	ids := make([]int64, 0)

	for _, s := range strings.Split(input, ",") {
		s = strings.Trim(s, "\r\n\t ")
		i, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			continue
		}
		ids = append(ids, i)
	}
	return ids
}
