// provides utility operations
package main

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"

	"github.com/huajiao-tv/qchat/client/saver"
	"gopkg.in/mgo.v2/bson"
)

const (
	ErrorNotFound  = "not found %s field"
	ErrorWrongType = "%s get wrong field type value(%T)"
)

/*
 * translate a byte array to a uint64
 * @param bytes is a byte array which should have 8 members at least
 * @return (uint64, nil) if no error occurs, otherwise (0, an error interface) is returned
 */
func Bytes2Uint64(bytes []byte) (uint64, error) {
	if len(bytes) < 8 {
		return 0, errors.New("invalid md5 value")
	}

	result := uint64(bytes[7])
	result = (result << 8) + uint64(bytes[6])
	result = (result << 8) + uint64(bytes[5])
	result = (result << 8) + uint64(bytes[4])
	result = (result << 8) + uint64(bytes[3])
	result = (result << 8) + uint64(bytes[2])
	result = (result << 8) + uint64(bytes[1])
	result = (result << 8) + uint64(bytes[0])

	return result, nil
}

/*
 * translate a string to uint64 value according to md5
 */
func Md5Uint64Hash(key string) (uint64, error) {
	hashMd5 := md5.New()
	io.WriteString(hashMd5, key)
	keyBytes := hashMd5.Sum(nil)
	return Bytes2Uint64(keyBytes)
}

func GetIntegerFieldValue(values *bson.M, fn string) (int64, error) {
	var res int64
	value, ok := (*values)[fn]
	if !ok {
		return 0, errors.New(fmt.Sprintf(ErrorNotFound, fn))
	}
	switch v := value.(type) {
	case int32:
		res = int64(v)
	case int:
		res = int64(v)
	case int8:
		res = int64(v)
	case int16:
		res = int64(v)
	case float32:
		res = int64(v)
	case float64:
		res = int64(v)
	case uint:
		res = int64(v)
	case uint8:
		res = int64(v)
	case uint16:
		res = int64(v)
	case uint32:
		res = int64(v)
	case uint64:
		res = int64(v)
	case int64:
		res = v
	default:
		return 0, errors.New(fmt.Sprintf(ErrorWrongType, "GetIntegerFieldValue", v))
	}
	return res, nil
}

func GetNumberFieldValue(values *bson.M, fn string) (float64, error) {
	var res float64
	value, ok := (*values)[fn]
	if !ok {
		return 0, errors.New(fmt.Sprintf(ErrorNotFound, fn))
	}
	switch v := value.(type) {
	case int:
		res = float64(v)
	case int64:
		res = float64(v)
	case float32:
		res = float64(v)
	case float64:
		res = v
	default:
		return 0, errors.New(fmt.Sprintf(ErrorWrongType, "GetNumberFieldValue", v))
	}
	return res, nil
}

func GetStringFieldValue(values *bson.M, fn string) (string, error) {
	var res string
	value, ok := (*values)[fn]
	if !ok {
		return "", errors.New(fmt.Sprintf(ErrorNotFound, fn))
	}
	switch v := value.(type) {
	case int32:
		res = strconv.FormatInt(int64(v), 10)
	case int:
		res = strconv.FormatInt(int64(v), 10)
	case int8:
		res = strconv.FormatInt(int64(v), 10)
	case int16:
		res = strconv.FormatInt(int64(v), 10)
	case uint:
		res = strconv.FormatInt(int64(v), 10)
	case uint8:
		res = strconv.FormatInt(int64(v), 10)
	case uint16:
		res = strconv.FormatInt(int64(v), 10)
	case uint32:
		res = strconv.FormatInt(int64(v), 10)
	case uint64:
		res = strconv.FormatInt(int64(v), 10)
	case int64:
		res = strconv.FormatInt(v, 10)
	case float32:
		res = strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		res = strconv.FormatFloat(v, 'f', -1, 64)
	case []byte:
		res = string(v)
	case string:
		res = v
	default:
		return "", errors.New(fmt.Sprintf(ErrorWrongType, "GetStringFieldValue", v))
	}
	return res, nil
}

func GetBinaryFieldValue(values *bson.M, fn string) ([]byte, error) {
	var res []byte
	value, ok := (*values)[fn]
	if !ok {
		return nil, errors.New(fmt.Sprintf(ErrorNotFound, fn))
	}
	switch v := value.(type) {
	case []byte:
		res = v
	case string:
		bytes := bytes.Buffer{}
		bytes.WriteString(v)
		res = bytes.Bytes()
	default:
		return nil, errors.New(fmt.Sprintf(ErrorWrongType, "GetBinaryFieldValue", v))
	}
	return res, nil
}

func GetDateFieldValue(values *bson.M, fn string) (time.Time, error) {
	value, ok := (*values)[fn]
	if !ok {
		return time.Now(), errors.New(fmt.Sprintf(ErrorNotFound, fn))
	}
	switch v := value.(type) {
	case time.Time:
		return v, nil
	default:
		return time.Now(), errors.New(fmt.Sprintf(ErrorWrongType, "GetDateFieldValue", v))
	}
}

func FormatCollection(collection string, appid uint16) string {
	if suffix, ok := netConf().AppidTablesSuffix[int(appid)]; ok {
		return collection + "_" + suffix
	}

	return collection
}

func TranslateBsonM2Message(record *bson.M, chatChannel string) (*saver.ChatMessage, error) {
	message := saver.ChatMessage{}

	var errorInfo string
	if msgId, err := GetIntegerFieldValue(record, FieldMsgId); err != nil {
		errorInfo = fmt.Sprint(errorInfo, " ", err.Error())
	} else {
		message.MsgId = uint64(msgId)
	}

	switch chatChannel {
	case saver.ChatChannelPublic:
		if to, err := GetStringFieldValue(record, FieldJid); err == nil {
			message.To = to
		}

		if from, err := GetStringFieldValue(record, FieldMsgFrom); err == nil {
			message.From = from
		}
	case saver.ChatChannelNotify:
		if to, err := GetStringFieldValue(record, FieldJid); err != nil {
			errorInfo = fmt.Sprint(errorInfo, " ", err.Error())
		} else {
			message.To = to
		}

		if from, err := GetStringFieldValue(record, FieldMsgFrom); err != nil {
			errorInfo = fmt.Sprint(errorInfo, " ", err.Error())
		} else {
			message.From = from
		}
	case saver.ChatChannelIMInbox:
		if to, err := GetStringFieldValue(record, FieldOwner); err != nil {
			errorInfo = fmt.Sprint(errorInfo, " ", err.Error())
		} else {
			message.To = to
		}

		if from, err := GetStringFieldValue(record, FieldMsgFrom); err != nil {
			errorInfo = fmt.Sprint(errorInfo, " ", err.Error())
		} else {
			message.From = from
		}
	case saver.ChatChannelIMOutbox:
		if to, err := GetStringFieldValue(record, FieldMsgTo); err != nil {
			errorInfo = fmt.Sprint(errorInfo, " ", err.Error())
		} else {
			message.To = to
		}

		if from, err := GetStringFieldValue(record, FieldOwner); err != nil {
			errorInfo = fmt.Sprint(errorInfo, " ", err.Error())
		} else {
			message.From = from
		}

		// mark the message is outbox message
		message.Box = 1
	}

	if content, err := GetStringFieldValue(record, FieldMsgData); err != nil {
		errorInfo = fmt.Sprint(errorInfo, " ", err.Error())
	} else {
		message.Content = content
	}

	if msgType, err := GetIntegerFieldValue(record, FieldMsgType); err != nil {
		errorInfo = fmt.Sprint(errorInfo, " ", err.Error())
	} else {
		message.Type = uint32(msgType)
	}

	if traceSn, err := GetIntegerFieldValue(record, FieldMsgSn); err != nil {
		errorInfo = fmt.Sprint(errorInfo, " ", err.Error())
	} else {
		message.TraceSN = traceSn
	}

	if interval, err := GetIntegerFieldValue(record, FieldExpireInterval); err != nil {
		errorInfo = fmt.Sprint(errorInfo, " ", err.Error())
	} else {
		message.ExpireInterval = int(interval)
	}

	if creation, err := GetDateFieldValue(record, FieldMsgCTime); err != nil {
		errorInfo = fmt.Sprint(errorInfo, " ", err.Error())
	} else {
		message.Creation = creation
	}

	if expire, err := GetDateFieldValue(record, FieldExpireTime); err != nil {
		errorInfo = fmt.Sprint(errorInfo, " ", err.Error())
	} else {
		message.ExpireTime = expire
	}

	if errorInfo != "" {
		return &message, errors.New(errorInfo)
	}
	return &message, nil
}

func isEqualStringSlice(a, b []string) bool {
	return isEqualSlice(ToSlice(a), ToSlice(b))
}

func ToSlice(s interface{}) []interface{} {
	val := reflect.ValueOf(s)

	switch val.Kind() {
	case reflect.Slice:
		len := val.Len()
		ret := make([]interface{}, len)
		for idx := 0; idx < len; idx++ {
			ret[idx] = val.Index(idx).Interface()
		}
		return ret
	default:
		return []interface{}{s}
	}
}

func isEqualSlice(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for idx, val := range a {
		if val != b[idx] {
			return false
		}
	}
	return true
}
