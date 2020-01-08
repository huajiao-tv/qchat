package llconn

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

//
// invoke HTTP interface to send peer|public|im|chatroom messages
//
type MessageSender struct {
	PeerUrl      string
	ChatroomUrl  string
	RobotJoinUrl string
	RobotQuitUrl string
	PublicUrl    string
	IMUrl        string
}

func NewMessageSender(center string) (*MessageSender, error) {
	if len(center) == 0 {
		return nil, errors.New("center host is empty")
	}

	imUrl := "http://" + center + "/push/chat"
	peerUrl := "http://" + center + "/huajiao/users"
	publicUrl := "http://" + center + "/huajiao/all"
	chatroomUrl := "http://" + center + "/send"
	robotJoinUrl := "http://" + center + "/joinroom"
	robotQuitUrl := "http://" + center + "/quitroom"

	ms := MessageSender{
		PeerUrl:      peerUrl,
		ChatroomUrl:  chatroomUrl,
		RobotJoinUrl: robotJoinUrl,
		RobotQuitUrl: robotQuitUrl,
		PublicUrl:    publicUrl,
		IMUrl:        imUrl,
	}

	return &ms, nil
}

func (this *MessageSender) base64Encode(m map[string]string) string {
	if nil == m {
		return ""
	}
	var s string
	for k, v := range m {
		s += "&" + k + "=" + v
	}
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func (this *MessageSender) checkResponse(resp *http.Response) string {
	if nil == resp {
		fmt.Println("empty response")
		return ""
	}

	b1, err1 := ioutil.ReadAll(resp.Body)
	if nil != err1 {
		fmt.Println("ReadAll failed")
		return ""
	}

	res := string(b1)

	fmt.Println("status code", resp.StatusCode, "response", res)
	return res
}

//
//访问为POST方式
//
//post数据内容：&m=111&b=Jm1zZz0xMjM0NV90ZXN0JnRyYWNlaWQ9MTIzNDU2Nzg5MCZyZWNlaXZlcnM9amlkXzEwMDAmc2VuZGVyPWppZF8yMDAw
//示例数据中b的内容为&msg=12345_test&traceid=1234567890&receivers=jid_1000&sender=jid_2000的base64编码
//请注意traceid的范围是无符号32位整数的范围
//m的值是b的未编码内容加上"wpms"后计算出来的md5，目前没有打开校验，可以随便填
//
//返回值为json格式字符串
//{"code":0,"reason":"ok","ios_offline_users":[],"msgids":[{"owner":"jid_1000","id":113,"box":"inbox"},{"owner":"jid_2000","id":24,"box":"outbox"}]}
//
func (this *MessageSender) SendPeerMessage(receivers string, content []byte, expiration, priority, appid int) (string, error) {
	if this == nil || len(receivers) == 0 || content == nil || len(content) == 0 {
		return "", errors.New("invalid arguments")
	}

	msg := string(content)
	values := make(map[string]string, 4)
	values["appid"] = strconv.Itoa(appid)
	values["msg"] = DefaultMessageMeter.WrapMessage(msg)
	values["traceid"] = "1234567890"
	values["receivers"] = receivers
	values["sender"] = "123456"

	b := this.base64Encode(values)

	body := "&m=123&b=" + b

	resp, err := http.Post(this.PeerUrl, "text", strings.NewReader(body))
	if nil != err {
		fmt.Println("http.Post failed", err)
		return "", err
	}

	res := this.checkResponse(resp)

	return res, nil
}

func (this *MessageSender) SendImMessage(receivers string, content []byte, expiration, priority, appid int) (string, error) {
	if this == nil || len(receivers) == 0 || content == nil || len(content) == 0 {
		return "", errors.New("invalid arguments")
	}

	msg := string(content)
	values := make(map[string]string, 4)
	values["appid"] = strconv.Itoa(appid)
	values["msg"] = DefaultMessageMeter.WrapMessage(msg)
	values["traceid"] = "1234567890"
	values["receivers"] = receivers
	values["sender"] = "123456"

	b := this.base64Encode(values)

	body := "&m=123&b=" + b

	resp, err := http.Post(this.IMUrl, "text", strings.NewReader(body))
	if nil != err {
		fmt.Println("http.Post failed", err)
		return "", err
	}

	res := this.checkResponse(resp)

	return res, nil
}

func (this *MessageSender) SendPublicMessage(content []byte, appid int) (string, error) {
	if this == nil || content == nil || len(content) == 0 {
		return "", errors.New("invalid arguments")
	}

	msg := string(content)
	msg = DefaultMessageMeter.WrapMessage(msg)
	values := make(map[string]string, 2)
	values["appid"] = strconv.Itoa(appid)
	values["msg"] = msg
	values["traceid"] = "1234567890"

	b := this.base64Encode(values)

	body := "&m=123&b=" + b

	reader := strings.NewReader(body)

	resp, err := http.Post(this.PublicUrl, "text", reader)
	if nil != err {
		fmt.Println("http.Post failed", err)
		return "", err
	}

	res := this.checkResponse(resp)

	return res, nil
}

func (this *MessageSender) SendChatroomMessage(roomid, sender string, content []byte, expiration, msgtype, priority, appid int) (string, error) {

	fmt.Println("url", this.ChatroomUrl)

	values := make(map[string][]string, 6)
	values["appid"] = []string{strconv.Itoa(appid)}
	values["roomid"] = []string{roomid}
	values["sender"] = []string{sender}
	values["content"] = []string{DefaultMessageMeter.WrapMessage(string(content))}
	values["expire"] = []string{strconv.Itoa(expiration)}
	values["type"] = []string{strconv.Itoa(msgtype)}
	values["priority"] = []string{strconv.Itoa(priority)}
	values["traceid"] = []string{"123456"}

	resp, err := http.PostForm(this.ChatroomUrl, values)
	if nil != err {
		fmt.Println("failed", err)
		return "", err
	}

	res := this.checkResponse(resp)

	return res, nil
}

func (this *MessageSender) SendRobotJoinMessage(roomid, userid string) (string, error) {
	//userid is robot's id, robot is also an user in qchat
	fmt.Println("url", this.RobotJoinUrl)

	values := make(map[string][]string, 6)
	values["rid"] = []string{roomid}
	values["uid"] = []string{userid}

	resp, err := http.PostForm(this.RobotJoinUrl, values)
	if nil != err {
		fmt.Println("failed", err)
		return "", err
	}

	res := this.checkResponse(resp)

	return res, nil
}

func (this *MessageSender) SendRobotQuitMessage(roomid, userid string) (string, error) {
	//userid is robot's id, robot is also an user in qchat
	fmt.Println("url", this.RobotQuitUrl)

	values := make(map[string][]string, 6)
	values["rid"] = []string{roomid}
	values["uid"] = []string{userid}

	resp, err := http.PostForm(this.RobotQuitUrl, values)
	if nil != err {
		fmt.Println("failed", err)
		return "", err
	}

	res := this.checkResponse(resp)

	return res, nil
}
