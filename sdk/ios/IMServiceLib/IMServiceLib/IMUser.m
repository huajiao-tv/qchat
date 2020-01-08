//
//  IMUser.m
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-7.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import <Foundation/NSException.h>
#import <Foundation/NSLock.h>
#import "IMServiceLib.h"
#import "IMUser.h"
#import "socket/GCDAsyncSocket.h"
#import "IMMessage.h"

#import "IMProtoHelper.h"
#import "IMFeatureWebInterface.h"
#import "IMFeatureUserInfo.h"
#import "IMFeaturePeerChat.h"
#import "IMFeatureGroupChat.h"
#import "IMFeatureChatRoom.h"
#import "IMFeatureCircleChat.h"
#import "IMFeatureRelation.h"
#import "IMFeaturePrivateChat.h"

#import "IMProtoRegistry.h"
#import "IMProtoWeb.h"
#import "IMProtoUserInfo.h"
#import "IMProtoMessage.h"
#import "IMProtoVoiceCallProxy.h"
#import "IMProtoServiceInner.h"
#import "IMProtoGroupChat.h"
#import "IMProtoChatRoom.h"
#import "IMProtoCircleChat.h"
#import "IMProtoRelation.h"
#import "IMProtoPrivateChat.h"

#import "IMTask.h"
#import "IMUtil.h"
#import "flog.h"
#import "GZIP.h"

#import "Base64.h"
#import "MD5Digest.h"
#import "RsaWorker.h"
#import "IM_Safe_cast.h"

#include <mach/semaphore.h>

#define _To_Dict(s)                 [IM_Safe_cast parseToDictionary:s]


#ifdef DEBUG
BOOL g_openIMLog = NO;
#else
BOOL g_openIMLog = NO;
#endif



/**
 * 定义socket tag 类型
 */
enum IMTagType
{
    /**
     * init login head
     */
    IM_Tag_InitLogin_Head = 0,

    /**
     * init login body
     */
    IM_Tag_InitLogin_Body = 1,

    /**
     * 登录请求的head
     */
    IM_Tag_Login_Head = 2,

    /**
     *登录请求的body
     */
    IM_Tag_Login_Body = 3,

    /**
     * init data head
     */
    IM_Tag_Data_Head = 4,

    /**
     * init data body
     */
    IM_Tag_Data_Body = 5
};
typedef enum IMTagType IMTagType;

enum IMThreadState
{
    /**
     * 发送线程因等待超时而离去,或者还没开始等待
     */
    IM_ThreadState_NoWaiting = 0,

    /**
     * 发送线程处于等待结果状态
     */
    IM_ThreadState_InWaiting = 1,

    /**
     * 接收线程收到了数据
     */
    IM_ThreadState_HasData = 2,

    /**
     * 发生了错误
     */
    IM_ThreadState_Error = 3
};
typedef enum IMThreadState IMThreadState;


/**
 * 定义私有函数
 */
@interface IMUser()<GCDAsyncSocketDelegate,NSStreamDelegate>
{
    //心跳就是4个字节的0
    NSData *_heartBeat;

    //上一次发心跳的时间
    NSDate *_lastHeartBeatTime;

    //上一次服务器下发心跳的时间
    NSDate *_lastHBAckTime;

    NSDate *_checkHBAckLock;

    //上一次检查_inboxLastGetInfoTimeDict的时间
    NSDate *_lastCheckGetInfoTime;

    //因为chatroom是多个，所以需要为每个chatroom保存一个时间戳
    NSMutableDictionary *_lastReadIDChatRoomDict;

    //因为chatroom是多个，所以需要为每个chatroom保存一个丢失的区间，每个区间NSRange
    NSMutableDictionary *_chatRoomMissedReadIDDict;

    //用户roomIDlostkey
    NSString *_roomIdMsgLostKey;

    NSInteger _lostIdCountOfChatRoom;



    //在配置系统里保存的用户配置key
    NSString *_userSettingKey;


    //用户丢失消息key
    NSString *_postMsgLostLogTimeKey;

    //计算服务器时间myServerTIme = （local2 - local1)/2 + serverTime;
    NSTimeInterval local1;
    NSTimeInterval local2;

}

/**
 * 是否停止服务
 */
@property (atomic) BOOL isStop;

/**
 * current user state
 */
@property (atomic, assign) IMUserState curState;

/**
 * IM_State_Login_Fail对应的errorID
 */
@property (atomic, assign) int loginFailReason;

/**
 * task queue
 */
@property (atomic, strong) NSOperationQueue *operationQueue;

/**
 * socket写和读线程间同步锁
 */
@property (atomic, strong) NSCondition *syncCondition;
/**
 * 因为NSCondition即使wait到信号也未必真正有数据，所以需要一个标志来做检查
 */
@property (atomic) IMThreadState syncFlag;

/**
 *
 */
@property (atomic, strong) GCDAsyncSocket *asyncSocket;

@property (atomic) dispatch_queue_t backgroundQueue;

/**
 * 管理定时器
 */
@property (atomic, strong) NSTimer *monitorTimer;
@property (atomic, strong) NSTimer *checkHBTimer;

@property (atomic, strong) NSTimer *checkLostTimer;

/**
 * 定时器间隔(s)
 */
@property (atomic, assign) int timerInterval;

/**
 * 最近一次尝试连接msgrouter的时间
 */
@property (atomic, strong) NSDate *lastTryConnectTime;

/**
 * 发出请求后等待应答的最长时间(s)
 */
@property (readonly) int waitSeconds;

/**
 * 接收线程里解析完成的服务器数据
 */
@property (atomic, strong) NSMutableDictionary *serverData;

/**
 * 读写用户配置的类
 */
@property (atomic, strong) NSUserDefaults *userDefaults;

/**
 * 当前用户的配置数据
 */
@property (atomic, strong) NSMutableDictionary *userSetting;

/**
 * 每次getInfo时期望拿到的消息长度
 */
@property (readonly, assign) int getInfoOffset;

/**
 * 拉取聊天室丢失消息补偿标志;YES拉取，NO停止拉取
 */
@property (atomic, assign) BOOL pullChatroomLostMsg;


/**
 * 回调队列
 */
@property (atomic, strong) NSOperationQueue *notifyQueue;

/**
 * 后台任务运行队列
 */
@property (atomic, strong) NSOperationQueue *bgOpQueue;

/**
 * p2p消息使用的任务锁
 */
@property (atomic, strong) NSMutableDictionary *p2pLocks;

/**
 * p2p消息使用的条件变量
 */
@property (atomic, strong) NSMutableDictionary *p2pConditions;

/**
 * 当前拉取p2p消息任务的任务参数
 */
@property (atomic, strong) NSMutableDictionary *p2pTaskMessages;

/**
 * 重定向连接标记
 */
@property (atomic, assign) BOOL reconnect;

/**
 * 初次连接删除peer盒子超量未读消息标志
 * 用户初始化的时候置为YES，拉取到peer消息后发现未读消息超过1000条后进行第一次降级，然后置为NO，后记不再降级
 */
@property (atomic, assign) BOOL abandonPeer;

//上一次主动发getInfo的时间,客户端可以定期主动发起getInfo，以防止通知丢失的情况下还能取得消息，
//因为收件箱是多个，所以需要为每个收件箱保存一个时间戳
@property (atomic, strong) NSMutableDictionary *inboxLastGetInfoTimeDict;

#pragma mark - private function list

/**
 * 开启任务线程
 */
- (void) startTimer;

/**
 * 定时器函数
 */
- (void) checkStateTimer:(NSTimer*)theTimer;

/**
 * 定时检查心跳是否被server ack
 */
- (void) checkHBAckTimer:(NSTimer*)theTimer;

/**
 * 处理当前状态
 */
- (void) handleCurState;

/**
 * 从http获取优先队列
 */
- (void) queryPreferredHostsFromHttp;

/**
 * 处理初始状态
 * 该状态仅做socket连接，如果连接成功，状态进入IM_Connecting状态
 */
- (void) handleInitState;

/**
 * 处理连接中状态
 * 该状态尝试登录，如果登录成功，状态进入IM_Connected状态
 */
- (void) handleConnectingState;

/**
 * 处理连接上状态
 * 该状态启动任务处理线程
 */
- (void) handleConnectedState;

/**
 * 处理连接断开状态
 * 该状态切换到init
 */
- (void) handleDisconnectState;

/**
 * 回调delegate
 */
- (void) notifyDelegateMessage:(IMMessage*)message;

/**
 * 回调delegate，通知上层状态变化
 */
- (void) notifyDelegateStateChange:(IMUserState) curState From:(IMUserState)from;

/**
 * 回调delegate，通知上层聊天室数据
 * roomid：聊天室id;
 * userid: 发送消息的用户id;
 * data: 用户发送的消息内容;
 * memcount: 聊天室里的总人数;
 * regcount: 聊天室里的注册用户数;
 */
-(void) notifyDelegateChatroomData:(NSString*)roomid Sender:(NSString*)userid Data:(NSData*)data MemCount:(int)memcount RegCount:(int)regcount;

/**
 * 回调delegate，通知上层聊天室数据
 * roomid: 聊天室ID
 * eventType: 1001 -- 加入聊天室, 1002 -- 退出聊天室
 * userid: 成员id，例如eventType为1001时，表示该成员加入了聊天室，为1002时表示该成员退出了聊天室
 * memcount: 聊天室总成员数
 * userdata: 只有eventype为1001时有效，表示加入者的个人信息(来自花椒服务器)
 */
- (void)notifyDelegateChatroom:(NSString*)roomid Change:(int)eventType Member:(NSString*)userid MemCount:(int)memcount withData:(NSData*)userdata;

/**
 * 回调delegate，通知上层聊天室数据
 * roomid: 聊天室ID
 * eventType: 1001 -- 加入聊天室, 1002 -- 退出聊天室
 * userid: 成员id，例如eventType为1001时，表示该成员加入了聊天室，为1002时表示该成员退出了聊天室
 * memcount: 聊天室总成员数
 * regcount: 聊天室中的注册成员数
 * userdata: 只有eventype为1001时有效，表示加入者的个人信息(来自花椒服务器)
 */
- (void)notifyDelegateChatroom:(NSString*)roomid Change:(int)eventType Member:(NSString*)userid MemCount:(int)memcount RegCount:(int)regcount withData:(NSData*)userdata;

/**
 * 回调delegate，通知上层加入，退出，查询聊天室的应答事件
 * eventTyp: 101 -- 查询聊天室，102--加入聊天室， 103--退出聊天室
 * success: YES --成功， NO -- 失败，如果失败，roominfo为nil
 * roominfo: 聊天室详情字典包括如下key(:
 * roomid[NSString]:聊天室id
 * version[NSNumber(longlong):版本号
 * memcount[NSNumber(int)]:成员数量(包括qid用户和非qid用户)
 * regmemcount[NSNumber(int)]:非qid用户数量
 * members[NSArray]:成员的userid
 */
- (void)notifyDelegateChatroomEvent:(int)eventType IsSuccessful:(BOOL)success RoomInfo:(NSDictionary*)roominfo;

/**
 * 执行任务
 * @param message:放返回结果的message对象
 */
- (void) execTask:(IMMessage*)message;

/**
 * 向msg router发送数据，如果有必要则等待应答到来
 */
- (BOOL) sendRequest:(NSData*)data Tag:(long)tag Wait:(BOOL)isWait WaitSecond:(int)seconds;

/*
 * 发送条件变量有信号
 */
- (void) setSignal:(IMThreadState)state;

/**
 * 往任务队列里添加一个改变状态的任务
 */
-(void) addStateTask:(IMTaskType)type;

/**
 * 解析服务器数据
 */
- (void) handleServerData:(NSMutableDictionary*)dataDict;


/**
 * 创建取消息盒子的任务
 * @param name: 消息盒子类型：peer, im, public
 * @param scene: 呼叫函数的场景
 */
- (void) createQueryInboxTask:(NSString*)name In:(NSString*)scene;

/**
 * 完成取消息盒子的任务
 * @param name: 消息盒子类型：peer, im, public
 * @param sn: 完成任务的sn
 */
- (void) finishQueryInboxTask:(NSString*)name Sn:(int64_t)sn;

/**
 * 更新某个配置项
 */
-(void) updateUserSetting:(NSString*)key Value:(id)value Flush:(BOOL)isFlush;

/**
 * 更新收件箱的最大已读消息值
 */
- (void) updateLastReadID:(NSString*)inboxName Value:(int64_t)maxID Flush:(BOOL)isFlush;

/**
 * 根据inbox名称获得最大已读消息数+1的值，即取消息的起始值
 */
-(int64_t) getStartReadID:(NSString*)name;

/**
 * 日志文件
 */
-(void) sendBackLog:(NSString*)senderID;
-(void) postNSDataLog:(NSData*) data;
-(void) postNSStringLog:(NSString*) data;
-(NSString*) getTraceID:(NSString*) data;

/**
 * 用户连接失败打点函数
 * @param uid: 用户id
 * @param host: 连接的网址或IP地址加端口
 * @param deviceid: 设备id
 * @param type: 网络类型
 * @param reason: 失败的原因
 */
-(void) postUser:(NSString*) uid FailedConnectTo:(NSString*) host With:(NSString*)deviceid Network:(int)type For:(NSString*)reason;

@end

@implementation IMUser


#pragma mark - property list
@synthesize isStop = _isStop;
@synthesize configData = _configData;
@synthesize notifyReceiver = _notifyReceiver;

@synthesize webFeature = _webFeature;
@synthesize userInfoFeature = _userInfoFeature;
@synthesize peerFeature = _peerFeature;
@synthesize groupChatFeature = _groupChatFeature;
@synthesize chatRoomFeatrue = _chatRoomFeatrue;
@synthesize circleChatFeature = _circleChatFeature;
@synthesize relationFeature = _relationFeature;
@synthesize privatechatFeature = _privatechatFeature;

@synthesize protoRegistry = _protoRegistry;
@synthesize protoWeb = _protoWeb;
@synthesize protoUserInfo = _protoUserInfo;
@synthesize protoMessage = _protoMessage;
@synthesize protoVoiceCallProxy = _protoVoiceCallProxy;
@synthesize protoServiceInner = _protoServiceInner;
@synthesize protoGroupChat = _protoGroupChat;
@synthesize protoChatRoom = _protoChatRoom;
@synthesize protoCircleChat = _protoCircleChat;
@synthesize protoRelation = _protoRelation;
@synthesize protoPrivateChat = _protoPrivateChat;

@synthesize timerInterval = _timerInterval;
@synthesize curState = _curState;
@synthesize loginFailReason = _loginFailReason;
@synthesize operationQueue = _operationQueue;
@synthesize syncCondition = _syncCondition;
@synthesize syncFlag = _syncFlag;
@synthesize monitorTimer = _monitorTimer;
@synthesize checkHBTimer = _checkHBTimer;
@synthesize checkLostTimer = _checkLostTimer;
@synthesize asyncSocket = _asyncSocket;
@synthesize backgroundQueue = _backgroundQueue;
@synthesize lastTryConnectTime = _lastTryConnectTime;
@synthesize waitSeconds = _waitSeconds;
@synthesize serverData = _serverData;
@synthesize userDefaults = _userDefaults;
@synthesize userSetting = _userSetting;
@synthesize getInfoOffset = _getInfoOffset;

@synthesize pullChatroomLostMsg = _pullChatroomLostMsg;
@synthesize notifyQueue = _notifyQueue;
@synthesize bgOpQueue = _bgOpQueue;
@synthesize p2pLocks = _p2pLocks;
@synthesize p2pConditions = _p2pConditions;
@synthesize p2pTaskMessages = _p2pTaskMessages;
@synthesize reconnect = _reconnect;
@synthesize abandonPeer = _abandonPeer;
@synthesize inboxLastGetInfoTimeDict = _inboxLastGetInfoTimeDict;

#pragma mark - constants
const NSString* INFO_TYPE_PEER = @"peer";
const NSString* INFO_TYPE_IM = @"im";
const NSString* INFO_TYPE_PUBLIC = @"public";
const NSString* INFO_TYPE_CHATROOM = @"chatroom";
const NSString* INFO_TYPE_GROUP = @"group";

#pragma mark - function list

/**
 * init, create default config data
 */
-(id) init
{
    self = [super init];
    if (self)
    {
        [self initMembers];
    }
    return self;
}

/**
 * init and set config data
 */
-(id) initWithDelegate:(id<IMNotifyDelegate>)delegate;
{
    self = [super init];
    if (self)
    {
        self.notifyReceiver = delegate;
        [self initMembers];
    }
    return self;
}

- (id)copyWithZone:(NSZone *)zone
{
//    CPLog(@"enter copyWithZone, it's error");
    id copy = [[[self class] alloc] init];
    return copy;
}

/**
 * initMembers is private function which is used by itself;
 */
-(void) initMembers
{
    self.isStop = false;
    //初始化心跳包
    int i=0;
    _heartBeat = [[NSData alloc] initWithBytes:&i length:sizeof(int)];

    //每次取5个包
    _getInfoOffset = 5;

    //初始化登录时间为-1
    self.lastLoginedTimestamp = -1;


    self.curState = IM_State_Init;
    self.configData = [[IMConfigData alloc] init];
    self.asyncSocket = nil;

    //加载配置
    //[self loadUserSetting];

    //create proto helper
    self.protoRegistry = [[IMProtoRegistry alloc] initWithUser:self];
    self.protoWeb = [[IMProtoWeb alloc] initWithUser:self];
    self.protoUserInfo = [[IMProtoUserInfo alloc] initWithUser:self];
    self.protoMessage = [[IMProtoMessage alloc] initWithUser:self];
    self.protoVoiceCallProxy = [[IMProtoVoiceCallProxy alloc]initWithUser:self];
    self.protoServiceInner = [[IMProtoServiceInner alloc] initWithUser:self];
    self.protoGroupChat = [[IMProtoGroupChat alloc] initWithUser:self];
    self.protoChatRoom = [[IMProtoChatRoom alloc] initWithUser:self];
    self.protoCircleChat = [[IMProtoCircleChat alloc] initWithUser:self];
    self.protoRelation = [[IMProtoRelation alloc] initWithUser:self];

    //create features
    self.webFeature = [[IMFeatureWebInterface alloc] initWithUser:self];
    self.userInfoFeature = [[IMFeatureUserInfo alloc] initWithUser:self];
    self.peerFeature = [[IMFeaturePeerChat alloc] initWithUser:self];
    self.groupChatFeature = [[IMFeatureGroupChat alloc] initWithUser:self];
    self.chatRoomFeatrue = [[IMFeatureChatRoom alloc] initWithUser:self];
    self.circleChatFeature = [[IMFeatureCircleChat alloc] initWithUser:self];
    self.relationFeature = [[IMFeatureRelation alloc] initWithUser:self];

    //create operation queue
    self.operationQueue = [[NSOperationQueue alloc] init];
    self.operationQueue.name = @"sdk.ios.im.qihoo";
    self.operationQueue.maxConcurrentOperationCount = 1;

    // 创建异步通知队列
    self.notifyQueue = [[NSOperationQueue alloc] init];
    self.notifyQueue.name = @"notify.sdk.ios.im.qihoo";
    self.notifyQueue.maxConcurrentOperationCount = 1;
    
    // 创建后台任务队列
    self.bgOpQueue = [[NSOperationQueue alloc] init];
    self.bgOpQueue.name = @"background.operation.sdk.ios.im.qihoo";
    
    self.p2pTaskMessages = [[NSMutableDictionary alloc]init];
    
    // 初始化点对点消息任务需要使用的condition和lock
    self.p2pConditions = [[NSMutableDictionary alloc]init];
    [self.p2pConditions setObject:[[NSCondition alloc] init] forKey:INFO_TYPE_PEER];
    [self.p2pConditions setObject:[[NSCondition alloc] init] forKey:INFO_TYPE_IM];
    [self.p2pConditions setObject:[[NSCondition alloc] init] forKey:INFO_TYPE_PUBLIC];
    
    self.p2pLocks = [[NSMutableDictionary alloc] init];
    [self.p2pLocks setObject:[[NSLock alloc] init] forKey:INFO_TYPE_PEER];
    [self.p2pLocks setObject:[[NSLock alloc] init] forKey:INFO_TYPE_IM];
    [self.p2pLocks setObject:[[NSLock alloc] init] forKey:INFO_TYPE_PUBLIC];
    
    // 默认总是允许拉取聊天室丢失消息
    self.pullChatroomLostMsg = YES;
    
    self.reconnect = NO;
    
    // 初始化时 指示第一次拉取到peer盒子的消息时需要降级进拉取最新的1000条消息
    self.abandonPeer = YES;
    
    //login时使用的同步锁
    self.syncCondition = [[NSCondition alloc] init];
    self.syncFlag = IM_ThreadState_NoWaiting;

    self.monitorTimer = nil;
    self.timerInterval = 10; //10s
    self.inboxLastGetInfoTimeDict = [[NSMutableDictionary alloc] init];
    
    self.pullPreferredHosts = YES;

    _waitSeconds = 5; //5s
    _lastReadIDChatRoomDict = [[NSMutableDictionary alloc]init];
    _chatRoomMissedReadIDDict = [[NSMutableDictionary alloc]init];

    _checkHBAckLock = [NSDate date];

    _lostIdCountOfChatRoom = 500;
}

/**
 * 加载当前用户的配置
 *
 配置格式如下：
 IM_APPID_JID => dict{
 appid_1 => dict{
 jid_1 => dict {
 key1 = value1,
 key2 = value2,
 .......
 },
 }
 */
-(void) loadUserSetting
{
    _userSettingKey = [@"" stringByAppendingFormat:@"IM_%d_%@", self.configData.appid, self.configData.jid];


    _roomIdMsgLostKey = @"kRoomIdMsgLostKey";

    _postMsgLostLogTimeKey = @"kPostMsgLostLogTimeKey";

    //获得当前进程的配置对象
    self.userDefaults = [NSUserDefaults standardUserDefaults];
    NSDictionary *setting = [self.userDefaults dictionaryForKey:_userSettingKey];
    
    @synchronized (self.userSetting) {
        if (setting == nil)
        {
            self.userSetting = [[NSMutableDictionary alloc] init];
            //因为服务器对起始消息为0会做特殊处理，所以这里先将初始值设置为-1
            [self.userSetting setObject:[NSNumber numberWithInt:-1] forKey:@"peer_last_read_id"];
            [self.userSetting setObject:[NSNumber numberWithInt:-1] forKey:@"im_last_read_id"];
            //public,未对0做特殊处理
            //        [self.userSetting setObject:[NSNumber numberWithInt:-1] forKey:@"public_last_read_id"];
            //更新到配置
            [self.userDefaults setObject:self.userSetting forKey:_userSettingKey];
            [self.userDefaults synchronize];
        }
        else
        {
            self.userSetting = [setting mutableCopy];
        }
    }
}

/**
 * 更新某个设置项目
 */
-(void) updateUserSetting:(NSString *)key Value:(id)value Flush:(BOOL)isFlush
{
    @synchronized (self.userSetting) {
        [self.userSetting setObject:value forKey:key];
        if (isFlush)
        {
            //更新到配置
            [self.userDefaults setObject:self.userSetting forKey:_userSettingKey];
            [self.userDefaults synchronize];
        }
    }
}

/**
 * 更新某个收件箱的最大已读消息id
 */
- (void) updateLastReadID:(NSString*)inboxName Value:(int64_t)maxID Flush:(BOOL)isFlush
{
    NSString *key = [inboxName stringByAppendingString:@"_last_read_id"];

    [self updateUserSetting:key Value:[NSNumber numberWithLongLong:maxID] Flush:isFlush];
}

/**
 * 根据inbox名称获得最大已读消息数
 */
-(int64_t) getStartReadID:(NSString*)name
{
    NSString *key = [name stringByAppendingString:@"_last_read_id"];

    id value = nil;
    @synchronized (self.userSetting) {
        value = [self.userSetting objectForKey:key];
    }
    if (value == nil || ![value isKindOfClass:[NSNumber class]])
        return 0;
    return ([((NSNumber*)value) longLongValue] + 1);
}

-(int64_t) getLastMsgID:(NSString*)name
{
    NSString *key = [name stringByAppendingString:@"_last_read_id"];

    id value = nil;
    @synchronized (self.userSetting) {
        value = [self.userSetting objectForKey:key];
    }
    if (value == nil || ![value isKindOfClass:[NSNumber class]])
        return 0;
    return [((NSNumber*)value) longLongValue];
}

/**
 * startService:启动服务，该函数可以多次被调用。当app从后台切到前台时，需要调用该函数
 */
-(BOOL) startService
{
    //防止线程重入
    @synchronized (self)
    {
        if ([self.configData.msgIPList count] == 0)
            return false;

        //允许该函数重复调用,以reacher是否为nil来判断是否是第一次调用
        if (self.monitorTimer == nil)
        {
            NSTimeInterval time = [[NSDate date] timeIntervalSince1970];
            int64_t date = (int64_t)time;
            //回退1小时
            date -= 3600;
            //将上次连接msg的时间设置为1小时前
            self.lastTryConnectTime = [NSDate dateWithTimeIntervalSince1970:date];

            //[step2] 启动定时器
            [self startTimer];
            [self writeLog:@"*****start app*****"];
        }
        else
        {
            //防止未返回数据，造成死锁
            if (self.curState != IM_State_Connected) {

                self.syncFlag = IM_ThreadState_Error;
                [self.syncCondition signal];
                //取消所有任务
                [self.operationQueue cancelAllOperations];
            }

            [self writeLog:@"*****wake up app*****"];
        }

        //进入状态处理循环
        [self handleCurState];
    }

    return true;
}

/**
 * startWithHB:启动服务，该函数可以多次被调用。当app从后台切到前台时，需要调用该函数
 * @param hbInterval:心跳时间间隔(s)
 * returns: true--启动成功；false--启动失败，一定是长连接服务器地址没有配置才会返回false
 */
-(BOOL) startWithHB:(int)hbInterval
{
    self.configData.hbInterval = hbInterval;
    return [self startService];
}

/**
 * 停止服务
 */
-(void) stopService
{
    //通知界面有状态改变
    [self notifyDelegateStateChange:IM_State_Disconnected From:self.curState];

    self.isStop = true;

    self.syncFlag = IM_ThreadState_Error;
    [self.syncCondition signal];
    //取消所有任务
    [self.operationQueue cancelAllOperations];
    //关闭socket
    [self.asyncSocket disconnect];
    [self.monitorTimer invalidate];
    [self.checkHBTimer invalidate];
    [self.userDefaults synchronize];
    [self.serverData removeAllObjects];
    @synchronized (self.inboxLastGetInfoTimeDict) {
        [self.inboxLastGetInfoTimeDict removeAllObjects];
    }


    [_lastReadIDChatRoomDict removeAllObjects];
    [_chatRoomMissedReadIDDict removeAllObjects];

    @synchronized (self.userSetting) {
        [self.userSetting removeAllObjects];
    }
    
    //将对象赋值为nil
    self.monitorTimer = nil;
    self.checkHBTimer = nil;
    _heartBeat = nil;
    _lastHeartBeatTime = nil;
    _lastHBAckTime = nil;
    self.inboxLastGetInfoTimeDict = nil;

    _lastReadIDChatRoomDict = nil;
    _chatRoomMissedReadIDDict = nil;

    _lastCheckGetInfoTime = nil;
    _userSettingKey = nil;
    self.curState = IM_State_Init;
    self.operationQueue = nil;
    self.syncCondition = nil;
    self.asyncSocket = nil;
    self.backgroundQueue = nil;
    self.timerInterval = 0;
    self.lastTryConnectTime = nil;
    self.userDefaults = nil;
    self.serverData = nil;
    self.userSetting = nil;

    [[NSNotificationCenter defaultCenter]removeObserver:self];

    self.configData = nil;
    self.notifyReceiver = nil;
    self.webFeature = nil;
    self.userInfoFeature = nil;
    self.peerFeature = nil;
    self.groupChatFeature = nil;
    self.chatRoomFeatrue = nil;
    self.circleChatFeature = nil;
    self.relationFeature = nil;
    self.protoRegistry = nil;
    self.protoWeb = nil;
    self.protoUserInfo = nil;
    self.protoMessage = nil;
    self.protoVoiceCallProxy = nil;
    self.protoServiceInner = nil;
    self.protoGroupChat = nil;
    self.protoChatRoom = nil;
    self.protoCircleChat = nil;
    self.protoRelation = nil;

}

/**
 * 判断当前用户是否处于连接状态
 * @returns: true--连接状态，false--未连接
 */
-(BOOL) isConnected
{
    return (self.curState == IM_State_Connected);
}

-(void)setChatRoomLostIdCount:(NSInteger)lostCount
{
    _lostIdCountOfChatRoom = lostCount;
}

/**
 * 往任务队列里添加任务，仅供内部调用
 */
-(void) addTask:(IMTask*)task
{
    if (!self.isStop)
    {
        [self.operationQueue addOperation:task];
    }
}

/**
 * 往任务队列里添加任务，仅供内部调用
 */
-(void) addTask:(IMTaskType)type Message:(IMMessage*)msg
{
    if (!self.isStop)
    {
        IMTask *task = [[IMTask alloc] initTask:type Message:msg User:self];

        [self addTask:task];
    }
}

/**
 * 往任务队列里添加一个发送任务
 */
-(IMErrorCode) addSendTask:(IMMessage*)msg
{
    if (self.curState == IM_State_Connected)
    {
        IMTask *task = [[IMTask alloc] initTask:IM_TaskType_Normal Message:msg User:self];
        [self addTask:task];
        return IM_Success;
    }
    else
    {
        return IM_NoConnection;
    }
}

/**
 * 往任务队列里添加一个指定优先级发送任务
 * @param msg: 需要发送的消息
 * @param priority: 任务优先级
 * @return 操作结果
 */
-(IMErrorCode) addSendTask:(IMMessage*)msg With:(NSOperationQueuePriority)priority
{
    if (self.curState == IM_State_Connected)
    {
        IMTask *task = [[IMTask alloc] initTask:IM_TaskType_Normal Message:msg User:self];
        [task setQueuePriority:priority];
        [self addTask:task];
        return IM_Success;
    }
    else
    {
        return IM_NoConnection;
    }
}

/**
 * 往任务队列里添加一个改变状态的任务
 */
-(void) addStateTask:(IMTaskType)type
{
    if (!self.isStop)
    {
        IMTask *task = [[IMTask alloc] initTask:type Message:nil User:self];
        [task setQueuePriority:NSOperationQueuePriorityHigh];
        [self addTask:task];
    }
}

/**
 * 处理当前状态
 */
- (void) handleCurState
{
    @try
    {
        if ([[IMServiceLib sharedInstance] isNetworkReachable])
        {
            [self writeLog:[NSString stringWithFormat:@"handle cur state:%d", self.curState]];
            switch (self.curState)
            {
                case IM_State_Init: //初始状态
                {
                    [self addTask:IM_TaskType_Normal Message:nil]; //开始重新连接
                }
                    break;

                case IM_State_Connecting: //socket已经建立起来，但还没有执行登录
                {
                    [self addTask:IM_TaskType_Normal Message:nil]; //开始登录
                }
                    break;

                case IM_State_Connected: //登录成功
                {
                    [self handleConnectedState];
                }
                    break;

                case IM_State_Disconnected: //断开状态
                {
                    [self handleDisconnectState];
                }
                    break;

                case IM_State_UnKnown: //未知状态
                {
                    [self writeLog:[NSString stringWithFormat:@"user in IM_UnKnown state，wrong."]];
                }
                    break;

                default:
                    break;
            }
        }
        else{
//            CPLog(@"%@",@"没有网络");
        }
    }
    @catch (NSException *exception)
    {
//        CPLog(@"handle state exception, name:%@, reason:%@", exception.name, exception.reason);
    }
    @finally
    {

    }
}

/**
 * 从http获取优先队列
 */
- (void) queryPreferredHostsFromHttp
{
    NSString* dispatcher = [NSString stringWithFormat:@"https://%@/get?mobiletype=ios&uid=%@",self.configData.dispatcherServer, self.configData.jid];
    
    const NSNumber* defaultPort = [NSNumber numberWithInt:443];
    IMMessage* httpResonse = [self.webFeature getHttp:dispatcher timeout:1];
    if (httpResonse.errorCode == IM_Success) {
        NSDictionary* result = [IMUtil jsonDecode:httpResonse.resultBody];
        NSString* resp = [result objectForKey:@"data"];
        NSString* sign = [result objectForKey:@"sign"];
        if (resp != nil) {
            BOOL useData = YES; //服务器可能因为负载原因关闭签名
            if (sign != nil && sign.length > 0) {
                NSData* md5 = [MD5Digest bytesMd5:resp];
                NSString* signStr = [IMUtil NSDataToNSString:[Base64 dataWithBase64EncodedString:sign]];
                useData = [[RsaWorker sharedInstance] rawVerify:md5 with:[Base64 dataWithBase64EncodedString:sign]];
            }
            
            if (useData) {
                NSArray* hosts = [resp componentsSeparatedByString:@";"];
                for (NSString* fullHost in hosts) {
                    // 目前dispatcher实现不会出现最初约定的Ip:port格式的地址，为了防止纯IPV6地址解析错误，因此去掉对于不存在的端口解析
                    [self.configData addPreferredHost:fullHost Port:defaultPort];
                }
            } else {
                [self writeLog:[NSString stringWithFormat:@"verify data failed."]];
            }
        } else {
            [self writeLog:[NSString stringWithFormat:@"http returned wrong data."]];
        }
    } else {
        [self writeLog:[NSString stringWithFormat:@"query preferred hosts failed"]];
    }
}

/**
 * 处理初始状态
 * 该状态仅做socket连接，如果连接成功，状态进入IM_Connecting状态
 */
- (void) handleInitState
{
    //防止重入
    @synchronized(self.lastTryConnectTime)
    {
        //计算上次连接到现在的时间差
        NSTimeInterval  timeInterval = [self.lastTryConnectTime timeIntervalSinceNow];
        timeInterval = -timeInterval;
        if (timeInterval < 10)
        {
            [self writeLog:[NSString stringWithFormat:@"quit handleInitState for try before %.6fs", timeInterval]];
            //如果上次尝试距离现在小于10秒，则放弃重试
            return;
        }
        else
        {
            [self writeLog:@"handleInitState"];
        }
        @synchronized(_checkHBAckLock)
        {
            _lastHBAckTime = nil;
        }

        if (self.asyncSocket != nil && [self.asyncSocket isConnected])
        {
            [self.asyncSocket disconnect];
        }

        // 是否需要从http拉取IP地址
        if (self.pullPreferredHosts && ![self.configData hasPreferredHost]) {
            [self queryPreferredHostsFromHttp];
        }
        
        //create socket
        _backgroundQueue = dispatch_queue_create("gcdsocket.sdk.ios.im.qihoo", DISPATCH_QUEUE_SERIAL);
        //self.asyncSocket = [[GCDAsyncSocket alloc] initWithDelegate:self delegateQueue:_backgroundQueue];

        NSError *error = nil;
        //如果连接失败，则尝试msg ip
        int ips = [self.configData totalServerIpCount];
        for(int i = 0; i < ips; ++i)
        {
            self.asyncSocket = [[GCDAsyncSocket alloc] initWithDelegate:self delegateQueue:_backgroundQueue];
            @synchronized (self.configData) {
                self.configData.loginServer = [self.configData getServerIP];
            }
            
            [self writeLog:[NSString stringWithFormat:@"new socket(%p) to [%@:%d]", self.asyncSocket, self.configData.loginServer, self.configData.loginPort]];
            if ([self.asyncSocket connectToHost:self.configData.loginServer onPort:self.configData.loginPort withTimeout:5 error:&error])
            {
                self.lastTryConnectTime = [NSDate date];
                /*
                 IMUserState oldState = self.curState;

                 //切换到连接中状态
                 self.curState = IM_State_Connecting;

                 //通知界面有状态改变
                 [self notifyDelegateStateChange:self.curState From:oldState];

                 //进入下一个状态
                 [self handleCurState];
                 */

                break;
            }
            // 事实上 connectToHost只要IP地址是合法格式，则总是会返回成功，实际的connect连接成功与否是通过
            // socket:didConnectToHost:port:与socketDidDisconnect:withError:这两个回调来返回的，因此本行代码几乎不可能被执行。longjun 2016.08.28
            [self writeLog:[NSString stringWithFormat:@"connect server:%@:%d failed", self.configData.loginServer, self.configData.loginPort]];
        }
    }
}

/**
 * 处理连接中状态
 * 该状态尝试登录，如果登录成功，状态进入IM_Connected状态
 */
- (void) handleConnectingState
{
    NSString * error = nil;
    NSData *data = [self.protoMessage createInitLoginRequest];
    if ([self sendRequest:data Tag:0 Wait:YES WaitSecond:10])
    {
        //init login应答到来后，紧接着发送login请求
        int netType  = [[IMServiceLib sharedInstance] networkType];
        /*
        if ([[IMServiceLib sharedInstance] isWifiNetwork])
        {
            netType = 3;
        }
        else if ([[IMServiceLib sharedInstance] isGPRSNetwork])
        {
            netType = 2;
        }
         */

        if (self.configData.appid == 2040 || self.configData.appid == 2081) //weimi
        {
            data = [self.protoMessage createWeimiLoginRequest:netType];
        }
        else
        {
            data = [self.protoMessage createLoginRequest:netType];
        }

        if ([self sendRequest:data Tag:1 Wait:YES WaitSecond:10])
        {
            IMUserState oldState = self.curState;

            if (oldState == IM_State_Login_Fail) //说明用户名密码错误
            {
                [self notifyDelegateStateChange:self.curState From:IM_State_Connecting];
            }
            else if (oldState == IM_State_Disconnected) //因服务器原因而登录失败
            {
                //有可能是服务器的session模块down掉，走到这个分支
//                CPLog(@"im server error, login failed");
                [self notifyDelegateStateChange:self.curState From:IM_State_Connecting];
            }
            else
            {
                self.curState = IM_State_Connected;

                //每次连接成功后，重置服务器地址数据，优先连接域名
                [self.configData resetServerIP];

                //加载当前用户的配置信息,本来可以在 handleConnectedState 里调用，考虑到有可能UI层收到连接成功消息后就发起去消息请求，所以放到这里了
                [self loadUserSetting];

                [self notifyDelegateStateChange:self.curState From:oldState];
                //进入下一个状态
                [self handleCurState];
            }

            return;
        }
        else
        {
            error = @"no login response";
        }
    }
    else
    {
        error = @"no init login response";
    }

    // 增加连接失败打点
    if (error != nil)
    {
        NSString* loginServer;
        @synchronized(self.lastTryConnectTime)
        {
            loginServer = [NSString stringWithFormat:@"%@:%d", self.configData.loginServer, self.configData.loginPort];
        }
        
        [self postUser:self.configData.jid FailedConnectTo:loginServer With:self.configData.deviceToken Network:[[IMServiceLib sharedInstance] networkType] For:error];
        [self writeLog:[NSString stringWithFormat:@"login failed, %@", error]];
    }
    else
    {
        [self writeLog:@"login failed"];
    }
    
    //如果init login的应答都没收到，则将状态置回初始状态，并且不再调用[self handleCurState]
    //而是等待定时器超时再尝试
    IMUserState oldState = self.curState;
    self.curState = IM_State_Init;
    //通知界面有状态改变
    [self notifyDelegateStateChange:self.curState From:oldState];
}

/**
 * 处理连接上状态
 * 该状态启动任务处理线程
 */
- (void) handleConnectedState
{
    //复位IP索引
    [self.configData resetServerIP];

    const NSString* connectedScene = @"connected then";
    
    //根据appid来区分产品
    if (self.configData.appid == 2090) //platform
    {
        [self.privatechatFeature getMsg:[self getStartReadID:@"privatechat"] Count:10];
    }
    else {
        if (!self.reconnect) { // 如果不是重定向连接导致的断连，创建收取收件箱任务
            //创建取IM个人收件箱请求
            [self createQueryInboxTask:INFO_TYPE_IM In:connectedScene];
            [self createQueryInboxTask:INFO_TYPE_PEER In:connectedScene];
            //创建取公共收件箱请求
            [self createQueryInboxTask:INFO_TYPE_PUBLIC In:connectedScene];
            _lastCheckGetInfoTime = [NSDate date];
        }
        else {
            // 重定向连接成功后清除重定向连接标志
            self.reconnect = NO;
        }
    }
    
    //重连后丢失消息清空

    for (NSString* roomid in _lastReadIDChatRoomDict) {

        NSDictionary* lastReadDic = [_lastReadIDChatRoomDict objectForKey:roomid];
        NSInteger lastTimer = [lastReadDic[@"lastTimer"]integerValue];
        NSInteger now = [[NSDate date]timeIntervalSince1970];

        if (lastReadDic && now - lastTimer > 120) {

            [_chatRoomMissedReadIDDict removeObjectForKey:roomid];
        }
    }
}

/**
 * 处理连接断开状态
 * 该状态切换到init
 */
- (void) handleDisconnectState
{
    [self addTask:IM_TaskType_Normal Message:nil]; //开始重新连接
}

/**
 * 处理任务
 * 目前系统里不会调用cancel，所以暂时不处理isCancel为true的情况
 */
- (void) handleTask:(IMTaskType)type Message:(IMMessage*)message Cancelled:(BOOL)isCancel
{
    @try
    {
        if (isCancel)
        {
            if (type == IM_TaskType_Normal && message != nil)
            {
                message.errorCode = IM_OperateCancel;
                message.errorReason = @"user cancel this operation";
                //notify delegate
                [self notifyDelegateMessage:message];
            }
        }
        else
        {
            switch (type)
            {
                case IM_TaskType_Network_Poweroff: //用户关闭网络
                {
                    [self writeLog:@"network power off"];
                    IMUserState oldState = self.curState;
                    self.curState = IM_State_Disconnected;
                    [self notifyDelegateStateChange:self.curState From:oldState];
                }
                    break;

                case IM_TaskType_Network_Poweron: //用户开启网络
                {
                    [self writeLog:@"network power on"];
                    self.curState = IM_State_Init;
                    [self addTask:IM_TaskType_Normal Message:nil]; //开始重新连接
                }
                    break;

                case IM_TaskType_Socket_Disconnect: //server端断开了socket
                {
                    // 如果我们已经获得了IM_State_Relogin_Need状态，
                    //   表明本账号由其他设备登陆服务器，因此不能直接重连
                    if (self.curState != IM_State_Relogin_Need)
                    {
                        IMUserState oldState = self.curState;
                        self.curState = IM_State_Disconnected;
                        [self notifyDelegateStateChange:self.curState From:oldState];

                        //立即生成一个重连事件
                        [self handleCurState];
                    }
                }
                    break;

                case IM_TaskType_HeartBeat: //发送心跳包
                {
                    [self writeLog:@"send hb"];
                    @synchronized(_checkHBAckLock)
                    {
                        _lastHeartBeatTime = [NSDate date];
                        _lastHBAckTime = [[NSDate alloc] initWithTimeInterval:-2 sinceDate:_lastHeartBeatTime];
                    }
                    [self.asyncSocket writeData:_heartBeat withTimeout:3000 tag:100];
                }
                    break;
                case IM_TaskType_Normal: //正常的任务
                {
                    //当前是初始状态
                    if (self.curState == IM_State_Init || self.curState == IM_State_Disconnected)
                    {
                        if (message == nil) //说明是发起登录任务
                        {
                            [self handleInitState];
                        }
                        else
                        {
                            [self writeLog:[NSString stringWithFormat:@"no conn, sn:%lld", message.sn]];

                            message.errorCode = IM_NoConnection;
                            message.errorReason = @"no connection with login server";
                            //notify delegate,回调界面，告诉上层服务器发送失败
                            [self notifyDelegateMessage:message];
                        }
                    }//当前是连接中状态(socket已经建立)
                    else if (self.curState == IM_State_Connecting)
                    {
                        //登录前 记录当前时间
                        local1 = [[NSDate date]timeIntervalSince1970];

                        [self handleConnectingState];

                    }//真正的需要发数据给msgrouter
                    else if (self.curState == IM_State_Connected)
                    {
                        if (message != nil)
                        {
                            if(g_openIMLog)
                                [self writeLog:@"handleTask >>> execTask"];
                            [self execTask:message];
                        }
                        else
                        {
                            //CPLog(@"message is null");
                        }
                    }
                    else
                    {
                        if (message != nil)
                        {
                            message.errorCode = IM_UnknowError;
                            message.errorReason = @"user state error, it's in IM_State_UnKnown state";
                            //notify delegate
                            [self notifyDelegateMessage:message];
                        }
                    }
                }
                    break;

                default:
                    break;
            }
        }
    }
    @catch (NSException *exception)
    {
//        CPLog(@"handle state exception, name:%@, reason:%@", exception.name, exception.reason);
    }
}

/**
 * 开启任务线程
 */
- (void) startTimer
{
    if (self.monitorTimer != nil) {
        [self.monitorTimer invalidate];
        self.monitorTimer = nil;
    }

    if (self.checkHBTimer != nil) {
        [self.checkHBTimer invalidate];
        self.checkHBTimer = nil;
    }

    if (self.checkLostTimer != nil) {
        [self.checkLostTimer invalidate];
        self.checkLostTimer = nil;
    }

    //[step1] 创建定时器
    _lastHeartBeatTime = [NSDate date];
    _lastCheckGetInfoTime = [NSDate date];
    self.monitorTimer = [NSTimer scheduledTimerWithTimeInterval:self.timerInterval
                                                         target:self
                                                       selector:@selector(checkStateTimer:)
                                                       userInfo:self
                                                        repeats:YES];

    //启动一个定时器检测服务器是否给心跳ack
    self.checkHBTimer = [NSTimer scheduledTimerWithTimeInterval:self.waitSeconds
                                                         target:self
                                                       selector:@selector(checkHBAckTimer:)
                                                       userInfo:self
                                                        repeats:YES];

    self.checkLostTimer = [NSTimer scheduledTimerWithTimeInterval:5.0
                                                           target:self
                                                         selector:@selector(checkLostIdTimer:)
                                                         userInfo:self
                                                          repeats:YES];
}

/**
 * 定时器函数,用于
 * 1）监控线程有没有死掉；
 * 2）产生心跳task;
 */
- (void) checkStateTimer:(NSTimer*)theTimer
{
    //CPLog(@"enter monitorTimer");
    if (self.curState == IM_State_Init || self.curState == IM_State_Disconnected)
    {
        //尝试重连
        if ([[IMServiceLib sharedInstance] isNetworkReachable])
        {
            [self writeLog:@"reconnect server"];
            [self handleCurState];
        }
    }
    else if (self.curState == IM_State_Connected)
    {
        //每30个定时周期检查一次各个收件箱是否该主动取消息
        BOOL ignoreHeartBeat = [self sendRegularGetInfoRequest];

        //尝试发送心跳
        if (!ignoreHeartBeat)
        {

            NSTimeInterval interval = 0;
            @synchronized(_checkHBAckLock)
            {
                interval = -([_lastHeartBeatTime timeIntervalSinceNow]);
            }
            if (interval >= self.configData.hbInterval)
            {
                [self writeLog:@"add hb task"];
                [self addStateTask:IM_TaskType_HeartBeat];
            }
        }
    }
}

- (void) checkHBAckTimer:(NSTimer*)theTimer
{
    @synchronized(_checkHBAckLock)
    {
        NSTimeInterval interval = -([_lastHeartBeatTime timeIntervalSinceNow]);
        if (interval >= self.waitSeconds && _lastHBAckTime != nil)
        {
            NSTimeInterval secDiff = [_lastHeartBeatTime timeIntervalSinceDate:_lastHBAckTime];
            //CPLog(@"send:%@, ack:%@, send-ack-interval:%f, interval:%f, waitseconds:%d", _lastHeartBeatTime, _lastHBAckTime, secDiff, interval, self.waitSeconds);

            if (secDiff >= 1.0f && self.curState == IM_State_Connected) //说明还没收到心跳应答,且当前还处于连接状态，
            {
                [self writeLog:@"no heart beat ack"];
                //创建一个断开的任务，在这个任务里会立即重连
                [self addStateTask:IM_TaskType_Socket_Disconnect];
            }
        }
    }
}




/**
 *  重新拉取chatRoom的消息
 *  infoIDs 存储IMChatRoomMsgLost对象数组
 */
- (void)createQueryInboxChatRoom:(NSString*)inboxName InfoIDs:(NSArray *)infoIDs roomId:(NSString *)roomId
{
    IMMessage *message = [[IMMessage alloc] init];
    message.msgID = ((IMChatRoomMsgLost *)infoIDs[0]).msgLostId;
    message.infoType = inboxName;
    message.featureID = IM_Feature_ChatRoom;
    message.sn = [IMUtil createSN];
    //临时纪录，roomId，range
    message.roomID = roomId;
    message.lostInfoIds = infoIDs;
    message.isWaitResp = NO;

    NSMutableArray * reloadArr = [NSMutableArray array];
    for (IMChatRoomMsgLost* lost in infoIDs) {
        [reloadArr addObject:[NSNumber numberWithInteger:lost.msgLostId]];
    }

    message.requestBody = [self.protoMessage createGetMultiInfosRequest:inboxName InfoIds:reloadArr RoomID:roomId Sn:message.sn];

    IMTask *task = [[IMTask alloc] initTask:message User:self];
    [self addTask:task];

    [self writeLog:[NSString stringWithFormat:@"+get lost cr msg task:%lld",message.sn]];
}

/**
 * 定期发送getInfo请求
 */
- (BOOL) sendRegularGetInfoRequest
{
    BOOL isDone = false;

    const NSString* regularScene = @"check timer";
    NSTimeInterval interval = -([_lastCheckGetInfoTime timeIntervalSinceNow]);
    //5 minutes
    if (interval >= self.configData.hbInterval * 10) {
        
        NSArray *keys;
        @synchronized (self.inboxLastGetInfoTimeDict) {
            keys = [self.inboxLastGetInfoTimeDict allKeys];
        }
        unsigned long count = [keys count];
        for(int i = 0; i < count; ++i) {
            NSString *name = [keys objectAtIndex: i];
            NSDate *lastTime;
            @synchronized (self.inboxLastGetInfoTimeDict) {
                lastTime = [self.inboxLastGetInfoTimeDict objectForKey:name];
            }
        
            if (lastTime == nil ||
                ((-1 * [lastTime timeIntervalSinceNow]) > self.configData.hbInterval * 10)) {
                [self createQueryInboxTask:name In:regularScene];
                //因为取消息请求会被当做心跳，所以后面的心跳包可以忽略掉
                isDone = true;
            }
        }
        _lastCheckGetInfoTime = [NSDate date];
    }
    return isDone;
}

/**
 * 系统网络切换回调函数
 * 0 -- power off
 * 1 -- power on with gprs
 * 2 -- power on with wifi
 */
- (void) reachabilityChanged:(int)status
{

    if(status == 0)
    {
        [self writeLog:@"network power off"];
        [self addStateTask:IM_TaskType_Network_Poweroff];
    }
    else
    {
        if (status == 1)
        {
            [self writeLog:@"network power on wifi"];
        }
        else if (status == 2)
        {
            [self writeLog:@"network power on mobile"];
        }
        [self addStateTask:IM_TaskType_Network_Poweron];

    }
}

/**
 * 回调代理接口
 * @param message:需要告诉代理的对象;
 * @returns void
 */
- (void) notifyDelegateMessage:(IMMessage*)message
{
    NSBlockOperation *notify = [NSBlockOperation blockOperationWithBlock:^{
        if (self.isStop)
            return;
        
        @try
        {
            //上行和下行的回调函数不同
            if (message.IsSend == true)
            {
                if(self.notifyReceiver != nil && [self.notifyReceiver respondsToSelector:@selector(onSendResult:User:)])
                {
                    [self.notifyReceiver onSendResult:message User:self];
                }
                else
                {
//                    CPLog(@"delegate is nil or does not implement onSendResult:User:");
                }
            }
            else
            {
                if (self.configData.appid > 0)
                {
                    if ([message.infoType isEqual:@"peer"] && [self.notifyReceiver respondsToSelector:@selector(onPeerchat:Data:)])
                    {
                        [self.notifyReceiver onPeerchat:message.msgID Data:message.resultBody];
                    }
                    else if ([message.infoType isEqual:@"im"] && [self.notifyReceiver respondsToSelector:@selector(onIMchat:Data:)])
                    {
                        [self.notifyReceiver onIMchat:message.msgID Data:message.resultBody];
                    }
                    else if ([message.infoType isEqual:@"public"] && [self.notifyReceiver respondsToSelector:@selector(onPublic:Data:)])
                    {
                        [self.notifyReceiver onPublic:message.msgID Data:message.resultBody];
                    }
                }
                else
                {
                    if(self.notifyReceiver != nil && [self.notifyReceiver respondsToSelector:@selector(onMessage:User:)])
                    {
                        [self.notifyReceiver onMessage:message User:self];
                    }
                    else
                    {
                        [self writeLog:@"wrong delegate"];
                    }
                }
            }
        }
        @catch (NSException *exception)
        {
//            CPLog(@"notifyDelegateMessage get exception,reason:%@", exception.reason);
        }
    }];
    
    [self.notifyQueue addOperation:notify];
}

/**
 * 群回调代理接口
 * @param :需要告诉代理的对象;
 */
- (void) notifyDelegateGroup:(NSMutableDictionary*)data
{
    NSBlockOperation *notify = [NSBlockOperation blockOperationWithBlock:^{
        if (self.isStop)
            return;
        
        @try
        {
            if(self.notifyReceiver != nil && [self.notifyReceiver respondsToSelector:@selector(onGroup:)])
            {
                if (data != nil)
                {
                    [self.notifyReceiver onGroup:data];
                }
                else
                {
//                    CPLog(@"notifyDelegateGroup get nil data");
                }
            }
            else
            {
//                CPLog(@"notifyDelegateGroup: delegate is nil or does not implement onGroup:");
            }
            
        }
        @catch (NSException *exception)
        {
//            CPLog(@"notifyDelegateGroup get exception,reason:%@", exception.reason);
        }
    }];
    
    [self.notifyQueue addOperation:notify];
}

- (void) notifyDelegateStateChange:(IMUserState) curState From:(IMUserState)from
{
    NSBlockOperation *notify = [NSBlockOperation blockOperationWithBlock:^{
        if (self.isStop)
            return;
        
        if (curState != from)
        {
            if(self.notifyReceiver != nil)
            {
                
                if ([self.notifyReceiver respondsToSelector:@selector(onStateChange:From:User:)])
                {
                    [self.notifyReceiver onStateChange:curState From:from User:self];
                }
                
                if ([self.notifyReceiver respondsToSelector:@selector(onStateChange:)])
                {
                    NSString *message = [@"" stringByAppendingFormat:@"state:%d(%@)->%d(%@)",
                                         from,
                                         [IMUtil getStateName:from],
                                         curState,
                                         [IMUtil getStateName:curState]];
                    [self writeLog:message];
                    [self.notifyReceiver onStateChange:message];
                }
            }
            else
            {
//                CPLog(@"delegate is nil or does not implement onStateChange:From:User:");
            }
        }
    }];
    
    // 状态切换通知拥有较高优先级,需要优先通知客户端
    notify.queuePriority = NSOperationQueuePriorityHigh;
    [self.notifyQueue addOperation:notify];
}

/**
 * 回调delegate，通知上层聊天室数据
 * roomid：聊天室id;
 * userid: 发送消息的用户id;
 * data: 用户发送的消息内容;
 * memcount: 聊天室里的总人数;
 * regcount: 聊天室里的注册用户数;
 */
-(void) notifyDelegateChatroomData:(NSString*)roomid Sender:(NSString*)userid Data:(NSData*)data MemCount:(int)memcount RegCount:(int)regcount
{
    NSBlockOperation *notify = [NSBlockOperation blockOperationWithBlock:^{
        if (self.isStop)
            return;
        
        if ([self.notifyReceiver respondsToSelector:@selector(onChatroomData:Sender:Data:MemCount:RegCount:)])
        {
            [self.notifyReceiver onChatroomData:roomid Sender:userid Data:data MemCount:memcount RegCount:regcount];
        }
        else if ([self.notifyReceiver respondsToSelector:@selector(onChatroomData:Sender:Data:)])
        {
            [self.notifyReceiver onChatroomData:roomid Sender:userid Data:data];
        }
    }];
    
    [self.notifyQueue addOperation:notify];
}

/**
 * 回调delegate，通知上层聊天室数据
 * roomid: 聊天室ID
 * eventType: 1001 -- 加入聊天室, 1002 -- 退出聊天室
 * userid: 成员id，例如eventType为1001时，表示该成员加入了聊天室，为1002时表示该成员退出了聊天室
 * memcount: 聊天室总成员数
 * userdata: 只有eventype为1001时有效，表示加入者的个人信息(来自花椒服务器)
 */
- (void)notifyDelegateChatroom:(NSString*)roomid Change:(int)eventType Member:(NSString*)userid MemCount:(int)memcount withData:(NSData*)userdata
{
    NSBlockOperation *notify = [NSBlockOperation blockOperationWithBlock:^{
        if (self.isStop)
            return;
        
        if ([self.notifyReceiver respondsToSelector:@selector(onChatroom:Change:Member:MemCount:withData:)])
        {
            [self.notifyReceiver onChatroom:roomid Change:eventType Member:userid MemCount:memcount withData:userdata];
        }
        else if ([self.notifyReceiver respondsToSelector:@selector(onChatroom:Change:Member:MemCount:)])
        {
            [self.notifyReceiver onChatroom:roomid Change:eventType Member:userid MemCount:memcount];
        }
    }];
    
    [self.notifyQueue addOperation:notify];
}

/**
 * 回调delegate，通知上层聊天室数据
 * roomid: 聊天室ID
 * eventType: 1001 -- 加入聊天室, 1002 -- 退出聊天室
 * userid: 成员id，例如eventType为1001时，表示该成员加入了聊天室，为1002时表示该成员退出了聊天室
 * memcount: 聊天室总成员数
 * regcount: 聊天室中的注册成员数
 * userdata: 只有eventype为1001时有效，表示加入者的个人信息(来自花椒服务器)
 */
- (void)notifyDelegateChatroom:(NSString*)roomid Change:(int)eventType Member:(NSString*)userid MemCount:(int)memcount RegCount:(int)regcount withData:(NSData*)userdata
{
    NSBlockOperation *notify = [NSBlockOperation blockOperationWithBlock:^{
        if (self.isStop)
            return;
        
        if ([self.notifyReceiver respondsToSelector:@selector(onChatroom:Change:Member:MemCount:RegCount:withData:)])
        {
            [self.notifyReceiver onChatroom:roomid Change:eventType Member:userid MemCount:memcount RegCount:regcount withData:userdata];
        }
        else if ([self.notifyReceiver respondsToSelector:@selector(onChatroom:Change:Member:MemCount:withData:)])
        {
            [self.notifyReceiver onChatroom:roomid Change:eventType Member:userid MemCount:memcount withData:userdata];
        }
        else if ([self.notifyReceiver respondsToSelector:@selector(onChatroom:Change:Member:MemCount:)])
        {
            [self.notifyReceiver onChatroom:roomid Change:eventType Member:userid MemCount:memcount];
        }
    }];
    
    [self.notifyQueue addOperation:notify];
}

/**
 * 回调delegate，通知上层加入，退出，查询聊天室的应答事件
 * eventTyp: 101 -- 查询聊天室，102--加入聊天室， 103--退出聊天室
 * success: YES --成功， NO -- 失败，如果失败，roominfo为nil
 * roominfo: 聊天室详情字典包括如下key(:
 * roomid[NSString]:聊天室id
 * version[NSNumber(longlong):版本号
 * memcount[NSNumber(int)]:成员数量(包括qid用户和非qid用户)
 * regmemcount[NSNumber(int)]:非qid用户数量
 * members[NSArray]:成员的userid
 */
- (void)notifyDelegateChatroomEvent:(int)eventType IsSuccessful:(BOOL)success RoomInfo:(NSDictionary*)roominfo
{
    NSBlockOperation *notify = [NSBlockOperation blockOperationWithBlock:^{
        if (self.isStop)
            return;
        
        [self.notifyReceiver onChatroomEvent:eventType IsSuccessful:success RoomInfo:roominfo];
    }];
    
    [self.notifyQueue addOperation:notify];
}

//消息补偿，每隔5s检查丢失队列是否有值，有值进行补偿拉取
//最多拉取近100条消息，多余消息不做拉取
- (void) checkLostIdTimer:(NSTimer*)theTimer
{
    @synchronized(self)
    {
        NSArray* allKeys = _chatRoomMissedReadIDDict.allKeys;
        NSInteger now = [[NSDate date]timeIntervalSince1970];

        for (NSString* key in allKeys)
        {

            NSDictionary* lastReadDic = [_lastReadIDChatRoomDict objectForKey:key];
            NSInteger lastTimer = [lastReadDic[@"lastTimer"]integerValue];

            if (lastReadDic) {
                //200秒未收到过新的消息，认为是垃圾数据，删除
                if (now - lastTimer > 200 ) {
                    [_chatRoomMissedReadIDDict removeObjectForKey:key];
                    [_lastReadIDChatRoomDict removeObjectForKey:key];
                }
            }
            
            NSMutableArray* lostIdArr = [[_chatRoomMissedReadIDDict objectForKey:key]mutableCopy];
            NSMutableArray* needToReload = [NSMutableArray array];
            
            //一旦超过100条，只拉最新的100条
            for (IMChatRoomMsgLost* lost in lostIdArr)
            {
                //未拉取过
                if (lost.msgReloadTime <= 0) {

                    if (now - lost.msgLostTime > 5.0  && needToReload.count < 100)
                    { //发现丢失超过5.0秒，并且数组长度小于100
                        [needToReload addObject:lost];
                        lost.msgReloadTime = [[NSDate date]timeIntervalSince1970];
                    }
                    else if(needToReload.count >= 100)////如果需要请求的过多，剩下的reloadTime小等于零的，标记为确保拉去过
                    {
                        lost.msgReloadTime = [[NSDate date]timeIntervalSince1970];
                    }
                }
            }

            //防止，空拉取
            if (needToReload.count)
            {
                [needToReload sortUsingComparator:^NSComparisonResult(id  _Nonnull obj1, id  _Nonnull obj2) {
                    return ((IMChatRoomMsgLost *)obj1).msgLostId > ((IMChatRoomMsgLost *)obj2).msgLostId;
                }];
                [self createQueryInboxChatRoom:@"chatroom" InfoIDs:needToReload roomId:key];
            }

        }

        [self cleanAllChatRoomMissedRead];
    }
}

- (BOOL)handleRoomId:(NSString*)roomid
               maxID:(NSNumber *)maxID
               msgID:(NSNumber *)msgID
           timeStamp:(NSNumber *)timeStamp
{
    @synchronized(self) {
        
        BOOL isNeedNotify = YES;
        
        NSInteger msgRead = msgID?[msgID integerValue]:-1;
        NSInteger maxRead = maxID?[maxID integerValue]:-1;
        
        //使用本地时间
        NSInteger now = [[NSDate date]timeIntervalSince1970];
        
        if (roomid) {
            
            //_lostIdCountOfChatRoom登录0，取消过滤重新拉取功能
            // 如果加入聊天室响应中设置了停止补偿拉取，也直接返回
            if (!self.pullChatroomLostMsg || _lostIdCountOfChatRoom <= 0) {
                return YES;
            }
            
            if (maxRead >= 0)
            {
                NSDictionary* lastReadDic = [_lastReadIDChatRoomDict objectForKey:roomid];
                NSInteger lastRead = -1;
                if(lastReadDic){
                    
                    lastRead = [lastReadDic[@"lastRead"]integerValue];
                    NSInteger lastTimer = [lastReadDic[@"lastTimer"]integerValue];
                    
                    if(msgRead <= lastRead && msgRead > 0){
                        NSMutableArray* missedArr = [[_chatRoomMissedReadIDDict objectForKey:roomid]mutableCopy];
                        if(missedArr){
                            IMChatRoomMsgLost* msgLost = [self findMsgLostByMsgID:msgRead inMissArr:missedArr];
                            
                            if (msgLost) {
                                [missedArr removeObject:msgLost];
                                [_chatRoomMissedReadIDDict setObject:missedArr forKey:roomid];
                                [self writeLog:[NSString stringWithFormat:@"rm lost cr msgid(rid:%@|max:%@|msgid:%@|lr:%ld)", roomid,
                                                maxID, msgID, lastRead]];
                            } else {
                                isNeedNotify = NO;
                                [self writeLog:[NSString stringWithFormat:@"give up cr msg(rid:%@|max:%@|msgid:%@|lr:%ld)", roomid,
                                                maxID, msgID, lastRead]];
                            }
                        }
                    }
                    
                    if (now - lastTimer > 120) {
                        [_lastReadIDChatRoomDict removeObjectForKey:roomid];
                        [_chatRoomMissedReadIDDict removeObjectForKey:roomid];
                        [self writeLog:[NSString stringWithFormat:@"rm lastr for 120s(rid:%@|max:%@|msgid:%@)", roomid,
                                        maxID, msgID]];
                    }
                    
                    if (maxRead - lastRead >= 1)
                    {
                        //如果当前是有msgID 的消息，max会增加，maxID 与 lastReadID 之间的差值>1 的部分认为是有消息丢失。
                        //如果当前是没有msgID 的消息，max不会增加。maxID 与 lastReadID 之间差值>=1 的部分认为是有消息丢失。
                        NSInteger maxEndIndex = msgRead > 0 ? maxRead : maxRead + 1;
                        [self addIMChatRoomMsgLostWithBeginIndex:lastRead endIndex:maxEndIndex roomID:roomid];
                    }
                }
                
                maxRead = maxRead > lastRead ? maxRead : lastRead;
                NSDictionary * temp = @{@"lastRead":[NSNumber numberWithInteger:maxRead], @"lastTimer":[NSNumber numberWithInteger:now]};
                [_lastReadIDChatRoomDict setObject:temp forKey:roomid];
                
            } else {
                [self writeLog:[NSString stringWithFormat:@"err msg id (rid:%@|max:%@|msgid:%@)", roomid,
                                maxID, msgID]];
            }
        }
        return isNeedNotify;
    }
}


-(IMChatRoomMsgLost*) findMsgLostByMsgID:(NSInteger) msgID inMissArr:(NSArray*) missedArr{
    
    IMChatRoomMsgLost* msgLost = nil;
    
    for (IMChatRoomMsgLost* temp in missedArr) {
        if (temp.msgLostId == msgID) {
            msgLost = temp;
            break;
        }
    }
    
    return msgLost;
}


-(void) cleanChatRoomMissedReadByRoomID:(NSString*) roomID{
    
    [_lastReadIDChatRoomDict removeObjectForKey:roomID];
    [_chatRoomMissedReadIDDict removeObjectForKey:roomID];
    [self writeLog:[NSString stringWithFormat:@"rm lastr(rid:%@)", roomID]];
    
}

-(void) cleanAllChatRoomMissedRead{
    
    NSArray* allKeys = _chatRoomMissedReadIDDict.allKeys;
    if (allKeys.count > 0) {
        [self writeLog:@"clean lastr"];
    }
    
    for (NSString* roomId in allKeys)
    {
        NSInteger now = [[NSDate date]timeIntervalSince1970];
        NSMutableArray* lostArr = [[_chatRoomMissedReadIDDict objectForKey:roomId]mutableCopy];
        
        //超时删除
        NSMutableArray *oldArr = [NSMutableArray array];
        for (IMChatRoomMsgLost* lost in lostArr) {
            if (now - lost.msgLostTime > 60) {
                [oldArr addObject:lost];
            }
        }
        for (IMChatRoomMsgLost* lost in oldArr) {
            [lostArr removeObject:lost];
        }
        
        [_chatRoomMissedReadIDDict setObject:lostArr forKey:roomId];
    }
    
    
}



- (void)addIMChatRoomMsgLostWithBeginIndex:(NSInteger)begin endIndex:(NSInteger)end roomID:(NSString *)roomid
{
    NSInteger now = [[NSDate date]timeIntervalSince1970];
    NSMutableArray* missedArr = [[_chatRoomMissedReadIDDict objectForKey:roomid]mutableCopy];

    if (missedArr == nil) missedArr = [[NSMutableArray alloc]init];

    if (end - begin <= 200) { //若丢失消息过多将不再拉取
        //按照msgLostTime丢失的顺序，从大到小排序，使得后统计到的丢失消息，优先拉取
        for (NSInteger i = begin+1; i < end; i++) {
            IMChatRoomMsgLost* lost = [[IMChatRoomMsgLost alloc]init];
            lost.msgLostTime = now;
            lost.msgLostId = i;

            //最多纪录500条
            if (missedArr.count <= _lostIdCountOfChatRoom) {
                [missedArr insertObject:lost atIndex:0];
            }
        }
    }

    [_chatRoomMissedReadIDDict setObject:missedArr forKey:roomid];
}

-(void) postLostMsgID:(NSString*)lostSum
{

    //send to http server
    NSString * urlStr = @"";


    urlStr = [urlStr stringByAppendingString:[NSString stringWithFormat:@"&uid=%@",self.configData.jid]];
    urlStr = [urlStr stringByAppendingString:lostSum];

    [self.webFeature postHttp:urlStr requestData:nil];
}


/**
 * 执行任务
 * @param message:放返回结果的message对象
 */
- (void) execTask:(IMMessage*)message
{
    [self writeLog:[NSString stringWithFormat:@"send req:%lld",message.sn]];

    if (self.curState == IM_State_Connected &&
        [self sendRequest:[self.protoMessage createSessionKeyOutDataWithStrData:message.requestBody] Tag:10 Wait:message.isWaitResp WaitSecond:10])
    {
        //重置心跳时间，因为发送消息也会被server视为收到心跳，客户端就无需再发心跳了
        @synchronized(_checkHBAckLock)
        {
            _lastHeartBeatTime = [NSDate date];
            _lastHBAckTime = nil;
        }

        //如果是无需等待应答的，则直接返回
        if (message.isWaitResp == NO)
        {
            return;
        }

        //通知代理
        message.errorCode = IM_Success;
        message.errorReason = @"success";
        int payloadType = [IMUtil getPayloadType:self.serverData];

        [self writeLog:[NSString stringWithFormat:@"get resp:%@，pt:%d",[self.serverData objectForKey:@"sn"], payloadType]];

        switch (payloadType)
        {
            case MSG_ID_RESP_EX1_QUERY_USER_STATUS:
            {
                /*
                 message RespEQ1User {
                 required string userid      = 1;
                 required string user_type   = 2;
                 required int32  status      = 3;           //0:not registry;  1:registry, offline, not reachable; 2:registry, offline, reachable; 3:registry, online, reachable
                 optional string jid         = 4;
                 optional uint32 app_id      = 5;
                 optional string platform    = 6;           // web, pc, mobile
                 optional string mobile_type = 7;       //android, ios
                 optional uint32 client_ver  = 8;
                 }

                 message Ex1QueryUserStatusResp {  //msgid = 200012
                 repeated RespEQ1User user_list = 1;
                 }
                 */
                NSMutableArray* users = [self.serverData objectForKey:@"users"];
                int errorCode =  [[self.serverData objectForKey:@"code"]intValue];
                NSMutableArray* statuses = [NSMutableArray array];

                if (nil != users) {
                    for (IMProtoUserInfo* user in users) {
//                        CPLog(@"Get User: %@\n", user);
                        if ([user.userType isEqualToString:@"phone"] && [self.configData.phone isEqualToString:user.userId]) {
                            self.configData.jid = user.jid;
                        }
                        [statuses addObject:[NSNumber numberWithInt:user.status]];
                    }
                }
                [self.notifyReceiver onPresenceWithSid:message.sessionID sn:message.sn result:errorCode!=0 users:users statuses:statuses];
            }
                break;

            case MSG_ID_RESP_SERVICE_CONTROL:
            {
                if ([[self.serverData objectForKey:@"service_id"]integerValue] == 10000007 )
                {
                    NSString* channel_id = [self.serverData objectForKey:@"channel_id"];
                    NSData* channel_info = [self.serverData objectForKey:@"channel_info"];

                    [self.notifyReceiver onChannelWithSid:message.sessionID sn:message.sn result:channel_info==nil channelId:channel_id channelInfo:channel_info];
                }
                if ([[self.serverData objectForKey:@"service_id"]integerValue] == SERVICE_ID_CHATROOM)
                {
                    NSMutableDictionary *chatroomDict = [self.serverData objectForKey:@"chatroom"];
                    int eventType = [[chatroomDict objectForKey:@"payload"] intValue];
                    int success = [[chatroomDict objectForKey:@"result"] intValue];
                    if (success != 0)
                    {
                        [self notifyDelegateChatroomEvent:eventType IsSuccessful:NO RoomInfo:chatroomDict];
                    }
                    else
                    {
                        id roomInfo = [chatroomDict objectForKey:@"room"];

                        if ([roomInfo objectForKey:@"partnerdata"] != nil)
                        {
                            NSMutableDictionary *tmpDict = [[NSMutableDictionary alloc] initWithDictionary:roomInfo copyItems:YES];

                            NSError *error = nil;
                            id jsonObj = [NSJSONSerialization
                                          JSONObjectWithData:[tmpDict objectForKey:@"partnerdata"]
                                          options:NSJSONReadingAllowFragments
                                          error:&error];
                            if (jsonObj != nil)
                            {
                                [tmpDict setObject:jsonObj forKey:@"partnerdata"];
                            }
                            else
                            {
                                [tmpDict removeObjectForKey:@"partnerdata"];
                            }
                            [self writeLog:[NSString stringWithFormat:@"cr resp>>t:%d,ri:%@", eventType, tmpDict]];
                        }
                        else
                        {
                            [self writeLog:[NSString stringWithFormat:@"cr resp>>t:%d,ri:%@", eventType, roomInfo]];
                        }

                        NSString *curRoomID = [roomInfo objectForKey:@"roomid"];
                        UInt64 longRoomID = [curRoomID longLongValue];
                        UInt64 maxmsgid = [[roomInfo objectForKey:@"maxmsgid"] longLongValue];
                        
                        switch (eventType) {
                            case PAYLOAD_JOIN_CHATROOM:
                            {
                                if (maxmsgid > 0 && longRoomID < 1000000000) {
                                    //缓存roomid对应的maxmsgid
                                    [self handleRoomId:curRoomID maxID:[NSNumber numberWithInteger:maxmsgid] msgID:nil timeStamp:nil];
                                }
                                
                                NSNumber* pl = [chatroomDict objectForKey:@"pull_lost"];
                                if (pl != nil && self.pullChatroomLostMsg != [pl boolValue]) {
                                    [self writeLog:[NSString stringWithFormat:@"pull_lost:%d->%@", self.pullChatroomLostMsg ? 1 : 0, pl]];
                                    self.pullChatroomLostMsg = [pl boolValue];
                                }
                            }
                                break;
                                
                            case PAYLOAD_QUIT_CHATROOM:
                                //删除缓存roomid对应的maxmsgid
                                [self cleanChatRoomMissedReadByRoomID:curRoomID];
                                break;
                                
                            default:
                                break;
                        }

                        [self notifyDelegateChatroomEvent:eventType IsSuccessful:YES RoomInfo:roomInfo];
                    }
                }
                else if ([[self.serverData objectForKey:@"service_id"]integerValue] == 10000013)
                {
                    if(g_openIMLog)
                        [self writeLog:@"execTask >>> handlePrivateNotify aaa >"];
                    [self handlePrivateNotify:self.serverData Inbox:@"privatechat" MsgID:0 SN:[[self.serverData objectForKey:@"sn"] longLongValue]];
                }
                else
                {
//                    CPLog(@" service_id_%ld", (long)[[self.serverData objectForKey:@"service_id"]integerValue]);
                }
            }
                break;

            case MSG_ID_RESP_CHAT:
            {
                message.errorCode = IM_Success;
                int result = [IMUtil getInt32FromDict:self.serverData Key:@"result"];
                if (result != IM_Success)
                {
                    message.errorCode = IM_SendDataFail;
                    message.errorReason = @"发送数据失败";
                }
                message.IsSend = true;

                //通知界面
                [self notifyDelegateMessage:message];

            }
                break;

            default:
//                CPLog(@"not support payload type:%d", payloadType);
                break;
        }
    }
    else
    {
        NSString *roomId = message.roomID;
        NSArray *infoIds = message.lostInfoIds;

        [self writeLog:[NSString stringWithFormat:@"send req:%lld timeout,%@", message.sn, message.infoType]];
        //通知代理
        message.errorCode = IM_OperateTimeout;
        message.errorReason = @"operation timeout";
        [self notifyDelegateMessage:message];
    }
}


/**
 * 发送请求到msg router
 */
- (BOOL) sendRequest:(NSData*)data Tag:(long)tag Wait:(BOOL)isWait WaitSecond:(int)seconds
{
    BOOL result = false;
    @try
    {
        //CPLog(@"call sendRequest, socket isconnected:%d", [self.asyncSocket isConnected]);
        if (self.asyncSocket != nil && [self.asyncSocket isConnected])
        {
            if (isWait) //如果是有应答的
            {
                [self.syncCondition lock];
                self.syncFlag = IM_ThreadState_InWaiting;
                [self.asyncSocket writeData:data withTimeout:5000 tag:tag];
                result = [self.syncCondition waitUntilDate:[NSDate dateWithTimeIntervalSinceNow:seconds]];
                if (self.syncFlag == IM_ThreadState_Error)
                {
                    result = false;
                }
                self.syncFlag = IM_ThreadState_NoWaiting;
                [self.syncCondition unlock];            }
            else
            {
//                CPLog(@"don't need wating");
                [self.asyncSocket writeData:data withTimeout:5000 tag:tag];
                result = TRUE;
            }
        }
    }
    @catch (NSException *exception) {
//        CPLog(@"exception, %@ %@", [exception name], [exception reason]);
    }

    return result;
}


/**
 * 给条件变量设置信号
 */
-(void) setSignal:(IMThreadState)state
{
    [self.syncCondition lock];

    //如果有线程正在等待数据，则设置条件变量为有信号
    if (self.syncFlag == IM_ThreadState_InWaiting)
    {
        self.syncFlag = state;
        [self.syncCondition signal];
    }

    [self.syncCondition unlock];
}

/*
 *建立连接
 */
- (void)socket:(GCDAsyncSocket *)sock didConnectToHost:(NSString *)host port:(uint16_t)port;
{
    [self writeLog:[NSString stringWithFormat:@"socket(%p) connected(%d) to %@ on %d,cur:%@", sock, [sock isConnected], host,port, [IMUtil getStateName:self.curState]]];
    //CPLog(@"self socket:%p, isConnected:%d", self.asyncSocket, [self.asyncSocket isConnected]);
    //因为从msgrouter下来的第一个包的头是6字节:flag(2bytes) + len(4bytes) + Message(protobuf)
    //tag:0表示期望得到第一个msg包，即：initLogin的response
    [sock readDataToLength:6 withTimeout:-1 tag:IM_Tag_InitLogin_Head];

    //将锁设置为有信号
    //[self setSignal];

    IMUserState oldState = self.curState;

    //切换到连接中状态
    self.curState = IM_State_Connecting;

    //通知界面有状态改变
    [self notifyDelegateStateChange:self.curState From:oldState];


    //进入下一个状态
    [self handleCurState];
}

/*
 *读取数据
 */
- (void)socket:(GCDAsyncSocket *)sock didReadData:(NSData *)data withTag:(long)tag;
{
    @try
    {
        //CPLog(@"socket receive data, length:%lu", (unsigned long)[data length]);
        if (tag == IM_Tag_InitLogin_Head) //如果是initLogin的head
        {
            //flag(2bytes) + len(4bytes) + Message(protobuf)
            NSRange range1 = NSMakeRange(0, 2);
            char flag[2];
            [data getBytes:flag range:range1];
            if (flag[0] == 'q' && flag[1] == 'h')
            {
                NSRange range = NSMakeRange(2, 4);
                int dataLength = 0;
                [data getBytes:&dataLength range:range];
                dataLength = htonl(dataLength);
                if (dataLength > 0)
                {
                    [sock readDataToLength:(dataLength-6) withTimeout:-1 tag:IM_Tag_InitLogin_Body];
                }
                else
                {
                    [self writeLog:@"invalid initLogin resp length"];
                    [sock disconnect];
                    
                    // 后面这两行代码看起来是不必要的, [sock disconnect]会触发socketDidDisconnect回调函数
                    // 该函数会呼叫[self addStateTask:IM_TaskType_Socket_Disconnect]，同时对于disconnect事件的处理会导致
                    // IMUser对象持有的socket对象发生变化，因此sock对象实际已经被抛弃了
                    // longjun 2016.08.28
                    [self addStateTask:IM_TaskType_Socket_Disconnect];
                    //重头开始
                    [sock readDataToLength:6 withTimeout:-1 tag:IM_Tag_InitLogin_Head];
                }
            }
            else
            {
                //it must have error
                [sock disconnect];
                [self writeLog:@"socket error"];
                [self addStateTask:IM_TaskType_Socket_Disconnect];
                [self setSignal:IM_ThreadState_Error];
            }
        }
        else if (tag == IM_Tag_InitLogin_Body)
        {
            NSMutableDictionary *dataDict = [self.protoMessage parseInitLoginResponse:data];
            if ([IMUtil hasSuccessFlag:dataDict])
            {
                [self setSignal:IM_ThreadState_HasData];
                [sock readDataToLength:4 withTimeout:-1 tag:IM_Tag_Login_Head];
            }
            else
            {
                //重头开始
                [sock readDataToLength:6 withTimeout:-1 tag:IM_Tag_InitLogin_Head];
            }
        }
        else if (tag == IM_Tag_Login_Head)
        {
            int dataLength = 0;
            [data getBytes:&dataLength length:4];
            dataLength = htonl(dataLength);
            if (dataLength > 0) //normal data
            {
                [sock readDataToLength:(dataLength-4) withTimeout:-1 tag:IM_Tag_Login_Body];
            }
            else
            {
                //CPLog(@"data length(%d) < 0", dataLength);
                //重头开始
                [sock readDataToLength:6 withTimeout:-1 tag:IM_Tag_InitLogin_Head];
            }
        }
        else if (tag == IM_Tag_Login_Body)
        {
            NSMutableDictionary *dataDict = [self.protoMessage parseLoginResponse:data];
            if ([IMUtil hasSuccessFlag:dataDict])
            {

                NSNumber *timestamp =  [dataDict objectForKey:@"timestamp"];

                local2 = [[NSDate date]timeIntervalSince1970];

                if (nil != timestamp){
                    int serverTime = [timestamp intValue];
                    self.lastLoginedTimestamp = (local2 - local1)/2 + serverTime;
                    [self writeLog:[NSString stringWithFormat:@"logined at:%ul", self.lastLoginedTimestamp]];

                }else{
//                    CPLog(@"----get serverTime failed!!!");
                }

                [self setSignal:IM_ThreadState_HasData];

                [sock readDataToLength:4 withTimeout:-1 tag:IM_Tag_Data_Head];
            }
            else
            {
                //说明登录失败
                int errorID = 0;

                if ([dataDict.allKeys containsObject:@"error"]) {
                    NSNumber* error = [dataDict objectForKey:@"error"];
                    errorID = [error intValue];
                }
                /*
                 用户名或者密码错误： -define(ERROR_USER_INVALID,       1008).

                 数据库太忙, 客户端需要等待;  超过负载后的登录最短间隔时间 5 分钟 , 超过负载后的登录最长间隔时间 10 分钟
                 -define(ERROR_DB_TOO_BUSY, 1012).
                 */
                if (errorID == 1008 )
                {
                    [self writeLog:@"login failed, uid or pwd incorrect"];
                    //说明用户名密码错误
                    //                    self.curState = IM_State_Login_Fail;
                    self.curState = IM_State_Disconnected;
                    self.loginFailReason = errorID;

                    //只为通知上层，不加入到重连逻辑
                    [self notifyDelegateStateChange:IM_State_Login_Fail From:IM_State_Connecting];

                }else{
                    if (errorID == 1002)
                    {
                        [self writeLog:@"login failed, server busy"];
                    }
                    else
                    {
                        [self writeLog:[NSString stringWithFormat:@"login failed, error:%d", errorID]];
                    }
                    self.loginFailReason = errorID;
                    self.curState = IM_State_Disconnected;
                }
                [self setSignal:IM_ThreadState_HasData];

                //重头开始
                [sock readDataToLength:6 withTimeout:-1 tag:IM_Tag_InitLogin_Head];
            }
        }
        else if (tag == IM_Tag_Data_Head)
        {
            int dataLength = 0;
            [data getBytes:&dataLength length:4];
            dataLength = htonl(dataLength);
            if (dataLength == 0) //heart beat
            {
                [self writeLog:@"receive hb ack"];
                @synchronized(_checkHBAckLock)
                {
                    _lastHBAckTime = [NSDate date];
                }
                [sock readDataToLength:4 withTimeout:-1 tag:IM_Tag_Data_Head];
            }
            else if (dataLength > 0) //normal data
            {
                [sock readDataToLength:(dataLength-4) withTimeout:-1 tag:IM_Tag_Data_Body];
            }
            else
            {
                //CPLog(@"data length(%d) < 0", dataLength);
                [sock readDataToLength:4 withTimeout:-1 tag:IM_Tag_Data_Head];
            }
        }
        else if (tag == IM_Tag_Data_Body)
        {
            //CPLog(@"data length:%lu", (unsigned long)[data length]);
            //self.serverData = [self.protoMessage parseServerData:data];
            NSMutableDictionary *recvDataDict = [self.protoMessage parseServerData:data];
            
            if(g_openIMLog){
                [self writeSocketData:recvDataDict];
            }
            
            if ([IMUtil hasSuccessFlag:recvDataDict])
            {
                [self handleServerData:recvDataDict];
            }
            else
            {
                [self writeLog:[NSString stringWithFormat:@"incorrect data, sn:%@", [recvDataDict objectForKey:@"sn"]]];
            }
            [sock readDataToLength:4 withTimeout:-1 tag:IM_Tag_Data_Head];
        }
        else
        {
            [self writeLog:[NSString stringWithFormat:@"unknown socket tag:%ld", tag]];
            [sock readDataToLength:4 withTimeout:-1 tag:IM_Tag_Data_Head];
        }
    }
    @catch (NSException *exception) {
//        CPLog(@"receive exception, %@, %@", [exception name], [exception reason]);
    }

}

-(void) writeSocketData:(NSMutableDictionary *)recvDataDict
{
    NSDictionary *chatRoom = _To_Dict([recvDataDict objectForKey:@"chatroom"]);
    NSDictionary *msgBody = @"";
    if(chatRoom && [chatRoom objectForKey:@"msgbody"]){
        msgBody = _To_Dict([NSJSONSerialization
                        JSONObjectWithData:[chatRoom objectForKey:@"msgbody"]
                        options:NSJSONReadingAllowFragments
                        error:nil]);
    }
    NSDictionary *chat_body = @"";
    NSDictionary* infos = _To_Dict([recvDataDict objectForKey:@"infos"]);
    if(infos && [infos objectForKey:@"chat_body"]){
        chat_body = _To_Dict([NSJSONSerialization
                          JSONObjectWithData:[infos objectForKey:@"chat_body"]
                          options:NSJSONReadingAllowFragments
                          error:nil]);
    }
    NSDictionary *partnerdata = @"";
    NSDictionary* room = _To_Dict([chatRoom objectForKey:@"room"]);
    if(chatRoom && room && [room objectForKey:@"partnerdata"])
    {
        partnerdata = _To_Dict([NSJSONSerialization
                            JSONObjectWithData:[room objectForKey:@"partnerdata"]
                            options:NSJSONReadingAllowFragments
                            error:nil]);
    }
    [self writeLog:[NSString stringWithFormat:@"????? recv Data Dict = %@ \r\n magBody = %@ \r\n chat_body = %@ \r\n partnerdata = %@ ", recvDataDict , msgBody , chat_body ,partnerdata]];
}








/*
 *遇到错误时关闭连接
 */
- (void)socketDidDisconnect:(GCDAsyncSocket *)sock withError:(NSError *)err;
{
    [self writeLog:[NSString stringWithFormat:@"socket(%p) disconnect: er:%@, cur:%@",sock, [err localizedFailureReason], [IMUtil getStateName:self.curState]]];

    if (self.curState == IM_State_Login_Fail) //说明密码错误
    {

    }
    else
    {
        BOOL needPostFailed = NO;
        NSString* loginServer;
        NSString* error = [err localizedFailureReason];
        // 确保连接失败后能够马上重试
        @synchronized(self.lastTryConnectTime)
        {
            if (sock == self.asyncSocket && (self.curState == IM_State_Init || self.curState == IM_State_Disconnected)) {
                self.lastTryConnectTime = [NSDate dateWithTimeIntervalSinceNow:-30];
                if (err != nil && [error containsString:@"connect"]) { // 这种情况下是连接失败
                    needPostFailed = YES;
                    loginServer = [NSString stringWithFormat:@"%@:%d", self.configData.loginServer, self.configData.loginPort];
                }
            }
        }
        
        if (needPostFailed){
            [self postUser:self.configData.jid FailedConnectTo:loginServer With:self.configData.deviceToken Network:[[IMServiceLib sharedInstance] networkType] For:error];
        }
        
        [self addStateTask:IM_TaskType_Socket_Disconnect];
        [self setSignal:IM_ThreadState_Error];
    }
}

/**
 * 用户连接失败打点函数
 * @param uid: 用户id
 * @param host: 连接的网址或IP地址加端口
 * @param deviceid: 设备id
 * @param type: 网络类型
 * @param reason: 失败的原因
 */
-(void) postUser:(NSString*) uid FailedConnectTo:(NSString*) host With:(NSString*)deviceid Network:(int)type For:(NSString*)reason
{
    /**
     线上打点: http://s.360.cn/huajiao/linkerr.html?ip=%s&rip=%s&net=%s&uid=%s&did=%s&plf=%s&r=%s
     */
    //send to http server
    NSString * urlStr = [NSString stringWithFormat:@"http://s.360.cn/huajiao/linkerr.html?ip=%@&rip=&net=%d&uid=%@&did=%@&plf=ios&r=%@", [host stringByAddingPercentEscapesUsingEncoding:NSUTF8StringEncoding], type, uid, [deviceid stringByAddingPercentEscapesUsingEncoding:NSUTF8StringEncoding], [reason stringByAddingPercentEscapesUsingEncoding:NSUTF8StringEncoding]];
    
    IMMessage* resp = [self.webFeature getHttp:urlStr timeout:1];
    [self writeLog:[NSString stringWithFormat:@"post data to url:%@, er:%d", urlStr, resp.errorCode]];
}

/*
 * 各种原因造成的断开socket连接
 */
- (void)socketDidCloseReadStream:(GCDAsyncSocket *)sock;
{
    [self writeLog:[NSString stringWithFormat:@"onSocketDidDisconnect:%p, %@",sock, [IMUtil getStateName:self.curState]]];
    [self addStateTask:IM_TaskType_Socket_Disconnect];
    [self setSignal:IM_ThreadState_Error];
}

/**
 * 发送数据完成后的回调方法
 */
- (void)socket:(GCDAsyncSocket *)sock didWriteDataWithTag:(long)tag;
{
    //CPLog(@"write data success, tag:%ld", tag);
}

- (void) handle_chatroom_send_msg:(NSDictionary *)roomDict SN:(int64_t)sendersn
{
    int memCount = 0;
    NSNumber *num = [roomDict objectForKey:@"memcount"];
    if (num != nil)
    {
        memCount = [num intValue];
    }
    int regCount = 0;
    num = [roomDict objectForKey:@"regmemcount"];
    if (num != nil)
    {
        regCount = [num intValue];
    }

    //纪录接受的消息msgid
    NSNumber* msgid = [roomDict objectForKey:@"msgid"];
    NSNumber* maxid= [roomDict objectForKey:@"maxid"];
    NSString* roomId = [roomDict objectForKey:@"roomid"];
    
    if ([roomDict objectForKey:@"msgbody"] != nil)
    {
        NSError *error = nil;
        NSDictionary *jsonObj = _To_Dict([NSJSONSerialization   //tg:20170111:收到msgbody是个数字0，导致崩溃
                                      JSONObjectWithData:[roomDict objectForKey:@"msgbody"]
                                      options:NSJSONReadingAllowFragments
                                      error:&error]);
        NSString *strType = @"";
        if ([jsonObj objectForKey:@"type"] != nil)
        {
            strType = [jsonObj objectForKey:@"type"];
        }
        NSString *strText = @"";
        if ([jsonObj objectForKey:@"text"] != nil)
        {
            strText = [jsonObj objectForKey:@"text"];
        }
        NSString *strTraceID = @"";
        if ([jsonObj objectForKey:@"traceid"] != nil)
        {
            strTraceID = [jsonObj objectForKey:@"traceid"];
        }
    }
    else
    {
        [self writeLog:[NSString stringWithFormat:@"cr msg>>sn:%lld|rid:%@|s:%@|c:%d|mid:%@|max:%@", sendersn, [roomDict objectForKey:@"roomid"],
                        [roomDict objectForKey:@"sender"], memCount, msgid, maxid]];
    }
    
    if ([self handleRoomId:roomId maxID:maxid msgID:msgid timeStamp:nil])
    {
        [self notifyDelegateChatroomData:[roomDict objectForKey:@"roomid"] Sender:[roomDict objectForKey:@"sender"] Data:[roomDict objectForKey:@"msgbody"] MemCount:memCount RegCount:regCount];
    }
    else
    {
        [self writeLog:[NSString stringWithFormat:@"giveup cr msg(sn:%lld|rid:%@|s:%@|max:%@|msgid:%@)", sendersn, [roomDict objectForKey:@"roomid"],
                        [roomDict objectForKey:@"sender"], maxid, msgid]];
    }
}

- (void) handle_chatroom_join_notify:(NSDictionary *)roomDict SN:(int64_t)sendersn
{
    NSArray *members = [roomDict objectForKey:@"members"];
    if ([members count] > 0)
    {
        int memCount = 0;
        NSNumber *num = [roomDict objectForKey:@"memcount"];
        if (num != nil)
        {
            memCount = [num intValue];
        }

        int regCount = 0;
        num = [roomDict objectForKey:@"regmemcount"];
        if (num != nil)
        {
            regCount = [num intValue];
        }

        NSDictionary *udataDict = [roomDict objectForKey:@"udatadict"];
        NSString *joinUserID = [members objectAtIndex:0];

        [self writeLog:[NSString stringWithFormat:@"+cr>>sn:%lld|rid:%@|join:%@|c:%d", sendersn, [roomDict objectForKey:@"roomid"],
                        joinUserID, memCount]];

        if (![joinUserID isEqualToString:self.configData.jid]) //如果是自己，需要过滤
        {
            NSData *userData = [udataDict objectForKey:joinUserID];
            [self notifyDelegateChatroom:[roomDict objectForKey:@"roomid"] Change:1001 Member:joinUserID MemCount:memCount RegCount:regCount withData:userData];
        }
    }
}

- (void) handle_chatroom_quit_notify:(NSDictionary*)roomDict SN:(int64_t)sendersn
{
    NSArray *members = [roomDict objectForKey:@"members"];
    if ([members count] > 0)
    {
        int memCount = 0;
        NSNumber *num = [roomDict objectForKey:@"memcount"];
        if (num != nil)
        {
            memCount = [num intValue];
        }

        NSString *quitUserID = [members objectAtIndex:0];

        [self writeLog:[NSString stringWithFormat:@"-cr>>sn:%lld|rid:%@|quit:%@|c:%d", sendersn, [roomDict objectForKey:@"roomid"],
                        quitUserID, memCount]];

        if (![quitUserID isEqualToString:self.configData.jid]) //如果是自己，需要过滤
        {
            [self notifyDelegateChatroom:[roomDict objectForKey:@"roomid"] Change:1002 Member:quitUserID MemCount:memCount withData:nil];
        }
    }
}

-(void) handlePrivateNotify:(NSMutableDictionary*)dataDict Inbox:(NSString*)inboxName MsgID:(int64_t)msgID SN:(int64_t)msgSN
{
    NSDictionary *privateDict = [dataDict objectForKey:@"privatechat"];
    if ([[privateDict objectForKey:@"payload"] intValue] == 10001)
    {
        int64_t lastReadID = [self getLastMsgID:inboxName];

        NSDictionary *msgDict = [privateDict objectForKey:@"msg"];
        id msgidObj = [msgDict objectForKey:@"msgid"];
        NSData *dataObj = [msgDict objectForKey:@"data"];
        int type = [[msgDict objectForKey:@"type"] intValue];
        NSString *srcID = [msgDict objectForKey:@"srcid"];
        int64_t sendTime = [[msgDict objectForKey:@"sendtime"] longLongValue];


        if (msgidObj != nil) //说明是没有保存的消息
        {
            msgID = [msgidObj longLongValue];

            if (msgID > lastReadID+1)
            {
                //发起get private chat msg
                [self.privatechatFeature getMsg:lastReadID+1 Count:10];
            }
            else
            {
                if(g_openIMLog)
                    [self writeLog:[NSString stringWithFormat:@"handlePrivateNotify 111 data dict %@ , inbox = %@ mssID = %lld" , dataDict , inboxName , msgID]];
                [self updateLastReadID:inboxName Value:msgID Flush:YES];
            }
        }

        if ([self.notifyReceiver respondsToSelector:@selector(onPrivateChat:SendTime:Msgid:Type:Data:SN:)])
        {
            [self.notifyReceiver onPrivateChat:srcID SendTime:sendTime Msgid:msgID Type:type Data:dataObj SN:msgSN];
        }
        else
        {
            [self writeLog:[NSString stringWithFormat:@"new pc msg notify, wrong delegate, (%@|%lld|%lld)", inboxName, msgID, msgSN]];
        }
    }
    else if ([[privateDict objectForKey:@"payload"] intValue] == 1001)
    {
        int64_t msgID = [[privateDict objectForKey:@"msgid"] longLongValue];
        int errCode = [[privateDict objectForKey:@"code"] intValue];
        NSString *reason = [privateDict objectForKey:@"reason"];
        if ([self.notifyReceiver respondsToSelector:@selector(onPrivateChat:Msgid:Code:Reason:)])
        {
            [self.notifyReceiver onPrivateChat:msgSN Msgid:msgID Code:errCode Reason:reason];
        }
    }
    else if([[privateDict objectForKey:@"payload"] intValue] == 1002)
    {
        int64_t maxMsgID = [[privateDict objectForKey:@"maxmsgid"] longLongValue];
        NSArray *msgList = [privateDict objectForKey:@"msglist"];
        BOOL hasCallback = false;
        if ([self.notifyReceiver respondsToSelector:@selector(onPrivateChat:SendTime:Msgid:Type:Data:SN:)])
        {
            hasCallback = true;
        }
        int64_t thisMaxMsgID = 0;

        for (int i=0; i<msgList.count; ++i)
        {
            NSDictionary *msgDict = [msgList objectAtIndex:i];

            NSData *dataObj = [msgDict objectForKey:@"data"];
            int type = [[msgDict objectForKey:@"type"] intValue];
            NSString *srcID = [msgDict objectForKey:@"srcid"];
            int64_t sendTime = [[msgDict objectForKey:@"sendtime"] longLongValue];
            if ([msgDict objectForKey:@"msgid"] != nil)
            {
                msgID = [[msgDict objectForKey:@"msgid"] longLongValue];
            }
            else
            {
                msgID = 0;
            }
            if (msgID > thisMaxMsgID)
            {
                thisMaxMsgID = msgID;
            }

            if (hasCallback)
            {
                [self.notifyReceiver onPrivateChat:srcID SendTime:sendTime Msgid:msgID Type:type Data:dataObj SN:0];
            }
        }
        
        if(g_openIMLog)
            [self writeLog:[NSString stringWithFormat:@"handlePrivateNotify 222 data dict %@ , inbox = %@ mssID = %lld" , dataDict , inboxName , msgID]];
        [self updateLastReadID:inboxName Value:thisMaxMsgID Flush:YES];
        if (thisMaxMsgID>0 && maxMsgID > thisMaxMsgID)
        {
            //发起get private chat msg
            [self.privatechatFeature getMsg:thisMaxMsgID+1 Count:10];
        }
    }
}

/**
 * 处理服务器发来的数据
 * @param dataDict: 解析后的数据字典
 */
- (void) handleServerData:(NSMutableDictionary*)dataDict
{
    int msgID = [IMUtil getPayloadType:dataDict];
    int64_t netSn = [IMUtil getInt64FromDict:dataDict Key:@"sn"];
    
    if (msgID != MSG_ID_NTF_NEW_MESSAGE) {
        [self writeLog:[NSString stringWithFormat:@"down,pt:%d,sn:%lld", msgID, netSn]];
    }

    switch (msgID)
    {
        case MSG_ID_NTF_NEW_MESSAGE: //NewMessageNotify,300000
        {
            int64_t msgSN = [IMUtil getInt64FromDict:dataDict Key:@"sn"];
            int64_t msgid = [IMUtil getInt64FromDict:dataDict Key:@"info_id"];
            NSString *inboxName = [IMUtil getStringFromDict:dataDict Key:@"info_type"];
            int64_t lastReadID = [self getLastMsgID:inboxName];
            BOOL isIgnore = false;
            if (msgid > 0 && lastReadID >= msgid)
            {
                [self writeLog:[NSString stringWithFormat:@"ignore new msg notify(%@|lastr:%lld|msgid:%lld|sn:%lld)", inboxName, lastReadID, msgid, msgSN]];
                return;
            }
            if (!isIgnore)
            {
                if ([inboxName isEqualToString:@"chatroom"]) //聊天室消息是直接放在通知里的
                {
                    
                    NSDictionary *chatroomDict = [dataDict objectForKey:@"chatroom"];
                    if ([[chatroomDict objectForKey:@"payload"] intValue] == 1000) //chatroom message
                    {
                        //上线时，需要将这个过滤自己发消息的逻辑加上，现在是为了上界面显示出来
                        //if (![self.configData.jid isEqualToString:[chatroomDict objectForKey:@"sender"]])
                        {
                            [self handle_chatroom_send_msg:chatroomDict SN:msgSN];
                        }
                    }
                    else if ([[chatroomDict objectForKey:@"payload"] intValue] == 1001) //join chatroom message
                    {
                        NSDictionary *roomDict = [chatroomDict objectForKey:@"room"];
                        [self handle_chatroom_join_notify:roomDict SN:msgSN];
                    }
                    else if ([[chatroomDict objectForKey:@"payload"] intValue] == 1002) //quit chatroom message
                    {
                        NSDictionary *roomDict = [chatroomDict objectForKey:@"room"];
                        [self handle_chatroom_quit_notify:roomDict SN:msgSN];
                    }
                    else if ([[chatroomDict objectForKey:@"payload"] intValue] == 1003) //quit chatroom message
                    {
                        NSArray *notifies = [chatroomDict objectForKey:@"roomlist"];
                        for (int i=0; i<[notifies count]; ++i)
                        {
                            NSDictionary *roomDict = notifies[i];
                            int type = [[roomDict objectForKey:@"payload" ] intValue];
                            if (type == 1000) //chatroom message
                            {
                                [self handle_chatroom_send_msg:roomDict SN:msgSN];
                            }
                            else if (type == 1001)
                            {
                                [self handle_chatroom_join_notify:roomDict SN:msgSN];
                            }
                            else if (type == 1002)
                            {
                                [self handle_chatroom_quit_notify:roomDict SN:msgSN];
                            }
                        }
                    }
                }
                else if ([inboxName isEqual: @"group"])
                {
                    //  群相关通知直接给上层appid
                    NSMutableDictionary* group = [dataDict objectForKey:@"group"];
                    [self notifyDelegateGroup:group];
                    [self writeLog:[NSString stringWithFormat:@"new gp msg notify"]];
                }
                else if ([inboxName isEqual: INFO_TYPE_PEER] || [inboxName isEqual:INFO_TYPE_IM] || [inboxName isEqualToString:INFO_TYPE_PUBLIC])
                {
                    const NSString* notifyScene = @"msg notify";
                    [self createQueryInboxTask:inboxName In:notifyScene];
                }
                else if ([inboxName isEqualToString:@"privatechat"]) //私聊的通知消息是带内容的
                {
                    if(g_openIMLog)
                        [self writeLog:@"handleServerData >>>> handlePrivateNotify"];
                    
                    [self handlePrivateNotify:dataDict Inbox:inboxName MsgID:msgid SN:msgSN];
                }
            }
            else
            {
                [self writeLog:[NSString stringWithFormat:@"ignore new msg notify(%@|lastr:%lld|msgid:%lld|sn:%lld)", inboxName, lastReadID, msgid, msgSN]];
            }
        }
            break;

        case MSG_ID_NTF_RELOGIN:
        {
            //CPLog(@"receive relogin data, id:%d", msgID);
            IMUserState oldState = self.curState;
            self.curState = IM_State_Relogin_Need;
            //            [self stopService];
            [self notifyDelegateStateChange:self.curState From:oldState];
            self.reconnect = YES;
            self.curState = IM_State_Disconnected;
        }
            break;
            
        case MSG_ID_NTF_RECONNECT:
        {
            NSNumber* port = [dataDict objectForKey:@"port"];
            NSMutableArray* ips = [dataDict objectForKey:@"ips"];
            
            if (port != nil && ips != nil) {
                [self.configData resetServerIP];
                for (NSString* host in ips) {
                    [self.configData addPreferredHost:host Port:port];
                }
                [self writeLog:[NSString stringWithFormat:@"need reconnect to [%@] on %@", ips, port]];
            }
            
            self.reconnect = YES;
            IMUserState oldState = self.curState;
            @synchronized(self.lastTryConnectTime) {
                // 确保能够马上重试
                self.lastTryConnectTime = [NSDate dateWithTimeIntervalSinceNow:-30];
            }
            self.curState = IM_State_Disconnected;
            [self notifyDelegateStateChange:self.curState From:oldState];
            [self handleCurState];
        }

        case MSG_ID_RESP_GET_INFO: //GetInfoResp
        case MSG_ID_RESP_GET_MULTI_INFOS:
        {
            NSMutableArray *infoList = [dataDict objectForKey:@"infos"];
            NSString *inboxName = [dataDict objectForKey:@"info_type"];
            int64_t lastReadID = 0;
            int64_t maxReadMsgID = 0;
            if (inboxName != nil && [inboxName length]>0 )
            {
                lastReadID = [self getLastMsgID:inboxName];
            }
            
            unsigned long infoCount = [infoList count];
            
            for (int i=0; i< infoCount; ++i)
            {
                IMMessage *message = [[IMMessage alloc] init];
                message.infoType = inboxName;
                NSMutableDictionary *oneMsg = [infoList objectAtIndex:i];
                /*
                 key options:
                 "info_id": msg id;
                 "chat_body": chat content;
                 "time_sent": send timestamp;
                 "msg_type": only weimi has this field(peer里的100表示字符串)
                 */
                int64_t curID = [IMUtil getInt64FromDict:oneMsg Key:@"info_id"];
                if (i == infoCount-1)
                {
                    maxReadMsgID = curID;
                }
                
                id msgType = [NSNumber numberWithInt:0];
                if ([oneMsg objectForKey:@"msg_type"] != nil)
                {
                    msgType = [oneMsg objectForKey:@"msg_type"];
                }
                int typeFlag = [(NSNumber*)msgType intValue];
                if (typeFlag == 200) //请求上传日志的命令
                {
                    /* // 目前客户端不使用SDK内置的日志上传机制，客户端会重定向SDK日志输出合并到客户端日志中，因此注销此特殊消息的响应
                    NSString *senderID = [self.protoMessage parseSenderID:[oneMsg objectForKey:@"chat_body"]];
                    if (senderID != nil)
                    {
                        CPLog(@"receive command from sender:%@", senderID);
                        [self sendBackLog:senderID];
                    }
                     */
                }
                else
                {
                    if ([inboxName isEqualToString:@"chatroom"])
                    {
                        if ([[oneMsg objectForKey:@"msg_valid"] integerValue] != 0)
                        {
                            NSDictionary* newMsgDic = [oneMsg objectForKey:@"chatroomnewmsg"];
                            
                            NSString* roomid = [newMsgDic objectForKey:@"roomid"];
                            NSNumber* msgID = [newMsgDic objectForKey:@"msgid"];
                            NSNumber* maxID = [newMsgDic objectForKey:@"maxid"];
                            NSString* sender = [newMsgDic objectForKey:@"sender"];

                            //历史旧消息，置为0
                            int memcount = 0;
                            int regmemcount = 0;
                            
                            NSData* chat_body = [newMsgDic objectForKey:@"msgbody"];
                            
                            if ([self handleRoomId:roomid maxID:maxID msgID:msgID timeStamp:nil]) {
                                [self notifyDelegateChatroomData:roomid Sender:sender Data:chat_body MemCount:memcount RegCount:regmemcount];
                            }
                        }
                    }
                    else
                    {
                        if (msgType != nil)
                        {
                            message.resultBody = [oneMsg objectForKey:@"chat_body"];
                            message.msgID = curID;
                        }else{
                            message.resultBody = nil;
                        }
                        
                        //调用代理, 之前必须去重
                        if (curID > lastReadID)
                        {
                            [self writeLog:[NSString stringWithFormat:@"notify %@:%lld", inboxName, curID]];
                            [self notifyDelegateMessage:message];
                        }
                        else
                        {
                            [self writeLog:[NSString stringWithFormat:@"got %@:%lld but lastr:%lld", inboxName, curID, lastReadID]];
                        }
                        
                        if (maxReadMsgID > 0)
                        {
                            //更新最大已读消息id
                            if(g_openIMLog)
                                [self writeLog:[NSString stringWithFormat:@"handleServerData: 111 >> inboxName = %@ , maxRedMsgID = %lld " , inboxName , maxReadMsgID]];
                            [self updateLastReadID:inboxName Value:maxReadMsgID Flush:true];
                        }
                    }
                    
                }
            }
            
            if (inboxName != nil && [inboxName isEqualToString:@"chatroom"] == NO)
            {
                // 通知任务完成
                [self finishQueryInboxTask:inboxName Sn:netSn];
                
                const NSString* responseScene = @"more";
                const int64_t ONLY_GET_LATEST = 1000;
                //如果还有更多消息，则发起新的请求任务
                int64_t lastInfoID = [IMUtil getInt64FromDict:dataDict Key:@"last_info_id"];
                if (infoCount > 0 && lastInfoID != -1)
                {
                    if (maxReadMsgID < lastInfoID) //说明还有没读干净的消息
                    {
                        // peer盒子可能大量未读消息导致影响其他操作，因此只拉取最近的部分消息，其他丢弃
                        if (self.abandonPeer && [inboxName isEqual: INFO_TYPE_PEER] && (lastInfoID - maxReadMsgID) > ONLY_GET_LATEST) {
                            self.abandonPeer = NO;
                            [self writeLog:[NSString stringWithFormat:@"peer msgs[%lld,%lld] is gived up for more unreads",
                                            (maxReadMsgID + 1), (lastInfoID - ONLY_GET_LATEST)]];
                            maxReadMsgID = lastInfoID - ONLY_GET_LATEST;
                            //更新最大已读消息id
                            
                            if(g_openIMLog)
                                [self writeLog:[NSString stringWithFormat:@"handleServerData: 222 >> inboxName = %@ , maxRedMsgID = %lld " , inboxName , maxReadMsgID]];
                            
                            [self updateLastReadID:inboxName Value:maxReadMsgID Flush:true];
                        }
                        
                        //更新希望读取的消息id
                        [self.serverData setObject:[NSNumber numberWithLong:maxReadMsgID+1] forKey:@"info_id"];
                        
                        [self createQueryInboxTask:inboxName In:responseScene];
                    }
                }
            }
        }
            break;

        case MSG_ID_RESP_SERVICE_CONTROL:
        {
            if ([[dataDict objectForKey:@"service_id"]integerValue] == SERVICE_ID_CHATROOM)
            {
                NSMutableDictionary *chatroomDict = [dataDict objectForKey:@"chatroom"];
                int eventType = [[chatroomDict objectForKey:@"payload"] intValue];
                
                if (eventType == PAYLOAD_SUBSCRIBE_CHATROOM) {
                    int success = [[chatroomDict objectForKey:@"result"] intValue];
                    if (success != 0)
                    {
                        [self notifyDelegateChatroomEvent:eventType IsSuccessful:NO RoomInfo:chatroomDict];
                    }
                    else
                    {
                        [self notifyDelegateChatroomEvent:eventType IsSuccessful:YES RoomInfo:chatroomDict];
                    }
                    break;
                }// else will fall through, then goto default, then exeTask will handle join/quit response
            }
            else if ([[dataDict objectForKey:@"service_id"]integerValue] == SERVICE_ID_GROUP)
            {
                NSMutableDictionary* group = [dataDict objectForKey:@"group"];
                // 通知group feature已经收到group请求的response
                int64_t sn = [IMUtil getInt64FromDict:dataDict Key:@"sn"];
                NSNumber* sleep = [group objectForKey:@"sendNextAfter"];
                if (sleep == nil || [sleep intValue] < 0)
                {
                    sleep = [NSNumber numberWithInt:0];
                }

                NSSet* reqSn = [self.groupChatFeature handleRequestReceipt:sn nextAfter:[sleep intValue]];
                if (reqSn != nil)
                {
                    [group setObject: reqSn forKey:@"sn"];
                    [self notifyDelegateGroup:group];
                }
                else
                {
                    // else will abandon the server data for cannot find out correspond client request trace id (might be removed for time out)
                    [self writeLog:[NSString stringWithFormat:@"sn:%lld,[%@] is gived up for tid missed.", sn, group]];
                }

                break;
            } // else will fall through, then goto default
        }

        default:
            self.serverData = dataDict;
            [self setSignal:IM_ThreadState_HasData];
            break;
    }
}

/**
 * 创建取消息盒子的任务
 * @param name: 消息盒子类型：peer, im, public
 * @param scene: 呼叫函数的场景
 */
- (void) createQueryInboxTask:(NSString*)name In:(NSString*)scene
{
    NSBlockOperation *operation = [NSBlockOperation blockOperationWithBlock:^{
        NSLock* lock = [self.p2pLocks objectForKey:name];
        NSCondition* cond = [self.p2pConditions objectForKey:name];
        if ([lock lockBeforeDate:[NSDate dateWithTimeIntervalSinceNow:0.05]]) {
            [cond lock];
            IMMessage *message = [[IMMessage alloc] init];
            message.msgID = [self getStartReadID:name];
            message.infoType = name;
            message.sn = [IMUtil createSN];
            message.isWaitResp = NO;
            
            [self writeLog:[NSString stringWithFormat:@"%@ added get %@ inbox[s:%lld|o:%d|sn:%lld]", scene, name, message.msgID, self.getInfoOffset, message.sn]];
            message.requestBody = [self.protoMessage createGetInfoRequest:name StartID:message.msgID Offset:self.getInfoOffset Sn:message.sn];
            
            // 目前IM/peer任务的优先级在大队列中为普通；public消息略低一级
            // 目前加入/退出聊天室以及状态改变任务为高优先级
            if ([name isEqualToString:INFO_TYPE_IM] || [name isEqualToString:INFO_TYPE_PEER]) {
                [self addSendTask:message];
            }
            else if ([name isEqualToString:INFO_TYPE_PUBLIC]){
                [self addSendTask:message With:NSOperationQueuePriorityLow];
            }
            else {
                [self addSendTask:message With:NSOperationQueuePriorityVeryLow];
            }
            
            // 记录任务信息，多个类型的inbox操作都会访问此变量，无论lock还是cond都是只针对相同类型的inbox 任务的竞争，此时lock和cond都无法避免竞争。加锁最保险
            @synchronized (self.p2pTaskMessages) {
                [self.p2pTaskMessages setObject:message forKey:name];
            }
           
            @synchronized (self.inboxLastGetInfoTimeDict) {
                [self.inboxLastGetInfoTimeDict setObject:[NSDate date] forKey:name];
            }
            
            if (![cond waitUntilDate:[NSDate dateWithTimeIntervalSinceNow:30]]) {
                [self writeLog:[NSString stringWithFormat:@"%@ get %@ inbox[s:%lld|sn:%lld] timeout", scene, name, message.msgID, message.sn]];
            }
            
            // 清除本次任务记录
            @synchronized (self.p2pTaskMessages) {
                [self.p2pTaskMessages removeObjectForKey:name];
            }
            
            [cond unlock];
            [lock unlock];
        }
        else {
            [self writeLog:[NSString stringWithFormat:@"%@ give up get %@ inbox[s:%lld]", scene, name, [self getStartReadID:name]]];
        }
    }];
    
    // 注意此处设置的高优先级仅仅是后台任务的优先级，并不涉及到往服务端发送的真实任务的优先级
    if ([name isEqualToString:INFO_TYPE_IM]) {
        [operation setQueuePriority:NSOperationQueuePriorityHigh];
    }
    [self.bgOpQueue addOperation:operation];
}

/**
 * 完成取消息盒子的任务
 * @param name: 消息盒子类型：peer, im, public
 * @param sn: 完成任务的sn
 */
- (void) finishQueryInboxTask:(NSString*)name Sn:(int64_t)sn
{
    NSBlockOperation *operation = [NSBlockOperation blockOperationWithBlock:^{
        BOOL failed = NO;
        NSCondition* cond = [self.p2pConditions objectForKey:name];
        [cond lock];
        @synchronized (self.p2pTaskMessages) {
            IMMessage* message = [self.p2pTaskMessages objectForKey:name];
            if (message != nil && message.sn == sn) {
                [cond signal];
            }
            else {
                failed = YES;
            }
        }
        [cond unlock];
        if (failed) {
            [self writeLog:[NSString stringWithFormat:@"can't finish get %@,sn:%lld", name, sn]];
        }
    }];
    
    // 需要尽快通知任务完成
    [operation setQueuePriority:NSOperationQueuePriorityHigh];
    [self.bgOpQueue addOperation:operation];
}

-(void) writeLog:(NSString*)data
{
    NSString *logData = [NSString stringWithFormat:@"[%@] %@\n", self.configData.jid, data];
    [[IMServiceLib sharedInstance] writeLog:logData];
}

-(void) sendBackLog:(NSString*)senderID
{
    //NSMutableData *logData = [[NSMutableData alloc] init];

    NSData *todayLog = [[IMServiceLib sharedInstance] getCurrentLogFileContent];
    //[logData appendData:todayLog];
    //post to http
    [self postNSDataLog:todayLog];

    //传以前的，系统启动时会删除60天前的日志
    for (int i=1; i<=60; ++i)
    {
        NSData *oldLog = [[IMServiceLib sharedInstance] getLogFileContent:[IMUtil getOldDateString:i]];
        if (oldLog == nil)
        {
            continue;
        }
        //[logData appendData:oldLog];
        //post to http
        [self postNSDataLog:oldLog];
    }

    /*
     //NSData *zipData = [IMUtil zipData:logData];
     NSData *zipData = logData;
     //zipData = [zipData gzippedData];
     int totalSize = [zipData length];
     NSRange range;
     range.location = 0;
     //切分数据,每块大小不能超过20k
     int blockSize = 1024*20;
     for(; ;)
     {
     if ((range.location + blockSize) <= totalSize)
     {
     range.length = blockSize;
     }
     else
     {
     range.length = totalSize - range.location;
     }
     NSData *fragData = [zipData subdataWithRange:range];
     //create a get info task
     IMMessage *message = [[IMMessage alloc] init];
     message.msgID = 0;
     message.infoType = @"cmdack";
     message.featureID = IM_Feature_Peer;
     [self writeLog:[NSString stringWithFormat:@"send command ack to %@", senderID]];
     message.requestBody = [self.protoMessage createChatRequest:senderID Data:fragData];
     IMTask *task = [[IMTask alloc] initTask:message User:self];
     [self addTask:task];
     range.location += range.length;
     if (range.location >= totalSize)
     break;
     }
     */

    //post to http
    //[self postNSDataLog:logData];
}

-(void) postNSDataLog:(NSData*) data
{
    NSData *zipData = [data gzippedData];
    //send to http server
    NSData *rc4Data = [IMUtil rc4EncodeFromNSData:zipData Key:@"sdklog#dcslc@3454$"];
    [self.webFeature postHttp:[NSString stringWithFormat:@"", self.configData.jid] requestData:rc4Data];
}

-(void) postNSStringLog:(NSString*) data
{
    [self postNSDataLog:[IMUtil NSStringToNSData:data]];
}

-(NSString*) getTraceID:(NSString*) data
{
    NSData *jsonData = [IMUtil NSStringToNSData:data];
    NSError *error = nil;
    id jsonObject = [NSJSONSerialization JSONObjectWithData:jsonData options:NSJSONReadingAllowFragments error:&error];
    if (jsonObject != nil && error == nil)
    {
        if ([jsonObject isKindOfClass:[NSDictionary class]])
        {
            NSDictionary *dataDict = (NSDictionary *)jsonObject;

            NSObject *object = [dataDict objectForKey:@"traceid"];
            if (object != nil && [object isKindOfClass:[NSString class]])
            {
                return (NSString*)object;
            }
        }
    }
    return @"";
}

-(void)dealloc{
//    CPLog(@"enter dealloc");
    [[NSNotificationCenter defaultCenter]removeObserver:self];
    
}


@end
