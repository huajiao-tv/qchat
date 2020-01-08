package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var httpClient *http.Client
var msgs map[string][]string
var contentTpl []string = []string{}
var realName []string = []string{}
var avatar []string = []string{}
var biggift []string = []string{}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	httpClient = &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(network, addr, time.Duration(1000)*time.Millisecond)
				if err != nil {
					return nil, err
				}
				return c, nil
			},
			MaxIdleConnsPerHost: 1000,
		},
		Timeout: time.Duration(2000) * time.Millisecond,
	}
	msgs = map[string][]string{
		"share": []string{
			`{"roomid":"$$roomid$$","type":42,"text":"\u5206\u4eab\u4e86\u4e3b\u64ad\uff0c\u4e3a\u4e0a\u70ed\u95e8\u505a\u51fa\u5de8\u5927\u8d21\u732e","time":1473606966,"expire":86400,"extends":{"liveid":"32072840","userid":"59960928","nickname":"\u3000  \u3065\uc655\uc790\u30c5 +.\ufe4e","avatar":"http:\/\/image.huajiao.com\/fdb9324d935768ab95d4103a87068091-100_100.jpg","verified":false,"verifiedinfo":{"credentials":"\u4e0d\u4e89\u4e0d\u62a2\u4e0d\u70ab\u8000 \u6211\u4f1a\u9ed8\u9ed8\u7684\u575a\u5f3a\u3002\u4e0d\u505a\u9650\u91cf\u7248\uff0c\u53ea\u60f3\u505a\u552f\u4e00\u3002","type":0,"realname":"\u3000  \u3065\uc655\uc790\u30c5 +.\ufe4e","status":0,"error":"","official":false},"verify_student":{"vs_status":0,"option_student":"Y","vs_realname":"","vs_school":""},"exp":212839,"level":18},"traceid":"$$traceid$$", "proirity":10}`,
		},
		"biggift": []string{
			`{"roomid":$$roomid$$,"type":30,"text":"","time":1473606261,"expire":86400,"extends":{"contents":"","limit_amount":-1,"largev":0,"creatime":"2016-09-11 23:04:21","receiver_balance":593,"receiver_income":"0","receiver_income_b":"415","receiver_income_p":"2825","sender_balance":5505,"receiver":{"avatar":"http:\/\/image.huajiao.com\/0f34c7f4bb1cf36d6d411dd49da62438-100_100.jpg","nickname":"\u79c0\u598dJoan\ud83d\udc95\ud83d\udc95","uid":$$receiver$$,"verified":false,"verifiedinfo":{"credentials":"\u6a21\u7279\u4e00\u679a \u4f1a\u5531\u6b4c \u4f1a\u8df3\u821e \u4f1a\u5520\u55d1\n\nvb\uff1a\u79c0\u598dJoan    \n\n\u611f\u8c22\u966a\u4f34\uff0c\u5fae\u4fe1\u60c5\u5230\u6df1\u5904\u81ea\u7136\u52a0\uff01","realname":"\u79c0\u598dJoan\ud83d\udc95\ud83d\udc95","type":0},"exp":115883,"level":15,"medal":[]},"sender":{"avatar":"http:\/\/image.huajiao.com\/3787d30a82cdb02401ca09e1422f7982-100_100.jpg","nickname":"\u561f\uff5e\u561f\uff5e","uid":65902420,"verified":false,"verifiedinfo":{"credentials":"","realname":"\u561f\uff5e\u561f\uff5e","type":0},"exp":66207,"level":13,"medal":[]},"giftinfo":{"giftid":"1441","giftname":"\u8ffd\u7231\u5170\u535a\u57fa\u5c3c","amount":"4888","icon":"http:\/\/static.huajiao.com\/huajiao\/gift\/zalbjn1256.png","pic":"http:\/\/static.huajiao.com\/huajiao\/gift\/zalbjn800.png","content":"","relativeInfo":{"property":{"repeatGift":0,"effectGift":1,"property_android":{"effectGift":1,"pausetime":"15000","desctop":"\u8ffd\u7231\u5170\u535a\u57fa\u5c3c","desc":"4888\u8c46","pic":"http:\/\/static.huajiao.com\/huajiao\/gift\/zalbjn800.png","points":1,"gif":"http:\/\/static.huajiao.com\/huajiao\/gifteffect\/10061_31.zip","tagImage":"http:\/\/static.huajiao.com\/huajiao\/gift\/xinjiaobao256.png"},"property_ios":{"effectPngGift":1,"screenshottime":"5","desctop":"\u8ffd\u7231\u5170\u535a\u57fa\u5c3c","desc":"4888\u8c46","pic":"http:\/\/static.huajiao.com\/huajiao\/gifteffect\/10061_3.zip","giftIdentity":"10061_3","points":1,"gif":"http:\/\/static.huajiao.com\/huajiao\/gifteffect\/10061_31.zip","tagImage":"http:\/\/static.huajiao.com\/huajiao\/gift\/xinjiaobao256.png"}}}},"title":"","scheme":"huajiao:\/\/huajiao.com\/goto\/wallet?userid=59581333"},"traceid":"$$traceid$$", "proirity":102}`,
		},
		"gift": []string{
			//`{"roomid":32306652,"type":30,"text":"","time":1473692403,"expire":86400,"extends":{"contents":"","limit_amount":-1,"largev":0,"creatime":"2016-09-12 23:00:03","receiver_balance":83,"receiver_income":"0","receiver_income_b":"107","receiver_income_p":"54141","sender_balance":10,"receiver":{"avatar":"http:\/\/image.huajiao.com\/3b3e9614091b1896b41d72ba325d06dc-100_100.jpg","nickname":"23\u53f7\u4e54\u4e54","uid":35166873,"verified":false,"verifiedinfo":{"credentials":"\u4f60\u90a3\u4e48\u53ef\u7231\u90a3\u4e48\u5e05 \u8fd8\u6765\u770b\u6211\u76f4\u64ad","realname":"23\u53f7\u4e54\u4e54","type":0},"exp":14911,"level":9,"medal":[]},"sender":{"avatar":"http:\/\/image.huajiao.com\/e1b1f1b4cdf78a0d34c9c9a4042ed17c-100_100.jpg","nickname":"\ud83d\udd25\u5433\u5c11\ud83d\udd25","uid":37724382,"verified":false,"verifiedinfo":{"credentials":"","realname":"\ud83d\udd25\u5433\u5c11\ud83d\udd25","type":0},"exp":878444,"level":23,"medal":[{"kind":"tuhao","medal":"2"}]},"giftinfo":{"giftid":"1453","giftname":"\u82f9\u679c7","amount":"7","icon":"http:\/\/static.huajiao.com\/huajiao\/gift\/pg256.png","pic":"http:\/\/static.huajiao.com\/huajiao\/gift\/pg256.png","content":"","relativeInfo":{"repeatId":"@37724382351668733230665214736924036364173","repeatNum":1,"property":{"repeatGift":1,"property_android":{"repeatGift":"1","desctop":"\u82f9\u679c7","desc":"7\u8c46","pic":"http:\/\/static.huajiao.com\/huajiao\/gift\/pg256.png","points":1,"gif":"http:\/\/static.huajiao.com\/huajiao\/20024_30.zip","tagImage":"http:\/\/static.huajiao.com\/huajiao\/gift\/xinjiaobao256.png"},"property_ios":{"repeatGift":"1","desctop":"\u82f9\u679c7","desc":"7\u8c46","pic":"http:\/\/static.huajiao.com\/huajiao\/gift\/pg256.png","points":1,"gif":"http:\/\/static.huajiao.com\/huajiao\/20024_30.zip","tagImage":"http:\/\/static.huajiao.com\/huajiao\/gift\/xinjiaobao256.png"}}}},"title":"","scheme":"huajiao:\/\/huajiao.com\/goto\/wallet?userid=35166873"},"traceid":"71235aed7db92c975871c4dcb5e5810d"}`,
			//`{"roomid":$$roomid$$,"type":30,"text":"","time":1473681606,"expire":86400,"extends":{"contents":"","limit_amount":-1,"largev":0,"creatime":"2016-09-12 20:00:06","receiver_balance":48,"receiver_income":"0","receiver_income_b":"703","receiver_income_p":"238501","sender_balance":8135,"receiver":{"avatar":"http:\/\/image.huajiao.com\/5f6d66d87c23248c98897db96a748e61-100_100.jpg","nickname":"*Unique-\u5c0f\u6a59\u5b50\ud83c\udf4a","uid":$$receiver$$,"verified":true,"verifiedinfo":{"credentials":"\ud83d\udc44\u611f\u8c22\u6bcf\u4e00\u4e2a\u2764\ufe0f\u559c\u6b22\u4e0e\u652f\u6301\u5c0f\u6a59\u5b50\u7684\u4eba\uff0c\u4e48\u4e48\u54d2\ud83d\ude18\u65b0\u6d6a\u5fae\u535a\uff1aUnique_\u5c0f\u6a59\u5b50\u6709\u66f4\u591a\u7f8e\u7167\u548c\u5927\u5bb6\u5206\u4eab\u54e6\ud83d\ude18\u5168\u6c11k\u6b4c\uff1a*Unique-\u5c0f\u6a59\u5b50\ud83c\udf4a \u6709\u4e3b\u64ad\u6240\u5531\u6b4c\u66f2\u5466\ud83d\ude18","realname":"*Unique-\u5c0f\u6a59\u5b50\ud83c\udf4a","type":1},"exp":123494,"level":16,"medal":[]},"sender":{"avatar":"http:\/\/image.huajiao.com\/70e6ae3070c45093c2db6e1730f50dd2-100_100.jpg","nickname":"\u6a59\u5b50\u4e5f\u53eb\u5965\u745e\u6a58orange\ud83c\udf4a","uid":29198087,"verified":false,"verifiedinfo":{"credentials":"\u4e00\u5fc3\u5b88\u62a4*Unique-\u5c0f\u6a59\u5b50\u5230\u6c38\u8fdc\uff01","realname":"\u6a59\u5b50\u4e5f\u53eb\u5965\u745e\u6a58orange\ud83c\udf4a","type":0},"exp":1233982,"level":25,"medal":[{"kind":"tuhao","medal":"3"}]},"giftinfo":{"giftid":"1077","giftname":"\u751c\u751c\u5708","amount":"1","icon":"http:\/\/static.huajiao.com\/huajiao\/gift\/tiantianquan256.png","pic":"http:\/\/static.huajiao.com\/huajiao\/gift\/tiantianquan256.png","content":"","relativeInfo":{"repeatId":"@1473681582.3555611815447293","repeatNum":$$i$$,"property":{"repeatGift":"1","points":1,"gif":"http:\/\/static.huajiao.com\/huajiao\/gifteffect\/20005_31.zip","pic":"https:\/\/static.huajiao.com\/huajiao\/gift\/tiantianquan256.png","desctop":"\u751c\u751c\u5708","desc":"1\u8c46","property_android":{"repeatGift":"1","desctop":"\u751c\u751c\u5708","desc":"1\u8c46","pic":"http:\/\/static.huajiao.com\/huajiao\/gift\/tiantianquan256.png","points":1,"gif":"http:\/\/static.huajiao.com\/huajiao\/gifteffect\/20005_31.zip"},"property_ios":{"repeatGift":"1","desctop":"\u751c\u751c\u5708","desc":"1\u8c46","pic":"http:\/\/static.huajiao.com\/huajiao\/gift\/tiantianquan256.png","points":1,"gif":"http:\/\/static.huajiao.com\/huajiao\/gifteffect\/20005_31.zip"}}}},"title":"","scheme":"huajiao:\/\/huajiao.com\/goto\/wallet?userid=54990053"},"traceid":"$$traceid$$"}`,
			//`{"roomid":$$roomid$$,"type":30,"text":"","time":1472620520,"expire":86400,"extends":{"contents":"","creatime":"2016-08-31 13:15:20","receiver_balance":7,"receiver_income":"0","receiver_income_b":"163","receiver_income_p":"6044","sender_balance":28,"receiver":{"avatar":"http:\/\/image.huajiao.com\/905a29437aa0710fc6dd8afee96f0d15-100_100.jpg","nickname":"\ud83d\udc59 Man \ud83d\udc59\ud83d\udc83\ud83c\udffb","uid":$$receiver$$,"verified":false,"verifiedinfo":{"credentials":"\u5f00\u5f00\u5fc3\u5fc3\u8fc7\u597d\u6bcf\u4e00\u5929\uff01\ud83d\udc44\ud83d\udc44\ud83d\udc44","realname":"\ud83d\udc59 Man \ud83d\udc59\ud83d\udc83\ud83c\udffb","type":0},"exp":15787,"level":9,"medal":[]},"sender":{"avatar":"http:\/\/image.huajiao.com\/167b2292573223e0fda7d0b92876e361-100_100.jpg","nickname":"\u6ce2\u52a8\u4e00\u4e0bmm","uid":35487498,"verified":false,"verifiedinfo":{"credentials":"","realname":"\u6ce2\u52a8\u4e00\u4e0bmm","type":0},"exp":6001,"level":7,"medal":[]},"giftinfo":{"giftid":"1091","giftname":"\u559c\u6b22\u4f60","amount":"5","icon":"http:\/\/static.huajiao.com\/huajiao\/gift\/woxihuanni2.png","pic":"http:\/\/static.huajiao.com\/huajiao\/gift\/woxihuanni2.png","content":"","relativeInfo":{"repeatId":"@35487498524105012979958814726205206574742","repeatNum":$$i$$,"property":{"repeatGift":1,"property_android":{"repeatGift":"1","desctop":"\u559c\u6b22\u4f60","desc":"5\u8c46","pic":"http:\/\/static.huajiao.com\/huajiao\/gift\/woxihuanni2.png","points":1,"gif":"http:\/\/static.huajiao.com\/huajiao\/gifteffect\/20006_30.zip"},"property_ios":{"repeatGift":"1","desctop":"\u559c\u6b22\u4f60","desc":"5\u8c46","pic":"http:\/\/static.huajiao.com\/huajiao\/gift\/woxihuanni2.png","points":1,"gif":"http:\/\/static.huajiao.com\/huajiao\/gifteffect\/20006_30.zip"}}}},"title":"","scheme":"huajiao:\/\/huajiao.com\/goto\/wallet?userid=52410501"},"traceid":"$$traceid$$","priority":101}`,
			//`{"roomid":$$roomid$$,"type":30,"text":"","time":1472620896,"expire":86400,"extends":{"contents":"","creatime":"2016-08-31 13:21:36","receiver_balance":19,"receiver_income":"0","receiver_income_b":"48","receiver_income_p":"22791","sender_balance":9,"receiver":{"avatar":"http:\/\/image.huajiao.com\/fda789b9672d08b3a59eca5480a295e7-100_100.jpg","nickname":"\u742a\u742ababy only\ud83d\udc83","uid":$$receiver$$,"verified":false,"verifiedinfo":{"credentials":"\u4e0d\u8981\u8ba9\u68a6\u60f3\u6bc1\u5728\u522b\u4eba\u7684\u5634\u91cc\uff0c\u56e0\u4e3a\u522b\u4eba\u4e0d\u4f1a\u4e3a\u4f60\u7684\u68a6\u60f3\u8d1f\u8d23\u3002\u76f8\u4fe1\u81ea\u5df1\uff0c\u52c7\u5f80\u76f4\u524d\uff01","realname":"\u742a\u742ababy only\ud83d\udc83","type":0},"exp":11217,"level":8,"medal":[]},"sender":{"avatar":"http:\/\/image.huajiao.com\/11567101378fc08988b38b8f0acb1f74-100_100.jpg","nickname":"\u6c5f\u5c71\u91cc","uid":23103771,"verified":false,"verifiedinfo":{"credentials":"","realname":"\u6c5f\u5c71\u91cc","type":0},"exp":13965,"level":9,"medal":[]},"giftinfo":{"giftid":"1431","giftname":"\u4f1f\u5927\u7684\u6c49\u5821","amount":"2","icon":"http:\/\/static.huajiao.com\/huajiao\/gift\/wddhb256.png","pic":"http:\/\/static.huajiao.com\/huajiao\/gift\/wddhb256.png","content":"","relativeInfo":{"repeatNum":$$i$$,"property":{"repeatGift":1,"property_ios":{"repeatGift":"1","desctop":"\u4f1f\u5927\u7684\u6c49\u5821","desc":"2\u8c46","pic":"http:\/\/static.huajiao.com\/huajiao\/gift\/wddhb256.png","points":1,"gif":"http:\/\/static.huajiao.com\/huajiao\/gifteffect\/20014_30.zip","tagImage":"http:\/\/static.huajiao.com\/huajiao\/gift\/xinjiaobao256.png"},"property_android":{"repeatGift":"1","desctop":"\u4f1f\u5927\u7684\u6c49\u5821","desc":"2\u8c46","pic":"http:\/\/static.huajiao.com\/huajiao\/gift\/wddhb256.png","points":1,"gif":"http:\/\/static.huajiao.com\/huajiao\/gifteffect\/20014_30.zip","tagImage":"http:\/\/static.huajiao.com\/huajiao\/gift\/xinjiaobao256.png"}},"repeatId":"@23103771634149182981587414726211369454915"}},"title":"","scheme":"huajiao:\/\/huajiao.com\/goto\/wallet?userid=63414918"},"traceid":"$$traceid$$"}`,
		},
		"join": []string{
			`{"roomid":"$$roomid$$","type":10,"text":"\u52a0\u5165\u76f4\u64ad\u4e86","time":1468247083,"expire":86400,"extends":{"liveid":"21086174","userid":"29108183","nickname":"\u5fc3\u7a9d\u6709\u5fd7","avatar":"http:\/\/image.huajiao.com\/b693b362724ee10cda698c359d966b5d-100_100.jpg","verified":false,"verifiedinfo":{"credentials":"","type":0,"realname":"\u5fc3\u7a9d\u6709\u5fd7","status":0,"error":"","official":false},"level":9,"exp":15164,"rank":15164,"watches":7018,"medal":[]},"traceid":"$$traceid$$"}`,
			`{"roomid":"$$roomid$$","type":10,"text":"\u52a0\u5165\u76f4\u64ad\u4e86","time":1468247085,"expire":86400,"extends":{"liveid":"21086174","userid":"51502253","nickname":"zjx1290c","avatar":"http:\/\/image.huajiao.com\/ff95f010a79141faf52cade1c239fe4c-100_100.jpg","verified":false,"verifiedinfo":{"credentials":"","type":0,"realname":"zjx1290c","status":0,"error":"","official":false},"level":3,"exp":401,"rank":401,"watches":7029,"medal":[]},"traceid":"$$traceid$$"}`,
			`{"roomid":"$$roomid$$","type":10,"text":"\u52a0\u5165\u76f4\u64ad\u4e86","time":1468247085,"expire":86400,"extends":{"liveid":"21086174","userid":"28340566","nickname":"\u5b59\u91d1\u5bcc","avatar":"http:\/\/image.huajiao.com\/0403c74860576527ef6a6fe53cf3b95b-100_100.jpg","verified":false,"verifiedinfo":{"credentials":"","type":0,"realname":"\u5b59\u91d1\u5bcc","status":0,"error":"","official":false},"level":5,"exp":1153,"rank":1153,"watches":7029,"medal":[]},"traceid":"$$traceid$$"}`,
		},
		"quit": []string{
			`{"roomid":$$roomid$$,"type":16,"text":"quit","time":1468246759,"expire":86400,"extends":{"liveid":21074619,"userid":27142549},"traceid":"$$traceid$$"}`,
		},
		"msg": []string{
			`{"roomid":"$$roomid$$","type":9,"text":"$$content$$","time":1468247085,"expire":86400,"extends":{"liveid":"21086174","userid":"29018148","nickname":"\u9738\u738b\u52cb\u3002","avatar":"http:\/\/image.huajiao.com\/9fb6ccaf155177e7c7047ba812215729-100_100.jpg","verified":false,"verifiedinfo":{"credentials":"","type":0,"realname":"$$realname$$","status":0,"error":"","official":false},"gift":0,"exp":270,"level":$$level$$,"medal":[]},"traceid":"$$traceid$$"}`,
		},
		"fly": []string{
			`{"roomid":"$$roomid$$","type":9,"text":"$$content$$","time":1468247085,"expire":86400,"extends":{"liveid":"21086174","userid":"29018148","nickname":"\u9738\u738b\u52cb\u3002","avatar":"$$avatar$$","verified":false,"verifiedinfo":{"credentials":"","type":0,"realname":"$$realname$$","status":0,"error":"","official":false},"gift":1,"exp":270,"level":$$level$$,"medal":[]},"traceid":"$$traceid$$","proirity":101}`,
		},
		"faceu": []string{
			`{"roomid":$$roomid$$,"type":30,"text":"","time":1472615028,"expire":86400,"extends":{"contents":"","creatime":"2016-08-31 11:43:48","receiver_balance":221,"receiver_income":"0","receiver_income_b":"56351","receiver_income_p":"0","sender_balance":18113,"receiver":{"avatar":"http:\/\/image.huajiao.com\/ef234547401ff12de7477f76db7b42d0-100_100.jpg","nickname":"\u4e0a\u5b98\u5e0c\u6587","uid":$$receiver$$,"verified":true,"verifiedinfo":{"credentials":"\u5357\u4eac\u51ef\u5929\u6587\u5316\u4f20\u5a92\u6709\u9650\u516c\u53f8\u5e02\u573a\u90e8\u603b\u76d1","realname":"$$realname$$","type":1},"exp":1098258,"level":$$level$$,"medal":[{"kind":"tuhao","medal":"2"}]},"sender":{"avatar":"http:\/\/image.huajiao.com\/02095ac977419b6e604cf00fb1bc3356-100_100.jpg","nickname":"\u99a8\u2618","uid":54309868,"verified":false,"verifiedinfo":{"credentials":"\u966a\u4f34\u662f\u6700\u957f\u60c5\u7684\u544a\u767d\uff0c\u99a8\u99a8\u76f8\u5e0c\ud83c\udf40","realname":"\u99a8\u2618","type":0},"exp":11809453,"level":37,"medal":[{"kind":"tuhao","medal":"6"}]},"giftinfo":{"giftid":"1408","giftname":"\u82b1\u6912\u5c0f\u91d1\u4eba","amount":"1333","icon":"http:\/\/static.huajiao.com\/huajiao\/gift\/huajiaoxiaojinren256.png","pic":"http:\/\/static.huajiao.com\/huajiao\/gift\/huajiaoxiaojinren800.png","content":"","relativeInfo":{"repeatId":"@1472614999.1994562134259106","repeatNum":$$i$$,"property":{"faceuGift":1,"points":1,"tagImage":"http:\/\/static.huajiao.com\/huajiao\/gift\/live_tag_new_event.png","pic":"http:\/\/static.huajiao.com\/huajiao\/faceugift\/90055_2.zip","gif":"http:\/\/static.huajiao.com\/huajiao\/gifteffect\/90055_30.zip","faceuRepeatNum":"2","desctop":"\u82b1\u6912\u5c0f\u91d1\u4eba","desc":"1333\u8c46","giftIdentity":"90055_2","property_android":{"faceuGift":1,"desctop":"1333\u8c46","desc":"\u82b1\u6912\u5c0f\u91d1\u4eba","tagImage":"http:\/\/static.huajiao.com\/huajiao\/gift\/live_tag_new_event.png","pic":"http:\/\/static.huajiao.com\/huajiao\/faceugift\/90055_20.zip","points":1,"gif":"http:\/\/static.huajiao.com\/huajiao\/gifteffect\/90055_30.zip"},"property_ios":{"faceuGift":1,"faceuRepeatNum":"2","desctop":"\u82b1\u6912\u5c0f\u91d1\u4eba","desc":"1333\u8c46","tagImage":"http:\/\/static.huajiao.com\/huajiao\/gift\/live_tag_new_event.png","pic":"http:\/\/static.huajiao.com\/huajiao\/faceugift\/90055_2.zip","giftIdentity":"90055_2","points":1,"gif":"http:\/\/static.huajiao.com\/huajiao\/gifteffect\/90055_30.zip"}}}},"title":"","scheme":"huajiao:\/\/huajiao.com\/goto\/wallet?userid=25698400"},"traceid":"$$traceid$$","priority":299}`,
			`{"roomid":$$roomid$$,"type":30,"text":"","time":1472614504,"expire":86400,"extends":{"contents":"","creatime":"2016-08-31 11:35:04","receiver_balance":2,"receiver_income":"0","receiver_income_b":"2451","receiver_income_p":"0","sender_balance":2547,"receiver":{"avatar":"http:\/\/image.huajiao.com\/1a3bb43887e1cd64226f2fdf3f0fd56e-100_100.jpg","nickname":"\u4ed9\u7237\u2763","uid":$$receiver$$,"verified":true,"verifiedinfo":{"credentials":"\u201c\u6211\u4f1a\u52aa\u529b\u6210\u4e3a\u4f60\u672a\u6765\u89c1\u5230\u4f1a\u540e\u6094\u6ca1\u6709\u73cd\u60dc\u7684\u4eba\u201d","realname":"$$realname$$","type":1},"exp":177656,"level":$$level$$,"medal":[]},"sender":{"avatar":"http:\/\/image.huajiao.com\/018e590c7e58923b696103fd194dd6e6-100_100.jpg","nickname":"\u542c\u6d77\u7684\u6b4c\uff5e\u4ed9\u65c5","uid":54305818,"verified":false,"verifiedinfo":{"credentials":"\u4e0d\u7ba1\u6f02\u6cca\u5230\u54ea\u91cc\uff0c\u8eab\u8fb9\u603b\u662f\u4f60\u7684\u7a7a\u4f4d\ud83d\ude47\ud83c\udf02","realname":"\u542c\u6d77\u7684\u6b4c\uff5e\u4ed9\u65c5","type":0},"exp":1830445,"level":26,"medal":[{"kind":"tuhao","medal":"3"}]},"giftinfo":{"giftid":"1407","giftname":"\u6d2a\u8352\u4e4b\u529b","amount":"126","icon":"http:\/\/static.huajiao.com\/huajiao\/gift\/hhzl1256.png","pic":"http:\/\/static.huajiao.com\/huajiao\/gift\/hhzl1800.png","content":"","relativeInfo":{"repeatId":"","repeatNum":$$i$$,"property":{"faceuGift":1,"points":1,"tagImage":"http:\/\/static.huajiao.com\/huajiao\/gift\/live_tag_upgrade.png","pic":"http:\/\/static.huajiao.com\/huajiao\/faceugift\/90052_1.zip","gif":"http:\/\/static.huajiao.com\/huajiao\/gifteffect\/90052_31.zip","faceuRepeatNum":"2","rand_pics":["http:\/\/static.huajiao.com\/huajiao\/faceugift\/90052_1.zip","http:\/\/static.huajiao.com\/huajiao\/faceugift\/90052_3.zip","http:\/\/static.huajiao.com\/huajiao\/faceugift\/90052_4.zip"],"desctop":"\u6d2a\u8352\u4e4b\u529b","desc":"126\u8c46","giftIdentity":"90052_1","property_android":{"faceuGift":1,"desctop":"\u6d2a\u8352\u4e4b\u529b","desc":"126\u8c46","pic":"http:\/\/static.huajiao.com\/huajiao\/faceugift\/90052_20.zip","points":1,"gif":"http:\/\/static.huajiao.com\/huajiao\/gifteffect\/90052_31.zip","tagImage":"http:\/\/static.huajiao.com\/huajiao\/gift\/live_tag_upgrade.png","rand_pic":"http:\/\/static.huajiao.com\/huajiao\/faceugift\/90052_40.zip"},"property_ios":{"faceuGift":1,"faceuRepeatNum":"2","desctop":"\u6d2a\u8352\u4e4b\u529b","desc":"126\u8c46","pic":"http:\/\/static.huajiao.com\/huajiao\/faceugift\/90052_1.zip","giftIdentity":"90052_1","points":1,"gif":"http:\/\/static.huajiao.com\/huajiao\/gifteffect\/90052_31.zip","tagImage":"http:\/\/static.huajiao.com\/huajiao\/gift\/live_tag_upgrade.png","rand_pic":"http:\/\/static.huajiao.com\/huajiao\/faceugift\/90052_4.zip"}}}},"title":"","scheme":"huajiao:\/\/huajiao.com\/goto\/wallet?userid=54203416"},"traceid":"$$traceid$$"}`,
		},
		"praise": []string{
			`{"roomid":$$roomid$$,"type":8,"text":"","time":1473681599,"expire":86400,"extends":{"liveid":"$$roomid$$","num":100,"userid":"69247689","total":100,"nickname":"\u82b1\u6912\u7528\u623709101146","verified":false,"verifiedinfo":{"credentials":"","type":0,"realname":"\u82b1\u6912\u7528\u623709101146","status":0,"error":"","official":false},"verify_student":{},"exp":0,"level":1},"traceid":"$$traceid$$"}`,
		},
		// 向指定房间内发世界礼物，不影响正常业务
		"world": []string{
			`{"roomid":44258983,"type":68,"text":"","time":1477520963,"expire":86400,"extends":{"contents":"","limit_amount":-1,"largev":0,"creatime":"2016-10-27 06:29:23","receiver_balance":16,"receiver_income":"9628","receiver_income_b":"432","receiver_income_p":"194393","sender_balance":9205,"receiver":{"avatar":"http:\/\/image.huajiao.com\/7c8a58ae7d6fc3bf685a93ac1ed6c119-100_100.jpg","nickname":"\u7f8e\u6a59Abby\ud83e\udd84","uid":21471618,"verified":true,"verifiedinfo":{"credentials":"\u4e0a\u6d77\u96e8\u5a77\u751f\u7269\u79d1\u6280\u6709\u9650\u516c\u53f8\u7ecf\u7406","realname":"\u7f8e\u6a59Abby\ud83e\udd84","type":1},"exp":28521,"level":11,"medal":[]},"sender":{"avatar":"http:\/\/image.huajiao.com\/9d6c4dd2da7fb985ef02b727c831555c-100_100.jpg","nickname":"\u5341\u6708\u65e0\u90aa","uid":58753594,"verified":false,"verifiedinfo":{"credentials":"","realname":"\u5341\u6708\u65e0\u90aa","type":0},"exp":2480869,"level":28,"medal":[{"kind":"tuhao","medal":"3"}]},"giftinfo":{"giftid":"1069","giftname":"\u68a6\u5e7b\u57ce\u5821","amount":"52000","icon":"https:\/\/static.huajiao.com\/huajiao\/gift\/chengbao1256.png","pic":"https:\/\/static.huajiao.com\/huajiao\/gifteffect\/10053_6.zip","content":"","relativeInfo":{"repeatId":"","repeatNum":1,"property":{"effectPngGift":1,"tagImage":"https:\/\/static.huajiao.com\/huajiao\/gift\/live_tag_world4.png","desctop":"\u68a6\u5e7b\u57ce\u5821","points":1,"gif":"https:\/\/static.huajiao.com\/huajiao\/gifteffect\/10053_32.zip","desc":"52000\u8c46","pic":"https:\/\/static.huajiao.com\/huajiao\/gifteffect\/10053_6.zip","worldDesc":"\u5f00\u542f\u68a6\u5e7b\u4e4b\u65c5","isWorldGift":"1","worldIcon":"http:\/\/static.huajiao.com\/huajiao\/gift\/chengbao1256.png","giftIdentity":"10053_6","screenshottime":"13.5","property_android":{"effectGift":1,"pausetime":"15000","desctop":"52000\u8c46","desc":"\u68a6\u5e7b\u57ce\u5821","pic":"https:\/\/static.huajiao.com\/huajiao\/gift\/chengbao800.png","points":1,"isWorldGift":"1","worldIcon":"http:\/\/static.huajiao.com\/huajiao\/gift\/chengbao1256.png","worldDesc":"\u5f00\u542f\u68a6\u5e7b\u4e4b\u65c5","gif":"https:\/\/static.huajiao.com\/huajiao\/gifteffect\/10053_32.zip","tagImage":"http:\/\/static.huajiao.com\/huajiao\/gift\/live_tag_world4.png"},"property_ios":{"effectPngGift":1,"screenshottime":"13.5","desctop":"\u68a6\u5e7b\u57ce\u5821","desc":"52000\u8c46","pic":"https:\/\/static.huajiao.com\/huajiao\/gifteffect\/10053_6.zip","giftIdentity":"10053_6","points":1,"isWorldGift":"1","worldIcon":"http:\/\/static.huajiao.com\/huajiao\/gift\/chengbao1256.png","worldDesc":"\u5f00\u542f\u68a6\u5e7b\u4e4b\u65c5","gif":"https:\/\/static.huajiao.com\/huajiao\/gifteffect\/10053_32.zip","tagImage":"https:\/\/static.huajiao.com\/huajiao\/gift\/live_tag_world4.png"}}}},"title":"","scheme":"huajiao:\/\/huajiao.com\/goto\/wallet?userid=21471618","feedid":"44258983","world_message":{"text":"#sender#\u7ed9#receiver#\u53d1\u9001\u4e86\u4e00\u4e2a#gift_name#\uff0c#gift_desc#\uff01\u5feb\u6765\u82b1\u6912\u53f7#author#\u56f4\u89c2\u5427\uff01","duration":"5"},"ts_id":"645099346"},"traceid":"$$traceid$$"}`,
		},
	}
	fileGift, err := os.Open("giftend")
	if err != nil {
		fmt.Println("os open fail")
		return
	}
	bfRd := bufio.NewReader(fileGift)
	for {
		line, err := bfRd.ReadBytes('\n')
		msgs["gift"] = append(msgs["gift"], string(line))
		if err != nil {
			if err == io.EOF {
				break
			}
		}
	}
	fileGift.Close()

	fileText, err := os.Open("yuliao")
	if err != nil {
		panic(err)
	}
	bfRd = bufio.NewReader(fileText)
	for {
		line, err := bfRd.ReadBytes('\n')
		linestr := strings.TrimSpace(string(line))
		contentTpl = append(contentTpl, linestr)
		if err != nil {
			if err == io.EOF {
				break
			}
		}
	}
	fileText.Close()

	fileName, err := os.Open("realname")
	if err != nil {
		panic(err)
	}
	bfRd = bufio.NewReader(fileName)
	for {
		line, err := bfRd.ReadBytes('\n')
		linestr := strings.TrimSpace(string(line))
		realName = append(realName, linestr)
		if err != nil {
			if err == io.EOF {
				break
			}
		}
	}
	fileName.Close()

	fileAvatar, err := os.Open("avatar")
	if err != nil {
		panic(err)
	}
	bfRd = bufio.NewReader(fileAvatar)
	for {
		line, err := bfRd.ReadBytes('\n')
		linestr := strings.TrimSpace(string(line))
		avatar = append(avatar, linestr)
		if err != nil {
			if err == io.EOF {
				break
			}
		}
	}
	fileAvatar.Close()

	filebig, err := os.Open("biggift")
	if err != nil {
		panic(err)
	}
	bfRd = bufio.NewReader(filebig)
	for {
		line, err := bfRd.ReadBytes('\n')
		msgs["biggift"] = append(msgs["biggift"], string(line))
		if err != nil {
			if err == io.EOF {
				break
			}
		}
	}
	filebig.Close()
}

/**
 * exec function f n times per second
 * m stand for max times
 * import(
 *     "time"
 * )
 */
func ExecNTimes(n int, m int64, f func()) {
	sleepTime := time.Second / time.Duration(n)
	next := time.Now()
	var i int64
	for m == 0 || i < m {
		i++
		next = next.Add(sleepTime)
		go f()
		left := next.Sub(time.Now())
		if left > 0 {
			time.Sleep(left)
		}
	}
}

var Succ, Fail uint64

func showInfo() {
	lock.RLock()
	if len(tasks) > 0 {
		fmt.Println("Succ:", atomic.LoadUint64(&Succ), ", Fail", atomic.LoadUint64(&Fail))
	}
	lock.RUnlock()
}

func send(addr, roomid, content, priority, typ string) bool {
	url := fmt.Sprintf("%s/chatroom/send", addr)
	values := make(map[string][]string, 6)
	values["roomid"] = []string{roomid}
	values["sender"] = []string{"admin"}
	values["traceid"] = []string{"1234567890"}
	values["content"] = []string{content}
	values["appid"] = []string{"2080"}
	values["priority"] = []string{priority}
	values["type"] = []string{typ}

	if _, err := httpClient.PostForm(url, values); err != nil {
		fmt.Println("SendChatRoomMsg err is ", err)
		return false
	}
	return true
}

var contentTpl2 []string = []string{
	"666",
}

var i int

func sendType(typ, center, roomid, receiver string) {
	arr := strings.Split(typ, "-")
	var t, proi string
	if len(arr) >= 2 {
		t = arr[1]
	}
	if len(arr) >= 3 {
		proi = arr[2]
	}
	i += 1
	msg := msgs[arr[0]][rand.Intn(len(msgs[arr[0]]))]
	msg = strings.Replace(msg, "$$roomid$$", roomid, -1)
	msg = strings.Replace(msg, "$$receiver$$", receiver, -1)
	msg = strings.Replace(msg, "$$traceid$$", fmt.Sprint(time.Now().UnixNano(), strconv.Itoa(rand.Int())), -1)
	msg = strings.Replace(msg, "$$content$$", contentTpl[rand.Int()%len(contentTpl)], -1)
	msg = strings.Replace(msg, "$$realname$$", realName[rand.Int()%len(realName)], -1)
	msg = strings.Replace(msg, "$$avatar$$", avatar[rand.Int()%len(avatar)], -1)
	msg = strings.Replace(msg, "$$level$$", strconv.Itoa(rand.Intn(35)), -1)
	//msg = strings.Replace(msg, "$$content$$", contentTpl[rand.Int()%len(contentTpl)], -1)
	msg = strings.Replace(msg, "$$i$$", strconv.Itoa(i), -1)
	fmt.Println(center, roomid, proi, t, msg)
	if send(center, roomid, msg, proi, t) {
		atomic.AddUint64(&Succ, 1)
	} else {
		atomic.AddUint64(&Fail, 1)
	}
}

func genSendType(typ, center, roomid, receiver string) func() {
	return func() {
		sendType(typ, center, roomid, receiver)
	}
}

var center string
var roomid string

func addHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	freq := map[string]int{}
	for k, v := range req.Form {
		i, _ := strconv.Atoi(v[0])
		if i > 0 && i < 5000 {
			freq[k] = i
		}

	}
	/*
		for k, _ := range msgs {
			i, _ := strconv.Atoi(req.Form.Get(k))
			if i > 0 && i < 5000 {
				freq[k] = i
			}
		}
	*/

	t := addTask(freq, req.Form.Get("center"), req.Form.Get("rid"), req.Form.Get("receiver"), 60)
	s, err := json.Marshal(t)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Write(s)
	}
}
func tasksHandler(w http.ResponseWriter, req *http.Request) {
	s, err := json.Marshal(getTasks())
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Write(s)
	}

}

func main() {
	fmt.Println("giftlen", len(msgs["gift"]))
	fmt.Println("yuliao len", len(contentTpl))

	go ExecNTimes(1, 0, showInfo)
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/tasks", tasksHandler)
	log.Fatal(http.ListenAndServe("0.0.0.0:7878", nil))
	select {}
}
