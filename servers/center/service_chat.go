package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/cryption"
)

const (
	ChatSuccess int = iota
	ChatFailed
	ChatNotFound
	ChatRecalledYet
	ChatSystemError    = 50000
	ChatParameterError = 50001
	//ChatNotInWhiteList    = 50002
	ChatSaverFailed = 50003
)

const (
	ChatMD5Suffix    = "wpms"
	ChatPublicSender = "admin"
)

const (
	ChatRequestMD5    = "m"
	ChatRequestBody   = "b"
	ChatSender        = "sender"
	ChatReceiver      = "receiver"
	ChatReceivers     = "receivers"
	ChatMessage       = "msg"
	ChatTraceId       = "traceid"
	ChatExpireTime    = "expire_time"
	ChatMsgType       = "msgtype"
	ChatRecallInboxId = "inid"
	ChatOwner         = "owner"
	ChatCh            = "ch"
	ChatCount         = "c"
	ChatStart         = "s"
)

const (
	ChatPushType       = 100
	ChatPublicPushType = 101
)

const (
	ChatInbox  = "inbox"
	ChatOutbox = "outbox"
)

type ChatMsgId struct {
	Owner string `json:"owner"`
	Id    uint64 `json:"id"`
	Box   string `json:"box"`
}

type ChatResponse struct {
	Code     int         `json:"code"`
	Reason   string      `json:"reason"`
	IosUsers []string    `json:"ios_offline_users"`
	MsgIds   []ChatMsgId `json:"msgids"`
}

/*
 * Formatting chat response as json string so that we can return it to http caller
 * @param cr is ChatResponse point which will be formatted as json response
 * @return formatted json string
 */
func (cr *ChatResponse) Response() string {
	if cr.IosUsers == nil {
		cr.IosUsers = make([]string, 0)
	}
	if cr.MsgIds == nil {
		cr.MsgIds = make([]ChatMsgId, 0)
	}

	if data, err := json.Marshal(*cr); err != nil {
		return fmt.Sprintf("{\"code\":%d, \"reason\":\"%s\"}", ChatSystemError, err.Error())
	} else {
		return string(data)
	}
}

type RetrieveChatResponse struct {
	Code       int                  `json:"code"`
	Reason     string               `json:"reason"`
	Inbox      []*saver.ChatMessage `json:"inbox messages"`
	Outbox     []*saver.ChatMessage `json:"outbox messages"`
	LatestID   uint64               `json:"latest id"`
	LastReadID uint64               `json:"last read id"`
}

/*
 * Formatting retrieve chat response as json string so that we can return it to http caller
 * @param rcr is RetrieveChatResponse point which will be formatted as json response
 * @return formatted json string
 */
func (rcr *RetrieveChatResponse) Response() string {
	if rcr.Inbox == nil {
		rcr.Inbox = make([]*saver.ChatMessage, 0)
	}
	if rcr.Outbox == nil {
		rcr.Outbox = make([]*saver.ChatMessage, 0)
	}

	if data, err := json.Marshal(*rcr); err != nil {
		return fmt.Sprintf("{\"code\":%d, \"reason\":\"%s\"}", ChatSystemError, err.Error())
	} else {
		return string(data)
	}
}

type SimpleChatResponse struct {
	Code   int    `json:"code"`
	Reason string `json:"reason"`
}

/*
 * Formatting simple chat response as json string so that we can return it to http caller
 * @param cr is ChatResponse point which will be formatted as json response
 * @return formatted json string
 */
func (scr *SimpleChatResponse) Response() string {
	if data, err := json.Marshal(*scr); err != nil {
		return fmt.Sprintf("{\"code\":%d, \"reason\":\"%s\"}", ChatSystemError, err.Error())
	} else {
		return string(data)
	}
}

// translate two responses
func (cr *ChatResponse) ToSimple() *SimpleChatResponse {
	return &SimpleChatResponse{Code: cr.Code, Reason: cr.Reason}
}

func (scr *SimpleChatResponse) ToCahtResponse(cr *ChatResponse) *ChatResponse {
	if cr == nil {
		cr = &ChatResponse{}
	}
	cr.Code, cr.Reason = scr.Code, scr.Reason
	return cr
}

/*
 * Chat error type
 */
type ChatError string

func (err ChatError) Error() string { return string(err) }

/*
 * Decodes chat request parameters from base64 then check md5 value of parameters string with
 *	suffix ChatMD5Suffix if need
 * @param req is a http.Request point which includes chat http request information
 * @return (decoded parameters string, nil) if decode and check successfully, otherwise (empty string, ChatError)
 */
func DecodeChatRequestParameters(req *http.Request) (string, error) {
	md5 := req.PostFormValue(ChatRequestMD5)
	body := req.PostFormValue(ChatRequestBody)
	if md5 == "" {
		return "", ChatError(fmt.Sprintf("client not send %s parameter", ChatRequestMD5))
	}
	if body == "" {
		return "", ChatError(fmt.Sprintf("client not send %s parameter", ChatRequestBody))
	}

	if bodyBytes, err := cryption.Base64Decode(body); err != nil {
		return "", ChatError("decode request failed")
	} else {
		body = string(bodyBytes)
	}

	if netConf().ChatCheckMd5 {
		// 当需要检查md5校验的时候,我们使用加盐校验
		CheckMd5 := cryption.Md5(body + ChatMD5Suffix)
		if md5 != CheckMd5 {
			Logger.Error("", "", "", "DecodeChatRequestParameters",
				"MD5 check failed", fmt.Sprint("client md5:", md5, "computed md5:", CheckMd5))
			return "", ChatError("decode request failed")
		}
	}

	return body, nil
}

/*
 * Check whether request parameters has specified parameter fields
 * @param form is a url.Values point which will be checked
 * @param args are checking field names
 * @return nil if check successfully, otherwise a ChatError indicates first unfound parameter
 */
func CheckChatParameter(form *url.Values, args ...string) error {
	for _, arg := range args {
		if _, ok := (*form)[arg]; !ok {
			return ChatError(fmt.Sprintf("client not send %s parameter", arg))
		}
	}
	return nil
}

/*
 * Get parameters of chat request
 * @param req is a http.Request point which includes chat http request information
 * @return (request parameters formatted as *url.Values, *ChatResponse)
 *  ChatResponse.Code is ChatSuccess indicates operation is successful, otherwise failed
 */
func GetChatRequestParameter(req *http.Request) (*url.Values, *ChatResponse) {
	var results *url.Values
	res := &ChatResponse{Code: ChatSuccess, Reason: "ok"}

	switch req.Method {
	case MethodPost:
		ct := req.Header.Get("Content-Type")
		// allows user uses text/xml format post value, so far http.Request only parses
		//  "application/x-www-form-urlencoded" type parameter
		if ct == "text/xml" || ct == "text" {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}

		// parse parameters
		if err := req.ParseForm(); err != nil {
			res.Code, res.Reason = ChatParameterError, "invalid request"
			break
		}

		if body, err := DecodeChatRequestParameters(req); err != nil {
			res.Code, res.Reason = ChatParameterError, err.Error()
		} else {
			if paras, err := url.ParseQuery(body); err != nil {
				res.Code, res.Reason = ChatParameterError, err.Error()
			} else {
				if paras.Get("appid") == "" {
					paras.Set("appid", logic.APPID_HUAJIAO_STR)
				} else if !logic.DynamicConf().Appids[paras.Get("appid")] {
					res.Code, res.Reason = ChatParameterError, "invalid appid"
					break
				}
				results = &paras
			}
		}

	default:
		res.Code, res.Reason = ChatParameterError, "not support http method, only POST"
	}

	return results, res
}

/*
 * make a notification or IM storage request from post parameters
 * @param parameters is pared post parameters
 * @param chatChannel is chat channel: saver.ChatChannelNotify or ChatChannelIM
 * @param appid is application id
 * @return (made request, receiver slice, traceid, chat response) if making is successful;
 *  otherwise (nil, nil, 0, chat response) is returned
 */
func MakeNotificationIMRequest(parameters *url.Values, chatChannel string, appid uint16) (
	Request *saver.StoreMessagesRequest,
	Receivers []string,
	TraceId int64,
	Response *ChatResponse) {
	res := &ChatResponse{Code: ChatSuccess, Reason: "ok"}

	// check required parameters first
	if err := CheckChatParameter(parameters, ChatSender, ChatReceivers, ChatMessage, ChatTraceId); err != nil {
		res.Code, res.Reason = ChatParameterError, err.Error()
		return nil, nil, 0, res
	}

	// get required parameters
	sender := parameters.Get(ChatSender)
	if sender == "" {
		res.Code, res.Reason = ChatParameterError, fmt.Sprintf("client not send %s parameter", ChatSender)
		return nil, nil, 0, res
	}
	receivers := strings.Split(parameters.Get(ChatReceivers), ",")
	receiverCount := len(receivers)
	if receiverCount == 0 {
		res.Code, res.Reason = ChatParameterError, fmt.Sprintf("client not send %s parameter", ChatReceivers)
		return nil, nil, 0, res
	}
	message := parameters.Get(ChatMessage)
	if message == "" {
		res.Code, res.Reason = ChatParameterError, fmt.Sprintf("client not send %s parameter", ChatMessage)
		return nil, nil, 0, res
	}
	traceId, err := strconv.ParseInt(parameters.Get(ChatTraceId), 10, 64)
	if err != nil {
		res.Code, res.Reason = ChatParameterError, "client not send trace information"
		return nil, nil, 0, res
	}

	// allow users does not send expire_time/msg_type parameter, will use default value in that case
	expireTime, err := strconv.ParseInt(parameters.Get(ChatExpireTime), 10, 32)
	if err != nil {
		expireTime = int64(netConf().ChatDefaultExpireTime)
	}
	msgType, err := strconv.ParseUint(parameters.Get(ChatMsgType), 10, 64)
	if err != nil {
		msgType = ChatPushType
	}

	// set im outbox storage flag according to config
	storeOutbox := uint8(0)
	if chatChannel == saver.ChatChannelIM && netConf().StoreImOutbox == 1 {
		storeOutbox = 1
	}

	// prepare to store notify channel message
	req := &saver.StoreMessagesRequest{Appid: appid,
		Messages:    make(map[string]*saver.ChatMessage, receiverCount),
		TraceSN:     fmt.Sprint(traceId),
		ChatChannel: chatChannel}

	for _, receiver := range receivers {
		req.Messages[receiver] = &saver.ChatMessage{Content: message, Type: uint32(msgType), To: receiver,
			From: sender, TraceSN: traceId, ExpireInterval: int(expireTime), StoreOutbox: storeOutbox}
	}

	return req, receivers, traceId, res
}

/*
 * this handles store notify/im message response and return a session query list and optional sender
 *  optional sender can be used to send outbox notification
 * @param req is store message request we have sent
 * @param resp is store response from saver
 * @param res is operation result taht will return to http caller
 *
 * @return (session query list, sender) if has IM out box value; otherwise (session query list, "")
 */
func HandleStoreMessageResponse(req *saver.StoreMessagesRequest, resp *saver.StoreMessagesResponse,
	res *ChatResponse) (*session.QuerySessionReq, string) {
	// exit if no message saved successfully
	storeds := len(resp.Inbox) + len(resp.Outbox)

	sender := ""
	hasOutbox := false
	querySession := session.QuerySessionReq{make([]*session.UserSession, 0, storeds)}
	res.MsgIds = make([]ChatMsgId, 0, storeds)
	for _, storeMsg := range req.Messages {
		// handle inbox message response
		if stored, ok := resp.Inbox[storeMsg.To]; !ok || stored == nil {
			err := storeMsg.ToString("To", "From", "TraceSN") +
				" is saved to storage failed for required parameter is invalid"
			summary := "Saved message to storage failed"
			/* should we check this too or we should believe saver
			   if (ok && stored == nil ) {
			       err = "BUG:" + storeMsg.String("To", "From", "TraceSN") +
			       " is saved to storage failed for saver set nill data"
			       summary = "Saver BUG"
			   }
			*/
			Logger.Error(storeMsg.To, req.Appid, storeMsg.TraceSN, "HandleStoreMessageResponse",
				summary, err)
			continue
		} else {
			userSession := session.UserSession{UserId: stored.To, AppId: req.Appid}
			querySession.QueryUserSessions = append(querySession.QueryUserSessions, &userSession)
			res.MsgIds = append(res.MsgIds, ChatMsgId{Owner: stored.To, Id: stored.MsgId, Box: ChatInbox})
		}

		// handle outbox if need
		if stored, ok := resp.Outbox[storeMsg.To]; ok {
			/* should we check this too or we should believe saver
			   if stored == nil {
			       err := "BUG:" + storeMsg.String("To", "From", "TraceSN") +
			       " is saved to storage failed for saver set nill data"
			       Logger.Error(storeMsg.To, logic.DEFAULT_APPID, storeMsg.TraceSN, "StoreMessageAndPushNotification",
			           "Saver BUG", err)
			       continue
			   }
			*/

			res.MsgIds = append(res.MsgIds, ChatMsgId{Owner: stored.From, Id: stored.MsgId, Box: ChatOutbox})
			hasOutbox = true
			sender = stored.From
		}
	}

	// if has outbox data, need to notify sender too
	if hasOutbox {
		userSession := session.UserSession{UserId: sender, AppId: req.Appid}
		querySession.QueryUserSessions = append(querySession.QueryUserSessions, &userSession)
	}

	return &querySession, sender
}

// 给inbox和outbox里的人发送有新消息通知, 前提是这些人在users里
func PushNotication(appid int, sender string, users []*session.UserSession, channel string, inbox, outbox map[string]uint64, traceId string) {
	// make push list
	pushes := make([]*router.ChatPushNotify, 0, len(users))
	for _, u := range users {
		Logger.Debug(*u, inbox, outbox)
		if msgid, ok := inbox[u.UserId]; ok {
			notify := &router.ChatPushNotify{
				Appid:       appid,
				Receiver:    u.UserId,
				GatewayAddr: u.GatewayAddr,
				ConnId:      u.ConnectionId,
				Sender:      sender,
				Channel:     channel,
				MsgId:       msgid,
				TraceSN:     traceId,
			}
			pushes = append(pushes, notify)
		}
		if msgid, ok := outbox[u.UserId]; ok {
			notify := &router.ChatPushNotify{
				Appid:       appid,
				Receiver:    u.UserId,
				GatewayAddr: u.GatewayAddr,
				ConnId:      u.ConnectionId,
				Sender:      sender,
				Channel:     channel,
				MsgId:       msgid,
				TraceSN:     traceId,
			}
			pushes = append(pushes, notify)
		}
	}

	// send push request only if there is at least one request
	if len(pushes) > 0 {
		if err := router.SendPushNotifyBatch(pushes); err != nil {
			// although we can not send notification for push failed,
			// we still send caller successful result for messages have been saved to storage yet
			// it does not matter for client will pull messages
			Logger.Warn("", appid, traceId, "StoreMessageAndPushNotification",
				"Push notification to user failed", err)
		}
	} else {
		Logger.Trace("", appid, traceId, "StoreMessageAndPushNotification",
			"No any receiver is online", "")
	}
}

/*
 * Stores formatted notification message to storage and then push notification to router if receivers are online
 * @param parameters is a url.Values point which includes decoded chat parameters
 * @param chatChannel is chat channel, so far support saver.ChatChannelNotify and ChatChannelIM
 * @appid application ID
 * @return a non-nil ChatResponse point,
 *  ChatResponse.Code is ChatSuccess indicates operation is successful, otherwise failed
 *  Handler can use func (cr *ChatResponse) Response() string as result to return to http caller
 */
func StoreMessageAndPushNotification(parameters *url.Values, chatChannel string, appid uint16) *ChatResponse {
	req, receivers, traceId, res := MakeNotificationIMRequest(parameters, chatChannel, appid)
	if res.Code != ChatSuccess {
		Logger.Error("", appid, "", "StoreMessageAndPushNotification",
			"Made request failed", res.Reason)
		return res
	}

	Logger.Trace(receivers, appid, traceId, "StoreMessageAndPushNotification",
		"Made request", req)
	// store notification message to storage and get stored message id from response
	if resp, err := saver.StoreChatMessages(req); err != nil {
		res.Code, res.Reason = ChatSaverFailed, err.Error()
		Logger.Error(receivers, appid, traceId, "StoreMessageAndPushNotification",
			"Saved message to storage failed", err.Error())
		return res
	} else {
		// exit if no message saved successfully
		if len(resp.Inbox) == 0 {
			res.Code, res.Reason = ChatSaverFailed, "failed to store message to storage"
			Logger.Error(receivers, appid, traceId, "StoreMessageAndPushNotification",
				"Saved message to storage failed", "All messages are saved failed.")
			return res
		}

		// handle response and get user list we need to query online state
		querySession, sender := HandleStoreMessageResponse(req, resp, res)

		// post notification asynchronously
		go func() {
			// query session users' online state
			if onlineUsers, err := session.Query(querySession.QueryUserSessions); err != nil {
				// although we can not send notification for query session failed,
				// we still send caller successful result for messages have been saved to storage yet
				// it does not matter for client will pull messages
				Logger.Warn(receivers, appid, traceId, "StoreMessageAndPushNotification",
					"Query user state from session failed", err)
			} else {
				maxOutboxId := uint64(0)
				for _, stored := range resp.Outbox {
					if stored.MsgId > maxOutboxId {
						maxOutboxId = stored.MsgId
					}
				}

				// make push list
				onlines := len(onlineUsers)
				pushes := make([]*router.ChatPushNotify, 0, onlines)
				for _, userSession := range onlineUsers {
					if stored, ok := resp.Inbox[userSession.UserId]; ok {
						notify := router.ChatPushNotify{Appid: int(appid), Receiver: stored.To,
							GatewayAddr: userSession.GatewayAddr, ConnId: userSession.ConnectionId,
							Sender: stored.From, Channel: chatChannel, MsgId: stored.MsgId,
							TraceSN: strconv.FormatInt(stored.TraceSN, 10)}
						pushes = append(pushes, &notify)
					}

					// if sender is in online list and has outbox message in response, we need send notification to sender to
					if userSession.UserId == sender && maxOutboxId > 0 {
						notify := router.ChatPushNotify{Appid: int(appid), Receiver: sender,
							GatewayAddr: userSession.GatewayAddr, ConnId: userSession.ConnectionId,
							Sender: sender, Channel: chatChannel, MsgId: maxOutboxId,
							TraceSN: strconv.FormatInt(traceId, 10)}
						pushes = append(pushes, &notify)
					}
				}

				// send push request only if there is at least one request
				if len(pushes) > 0 {
					if err = router.SendPushNotifyBatch(pushes); err != nil {
						// although we can not send notification for push failed,
						// we still send caller successful result for messages have been saved to storage yet
						// it does not matter for client will pull messages
						Logger.Warn("", appid, traceId, "StoreMessageAndPushNotification",
							"Push notification to user failed", err)
					}
					Logger.Trace(receivers, appid, traceId, "StoreMessageAndPushNotification",
						"Send push notification request", "")
				} else {
					Logger.Trace(receivers, appid, traceId, "StoreMessageAndPushNotification",
						"No any receiver is online", "")
				}
			}
		}()
	}

	Logger.Trace(receivers, appid, traceId, "StoreMessageAndPushNotification",
		"Response to caller", res.Response())
	return res
}

/*
 * make a public or hot message storage request from post parameters
 * @param parameters is pared post parameters
 * @param appid is application id
 * @return (made request, receiver slice, traceid, chat response) if making is successful;
 *  otherwise (nil, nil, 0, chat response) is returned
 */
func MakePublicRequest(parameters *url.Values, hot bool, appid uint16) (
	Request *saver.StoreMessagesRequest,
	TraceId int64,
	Response *SimpleChatResponse) {
	res := &SimpleChatResponse{Code: ChatSuccess, Reason: "ok"}

	// check required parameters first
	if err := CheckChatParameter(parameters, ChatMessage, ChatTraceId); err != nil {
		res.Code, res.Reason = ChatParameterError, err.Error()
		return nil, 0, res
	}

	// get required parameters
	message := parameters.Get(ChatMessage)
	if message == "" {
		res.Code, res.Reason = ChatParameterError, fmt.Sprintf("client not send %s parameter", ChatMessage)
		return nil, 0, res
	}
	traceId, err := strconv.ParseInt(parameters.Get(ChatTraceId), 10, 64)
	if err != nil {
		res.Code, res.Reason = ChatParameterError, "client not send trace information"
		return nil, 0, res
	}

	// allow users does not send expire_time/msg_type parameter, will use default value in that case
	expireTime, err := strconv.ParseInt(parameters.Get(ChatExpireTime), 10, 32)
	if err != nil {
		if hot {
			expireTime = int64(netConf().ChatDefaultHotExpireTime)
		} else {
			expireTime = int64(netConf().ChatDefaultExpireTime)
		}
	}
	msgType, err := strconv.ParseUint(parameters.Get(ChatMsgType), 10, 64)
	if err != nil {
		msgType = ChatPushType
	}

	// prepare to store notify channel message
	req := &saver.StoreMessagesRequest{Appid: appid,
		Messages:    make(map[string]*saver.ChatMessage, 1),
		TraceSN:     fmt.Sprint(traceId),
		ChatChannel: saver.ChatChannelPublic}

	req.Messages[strconv.FormatUint(uint64(appid), 10)] = &saver.ChatMessage{To: saver.ChatChannelPublic,
		Content: message, Type: uint32(msgType), TraceSN: traceId, ExpireInterval: int(expireTime)}

	return req, traceId, res
}

const (
	defaultSendPublicNotificationAfter = 2
	maxSendPublicNotificationAfter     = 10
)

/*
 * Stores formatted public message to storage and then push notification to online users
 * @param parameters is a url.Values point which includes decoded chat parameters
 * @param hot is indicates whether is hot public message
 * @appid application ID
 * @return a non-nil ChatResponse point,
 *  ChatResponse.Code is ChatSuccess indicates operation is successful, otherwise failed
 *  Handler can use func (cr *ChatResponse) Response() string as result to return to http caller
 */
func StorePublicAndHot(parameters *url.Values, hot bool, appid uint16) *SimpleChatResponse {
	req, traceId, res := MakePublicRequest(parameters, hot, appid)
	if res.Code != ChatSuccess {
		Logger.Error("public", appid, traceId, "StorePublicAndHot",
			"Make request failed", res.Response())
		return res
	}

	Logger.Trace("public", appid, traceId, "StorePublicAndHot",
		"Made request", req)

	// store notification message to storage and get stored message id from response
	if resp, err := saver.StoreChatMessages(req); err != nil {
		res.Code, res.Reason = ChatSaverFailed, err.Error()
		Logger.Error("public", appid, traceId, "StorePublicAndHot",
			"Saved message to storage failed", err.Error())
		return res
	} else {
		// exit if no message saved successfully
		if len(resp.Inbox) == 0 {
			res.Code, res.Reason = ChatSaverFailed, "failed to store message to storage"
			Logger.Error("public", appid, traceId, "StorePublicAndHot",
				"Saved message to storage failed", "All messages are saved failed.")
			return res
		}

		appidStr := strconv.FormatUint(uint64(appid), 10)

		// post notification asynchronously
		go func() {
			sendAfter := netConf().SendPublicNotificationAfter
			if sendAfter < defaultSendPublicNotificationAfter {
				sendAfter = defaultSendPublicNotificationAfter
			} else if sendAfter > maxSendPublicNotificationAfter {
				sendAfter = maxSendPublicNotificationAfter
			}
			time.Sleep(time.Second * time.Duration(sendAfter))

			for _, stored := range resp.Inbox {
				// send notification to online users
				if err := router.SendPushTags([]string{appidStr}, []string{}, ChatPublicSender,
					saver.ChatChannelPublic, strconv.FormatInt(traceId, 10), stored.MsgId); err != nil {
					// although we can not send notification for push failed,
					// we still send caller successful result for messages have been saved to storage yet
					// it does not matter for client will pull messages
					Logger.Error("", appid, traceId, "StorePublicAndHot",
						"Push notification to user failed", err)
				}
			}
		}()
	}

	Logger.Trace("public", appid, traceId, "StorePublicAndHot",
		"Response to caller", res.Response())
	return res
}

/*
 * recall IM message for specified owner
 * @param parameters is a url.Values point which includes decoded chat parameters
 * @param hot is indicates whether is hot public message
 * @appid application ID
 * @return a non-nil ChatResponse point,
 *  ChatResponse.Code is ChatSuccess indicates operation is successful, otherwise failed
 *  Handler can use func (cr *ChatResponse) Response() string as result to return to http caller
 */
func RecallImMessage(parameters *url.Values, appid uint16) *SimpleChatResponse {
	res := &SimpleChatResponse{Code: ChatSuccess, Reason: "ok"}

	// check required parameters first
	if err := CheckChatParameter(parameters, ChatReceiver, ChatSender, ChatRecallInboxId, ChatTraceId); err != nil {
		res.Code, res.Reason = ChatParameterError, err.Error()
		return res
	}

	// get required parameters
	sender := parameters.Get(ChatSender)
	if sender == "" {
		res.Code, res.Reason = ChatParameterError, fmt.Sprintf("client not send %s parameter", ChatSender)
		return res
	}
	receiver := parameters.Get(ChatReceiver)
	if receiver == "" {
		res.Code, res.Reason = ChatParameterError, fmt.Sprintf("client not send %s parameter", ChatReceiver)
		return res
	}
	inboxId, err := strconv.ParseInt(parameters.Get(ChatRecallInboxId), 10, 64)
	if err != nil || inboxId == 0 {
		res.Code, res.Reason = ChatParameterError, fmt.Sprintf("client not send %s parameter", ChatRecallInboxId)
		return res
	}
	traceId := parameters.Get(ChatTraceId)
	if traceId == "" {
		res.Code, res.Reason = ChatParameterError, "client not send trace information"
		return res
	}

	// so far we only recall IM message
	req := &saver.RecallMessagesRequest{
		Appid:       appid,
		ChatChannel: saver.ChatChannelIM,
		Sender:      sender,
		Receiver:    receiver,
		InboxId:     uint64(inboxId),
		TraceSN:     traceId}
	Logger.Trace(receiver, appid, traceId, "RecallImMessage", "Made request", req)

	// recall im message
	if resp, err := saver.RecallChatMessages(req); err != nil {
		if strings.Contains(err.Error(), saver.ErrNotFound.Error()) {
			res.Code, res.Reason = ChatNotFound, saver.ErrNotFound.Error()
		} else if strings.Contains(err.Error(), saver.ErrRecalledYet.Error()) {
			res.Code, res.Reason = ChatRecalledYet, saver.ErrRecalledYet.Error()
		} else {
			res.Code, res.Reason = ChatFailed, err.Error()
			Logger.Error(receiver, appid, traceId, "RecallImMessage", "Recall message failed", err)
		}

		Logger.Trace(receiver, appid, traceId, "RecallImMessage", "Recall message failed", err)
		return res
	} else {
		// exit if no message saved successfully
		if len(resp.Inbox) == 0 && len(resp.Outbox) == 0 {
			res.Code, res.Reason = ChatFailed, "failed to store message to storage"
			Logger.Error(receiver, appid, traceId, "RecallImMessage",
				"Saved message to storage failed", "All messages are saved failed.")
		} else {
			// post notification asynchronously
			go func() {
				// try to send notification
				inbox := map[string]uint64{}
				outbox := map[string]uint64{}
				reqSess := []*session.UserSession{}
				for _, m := range resp.Inbox {
					inbox[m.To] = m.MsgId
					reqSess = append(reqSess, &session.UserSession{UserId: m.To, AppId: uint16(appid)})
				}

				for _, m := range resp.Outbox {
					outbox[m.From] = m.MsgId
					reqSess = append(reqSess, &session.UserSession{UserId: m.From, AppId: uint16(appid)})
				}

				if onlineUsers, err := session.Query(reqSess); err != nil {
					Logger.Error(sender, appid, traceId, "RecallImMessage", "session.Query error", err)
				} else {
					PushNotication(int(appid), sender, onlineUsers, saver.ChatChannelIM, inbox, outbox, traceId)
					Logger.Trace(sender, appid, traceId, "RecallImMessage", "Query Session Result",
						onlineUsers)
				}
			}()
		}
	}

	Logger.Trace(receiver, appid, traceId, "RecallImMessage", "Response to caller", res.Response())
	return res
}

/*
 * recall IM message for specified owner
 * @param parameters is a url.Values point which includes decoded chat parameters
 * @param hot is indicates whether is hot public message
 * @appid application ID
 * @return a non-nil ChatResponse point,
 *  ChatResponse.Code is ChatSuccess indicates operation is successful, otherwise failed
 *  Handler can use func (cr *ChatResponse) Response() string as result to return to http caller
 */
func RetrieveMessages(parameters *url.Values, appid uint16) *RetrieveChatResponse {
	res := &RetrieveChatResponse{Code: ChatSuccess, Reason: "ok"}

	// check required parameters first
	if err := CheckChatParameter(parameters, ChatOwner, ChatTraceId); err != nil {
		res.Code, res.Reason = ChatParameterError, err.Error()
		return res
	}

	// get required parameters
	owner := parameters.Get(ChatOwner)
	if owner == "" {
		res.Code, res.Reason = ChatParameterError, fmt.Sprintf("client not send %s parameter", ChatOwner)
		return res
	}
	traceId := parameters.Get(ChatTraceId)
	if traceId == "" {
		res.Code, res.Reason = ChatParameterError, "client not send trace information"
		return res
	}
	channel := parameters.Get(ChatCh)
	if channel == "" {
		channel = saver.ChatChannelNotify
	}
	count, _ := strconv.ParseInt(parameters.Get(ChatCount), 10, 64)
	start, _ := strconv.ParseInt(parameters.Get(ChatStart), 10, 64)

	chnnels := map[string]*saver.RetrieveChannel{
		channel: &saver.RetrieveChannel{Channel: channel, StartMsgId: start, MaxCount: int(count)}}
	req := &saver.RetrieveMessagesRequest{Appid: appid, Owner: owner, ChatChannels: chnnels,
		TraceSN: traceId}
	Logger.Trace("public", appid, traceId, "RetrieveMessages", "Made request", req)
	if resp, err := saver.RetrieveChatMessages(req); err != nil {
		res.Code, res.Reason = ChatFailed, err.Error()
		Logger.Error(owner, appid, traceId, "RetrieveMessages", "Retrieve messages failed", err)
		Logger.Trace(owner, appid, traceId, "RetrieveMessages", "Retrieve messages failed", err)
		return res
	} else {
		res.LatestID, res.LastReadID = resp.LatestID[channel], resp.LastReadID[channel]
		if messages, ok := resp.Inbox[channel]; ok && len(messages) > 0 {
			res.Inbox = messages
		}

		if messages, ok := resp.Outbox[channel]; ok && len(messages) > 0 {
			res.Outbox = messages
		}
	}

	Logger.Trace(owner, appid, traceId, "Response to callser", res.Response())
	return res
}
