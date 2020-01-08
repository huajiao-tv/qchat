package main

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/mgo.v2"
)

const (
	MongoServerPrefix      = "mongodb://"
	MongoServerPlaceholder = "%mongo_server%"
)

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

/*
 * format mgo connect string with or without sensitive information
 * @param server is mongodb server address
 * @param cs is connect string template which might has mongo server placeholder
 * @param withPwd indicates whether returned connect string includes sensitive information
 * @return connect string with or without sensitive information
 */
func formatConnectString(server, cs string, withPwd bool) string {
	// replace mongo server placeholder with mongo server hostname/ip
	mongo := cs
	if len(server) > 0 {
		mongo = strings.Replace(cs, MongoServerPlaceholder, server, -1)
	}

	// check whether mongos address starts with "mongodb://"
	// TODO: disable for aliyun
	//mongo = mongo + "?authMechanism=MONGODB-CR"
	if !strings.HasPrefix(mongo, MongoServerPrefix) {
		mongo = MongoServerPrefix + mongo
	}

	if withPwd {
		return mongo
	}

	// remove sensitive information
	return mongo[strings.Index(mongo, "@")+1:]
}

/*
 * dial to mongodb server and returned mgo session if successfully
 * @param server is mongodb server address
 * @param cs is connect string template which might has mongo server placeholder
 * @return connect string with or without sensitive information
 */
func getMgoSession(server, cs string) (*mgo.Session, error) {
	// get mongo connect string with password if there is
	mongo := formatConnectString(server, cs, true)

	timeout := netConf().MongoTimeout
	if timeout < 1 {
		timeout = 1
	}
	session, err := mgo.DialWithTimeout(mongo, time.Duration(timeout)*time.Second)
	if err != nil || session == nil {
		// because mongo might include sensitive information, so need to remove them if there is
		return nil, errors.New(fmt.Sprint("dialed ", mongo[strings.Index(mongo, "@")+1:], " failed, error: ", err))
	}

	return session, nil
}

/*
 * initialize mgo.Session object arrays for specified mongo server group
 * @param mongoServers is mongodb server list
 * @param mongoPorts is mongodb connect string template, it can use server placeholder
 *      so that we can use server in mongoServers to build final connect string
 * @return (map[string][]*mgoSession) which is initialized map stored nil mgo session objects for each server
 */
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
		Logger.Debug("", "", "", "initMgoSessions",
			fmt.Sprint("initialized session objects to server ", server))
	}

	ports := removeSensitiveMongoInfo(mongoPorts)

	if len(sessions) == 0 {
		Logger.Error("", "", "", "initMgoSessions", "invalid parameters",
			fmt.Sprint("servers: ", mongoServers, ", ports: ", ports))
	} else {
		Logger.Trace("", "", "", "initMgoSessions", "initialized mgo session arrays",
			fmt.Sprint("servers: ", mongoServers, ", ports: ", ports))
	}

	return sessions
}

/*
 * initialize mongo sessions for service, we need to lock data to ensure that thread safe
 */
func initializeMongoSessions() {
	servers := netConf().MessageMongo
	ports := netConf().MessageMongoPorts
	alternativeServers := netConf().AlternativeMessageMongo
	alternativePorts := netConf().AlternativeMessageMongoPorts
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

/*
 * initialize mongo sessions for service
 * @return nil if no error occurs, otherwise an error interface is returned
 *      Note: so far we always return nil even though there is any initialization error
 */
func initMongoSession() error {
	// init mongo sessions first
	initializeMongoSessions()

	// start session checking thread
	go checkSessions()

	return nil
}

/*
 * remove sensitive mongo information from original slice
 * @param s is original slice might with sensitive mongo information
 * @return new slice include mongo information which have been removed sensitive information
 */
func removeSensitiveMongoInfo(s []string) []string {
	ret := make([]string, 0, len(s))
	for _, val := range s {
		val = val[strings.Index(val, "@")+1:] // 去掉用户名密码等敏感信息
		ret = append(ret, val)
	}

	return ret
}

/*
 * check whether we need to update sessions to mongo servers for config is changed
 */
func checkSessions() {
	if netConf().SessionCheckInterval > 0 {
		sessionCheckInterval = netConf().SessionCheckInterval
	}

	ticker := time.NewTicker(time.Second * time.Duration(sessionCheckInterval))
	defer func() {
		ticker.Stop()
	}()

	for {
		<-ticker.C // wait a moment

		if !(isEqualStringSlice(msgMongoServers, netConf().MessageMongo) &&
			isEqualStringSlice(msgMongoPorts, netConf().MessageMongoPorts) &&
			isEqualStringSlice(alternativeMsgMongoServers, netConf().AlternativeMessageMongo) &&
			(isEqualStringSlice(alternativeMsgMongoPorts, netConf().AlternativeMessageMongoPorts) ||
				isEqualStringSlice(alternativeMsgMongoPorts, netConf().MessageMongoPorts))) {
			Logger.Debug("", "", "", "checkSessions", "sessions config is changed",
				fmt.Sprint("msgMongoServers:", msgMongoServers,
					" msgMongoPorts:", removeSensitiveMongoInfo(msgMongoPorts),
					" alternativeMsgMongoServers:", alternativeMsgMongoServers,
					" alternativemsgMongoPorts:", removeSensitiveMongoInfo(alternativeMsgMongoPorts),
					" netConf().MessageMongo:", netConf().MessageMongo,
					" netConf().MessageMongoPorts:", removeSensitiveMongoInfo(netConf().MessageMongoPorts),
					" netConf().AlternativeMessageMongo:", netConf().AlternativeMessageMongo,
					" netConf().AlternativeMessageMongoPorts:",
					removeSensitiveMongoInfo(netConf().AlternativeMessageMongoPorts)))

			close(exit)
			sessionWG.Wait()

			Logger.Debug("", "", "", "checkSessions", "reinitializing sessions for mongodb config is changed")

			// reinitialize mongo sessions
			initializeMongoSessions()
		}
	}
}

/*
 * we have hashed users to mongo stores, this allow us to find what message store has owner's data
 * @param owner is an unique string to identify a owner
 * @param appid is application ID
 * @param traceSn is trace SN string
 * @return ([]*mgo.Session, nil) if no error occurs, otherwise (nil, error) is returned
 *
 * @Note: we need to close the session after we use it
 */
func GetMessageMongoStore(owner string, appid uint16, traceSn string) ([]*mgo.Session, error) {
	sessionDataRWLock.RLock()
	preferredServers, preferredSessions := msgMongoServers, msgMgoSessions
	alternativeServers, alternativeSessions := alternativeMsgMongoServers, alternativeMsgMgoSessions
	sessionDataRWLock.RUnlock()

	if len(preferredSessions) == 0 && len(alternativeSessions) == 0 {
		return nil, errors.New("no available mongo connection")
	}

	ret := make([]*mgo.Session, 0, 2)
	session, info := getMongoSession(owner, preferredServers, preferredSessions)
	if session == nil {
		Logger.Error(owner, appid, traceSn, "GetMessageMongoStore", "Get mongo preferred storage failed", info)
	} else {
		ret = append(ret, session)
	}

	session, alternative := getMongoSession(owner, alternativeServers, alternativeSessions)
	info = info + "; " + alternative
	if session == nil {
		Logger.Error(owner, appid, traceSn, "GetMessageMongoStore", "Get mongo alternative storage failed", alternative)
	} else {
		ret = append(ret, session)
	}

	if len(ret) > 0 {
		/*Logger.Debug(owner, appid, traceSn, "GetMessageMongoStore",
		"Get mongo storage", info)*/
		return ret, nil
	}

	return nil, errors.New(info)
}

/*
 * close mgo.Session objects
 * @param sessions is mgo.Session objects which will be closed
 */
func CloseMgoSessions(sessions []*mgo.Session) {
	for idx := 0; idx < len(sessions); idx++ {
		//Logger.Debug("", "", "", "CloseMgoSessions", fmt.Sprint("close session to servers ", sessions[idx].LiveServers()))
		sessions[idx].Close()
	}
}

/*
 * get session from specified session groups
 * @param owner is an unique string to identify a owner
 * @param servers is specified mongo servers
 * @param sessionMap is sessions group
 * @return (*mgo.Session, information) if no error occurs, otherwise (nil, information) is returned
 *      usually caller uses information to log
 *
 * @Note: we need to close the session after we use it
 */
func getMongoSession(owner string, servers []string, sessionMap map[string][]*mgoSession) (*mgo.Session, string) {
	if len(sessionMap) == 0 {
		return nil, "no available mongo connection"
	}
	timeout := netConf().MongoTimeout
	if timeout < 1 {
		timeout = 1
	}

	result := ""
	if hash, err := Md5Uint64Hash(owner); err != nil {
		return nil, err.Error()
	} else {
		server := servers[rand.Intn(len(servers))]
		if sessions, ok := sessionMap[server]; ok {
			// get hash index
			idx := hash % uint64(len(sessions))
			// get mgo.Session copy thread safe
			session := sessions[idx].GetCopy()

			result = fmt.Sprint("server:", server, " port index:", idx)
			if session != nil {
				// use net config for each session so that new session can use latest config
				session.SetSyncTimeout(time.Duration(timeout) * time.Second)
				session.SetSocketTimeout(time.Duration(timeout) * time.Second)
				return session, result
			} else {
				result = fmt.Sprint("the session to server:", server, " and port index: ", idx, " is nil")
			}
		} else {
			result = fmt.Sprint("servers:", servers, " selected server:", server, " sessions:", sessions)
		}

		// try to find an available session
		for server, sessions := range sessionMap {
			idx := hash % uint64(len(sessions))
			// get mgo.Session copy thread safe
			session := sessions[idx].GetCopy()

			if session != nil {
				// use net config for each session so that new session can use latest config
				session.SetSyncTimeout(time.Duration(timeout) * time.Second)
				session.SetSocketTimeout(time.Duration(timeout) * time.Second)
				result = result + fmt.Sprint("; new server:", server, " port index:", idx)
				return session, result
			}
		}
	}

	return nil, result
}

/*
 * start session state monitor threads
 * @param mgoSessions is initialized sessions map
 * @param mongoPorts is mongo ports array
 */
func monitorSessionsState(mgoSessions map[string][]*mgoSession, mongoPorts []string,
	wg *sync.WaitGroup, quit chan bool) {
	for server, sessions := range mgoSessions {
		for idx, _ := range sessions {
			go monitorSessionState(mgoSessions, mongoPorts, server, idx, wg, quit)
		}
	}
}

/*
 * session state monitor thread function. This monitors specified session object (only one,
 *  so that will not block others), if session object has been created, this will check session
 *  health by query; otherwise will try to create new session object to specified mongo server
 * @param mgoSessions is initialized sessions map
 * @param mongoPorts is mongo ports array
 * @param server is the specified session should connect to
 * @param idx is index of mongo port the specified session should connect to
 */
func monitorSessionState(mgoSessions map[string][]*mgoSession, mongoPorts []string,
	server string, idx int, wg *sync.WaitGroup, quit chan bool) {
	// get mongo service server and port information for log
	cs := formatConnectString(server, mongoPorts[idx], false)

	wg.Add(1)
	defer wg.Done()

	ticker := time.NewTicker(time.Second * time.Duration(sessionCheckInterval))
	defer func() {
		ticker.Stop()
	}()

	for {
		if sessions, ok := mgoSessions[server]; ok {
			if idx < len(sessions) && idx < len(mongoPorts) {
				session := sessions[idx].GetCopy()
				// check whether session state is OK
				if session != nil {
					lives := session.LiveServers()
					session.Close() // will not use this session
					if len(lives) == 0 {
						session = nil
						Logger.Error("", "", "", "monitorSessionState", fmt.Sprint("the #", idx,
							" session to server [", cs, "] is in BAD state"))
					} else {
						Logger.Debug("", "", "", "monitorSessionState", fmt.Sprint("the #", idx,
							" session to server [", cs, "] is in OK state"))
					}
				}

				// dial to mongo server if need
				mongoClean := atomic.LoadInt32(&MongoClean)
				if session == nil || mongoClean == 1 {
					atomic.StoreInt32(&MongoClean, 0)
					if session != nil {
						session.Close()
					}

					Logger.Debug("", "", "", "monitorSessionState", "dialing "+cs)
					session, err := getMgoSession(server, mongoPorts[idx])

					// this is thread safe
					sessions[idx].Set(session)

					if err != nil {
						Logger.Error("", "", "", "monitorSessionState", "dialed mongo server failed", err)
						Logger.Debug("", "", "", "monitorSessionState", "dialed mongo server failed", err)
					} else {
						Logger.Debug("", "", "", "monitorSessionState",
							fmt.Sprint("dialed ", cs, " successfully"))
					}
				}
			} else {
				Logger.Error("", "", "", "monitorSessionState", "wrong session index",
					fmt.Sprint("session index ", idx,
						" is larger than the count of sessions for server [", server, "]"))
			}
		} else {
			Logger.Error("", "", "", "monitorSessionState", "wrong parameter",
				fmt.Sprint("there is not sessions for server [", server, "]"))
		}

		// update tick timer if config is changed
		if netConf().SessionCheckInterval != sessionCheckInterval && netConf().SessionCheckInterval > 0 {
			ticker.Stop() // should stop old ticker explicitly
			Logger.Debug("", "", "", "monitorSessionState",
				fmt.Sprintf("check session interval is changed from %v to %v.",
					sessionCheckInterval, netConf().SessionCheckInterval))

			// update session state check interval to new value
			sessionCheckInterval = netConf().SessionCheckInterval

			// make new time ticker
			ticker = time.NewTicker(time.Second * time.Duration(sessionCheckInterval))
		}

		select {
		case <-ticker.C: // wait a moment
		case <-quit:
			Logger.Debug("", "", "", "monitorSessionState",
				fmt.Sprint("the monitor thread for ", cs, " received exit singal."))
			return
		}
	}
}
