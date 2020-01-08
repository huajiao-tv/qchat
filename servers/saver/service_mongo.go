// provides mongo message db operations
package main

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/huajiao-tv/qchat/client/saver"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	MongoErrorDuplicateCode = "E11000"
)

const (
	OpInc    = "$inc"
	OpSet    = "$set"
	OpLte    = "$lte"
	OpLt     = "$lt"
	OpGte    = "$gte"
	OpExists = "$exists"
	OpNe     = "$ne"
	OpOr     = "$or"
)

const (
	FieldJid            = "jid"
	FieldLatest         = "latest"
	FieldLatestModified = "latest_modified"
	FieldMsgId          = "msg_id"
	FieldMsgType        = "msg_t"
	FieldMsgFrom        = "msg_from"
	FieldMsgCTime       = "msg_ctime"
	FieldExpireTime     = "expire_time"
	FieldExpireInterval = "expire_interval"
	FieldCountTime      = "count_time"
	FieldMsgSn          = "msg_sn"
	FieldModified       = "modified"
	FieldMsgData        = "msg_data"
	FiedlLastRead       = "last_r"
	FieldRecalled       = "recalled"
	FieldOwner          = "owner"
	FieldMsgTo          = "msg_to"
	FieldInboxMsgId     = "inbox_msg_id"
)

const (
	NotificationIdDB        = "msg_messageid"
	NotificationIdCol       = "latest_id"
	NotificationLastReadDB  = "msg_lastr"
	NotificationLastReadCol = "last_reads"
	NotificationMsgDB       = "msg_messages"
	NotificationMsgCol      = "message_data"
	PubligMsgDB             = "msg_publics"
	PublicMsgCol            = "public_message_data"
	ImIdDB                  = "msg_imid"
	ImIdCol                 = "im_latest_id"
	ImInboxDB               = "msg_im_inboxes"
	ImInboxCol              = "msg_im_inbox"
	ImOutboxDB              = "msg_im_outboxes"
	ImOutBoxCol             = "msg_im_outbox"
	ImLastReadDB            = "msg_im_lastr"
	ImLastReadCol           = "im_last_reads"
)

const (
	MaxRetryTime       = 3
	DefaultReturnCount = 5
	RecallMsgType      = 10000
	RecalledFlag       = 1
	MaxExpireTime      = 8640000 // 最大消息过期时间
)

/*
 * Stores chat messages to mongo storage
 * @param req is a saver.StoreMessagesRequest point which include messages information need to store
 * @param resp is a saver.StoreMessagesResponse point which includes response
 * @return nil if no error occurs, otherwise an error interface is returned
 */
func StoreMongoChatMessages(req *saver.StoreMessagesRequest,
	resp *saver.StoreMessagesResponse) (ret error) {
	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countP2pResponseTime("", req.TraceSN, "StoreMongoChatMessages", req.Appid)
		defer countFunc()
	}

	Logger.Trace("", req.Appid, "", "StoreMongoChatMessages",
		"Get store message request", req)

	switch req.ChatChannel {
	case saver.ChatChannelNotify:
		ret = SaveMongoNotifyChannel(req, resp)
		// count peer store operation
		if ret != nil {
			requestStat.AtomicAddStorePeerFails(1)
		} else {
			requestStat.AtomicAddStorePeers(1)
		}
	case saver.ChatChannelPublic:
		ret = SaveMongoPublicChannel(req, resp)
		// count public store operation, because public operation is very few,
		// so just treat it as peer for they are both send by our system
		if ret != nil {
			requestStat.AtomicAddStorePeerFails(1)
		} else {
			requestStat.AtomicAddStorePeers(1)
		}
	case saver.ChatChannelIM:
		ret = SaveMongoImChannel(req, resp)
		// count IM store operation
		if ret != nil {
			requestStat.AtomicAddStoreIMFails(1)
		} else {
			requestStat.AtomicAddStoreIMs(1)
		}
	default:
		err := fmt.Sprintf("%s channel is not supported now", req.ChatChannel)
		Logger.Error("", req.Appid, "", "StoreMongoChatMessages",
			"Saved message to storage failed", err)
		ret = errors.New(err)
	}
	return
}

/*
 * get latest message id, will retry MaxRetryTime - 1 times if failed
 * @param owner owner is who retrieve latest id
 * @param db is database include owner's data
 * @param collection is collection stores owner's data
 * @param session is mongo session
 * @param inc is increment of counter
 * @param appid is application ID
 * @param res_debug is mongo findAndModify command raw result
 *
 * @return (latest id, nil) if succeeded, otherwise (0, error) is returned
 */
func GetNewMessageID(owner string, db string, collection string, session *mgo.Session, inc uint, appid uint16,
	res_debug *bson.M) (uint64, error) {
	change := mgo.Change{
		Update:    bson.M{OpInc: bson.M{FieldLatest: inc}, OpSet: bson.M{FieldLatestModified: time.Now()}},
		ReturnNew: true,
	}

	errInfo := ""
	for i := 0; i < MaxRetryTime; i++ {
		var result bson.M
		if _, err := session.DB(db).C(collection).Find(&bson.M{FieldJid: owner}).Apply(change, &result); err != nil {
			if err == mgo.ErrNotFound {
				if err := session.DB(db).C(collection).Insert(
					bson.D{bson.DocElem{FieldJid, owner}, bson.DocElem{FieldLatest, 0},
						bson.DocElem{FieldLatestModified, time.Now()}}); err != nil {
					errInfo = fmt.Sprintf("insert new record for owner %s failed, err: %s", owner, err.Error())

					// return if get error that is not duplicate error
					if !strings.Contains(errInfo, MongoErrorDuplicateCode) {
						Logger.Error(owner, appid, "", "GetMessageID", "get message id failed", errInfo)
						goto ExitWithError
					}

					Logger.Debug(owner, appid, "", "GetMessageID", "insert message id failed for duplicate", errInfo)
				}

				continue // continue to fAndM if no error or get insert duplicate error
			}

			errInfo = fmt.Sprintf("get message id for owner %s failed, err: %s", owner, err.Error())
			Logger.Error(owner, appid, "", "GetMessageID", "get message id failed", errInfo)
			goto ExitWithError

		}

		if res_debug != nil {
			*res_debug = result
		}
		Logger.Debug(fmt.Sprintf("GetMessageID get result[%v] of owner %s", result, owner))

		if res, err := GetIntegerFieldValue(&result, FieldLatest); err != nil {
			errInfo = fmt.Sprintf("get message id field failed for owner %s failed, err: %s", owner, err.Error())
			Logger.Error(owner, appid, "", "GetMessageID", "get message id field failed",
				errInfo)
			goto ExitWithError
		} else {
			return uint64(res), nil
		}
	}

	errInfo = fmt.Sprintf("%s have retried %d times but all failed", "GetMessageID", MaxRetryTime)
	Logger.Error(owner, appid, "", "GetMessageID", "get message id failed", errInfo)

ExitWithError:
	return 0, errors.New(errInfo)
}

/*
 * saves single notification record to correct DB and collection
 * @param message is message that will be saved
 * @param session is mongo session
 * @param appid is application id
 *
 * @return nil if saved successfully; otherwise a error is returned
 */
func SaveChatNotificationRecord(message *saver.ChatMessage, session *mgo.Session, appid uint16) error {
	message.Creation = time.Now()
	message.ExpireTime = message.Creation.Add(time.Duration(message.ExpireInterval) * 1000 * 1000 * 1000)
	countTime := message.ExpireTime.Add(time.Duration(netConf().DefaultExpire*-1) * 1000 * 1000 * 1000)

	b := bson.D{bson.DocElem{FieldJid, message.To}, bson.DocElem{FieldMsgId, message.MsgId},
		bson.DocElem{FieldMsgType, message.Type}, bson.DocElem{FieldMsgFrom, message.From},
		bson.DocElem{FieldMsgCTime, message.Creation}, bson.DocElem{FieldExpireTime, message.ExpireTime},
		bson.DocElem{FieldExpireInterval, message.ExpireInterval}, bson.DocElem{FieldCountTime, countTime},
		bson.DocElem{FieldMsgSn, message.TraceSN}, bson.DocElem{FieldModified, message.Creation},
		bson.DocElem{FieldMsgData, []byte(message.Content)}}

	collection := FormatCollection(NotificationMsgCol, appid)
	if err := session.DB(NotificationMsgDB).C(collection).Insert(&b); err != nil {
		return err
	}
	return nil
}

/*
 * saves notification messages in request to correct DB and collection
 * @param req is store message request
 * @param resp is response that will be sent to caller
 *
 * @return nil if saved successfully; otherwise a error is returned
 */
func SaveMongoNotifyChannel(req *saver.StoreMessagesRequest,
	resp *saver.StoreMessagesResponse) error {
	count := len(req.Messages)
	resp.Inbox = make(map[string]*saver.ChatMessage, count)

	for _, message := range req.Messages {
		var saved saver.ChatMessage
		saved = *message
		if err := SavePeerOrPublicRecord(&saved, req.Appid, req.ChatChannel); err != nil {
			Logger.Error(message.To, req.Appid, message.TraceSN, "SaveMongoNotifyChannel",
				"Saved message to storage failed", err)
			continue
		} else {
			resp.Inbox[message.To] = &saved
			Logger.Trace(message.To, req.Appid, message.TraceSN, "SaveMongoNotifyChannel",
				"Saved peer message successfully", saved.ToString("MsgID", "Type", "To", "From", "TraceSN"))
		}
	}
	if len(resp.Inbox) == 0 {
		err := fmt.Sprint("No any one message has been saved successfully, request message count: ", count)
		Logger.Error("", req.Appid, "", "SaveMongoNotifyChannel",
			"Saved message to storage failed", err)
		return errors.New(err)
	}
	return nil
}

/*
 * generic message save function that can save saver.ChatChannelPublic and saver.ChatChannelNotify messages
 * @param message is message that will be saved
 * @param appid is application ID
 * @param chatChannel is channel identifier, only support saver.ChatChannelPublic and saver.ChatChannelNotify
 *
 * @return nil if saved successfully; otherwise a error is returned
 */
func SavePeerOrPublicRecord(message *saver.ChatMessage, appid uint16, chatChannel string) (ret error) {
	if chatChannel == saver.ChatChannelPublic {
		message.To = saver.ChatChannelPublic
	}

	var errInfo string
	if sessions, err := GetMessageMongoStore(message.To, appid, fmt.Sprint(message.TraceSN)); err != nil {
		Logger.Error(message.To, appid, message.TraceSN, "SavePeerOrPublicMessage",
			"Saved message to storage failed", err)
		ret = err
	} else if len(sessions) == 0 {
		Logger.Error(message.To, appid, message.TraceSN, "SavePeerOrPublicMessage",
			"Saved message to storage failed", "no available mongo connection")
		ret = errors.New("no available mongo connection")
	} else {
		// need to close sessions to release socket connection
		defer CloseMgoSessions(sessions)

		// ensure that we have a valid expire interval
		if message.ExpireInterval <= 0 {
			message.ExpireInterval = netConf().DefaultExpire
		} else if message.ExpireInterval > MaxExpireTime {
			message.ExpireInterval = MaxExpireTime
		}

		for _, session := range sessions {
			var res_debug bson.M
			if msgId, err := GetNewMessageID(message.To, NotificationIdDB, FormatCollection(NotificationIdCol, appid),
				session, 1, appid, &res_debug); err != nil {
				errInfo = strings.Join([]string{errInfo, err.Error()}, "; ")

				continue
			} else {
				message.MsgId = msgId
				if chatChannel == saver.ChatChannelPublic {
					// store public message
					if err := SaveChatPublicRecord(message, session, appid); err != nil {
						info := fmt.Sprintf("save public message error: [%v], GetMessageID returned mongo operatation result:[%v]",
							err, res_debug)
						errInfo = strings.Join([]string{errInfo, info}, "; ")

						continue
					}
				} else {
					// store notify channel message
					if err := SaveChatNotificationRecord(message, session, appid); err != nil {
						info := fmt.Sprintf("save peer message error: [%v], GetMessageID returned mongo operatation result:[%v]",
							err, res_debug)
						errInfo = strings.Join([]string{errInfo, info}, "; ")

						continue
					} else {
						Logger.Trace(message.To, appid, message.TraceSN, "SavePeerOrPublicMessage", "Saved message",
							message.ToString("MsgID", "Type", "To", "From", "TraceSN"))
					}
				}
			}

			if len(errInfo) > 0 {
				Logger.Error(message.To, appid, message.TraceSN, "SavePeerOrPublicMessage",
					"Saved message failed", errInfo)
			}
			// clear error for now operation is successful
			errInfo = ""
			ret = nil
			break
		}
	}

	if len(errInfo) > 0 {
		ret = errors.New(errInfo)
		Logger.Error(message.To, appid, message.TraceSN, "SavePeerOrPublicMessage",
			"Saved message failed", errInfo)
	}

	return
}

/*
 * saves public messages in request to correct DB and collection
 * @param req is store message request
 * @param resp is response that will be sent to caller
 *
 * @return nil if saved successfully; otherwise a error is returned
 */
func SaveMongoPublicChannel(req *saver.StoreMessagesRequest,
	resp *saver.StoreMessagesResponse) error {
	count := len(req.Messages)

	resp.Inbox = make(map[string]*saver.ChatMessage, count)
	for key, message := range req.Messages {
		var saved saver.ChatMessage
		saved = *message
		if err := SavePeerOrPublicRecord(&saved, req.Appid, req.ChatChannel); err != nil {
			Logger.Error(message.To, req.Appid, message.TraceSN, "SaveMongoPublicChannel",
				"Saved message to storage failed", err)
			continue
		} else {
			resp.Inbox[key] = &saved
			Logger.Trace(message.To, req.Appid, message.TraceSN, "SaveMongoPublicChannel",
				"Saved public message successfully", saved.ToString("MsgID", "Type", "To", "From", "TraceSN"))
		}
	}
	if len(resp.Inbox) == 0 {
		err := fmt.Sprint("No any one message has been saved successfully, request message count: ", count)
		Logger.Error("", req.Appid, "", "SaveMongoPublicChannel", "Saved message to storage failed", err)
		return errors.New(err)
	}

	return nil
}

/*
 * saves single public message record to correct DB and collection
 * @param message is message that will be saved
 * @param session is mongo session
 * @param appid is application id
 *
 * @return nil if saved successfully; otherwise a error is returned
 */
func SaveChatPublicRecord(message *saver.ChatMessage, session *mgo.Session, appid uint16) error {
	message.Creation = time.Now()
	message.ExpireTime = message.Creation.Add(time.Duration(message.ExpireInterval) * 1000 * 1000 * 1000)
	countTime := message.ExpireTime.Add(time.Duration(netConf().DefaultExpire*-1) * 1000 * 1000 * 1000)

	b := bson.D{bson.DocElem{FieldMsgId, message.MsgId}, bson.DocElem{FieldMsgType, message.Type},
		bson.DocElem{FieldMsgCTime, message.Creation}, bson.DocElem{FieldExpireTime, message.ExpireTime},
		bson.DocElem{FieldExpireInterval, message.ExpireInterval}, bson.DocElem{FieldCountTime, countTime},
		bson.DocElem{FieldMsgSn, message.TraceSN}, bson.DocElem{FieldModified, message.Creation},
		bson.DocElem{FieldMsgData, []byte(message.Content)}}

	collection := FormatCollection(PublicMsgCol, appid)
	if err := session.DB(PubligMsgDB).C(collection).Insert(&b); err != nil {
		return err
	}
	return nil
}

/*
 * get last read id of owner at specified channel
 * @param appid is application id
 * @param owner is owner of last read id
 * @param chatChannel is chat channel, only support saver.ChatChannelNotify and saver.ChatChannelIM
 * @param session is mongo session
 *
 * @return (last read id, nil) if operation is successful, otherwise (0, error) is returned
 */
func GetLastRead(appid uint16, owner, chatChannel string, sessions []*mgo.Session) (int64, error) {
	var db, collection string
	switch chatChannel {
	case saver.ChatChannelNotify:
		db = NotificationLastReadDB
		collection = FormatCollection(NotificationLastReadCol, appid)
	case saver.ChatChannelIMInbox:
		fallthrough
	case saver.ChatChannelIMOutbox:
		fallthrough
	case saver.ChatChannelIM:
		db = ImLastReadDB
		collection = FormatCollection(ImLastReadCol, appid)
	default:
		err := "GetLastRead: " + fmt.Sprintf("%s channel is not supported now", chatChannel)
		Logger.Error(owner, appid, "", "GetLastRead", err, "")
		return 0, errors.New(err)
	}

	for i, session := range sessions {
		var result bson.M
		if err := session.DB(db).C(collection).Find(&bson.M{FieldJid: owner}).One(&result); err != nil {
			if err != mgo.ErrNotFound {
				Logger.Error(owner, appid, "", "GetLastRead", "find last read id failed",
					fmt.Sprintf("Error: %v, session index: %v", err, i))
				continue
			} else {
				// we can not find the last read id record, just set it as 0
				return 0, nil
			}
		}

		if res, err := GetIntegerFieldValue(&result, FiedlLastRead); err != nil {
			errorInfo := fmt.Sprintf("Error: %v, record: %v", err, result)
			Logger.Error(owner, appid, "", "GetLastRead", "get last read id value failed",
				errorInfo)
			return 0, errors.New(errorInfo)
		} else {
			// return correct value
			return res, nil
		}
	}

	err := fmt.Sprintf("%s have retried %d times but all failed", "GetLastRead", MaxRetryTime)
	Logger.Error(owner, appid, "", "GetLastRead", "find last read id failed", err)

	return 0, errors.New(err)
}

/*
 * update last read id of owner at specified channel
 * @param appid is application id
 * @param owner is owner of last read id
 * @param chatChannel is chat channel, only support saver.ChatChannelNotify and saver.ChatChannelIM
 * @param sessions is mongo session group
 * @param newLastRead is new will update to database
 *
 * @return nil if operation is successful, otherwise an error is returned
 */
func UpdateLastRead(appid uint16, owner, chatChannel string, sessions []*mgo.Session, newLastRead int64) error {
	var db, collection string
	switch chatChannel {
	case saver.ChatChannelNotify:
		db = NotificationLastReadDB
		collection = FormatCollection(NotificationLastReadCol, appid)
	case saver.ChatChannelIMInbox:
		fallthrough
	case saver.ChatChannelIMOutbox:
		fallthrough
	case saver.ChatChannelIM:
		db = ImLastReadDB
		collection = FormatCollection(ImLastReadCol, appid)
	default:
		err := "UpdateLastRead: " + fmt.Sprintf("%s channel is not supported now", chatChannel)
		Logger.Error(owner, appid, "", "GetLastRead", err, "")
		return errors.New(err)
	}

	change := mgo.Change{
		Update: bson.M{OpSet: bson.M{FiedlLastRead: newLastRead, FieldLatestModified: time.Now()}},
		Upsert: true}

	var errInfo string
	// we only update data when new last read id is more than DB value
	query := bson.D{bson.DocElem{FieldJid, owner},
		bson.DocElem{FiedlLastRead, bson.D{bson.DocElem{OpLte, newLastRead}}}}

	for idx, session := range sessions {
		if _, err := session.DB(db).C(collection).Find(&query).Apply(change, &bson.M{}); err != nil {
			// only log errors that not duplicate error
			if !strings.Contains(err.Error(), MongoErrorDuplicateCode) {
				info := fmt.Sprint("session index: ", idx, ", error: ", err)
				Logger.Error(owner, appid, "", "UpdateLastRead", "update last read id failed", info)
				errInfo = strings.Join([]string{errInfo, info}, "; ")
				continue // uses alternative session to retry if there is
			} else {
				Logger.Debug(owner, appid, "", "UpdateLastRead", "duplicate error", err)
			}
		}

		if len(errInfo) > 0 {
			Logger.Error(owner, appid, "", "UpdateLastRead",
				"UpdateLastRead operation is successful with some error", errInfo)
		}
		errInfo = ""

		break
	}

	if len(errInfo) > 0 {
		return errors.New(errInfo)
	}

	return nil
}

/*
 * get latest message id of owner at specified channel
 * @param appid is application id
 * @param owner is owner of latest read id
 * @param chatChannel is chat channel, only support saver.ChatChannelNotify and saver.ChatChannelIM
 * @param session is mongo session
 *
 * @return nil if operation is successful, otherwise an error is returned
 */
func GetLatestMessageId(appid uint16, owner, chatChannel string, sessions []*mgo.Session) (uint64, error) {
	var db, collection string
	switch chatChannel {
	case saver.ChatChannelPublic:
		fallthrough // public latest id shares peer message DB
	case saver.ChatChannelNotify:
		db = NotificationIdDB
		collection = FormatCollection(NotificationIdCol, appid)
	case saver.ChatChannelIMInbox:
		fallthrough
	case saver.ChatChannelIMOutbox:
		fallthrough
	case saver.ChatChannelIM:
		db = ImIdDB
		collection = FormatCollection(ImIdCol, appid)
	default:
		err := "GetLatestMessageId: " + fmt.Sprintf("%s channel is not supported now", chatChannel)
		Logger.Error(owner, appid, "", "GetLatestMessageId", err, "")
		return 0, errors.New(err)
	}

	for i, session := range sessions {
		var result bson.M
		if err := session.DB(db).C(collection).Find(&bson.M{FieldJid: owner}).One(&result); err != nil {
			if err != mgo.ErrNotFound {
				Logger.Error(owner, appid, "", "GetLatestMessageId", "find latest id failed",
					fmt.Sprintf("Error: %v, session index: %v", err, i))
				continue
			} else {
				// we can not find the latest id record, just set it as 0
				return 0, nil
			}
		}
		if res, err := GetIntegerFieldValue(&result, FieldLatest); err != nil {
			errorInfo := fmt.Sprintf("Error: %v, record: %v", err, result)
			Logger.Error(owner, appid, "", "GetLatestMessageId", "get latest id value failed",
				errorInfo)
			return 0, errors.New(errorInfo)
		} else {
			// return correct value
			return uint64(res), nil
		}
	}

	err := fmt.Sprintf("%s have retried %d times but all failed", "GetLatestMessageId", MaxRetryTime)
	Logger.Error(owner, appid, "", "GetLatestMessageId", "find latest id failed", err)

	return 0, errors.New(err)
}

/*
 * Retrieve chat messages from mongo storage
 * @param req is a saver.RetrieveMessagesRequest point which include required information to retrieve messages
 * @param resp is a saver.RetrieveMessagesResponse point which includes response
 * @return nil if no error occurs, otherwise an error interface is returned
 */
func RetrieveMongoChatMessages(req *saver.RetrieveMessagesRequest,
	resp *saver.RetrieveMessagesResponse) error {

	channels := len(req.ChatChannels)
	if channels == 0 {
		Logger.Error(req.Owner, req.Appid, "", "RetrieveMongoChatMessages",
			"Retrieve messages from storage failed", "there isn't any channel information")
		return errors.New("there isn't any channel information")
	}

	resp.Inbox = make(map[string][]*saver.ChatMessage, channels)
	resp.LatestID = make(map[string]uint64, channels)
	resp.LastReadID = make(map[string]uint64, channels)
	errorInfo := ""
	for key, info := range req.ChatChannels {
		info.Channel = key //ensure that data consistency
		switch key {
		case saver.ChatChannelIMInbox, saver.ChatChannelIMOutbox, saver.ChatChannelIM:
			// we don't do anything if get empty owner, just return as operation is successful but count it as failed
			if req.Owner == "" {
				requestStat.AtomicAddRetrieveImFails(1)
				continue
			}
			resp.Outbox = make(map[string][]*saver.ChatMessage, channels)
			if err := RetrieveImRecords(req.Appid, req.Owner, info, req.TraceSN, resp); err != nil {
				errorInfo = fmt.Sprint(errorInfo, "IM Error[", err.Error(), "] ")
				requestStat.AtomicAddRetrieveImFails(1)
			} else {
				requestStat.AtomicAddRetrieveIms(1)
			}
		case saver.ChatChannelNotify:
			// we don't do anything if get empty owner, just return as operation is successful but count it as failed
			if req.Owner == "" {
				requestStat.AtomicAddRetrievePeerFails(1)
				continue
			}
			if err := RetrieveNotificationRecords(req.Appid, req.Owner, info, req.TraceSN, resp); err != nil {
				errorInfo = fmt.Sprint(errorInfo, "Peer Error[", err.Error(), "] ")
				requestStat.AtomicAddRetrievePeerFails(1)
			} else {
				requestStat.AtomicAddRetrievePeers(1)
			}
		case saver.ChatChannelPublic:
			if err := RetrievePublicRecordsFromCache(req.Appid, req.Owner, info, req.TraceSN, resp); err != nil {
				errorInfo = fmt.Sprint(errorInfo, "Public/hot Error[", err.Error(), "] ")
			} else {
				requestStat.AtomicAddRetrievePublics(1)
			}
		default:
			err := fmt.Sprint(fmt.Sprintf("%s channel is not supported now", info.Channel),
				" Invalid channel information:", info)
			errorInfo = fmt.Sprint(errorInfo, "Unsupported channel:[", err, "] ")
		}
	}

	if errorInfo == "" {
		Logger.Trace("", req.Appid, "", "RetrieveMongoChatMessages",
			"Request", req, "Response", resp)
		return nil
	}
	Logger.Error(req.Owner, req.Appid, "", "RetrieveMongoChatMessages", "Retrieve message operation has error",
		errorInfo)
	return errors.New(errorInfo)
}

// 获取未读消息数量
func RetrieveMongoUnreadCount(req *saver.RetrieveMessagesRequest) (*saver.RetrieveMessagesResponse, error) {
	channels := len(req.ChatChannels)
	if channels == 0 {
		Logger.Error(req.Owner, req.Appid, "", "RetrieveMongoUnreadCount",
			"Retrieve unread count from storage failed", "there isn't any channel information")
		return nil, errors.New("there isn't any channel information")
	}
	resp := &saver.RetrieveMessagesResponse{
		Appid:      req.Appid,
		Owner:      req.Owner,
		LatestID:   make(map[string]uint64, channels),
		LastReadID: make(map[string]uint64, channels),
	}

	errorInfo := ""
	for key, info := range req.ChatChannels {
		info.Channel = key //ensure that data consistency
		switch key {
		case saver.ChatChannelIMInbox, saver.ChatChannelIMOutbox, saver.ChatChannelIM:
			// we don't do anything if get empty owner, just return as operation is successful but count it as failed
			if req.Owner == "" {
				requestStat.AtomicAddRetrieveImFails(1)
				continue
			}
			if err := RetrieveImUnreadCount(req.Appid, req.Owner, info, req.TraceSN, resp); err != nil {
				errorInfo = fmt.Sprint(errorInfo, "IM Error[", err.Error(), "] ")
				requestStat.AtomicAddRetrieveImFails(1)
			} else {
				requestStat.AtomicAddRetrieveIms(1)
			}
		case saver.ChatChannelNotify:
			// we don't do anything if get empty owner, just return as operation is successful but count it as failed
			if req.Owner == "" {
				requestStat.AtomicAddRetrievePeerFails(1)
				continue
			}
			if err := RetrieveNotificationUnreadCount(req.Appid, req.Owner, info, req.TraceSN, resp); err != nil {
				errorInfo = fmt.Sprint(errorInfo, "Peer Error[", err.Error(), "] ")
				requestStat.AtomicAddRetrievePeerFails(1)
			} else {
				requestStat.AtomicAddRetrievePeers(1)
			}
		case saver.ChatChannelPublic:
			// public 不支持查询
			fallthrough
		default:
			err := fmt.Sprint(fmt.Sprintf("%s channel is not supported now", info.Channel),
				" Invalid channel information:", info)
			errorInfo = fmt.Sprint(errorInfo, "Unsupported channel:[", err, "] ")
		}
	}

	if errorInfo == "" {
		Logger.Trace("", req.Appid, "", "RetrieveMongoUnreadCount",
			"Request", req, "Response", resp)
		return resp, nil
	}
	Logger.Error(req.Owner, req.Appid, "", "RetrieveMongoUnreadCount", "Retrieve unread count operation has error",
		errorInfo)
	return nil, errors.New(errorInfo)
}

/*
 * retrieve notification messages
 * @param appid is application id
 * @param owner is owner of messages
 * @param channelInfo includes required information to retrieve messages
 * @param traceSn is used to trace procedure
 * @param resp is response will return to rpc caller
 *
 * @return nil if operation is successful; otherwise an error is returned
 */
func RetrieveNotificationRecords(appid uint16, owner string, channelInfo *saver.RetrieveChannel,
	traceSn string, resp *saver.RetrieveMessagesResponse) error {
	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countP2pResponseTime(owner, traceSn, "RetrieveNotificationRecords", appid)
		defer countFunc()
	}

	if sessions, err := GetMessageMongoStore(owner, appid, traceSn); err != nil {
		Logger.Error(owner, appid, traceSn, "RetrieveNotificationRecords",
			"Retrieve message from storage failed", err)
		return err
	} else if len(sessions) == 0 {
		Logger.Error(owner, appid, traceSn, "RetrieveNotificationRecords",
			"Retrieve message from storage failed", "no available mongo connection")
		return errors.New("no available mongo connection")
	} else {
		// need to close sessions to release socket connection
		defer CloseMgoSessions(sessions)

		sortField := FieldMsgId
		if channelInfo.MaxCount == 0 {
			channelInfo.MaxCount = DefaultReturnCount
		} else if channelInfo.MaxCount < 0 {
			// if channelInfo.StartMsgId and channelInfo.MaxCount both are negative, we set sort field later
			if channelInfo.StartMsgId > 0 {
				sortField = "-" + sortField
			}
			channelInfo.MaxCount = -1 * channelInfo.MaxCount
		}

		if channelInfo.StartMsgId == 0 {
			if lastRead, err := GetLastRead(appid, owner, channelInfo.Channel, sessions); err == nil && lastRead > 0 {
				channelInfo.StartMsgId = lastRead + 1
			}
		}

		// makes complex query:
		// condition is that:
		// owner is specified owner,
		// no recall field or recall field value is not 1
		// sorted field is message id field
		var query *bson.D
		if channelInfo.StartMsgId >= 0 {
			if strings.HasPrefix(sortField, "-") {
				// query condition: message id is less than start id
				// retrieves messages in descending order of message id,
				//  result does not include message which id is channelInfo.StartMsgId.
				query = &bson.D{bson.DocElem{FieldJid, owner},
					bson.DocElem{FieldMsgId, bson.D{bson.DocElem{OpLt, channelInfo.StartMsgId}}},
					bson.DocElem{OpOr, []bson.D{
						{bson.DocElem{FieldRecalled, bson.D{bson.DocElem{OpExists, false}}}},
						{bson.DocElem{FieldRecalled, bson.D{bson.DocElem{OpNe, 1}}}}}}}
			} else {
				// query condition: message id is equal or lager than start id,
				// retrieves messages in ascending order of message id,
				//  result includes message which id is channelInfo.StartMsgId if there is.
				query = &bson.D{bson.DocElem{FieldJid, owner},
					bson.DocElem{FieldMsgId, bson.D{bson.DocElem{OpGte, channelInfo.StartMsgId}}},
					bson.DocElem{OpOr, []bson.D{
						{bson.DocElem{FieldRecalled, bson.D{bson.DocElem{OpExists, false}}}},
						{bson.DocElem{FieldRecalled, bson.D{bson.DocElem{OpNe, 1}}}}}}}
			}
		} else {
			// retrieve latest messages in descending order from latest message
			query = &bson.D{bson.DocElem{FieldJid, owner},
				bson.DocElem{OpOr, []bson.D{
					{bson.DocElem{FieldRecalled, bson.D{bson.DocElem{OpExists, false}}}},
					{bson.DocElem{FieldRecalled, bson.D{bson.DocElem{OpNe, 1}}}}}}}
			sortField = "-" + sortField
		}

		messages := make([]*saver.ChatMessage, 0, channelInfo.MaxCount)
		if records, err := find(query, channelInfo.MaxCount, sessions, NotificationMsgDB,
			FormatCollection(NotificationMsgCol, appid), sortField); err != nil {
			return err
		} else {
			for _, record := range records {
				// translate a record to saver.ChatMessage object; if an error occurred, translated message will not
				// return to caller but only log it to error log
				if message, err := TranslateBsonM2Message(&record, channelInfo.Channel); err != nil {
					Logger.Error(owner, appid, traceSn, "RetrieveNotificationRecords",
						"Retrieve message from storage failed",
						fmt.Sprint("error:[", err.Error(), "] translated message:[", message, "]"))
					continue
				} else {
					messages = append(messages, message)
				}
			}
		}

		resp.Inbox[channelInfo.Channel] = messages

		// get max message id/count retrieved
		maxMsgId := uint64(0)
		queriedCount := len(resp.Inbox[channelInfo.Channel])
		if queriedCount > 0 {
			maxMsgId = resp.Inbox[channelInfo.Channel][queriedCount-1].MsgId
			if resp.Inbox[channelInfo.Channel][0].MsgId > maxMsgId {
				maxMsgId = resp.Inbox[channelInfo.Channel][0].MsgId
			}
		}

		Logger.Debug(owner, appid, traceSn, "RetrieveNotificationRecords", "Retrieved notification records",
			fmt.Sprintf("Retrieved messages of %s, start: %v, request len: %v, return: %v, retrun max id: %v",
				owner, channelInfo.StartMsgId, channelInfo.MaxCount, queriedCount, maxMsgId))

		// get latest id of channel
		latest, err := GetLatestMessageId(appid, owner, channelInfo.Channel, sessions)
		if err != nil {
			Logger.Error(owner, appid, traceSn, "RetrieveNotificationRecords",
				"get latest id failed", err)
		} else {
			resp.LatestID[channelInfo.Channel] = latest
		}

		// try to update last read id, update will failed if new last read is less than db record
		lastRead, _ := GetLastRead(appid, owner, channelInfo.Channel, sessions)
		if lastRead < (channelInfo.StartMsgId-1) && latest > 0 {
			lastRead = channelInfo.StartMsgId - 1
			if lastRead > int64(latest) {
				lastRead = int64(latest) // 确保最后读取ID是一个有效的ID
			}

			UpdateLastRead(appid, owner, channelInfo.Channel, sessions, lastRead)
		}
		resp.LastReadID[channelInfo.Channel] = uint64(lastRead)
	}
	return nil
}

// 获取 peer 通道未读消息数量
func RetrieveNotificationUnreadCount(appid uint16, owner string, channelInfo *saver.RetrieveChannel,
	traceSn string, resp *saver.RetrieveMessagesResponse) error {
	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countP2pResponseTime(owner, traceSn, "RetrieveNotificationUnreadCount", appid)
		defer countFunc()
	}

	if sessions, err := GetMessageMongoStore(owner, appid, traceSn); err != nil {
		Logger.Error(owner, appid, traceSn, "RetrieveNotificationUnreadCount",
			"Retrieve unread count from storage failed", err)
		return err
	} else if len(sessions) == 0 {
		Logger.Error(owner, appid, traceSn, "RetrieveNotificationUnreadCount",
			"Retrieve unread count from storage failed", "no available mongo connection")
		return errors.New("no available mongo connection")
	} else {
		// need to close sessions to release socket connection
		defer CloseMgoSessions(sessions)

		lastRead, err := GetLastRead(appid, owner, channelInfo.Channel, sessions)
		if err != nil {
			Logger.Error(owner, appid, traceSn, "RetrieveNotificationUnreadCount", "get lastRead id failed", err)
			return err
		}

		// get latest id of channel
		latest, err := GetLatestMessageId(appid, owner, channelInfo.Channel, sessions)
		if err != nil {
			Logger.Error(owner, appid, traceSn, "RetrieveNotificationUnreadCount", "get latest id failed", err)
			return err
		}

		Logger.Debug(owner, appid, traceSn, "RetrieveNotificationUnreadCount", lastRead, latest)

		resp.LatestID[channelInfo.Channel] = latest
		resp.LastReadID[channelInfo.Channel] = uint64(lastRead)
	}
	return nil
}

/*
 * retrieve public messages
 * @param appid is application id
 *
 * @return (PublicMsgCache, nil) if operation is successful; otherwise a (nil, error) pair is returned
 */
func RetrievePublicRecords(appid uint16) (*PublicMsgCache, error) {
	sessions, err := GetMessageMongoStore(saver.ChatChannelPublic, appid, "")
	if err != nil {
		Logger.Error("", appid, "", "RetrievePublicRecords",
			"Retrieve message from public storage failed", err)
		return nil, err
	} else if len(sessions) == 0 {
		Logger.Error("", appid, "", "RetrievePublicRecords",
			"Retrieve message from public storage failed", "no available mongo connection")
		return nil, errors.New("no available mongo connection")
	}
	// need to close sessions to release socket connection
	defer CloseMgoSessions(sessions)

	ret := &PublicMsgCache{messages: make(map[uint64]*saver.ChatMessage)}
	// query all, sorted field is message id field desc
	query := &bson.D{bson.DocElem{FieldMsgId, bson.D{bson.DocElem{OpGte, 0}}}}
	if records, err := find(query, maxPublicMessageCache, sessions, PubligMsgDB,
		FormatCollection(PublicMsgCol, appid), "-"+FieldMsgId); err != nil {
		return nil, err
	} else {
		for _, record := range records {
			// translate a record to saver.ChatMessage object; if an error occurred, translated message will not
			// return to caller but only log it to error log
			message, err := TranslateBsonM2Message(&record, saver.ChatChannelPublic)
			if err != nil {
				Logger.Debug("", appid, "", "RetrievePublicRecords", "Retrieve message from storage failed",
					fmt.Sprint("error:[", err.Error(), "] translated message:[", message, "]"))
			} else {
				ret.messages[message.MsgId] = message
				if message.MsgId > ret.maxMsgId {
					ret.maxMsgId = message.MsgId // get max message id/count retrieved
				}
				if message.MsgId < ret.minMsgId || ret.minMsgId == 0 {
					ret.minMsgId = message.MsgId
				}
			}
		}
	}

	Logger.Debug("", appid, "", "RetrievePublicRecords", "Retrieved public records",
		fmt.Sprintf("Retrieved public messages, request len: %v, return: %v, retrun max id: %v",
			maxPublicMessageCache, len(ret.messages), ret.maxMsgId))

	// get latest id of public
	if latest, err := GetLatestMessageId(appid, saver.ChatChannelPublic, saver.ChatChannelPublic, sessions); err != nil {
		Logger.Error("", appid, "", "RetrievePublicRecords", "get latest id failed", err)
	} else {
		ret.latest = latest
	}

	return ret, nil
}

/*
 * saves IM messages in request to correct DB and collection
 * @param req is store message request
 * @param resp is response that will be sent to caller
 *
 * @return nil if saved successfully; otherwise a error is returned
 */
func SaveMongoImChannel(req *saver.StoreMessagesRequest,
	resp *saver.StoreMessagesResponse) error {
	resp.Inbox = make(map[string]*saver.ChatMessage, len(req.Messages))
	resp.Outbox = make(map[string]*saver.ChatMessage, len(req.Messages))
	for key, message := range req.Messages {
		if err := SaveImMessage(message, req.Appid, resp, key); err != nil {
			Logger.Error(message.To, req.Appid, message.TraceSN, "SaveMongoImChannel",
				"Saved message to storage failed", err)
			continue
		} else {
			Logger.Trace(message.To, req.Appid, message.TraceSN, "SaveMongoImChannel",
				"Saved IM message with no error", "")
		}
	}
	if len(resp.Inbox) == 0 {
		Logger.Error("", req.Appid, "", "SaveMongoImChannel",
			"Saved message to storage failed", "No any one message has been saved successfully")
		return errors.New("No any one message has been saved successfully")
	}
	return nil
}

/*
 * saves im message record to inbox
 * @param message is message that will be saved
 * @param session is mongo session
 * @param appid is application id
 *
 * @return nil if saved successfully; otherwise a error is returned
 */
func saveChatImInboxRecord(message *saver.ChatMessage, session *mgo.Session, appid uint16) error {
	message.Creation = time.Now()
	message.ExpireTime = message.Creation.Add(time.Duration(message.ExpireInterval) * 1000 * 1000 * 1000)
	countTime := message.ExpireTime.Add(time.Duration(netConf().DefaultExpire*-1) * 1000 * 1000 * 1000)

	b := bson.D{bson.DocElem{FieldOwner, message.To}, bson.DocElem{FieldMsgId, message.MsgId},
		bson.DocElem{FieldMsgType, message.Type}, bson.DocElem{FieldMsgFrom, message.From},
		bson.DocElem{FieldMsgCTime, message.Creation}, bson.DocElem{FieldExpireTime, message.ExpireTime},
		bson.DocElem{FieldExpireInterval, message.ExpireInterval}, bson.DocElem{FieldCountTime, countTime},
		bson.DocElem{FieldMsgSn, message.TraceSN}, bson.DocElem{FieldModified, message.Creation},
		bson.DocElem{FieldMsgData, []byte(message.Content)}}

	collection := FormatCollection(ImInboxCol, appid)
	if err := session.DB(ImInboxDB).C(collection).Insert(&b); err != nil {
		return err
	}
	return nil
}

/*
 * saves single public message record to outbox
 * @param message is message that will be saved
 * @param session is mongo session
 * @param appid is application id
 *
 * @return nil if saved successfully; otherwise a error is returned
 */
func saveChatImOutboxRecord(message *saver.ChatMessage, inboxId uint64, session *mgo.Session, appid uint16) error {
	message.Creation = time.Now()
	message.ExpireTime = message.Creation.Add(time.Duration(message.ExpireInterval) * 1000 * 1000 * 1000)
	countTime := message.ExpireTime.Add(time.Duration(netConf().DefaultExpire*-1) * 1000 * 1000 * 1000)
	var buf bytes.Buffer
	buf.WriteString(message.Content)
	b := bson.D{bson.DocElem{FieldOwner, message.From}, bson.DocElem{FieldMsgId, message.MsgId},
		bson.DocElem{FieldMsgType, message.Type}, bson.DocElem{FieldMsgTo, message.To},
		bson.DocElem{FieldMsgCTime, message.Creation}, bson.DocElem{FieldExpireTime, message.ExpireTime},
		bson.DocElem{FieldExpireInterval, message.ExpireInterval}, bson.DocElem{FieldCountTime, countTime},
		bson.DocElem{FieldMsgSn, message.TraceSN}, bson.DocElem{FieldModified, message.Creation},
		bson.DocElem{FieldInboxMsgId, inboxId}, bson.DocElem{FieldMsgData, buf.Bytes()}}

	collection := FormatCollection(ImOutBoxCol, appid)
	if err := session.DB(ImOutboxDB).C(collection).Insert(&b); err != nil {
		return err
	}
	return nil
}

/*
 * im message save function that can save im message to inbox and outbox if need
 * @param message is message that will be saved
 * @param appid is application ID
 * @param resp is response to rpc caller, which will stores stored message
 * @param key is used to identify message response
 *
 * @return nil if saved successfully; otherwise a error is returned
 */
func SaveImMessage(message *saver.ChatMessage, appid uint16, resp *saver.StoreMessagesResponse, key string) error {
	inboxId := uint64(0)
	// first step, always save message to receiver's inbox
	if sessions, err := GetMessageMongoStore(message.To, appid, fmt.Sprint(message.TraceSN)); err != nil {
		Logger.Error(message.To, appid, message.TraceSN, "SaveImMessage",
			"Saved message to inbox storage failed", err)
		return err
	} else if len(sessions) == 0 {
		Logger.Error(message.To, appid, message.TraceSN, "SaveImMessage",
			"Saved message to inbox storage failed", "no available mongo connection")
		return errors.New("no available mongo connection")
	} else {
		// need to close sessions to release socket connection
		defer CloseMgoSessions(sessions)

		var saveMsg saver.ChatMessage
		saveMsg = *message
		if err := saveImChatRecord(appid, sessions, &saveMsg, nil); err != nil {
			return err
		} else {
			// store saved message to response
			resp.Inbox[key] = &saveMsg
			inboxId = saveMsg.MsgId // it will be used when save outbox
		}
	}

	// second step: save message to sender's outbox if need
	if inboxId > 0 && message.StoreOutbox == 1 {
		if sessions, err := GetMessageMongoStore(message.From, appid, fmt.Sprint(message.TraceSN)); err != nil {
			Logger.Error(message.From, appid, message.TraceSN, "SaveImMessage",
				"Saved message to outbox storage failed", err)
			return err
		} else if len(sessions) == 0 {
			Logger.Error(message.From, appid, message.TraceSN, "SaveImMessage",
				"Saved message to outbox storage failed", "no available mongo connection")
			return errors.New("no available mongo connection")
		} else {
			// need to close sessions to release socket connection
			defer CloseMgoSessions(sessions)

			var saveMsg saver.ChatMessage
			saveMsg = *message

			if err := saveImChatRecord(appid, sessions, &saveMsg, &inboxId); err != nil {
				return err
			} else {
				// store saved message to response
				resp.Outbox[key] = &saveMsg
			}
		}
	}

	return nil
}

/*
 * Save im message record to inbox/outbox
 * @param appid is application ID
 * @param sessions is mongo session group
 * @param message is message that will be saved
 * @param inboxId is related inbox message id when save outbox message;
 *      Note that this also indicates which box should message is saved to;
 *      nil indicates that message should be saved to inbox; otherwise message should be saved to outbox
 *
 * @return nil if saved successfully; otherwise a error is returned
 */
func saveImChatRecord(appid uint16, sessions []*mgo.Session, message *saver.ChatMessage, inboxId *uint64) error {
	owner := ""
	if inboxId == nil {
		owner = message.To
	} else {
		owner = message.From
	}

	// ensure that we have a valid expire interval
	if message.ExpireInterval <= 0 {
		message.ExpireInterval = netConf().DefaultExpire
	} else if message.ExpireInterval > MaxExpireTime {
		message.ExpireInterval = MaxExpireTime
	}

	lastErr := ""
	for _, session := range sessions {
		var res_debug bson.M
		if msgId, err := GetNewMessageID(owner, ImIdDB, FormatCollection(ImIdCol, appid),
			session, 2, appid, &res_debug); err != nil {
			lastErr = strings.Join([]string{lastErr, err.Error()}, "; ")
			continue
		} else {
			// in the past, inbox is odd number, outbox is even number;
			// in that case client might retrieve twice for inbox for latest is larger than latest inbox id
			// now we set inbox as even number while set outbox as odd number
			if (msgId%2 == 0 && inboxId == nil) || (msgId%2 == 1 && inboxId != nil) {
				message.MsgId = msgId
			} else {
				message.MsgId = msgId - 1
			}

			if inboxId == nil {
				// store im channel message to inbox
				if err := saveChatImInboxRecord(message, session, appid); err != nil {
					Logger.Error(owner, appid, message.TraceSN, "saveImChatRecord",
						"Saved message to inbox failed",
						fmt.Sprintf("error: [%v], GetMessageID returned mongo operatation result:[%v]", err, res_debug))

					lastErr = strings.Join([]string{lastErr, err.Error()}, "; ")
					continue
				} else {
					Logger.Trace(owner, appid, message.TraceSN, "saveImChatRecord", "Saved message to inbox",
						message.ToString("MsgID", "Type", "To", "From", "TraceSN"))
				}
			} else {
				// store im channel message to outbox
				if err := saveChatImOutboxRecord(message, *inboxId, session, appid); err != nil {
					Logger.Error(owner, appid, message.TraceSN, "saveImChatRecord",
						"Saved message to outbox failed",
						fmt.Sprintf("error: [%v], GetMessageID returned mongo operatation result:[%v]", err, res_debug))

					lastErr = strings.Join([]string{lastErr, err.Error()}, "; ")
					continue
				} else {
					Logger.Trace(owner, appid, message.TraceSN, "saveImChatRecord", "Saved message to outbox",
						message.ToString("MsgID", "Type", "To", "From", "TraceSN"))
				}
			}
		}

		if len(lastErr) > 0 {
			Logger.Error(owner, appid, message.TraceSN, "saveImChatRecord",
				"Saved im message successfully with some error", lastErr)
		}
		return nil // just return for we actually saved message without error
	}

	err := fmt.Sprintf("have tried %d times but all failed, error: %s", len(sessions), lastErr)
	Logger.Error(owner, appid, message.TraceSN, "saveImChatRecord",
		"Saved im message failed", err)
	return errors.New(err)
}

/*
 * retrieve im messages
 * @param appid is application id
 * @param owner is owner of messages
 * @param channelInfo includes required information to retrieve messages
 * @param traceSn is used to trace procedure
 * @param resp is response will return to rpc caller
 *
 * @return nil if operation is successful; otherwise an error is returned
 */
func RetrieveImRecords(appid uint16, owner string, channelInfo *saver.RetrieveChannel,
	traceSn string, resp *saver.RetrieveMessagesResponse) error {
	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countP2pResponseTime(owner, traceSn, "RetrieveImRecords", appid)
		defer countFunc()
	}

	if sessions, err := GetMessageMongoStore(owner, appid, traceSn); err != nil {
		Logger.Error(owner, appid, traceSn, "RetrieveImRecords",
			"Retrieve message from storage failed", err)
		return err
	} else if len(sessions) == 0 {
		Logger.Error(owner, appid, traceSn, "RetrieveImRecords",
			"Retrieve message from storage failed", "no available mongo connection")
		return errors.New("no available mongo connection")
	} else {
		// need to close sessions to release socket connection
		defer CloseMgoSessions(sessions)

		sortField := FieldMsgId
		if channelInfo.MaxCount == 0 {
			channelInfo.MaxCount = DefaultReturnCount
		} else if channelInfo.MaxCount < 0 {
			// if channelInfo.StartMsgId and channelInfo.MaxCount both are negative, we set sort field later
			if channelInfo.StartMsgId > 0 {
				sortField = "-" + sortField
			}
			channelInfo.MaxCount = -1 * channelInfo.MaxCount
		}

		if channelInfo.StartMsgId == 0 {
			if lastRead, err := GetLastRead(appid, owner, channelInfo.Channel, sessions); err == nil && lastRead > 0 {
				channelInfo.StartMsgId = lastRead + 1
			}
		}

		// makes complex query:
		// condition is that:
		// owner is specified owner,
		// message id is lager than start id,
		// no recall field or recall field value is 1
		// sorted field is message id field
		var query *bson.D
		if channelInfo.StartMsgId >= 0 {
			// retrieve messages forwardly, not inlcude message which id is channelInfo.StartMsgId
			if strings.HasPrefix(sortField, "-") {
				query = &bson.D{bson.DocElem{FieldOwner, owner},
					bson.DocElem{FieldMsgId, bson.D{bson.DocElem{OpLt, channelInfo.StartMsgId}}},
					bson.DocElem{OpOr, []bson.D{
						{bson.DocElem{FieldRecalled, bson.D{bson.DocElem{OpExists, false}}}},
						{bson.DocElem{FieldRecalled, bson.D{bson.DocElem{OpNe, 1}}}}}}}
			} else {
				query = &bson.D{bson.DocElem{FieldOwner, owner},
					bson.DocElem{FieldMsgId, bson.D{bson.DocElem{OpGte, channelInfo.StartMsgId}}},
					bson.DocElem{OpOr, []bson.D{
						{bson.DocElem{FieldRecalled, bson.D{bson.DocElem{OpExists, false}}}},
						{bson.DocElem{FieldRecalled, bson.D{bson.DocElem{OpNe, 1}}}}}}}
			}
		} else {
			// retrieve latest messages
			query = &bson.D{bson.DocElem{FieldOwner, owner},
				bson.DocElem{OpOr, []bson.D{
					{bson.DocElem{FieldRecalled, bson.D{bson.DocElem{OpExists, false}}}},
					{bson.DocElem{FieldRecalled, bson.D{bson.DocElem{OpNe, 1}}}}}}}
			sortField = "-" + sortField
		}

		inboxMsgs := make([]*saver.ChatMessage, 0, channelInfo.MaxCount)

		if channelInfo.Channel == saver.ChatChannelIM || channelInfo.Channel == saver.ChatChannelIMInbox {
			if records, err := find(query, channelInfo.MaxCount, sessions, ImInboxDB,
				FormatCollection(ImInboxCol, appid), sortField); err != nil {
				return err
			} else {
				for _, record := range records {
					// translate a record to saver.ChatMessage object; if an error occurred, translated message will not
					// return to caller but only log it to error log
					if message, err := TranslateBsonM2Message(&record, saver.ChatChannelIMInbox); err != nil {
						Logger.Error(owner, appid, traceSn, "RetrieveImRecords",
							"Retrieve message from storage failed",
							fmt.Sprint("error:[", err.Error(), "] translated message:[", message.ToString(), "]"))
						continue
					} else {
						inboxMsgs = append(inboxMsgs, message)
					}
				}
			}
		}

		outboxMsgs := make([]*saver.ChatMessage, 0, channelInfo.MaxCount)

		if channelInfo.Channel == saver.ChatChannelIM || channelInfo.Channel == saver.ChatChannelIMOutbox {
			if records, err := find(query, channelInfo.MaxCount, sessions, ImOutboxDB,
				FormatCollection(ImOutBoxCol, appid), FieldMsgId); err != nil {
				return err
			} else {
				for _, record := range records {
					// translate a record to saver.ChatMessage object; if an error occurred, translated message will not
					// return to caller but only log it to error log
					if message, err := TranslateBsonM2Message(&record, saver.ChatChannelIMOutbox); err != nil {
						Logger.Error(owner, appid, traceSn, "RetrieveImRecords",
							"Retrieve message from storage failed",
							fmt.Sprint("error:[", err.Error(), "] translated message:[", message.ToString(), "]"))
						continue
					} else {
						outboxMsgs = append(outboxMsgs, message)
					}
				}
			}
		}

		resp.Inbox[channelInfo.Channel] = inboxMsgs
		resp.Outbox[channelInfo.Channel] = outboxMsgs

		// get max message id/count retrieved
		maxMsgId := uint64(0)
		queriedInbox := len(resp.Inbox[channelInfo.Channel])
		queriedOutbox := len(resp.Outbox[channelInfo.Channel])
		queriedCount := queriedInbox + queriedOutbox
		if queriedInbox > 0 {
			maxMsgId = resp.Inbox[channelInfo.Channel][queriedInbox-1].MsgId
			if resp.Inbox[channelInfo.Channel][0].MsgId > maxMsgId {
				maxMsgId = resp.Inbox[channelInfo.Channel][0].MsgId
			}
		}

		if queriedOutbox > 0 &&
			(resp.Outbox[channelInfo.Channel][queriedOutbox-1].MsgId > maxMsgId ||
				resp.Outbox[channelInfo.Channel][0].MsgId > maxMsgId) {
			maxMsgId = resp.Outbox[channelInfo.Channel][queriedOutbox-1].MsgId
			if resp.Outbox[channelInfo.Channel][0].MsgId > maxMsgId {
				maxMsgId = resp.Outbox[channelInfo.Channel][0].MsgId
			}
		}

		Logger.Debug(owner, appid, traceSn, "RetrieveImRecords", "Retrieved notification records",
			fmt.Sprintf("Retrieved messages of %s, start: %v, request len: %v, return: %v (inbox: %v, outbox %v), retrun max id: %v",
				owner, channelInfo.StartMsgId, channelInfo.MaxCount, queriedCount, queriedInbox, queriedOutbox, maxMsgId))

		// get latest id of channel
		latest, err := GetLatestMessageId(appid, owner, channelInfo.Channel, sessions)
		if err != nil {
			Logger.Error(owner, appid, traceSn, "RetrieveImRecords",
				"get latest id failed", err)
		} else {
			resp.LatestID[channelInfo.Channel] = latest
		}

		// try to update last read id, update will failed if new last read is less than db record
		lastRead, _ := GetLastRead(appid, owner, channelInfo.Channel, sessions)
		if lastRead < (channelInfo.StartMsgId-1) && latest > 0 {
			lastRead = channelInfo.StartMsgId - 1
			if lastRead > int64(latest) {
				lastRead = int64(latest) // 确保最后读取ID是一个有效的ID
			}

			UpdateLastRead(appid, owner, channelInfo.Channel, sessions, lastRead)
		}
		resp.LastReadID[channelInfo.Channel] = uint64(lastRead)
	}
	return nil
}

// 取 im 通道未读数量
func RetrieveImUnreadCount(appid uint16, owner string, channelInfo *saver.RetrieveChannel,
	traceSn string, resp *saver.RetrieveMessagesResponse) error {
	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countP2pResponseTime(owner, traceSn, "RetrieveImUnreadCount", appid)
		defer countFunc()
	}

	if sessions, err := GetMessageMongoStore(owner, appid, traceSn); err != nil {
		Logger.Error(owner, appid, traceSn, "RetrieveImUnreadCount",
			"Retrieve unread count from storage failed", err)
		return err
	} else if len(sessions) == 0 {
		Logger.Error(owner, appid, traceSn, "RetrieveImUnreadCount",
			"Retrieve unread count from storage failed", "no available mongo connection")
		return errors.New("no available mongo connection")
	} else {
		// need to close sessions to release socket connection
		defer CloseMgoSessions(sessions)

		lastRead, err := GetLastRead(appid, owner, channelInfo.Channel, sessions)
		if err != nil {
			Logger.Error(owner, appid, traceSn, "RetrieveImUnreadCount", "get lastRead id failed", err)
			return err
		}

		// get latest id of channel
		latest, err := GetLatestMessageId(appid, owner, channelInfo.Channel, sessions)
		if err != nil {
			Logger.Error(owner, appid, traceSn, "RetrieveImUnreadCount", "get latest id failed", err)
			return err
		}

		Logger.Debug(owner, appid, traceSn, "RetrieveImUnreadCount", lastRead, latest)

		resp.LatestID[channelInfo.Channel] = latest
		resp.LastReadID[channelInfo.Channel] = uint64(lastRead)
	}
	return nil
}

/*
 * Recall chat messages
 * @param req is a saver.RetrieveMessagesRequest point which include required information to retrieve messages
 * @param resp is a saver.RetrieveMessagesResponse point which includes response
 * @return nil if no error occurs, otherwise an error interface is returned
 */
func RecallMongoChatMessages(req *saver.RecallMessagesRequest,
	resp *saver.StoreMessagesResponse) error {
	Logger.Trace(req.Receiver, req.Appid, req.TraceSN, "RecallMongoChatMessages",
		"Get recall message request", req)

	switch req.ChatChannel {
	case saver.ChatChannelNotify:
		return recallNotifyChannelMessage(req, resp)
	case saver.ChatChannelIM:
		return recallImChannelMessage(req, resp)
	case saver.ChatChannelPublic:
	}

	err := fmt.Sprintf("%s channel is not supported now", req.ChatChannel)
	Logger.Error(req.Receiver, req.Appid, req.TraceSN, "RecallMongoChatMessages",
		"Recall message failed", err)
	return errors.New(err)
}

/*
 * Recall notification chat messages
 * @param req is a saver.RetrieveMessagesRequest point which include required information to retrieve messages
 * @param resp is a saver.RetrieveMessagesResponse point which includes response
 * @return nil if no error occurs, otherwise an error interface is returned
 */
func recallNotifyChannelMessage(req *saver.RecallMessagesRequest,
	resp *saver.StoreMessagesResponse) error {
	if sessions, err := GetMessageMongoStore(req.Receiver, req.Appid, req.TraceSN); err != nil {
		Logger.Error(req.Receiver, req.Appid, req.TraceSN, "RecallNotifyChannelMessage",
			"Recall message failed", err)
		return err
	} else if len(sessions) == 0 {
		Logger.Error(req.Receiver, req.Appid, req.TraceSN, "RecallNotifyChannelMessage",
			"Recall message failed", "no available mongo connection")
		return errors.New("no available mongo connection")
	} else {
		// need to close sessions to release socket connection
		defer CloseMgoSessions(sessions)

		query := bson.D{
			bson.DocElem{FieldJid, req.Receiver},
			bson.DocElem{FieldMsgFrom, req.Sender},
			bson.DocElem{FieldMsgId, req.InboxId}}

		collection := FormatCollection(NotificationMsgCol, req.Appid)
		if _, err := findRecallMessage(query, sessions, NotificationMsgDB, collection); err != nil {
			Logger.Error(req.Receiver, req.Appid, req.TraceSN, "RecallNotifyChannelMessage",
				"Recall message failed", err)
			return err
		} else {
			// store a recall message first
			message := &saver.ChatMessage{
				To:      req.Receiver,
				From:    req.Sender,
				Type:    RecallMsgType,
				TraceSN: time.Now().UnixNano(),
				Content: strconv.FormatUint(req.InboxId, 10)}
			if err := SavePeerOrPublicRecord(message, req.Appid, req.ChatChannel); err != nil {
				Logger.Error(req.Receiver, req.Appid, req.TraceSN, "RecallNotifyChannelMessage",
					"Recall message failed when store new recall message", err)
				return err
			}
			resp.Inbox = make(map[string]*saver.ChatMessage)
			resp.Inbox[req.Receiver] = message

			// set recall flag for message which need to be recalled
			update := bson.M{OpSet: bson.M{FieldRecalled: RecalledFlag, FieldModified: time.Now()}}
			if err := findAndModify(sessions, NotificationMsgDB, collection, query, update, false, false,
				nil); err != nil {
				Logger.Error(req.Receiver, req.Appid, req.TraceSN, "RecallNotifyChannelMessage",
					"Set recall message flag failed", err)
			}
		}
	}

	Logger.Trace(req.Receiver, req.Appid, req.TraceSN, "RecallNotifyChannelMessage",
		"Request", req, "Response", resp)
	return nil
}

/*
 * try to find recall message according to query
 * @param query is query conditions
 * @param session is mongo session
 * @param db is database that query is executed on
 * @param collection is collection that query is executed on
 * @return (record, nil) if message is found and it is not recalled yet;
 *  (nil, saver.ErrNotFound) if not found;
 *  (nil, saver.ErrRecalledYet) if message has been recalled yet;
 *  otherwise (nil, error) is returned
 */
func findRecallMessage(query interface{}, sessions []*mgo.Session, db, collection string) (*bson.M, error) {
	if record, err := findOne(query, sessions, db, collection); err != nil {
		if err == mgo.ErrNotFound {
			return nil, saver.ErrNotFound
		}
		return nil, err
	} else {
		if v, err := GetIntegerFieldValue(record, FieldRecalled); err == nil {
			if v == RecalledFlag {
				return nil, saver.ErrRecalledYet
			}
		}

		return record, nil
	}
}

/*
 * Recall im chat messages recall to im channel mongo storage
 * @param req is a saver.RetrieveMessagesRequest point which include required information to retrieve messages
 * @param resp is a saver.RetrieveMessagesResponse point which includes response
 * @return nil if no error occurs, otherwise an error interface is returned
 */
func recallImChannelMessage(req *saver.RecallMessagesRequest,
	resp *saver.StoreMessagesResponse) error {

	recallInboxId := uint64(0)

	// first step, recall message in receiver's inbox
	if sessions, err := GetMessageMongoStore(req.Receiver, req.Appid, req.TraceSN); err != nil {
		Logger.Error(req.Receiver, req.Appid, req.TraceSN, "RecallImChannelMessage",
			"Recall inbox message failed", err)
		return err
	} else if len(sessions) == 0 {
		Logger.Error(req.Receiver, req.Appid, req.TraceSN, "RecallImChannelMessage",
			"Recall inbox message failed", "no available mongo connection")
		return errors.New("no available mongo connection")
	} else {
		// need to close sessions to release socket connection
		defer CloseMgoSessions(sessions)

		query := bson.D{
			bson.DocElem{FieldOwner, req.Receiver},
			bson.DocElem{FieldMsgFrom, req.Sender},
			bson.DocElem{FieldMsgId, req.InboxId}}

		collection := FormatCollection(ImInboxCol, req.Appid)
		if _, err := findRecallMessage(query, sessions, ImInboxDB, collection); err != nil {
			Logger.Error(req.Receiver, req.Appid, req.TraceSN, "RecallImChannelMessage",
				"Recall inbox message failed", err)
			return err
		} else {
			// store a recall message first
			message := &saver.ChatMessage{
				To:      req.Receiver,
				From:    req.Sender,
				Type:    RecallMsgType,
				TraceSN: time.Now().UnixNano(),
				Content: strconv.FormatUint(req.InboxId, 10)}
			if err := saveImChatRecord(req.Appid, sessions, message, nil); err != nil {
				Logger.Error(req.Receiver, req.Appid, req.TraceSN, "RecallImChannelMessage",
					"Recall inbox message failed when store new recall message", err)
				return err
			}
			resp.Inbox = make(map[string]*saver.ChatMessage)
			resp.Inbox[req.Receiver] = message
			recallInboxId = message.MsgId

			// set recall flag for message which need to be recalled
			update := bson.M{OpSet: bson.M{FieldRecalled: RecalledFlag, FieldModified: time.Now()}}
			if err := findAndModify(sessions, ImInboxDB, collection, query, update, false, false,
				nil); err != nil {
				Logger.Error(req.Receiver, req.Appid, req.TraceSN, "RecallImChannelMessage",
					"Set recall message flag failed", err)
			}
		}
	}

	// second step, recall message in sender's outbox if need
	if recallInboxId > 0 {
		if sessions, err := GetMessageMongoStore(req.Sender, req.Appid, req.TraceSN); err != nil {
			Logger.Error(req.Sender, req.Appid, req.TraceSN, "RecallImChannelMessage",
				"Recall outbox message failed", err)
			// here we don't return error even though we reached error when recall outbox message
			// for our main purpose is that recall inbox message
		} else if len(sessions) == 0 {
			Logger.Error(req.Sender, req.Appid, req.TraceSN, "RecallImChannelMessage",
				"Recall outbox message failed", "no available mongo connection")
		} else {
			// need to close sessions to release socket connection
			defer CloseMgoSessions(sessions)

			query := bson.D{
				bson.DocElem{FieldOwner, req.Sender},
				bson.DocElem{FieldMsgTo, req.Receiver},
				bson.DocElem{FieldInboxMsgId, req.InboxId}}

			collection := FormatCollection(ImOutBoxCol, req.Appid)
			if record, err := findRecallMessage(query, sessions, ImOutboxDB, collection); err != nil {
				Logger.Error(req.Sender, req.Appid, req.TraceSN, "RecallImChannelMessage",
					"Recall outbox message failed", err)
			} else {
				if outboxId, err := GetIntegerFieldValue(record, FieldMsgId); err != nil {
					Logger.Error(req.Sender, req.Appid, req.TraceSN, "RecallImChannelMessage",
						"Recall outbox message failed", err)
				} else {
					// store a recall message first
					message := &saver.ChatMessage{
						To:      req.Receiver,
						From:    req.Sender,
						Type:    RecallMsgType,
						TraceSN: time.Now().UnixNano(),
						Content: strconv.FormatInt(outboxId, 10)}
					if err := saveImChatRecord(req.Appid, sessions, message, &recallInboxId); err != nil {
						Logger.Error(req.Sender, req.Appid, req.TraceSN, "RecallImChannelMessage",
							"Recall outbox message failed when store new recall message", err)
					} else {
						resp.Outbox = make(map[string]*saver.ChatMessage)
						resp.Outbox[req.Receiver] = message

						query := bson.D{
							bson.DocElem{FieldOwner, req.Sender},
							bson.DocElem{FieldMsgId, outboxId}}

						// set recall flag for message which need to be recalled
						update := bson.M{OpSet: bson.M{FieldRecalled: RecalledFlag, FieldModified: time.Now()}}
						if err := findAndModify(sessions, ImOutboxDB, collection, query, update, false, false,
							nil); err != nil {
							Logger.Error(req.Sender, req.Appid, req.TraceSN, "RecallImChannelMessage",
								"Set outbox recall message flag failed", err)
						}
					}
				}
			}
		}
	}

	Logger.Trace(req.Receiver, req.Appid, req.TraceSN, "RecallImChannelMessage",
		"Request", req, "Response", resp)
	return nil
}
