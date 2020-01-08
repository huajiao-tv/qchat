package pb

import (
	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/gzipPool"
	"github.com/huajiao-tv/qchat/utility/network"
)

const (
	LOGIN_REQ                 = 100001
	CHAT_REQ                  = 100002
	GET_INFO_REQ              = 100004
	LOGOUT_REQ                = 100005
	INIT_LOGIN_REQ            = 100009
	SERVICE_REQ               = 100011
	EX1_QUERY_USER_STATUS_REQ = 100012
	GET_MULITI_INFOS_REQ      = 100100
)

const (
	LOGIN_RESP                 = 200001
	CHAT_RESP                  = 200002
	GET_INFO_RESP              = 200004
	LOGOUT_RESP                = 200005
	INIT_LOGIN_RESP            = 200009
	SERVICE_RESP               = 200011
	EX1_QUERY_USER_STATUS_RESP = 200012
	GET_MULTI_INFOS_RESP       = 200100
)
const (
	NEW_MESSAGE_NOTIFY = 300000
	RE_LOGIN_NOTIFY    = 300001
	RE_CONNECT_NOTIFY  = 300002
)

const (
	CHATROOM_SERVICE_ID = 10000006
	CR_PAYLOAD_QUERY    = 101 // 查询聊天室信息
	CR_PAYLOAD_JOIN     = 102 // 加入聊天室
	CR_PAYLOAD_QUIT     = 103 // 退出聊天室
	CR_PAYLOAD_SUB      = 109 // 订阅聊天室消息
	CR_PAYLOAD_MESSAGE  = 300 // 聊天室消息

	CR_PAYLOAD_INCOMING_MSG   = 1000 // 消息通知
	CR_PAYLOAD_MEMBER_ADDED   = 1001 // 加入通知
	CR_PAYLOAD_MEMBER_REMOVED = 1002 // 退出通知
	CR_PAYLOAD_COMPRESSED     = 1003 // 压缩协议

	CR_PAYLOAD_JOIN_RESP = 102 //加入聊天室回包
	CR_PAYLOAD_QUIT_RESP = 103 // 退了聊天室回包

	GROUP_SERVICE_ID          = 10000001
	GC_PAYLOAD_SYNC           = 108 // 同步群的最大消息id和版本号
	GC_PAYLOAD_MSG_REQ_RESP   = 109 // 拉取消息返回
	GC_PAYLOAD_NEW_MSG_NOTIFY = 200 // 新消息
)

const (

	/**
	 * 严重错误， 发送该错误， 给客户端发送错误后需要断连
	 */
	ERR_BAD_SEQUENCE   = 1001 // 客户端发包的次序不正确， 比如登录需要1: InitLogin --> Login, 如果不是这个顺序报该错误
	ERR_INVALID_SENDER = 1003 // 用于站内(上行也走长连接)， 发消息时， 如果sender和对应的sender_type不匹配报该错误
	ERR_TOO_FREQUENTLY = 1004 // 客户端上行包（pb.Message）的频度超过每分钟200个时， 报该错误然后断连

	ERR_DBA_EXCEPTION     = 1006 // 访问dba抛异常， 客户端登录时收到该错误码后下次登录至少在5分钟之后
	ERR_SESSION_EXCEPTION = 1007 // 访问session抛异常， 客户端登录时收到该错误码后下次登录至少在5分钟之后

	ERR_USER_INVALID     = 1008 // 登录失败， 用户名和密码不匹配
	ERR_ROUTER_EXCEPTION = 1009 // router 异常报错

	ERR_SESSION_REFUSED = 1011 // 访问session抛异常， 客户端登录时收到该错误码后下次登录至少在5分钟之后
	ERR_DBA_TOO_BUSY    = 1012 // DBA过于繁忙， 客户端登录时收到该错误码后下次登录至少在5分钟之后
	ERR_BAD_PARAM       = 1013 // 现在的参数错误，原来是SRM client抛出异常

	ERR_SERVER_OVERLOADED   = 1015 // 服务器过载， 客户端登录时收到该错误码后下次登录至少在5分钟之后
	ERR_LOGGED_IN_ELSEWHERE = 1016 // 客户端在别处登录， 如果发现同一实例在别处登录， 那么先给要被下线的客户度该错误， 然后断连

	/**
	 * 普通错误， 如果登录过程中碰到需要断连， 否则不需要
	 */
	ERR_DBA_GENERIC        = 2000 // DB报错
	ERR_SESSION_GENERIC    = 2001 // session报错
	ERR_SRM_CLIENT_GENERIC = 2002 // SRM client 报错

	/**
	 * 可以忽略的错误
	 */
	ERR_USER_NOT_FOUND      = 3000 // 发送消息时， 未找到接收者
	ERR_INVALID_QUERY_PARAM = 3001 // 查询用户状态参数不正确
	ERR_RECEIVER_TYPE       = 3002 // 消息接受者类型不正确
	ERR_REDIS_GENERIC       = 3100 // 获取聊天消息发生错误
)

const (
	CompressPoolCap = 500
)

var (
	compressPool *gzipPool.CompressPool
)

func init() {
	compressPool = gzipPool.NewGzipCompressPool(CompressPoolCap)
}

func CompressChatRoomNewMsg(msg *logic.ChatRoomMessage) ([]byte, error) {
	if msg == nil {
		return nil, nil
	}
	msgid := proto.Uint32(uint32(msg.MsgID))
	if msg.MsgID == 0 {
		msgid = nil
	}
	content := &ChatRoomNewMsg{
		Roomid: []byte(msg.RoomID),
		Sender: &CRUser{
			Userid: []byte(msg.Sender),
		},
		Msgtype:     proto.Int(msg.MsgType),
		Msgcontent:  msg.MsgContent,
		Regmemcount: proto.Int(msg.RegMemCount),
		Memcount:    proto.Int(msg.MemCount),
		Msgid:       msgid,
		Maxid:       proto.Uint32(uint32(msg.MaxID)),
		Timestamp:   proto.Uint64(uint64(msg.TimeStamp)),
	}
	ds, err := proto.Marshal(content)
	if err != nil {
		return nil, errors.New("pb marshal error:" + err.Error())
	}

	return compressPool.Compress(ds), nil
}

func CreateMsgNotify(infoType string, infoContent []byte, infoId int64, sender, receiver string, queryInterval uint32) (*network.XimpBuffer, error) {
	sn := logic.GetSn()
	msgNotify := &Message{
		Msgid:    proto.Uint32(NEW_MESSAGE_NOTIFY),
		Sn:       &sn,
		Sender:   &sender,
		Receiver: &receiver,
		Notify: &Notify{
			NewinfoNtf: &NewMessageNotify{
				InfoType:          &infoType,
				InfoContent:       infoContent,
				InfoId:            &infoId,
				QueryAfterSeconds: &queryInterval,
			},
		},
	}
	pbBytes, err := proto.Marshal(msgNotify)
	if err != nil {
		return nil, err
	}

	ximp := &network.XimpBuffer{
		IsDecrypt:  true,
		DataStream: pbBytes,
	}
	return ximp, nil
}
