package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"gopkg.in/mgo.v2"
)

const (
	MongoServerPrefix      = "mongodb://"
	MongoServerPlaceholder = "%mongo_server%"
)

func main() {
	initializeMongoSessions()

	time.Sleep(time.Hour)
}

func formatConnectString(server, cs string, withPwd bool) string {
	// replace mongo server placeholder with mongo server hostname/ip
	mongo := cs
	if len(server) > 0 {
		mongo = strings.Replace(cs, MongoServerPlaceholder, server, -1)
	}

	// check whether mongos address starts with "mongodb://"
	//mongo = mongo + "&authMechanism=MONGODB-CR"
	if !strings.HasPrefix(mongo, MongoServerPrefix) {
		mongo = MongoServerPrefix + mongo
	}

	if withPwd {
		return mongo
	}

	// remove sensitive information
	return mongo[strings.Index(mongo, "@")+1:]
}

func getMgoSession(server, cs string) (*mgo.Session, error) {
	// get mongo connect string with password if there is
	mongo := formatConnectString(server, cs, true)
	fmt.Println(1234567, mongo)

	timeout := 5
	session, err := mgo.DialWithTimeout(mongo, time.Duration(timeout)*time.Second)
	if err != nil || session == nil {
		// because mongo might include sensitive information, so need to remove them if there is
		return nil, errors.New(fmt.Sprint("dialed ", mongo[strings.Index(mongo, "@")+1:], " failed, error: ", err))
	}

	return session, nil
}

var (
	msgMongoServers            []string                 // 消息mongo的server
	msgMongoPorts              []string                 // 消息mongo的端口信息
	msgMgoSessions             map[string][]*mgoSession // 消息mongo会话
	alternativeMsgMongoServers []string                 // 消息mongo的备用server
	alternativeMsgMongoPorts   []string                 // 消息mongo的端口信息
	alternativeMsgMgoSessions  map[string][]*mgoSession // 消息mongo备用会话
	sessionDataRWLock          sync.RWMutex             // session数据读写锁
	sessionCheckInterval       = 1                      // session默认检查间隔,单位秒
	sessionWG                  *sync.WaitGroup          // session检查线程退出同步器
	exit                       chan bool                // session检查线程退出信号
)

func initializeMongoSessions() {
	servers := []string{"127.0.0.1:9090"}
	ports := []string{"root:123456@127.0.0.1:27017"}
	alternativeServers := []string{"127.0.0.1:9091"}
	alternativePorts := []string{}
	if len(alternativePorts) == 0 ||
		(len(alternativePorts) == 1 && len(alternativePorts[0]) == 0) {
		alternativePorts = ports
	}

	// init preference mongo sessions
	sessions := initMgoSessions(servers, ports)
	// init alternative mongo sessions
	alternativeSessions := initMgoSessions(alternativeServers, alternativePorts)

	// ensure the thread safe
	sessionDataRWLock.Lock()
	msgMongoServers, msgMongoPorts = servers, ports
	alternativeMsgMongoServers, alternativeMsgMongoPorts = alternativeServers, alternativePorts
	msgMgoSessions, alternativeMsgMgoSessions = sessions, alternativeSessions
	sessionDataRWLock.Unlock()

	sessionWG = new(sync.WaitGroup)
	exit = make(chan bool)
	// start session checking thread
	monitorSessionsState(sessions, ports, sessionWG, exit)
	monitorSessionsState(alternativeSessions, alternativeMsgMongoPorts, sessionWG, exit)
}

func initMgoSessions(mongoServers, mongoPorts []string) map[string][]*mgoSession {
	servers := len(mongoServers)
	mongos := len(mongoPorts)
	sessions := make(map[string][]*mgoSession, servers)

	// 我们只初始化session对象数组,实际mgo.Session对象留待monitorSessionState处理
	for _, server := range mongoServers {
		sessions[server] = make([]*mgoSession, mongos)
		for idx := 0; idx < mongos; idx++ {
			sessions[server][idx] = &mgoSession{}
		}
		fmt.Println("", "", "", "initMgoSessions",
			fmt.Sprint("initialized session objects to server ", server))
	}

	ports := removeSensitiveMongoInfo(mongoPorts)

	if len(sessions) == 0 {
		fmt.Println("", "", "", "initMgoSessions", "invalid parameters",
			fmt.Sprint("servers: ", mongoServers, ", ports: ", ports))
	} else {
		fmt.Println("", "", "", "initMgoSessions", "initialized mgo session arrays",
			fmt.Sprint("servers: ", mongoServers, ", ports: ", ports))
	}

	return sessions
}

func monitorSessionsState(mgoSessions map[string][]*mgoSession, mongoPorts []string,
	wg *sync.WaitGroup, quit chan bool) {
	for server, sessions := range mgoSessions {
		for idx, _ := range sessions {
			if sessions, ok := mgoSessions[server]; ok {
				session, err := getMgoSession(server, mongoPorts[idx])
				if err != nil {
					fmt.Println("get session err", err)
					continue
				}
				n, e := session.DatabaseNames()
				fmt.Println("xxxx", n, e)
				sessions[idx].Set(session)
			}

		}
	}
}

func removeSensitiveMongoInfo(s []string) []string {
	ret := make([]string, 0, len(s))
	for _, val := range s {
		val = val[strings.Index(val, "@")+1:] // 去掉用户名密码等敏感信息
		ret = append(ret, val)
	}

	return ret
}

/*
 * thread safe mgo.Session encapsulation
 */
type mgoSession struct {
	session *mgo.Session // session object
	lock    sync.RWMutex // session数据读写锁
}

/*
 * get mgo.Session copy thread safe if there is available session object
 */
func (ses *mgoSession) GetCopy() *mgo.Session {
	ses.lock.RLock()
	defer ses.lock.RUnlock()
	if ses.session == nil {
		return nil
	}
	return ses.session.Copy()
}

/*
 * set mgo.Session point thread safe
 */
func (ses *mgoSession) Set(newSession *mgo.Session) {
	ses.lock.Lock()
	defer ses.lock.Unlock()
	if ses.session != nil {
		ses.session.Close()
	}
	ses.session = newSession
}
