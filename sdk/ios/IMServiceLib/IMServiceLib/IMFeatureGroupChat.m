//
//  IMFeatureGroupChat.m
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-25.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMFeatureGroupChat.h"
#import "IMUtil.h"
#import "IMUser.h"
#import "IMProtoGroupChat.h"
#import "IMProtoMessage.h"


@interface IMGetGroupMsgReq()
@property (atomic, strong) NSMutableArray* groups;

@end

@implementation IMGetGroupMsgReq
/**
 * 重载初始函数
 */
-(id) init
{
    self = [super init];
    if (self)
    {
        self.groups = [[NSMutableArray alloc] init];
    }
    return self;
}

/**
 * 添加一个需要取消息的请求信息
 * @param groupId: 群ID
 * @param msgId:获取消息请求中的消息起始ID
 * @param offset:请求获取群消息的最大数量
 * @param tid:操作追踪序号
 * @return 结果代码
 */
-(IMErrorCode) addRequestGroup:(NSString*) groupId Start:(unsigned long long) msgId Offset: (int) offset Trace:(int64_t)tid
{
    if (groupId == nil || groupId.length == 0)
    {
        return IM_InvalidParam;
    }

    GetGroupMsgParam* group = [[GetGroupMsgParam alloc]init];
    group.groupId = groupId;
    group.startId = msgId;
    group.offset = offset;
    [group.traceIds addObject:[NSNumber numberWithLongLong:tid]];

    [self.groups addObject:group];
    return IM_Success;
}

/**
 * 获取群消息请求信息
 * @return 请求摘要数组
 */
-(NSArray*) requestSummary
{
    return [self.groups copy];
}
@end

/**
 * IMFeatureGroupChat私有成员声明
 */
@interface IMFeatureGroupChat()

/**
 * 请求任务队列
 */
@property (atomic, strong) NSOperationQueue* taskQueue;
/**
 * 请求任务同步锁
 */
@property (atomic, strong) NSCondition *syncCondition;
/**
 * 相邻两个请求之间需要暂停的秒数，默认是0
 */
@property (atomic, assign) int sendNextAfter;
/**
 * 取消息请求队列，会按照群组合并
 */
@property(atomic, strong) NSMutableDictionary* getMsgReqQueue;
/**
 * 同步群概要信息请求队列
 */
@property(atomic, strong) NSMutableArray* syncGroupQueue;
/**
 * 请求对应列表
 * key是实际请求的sn，值为APP层请求的NSSet集合（集合值类型为NSNumber）
 */
@property(atomic, strong) NSMutableDictionary* requestSnDic;

/**
 * 添加一个需要取消息的请求信息
 * @param data: 需要发送的群请求包GroupUpPacket的二进制流
 * @param sn: 需要发送的群请求包对应的sn
 * @return 结果代码
 */
-(IMErrorCode)sendGroupRequest:(NSData*)data Sn:(int64_t)sn;

/**
 * 将取消息的请求放入排队队列中
 * @param groups: IMGetGroupMsgReq对象的groups属性拷贝
 * @return 操作结果
 */
-(BOOL)queueGetMsgsReq:(NSArray*) groups;

/**
 * 将同步群概要的请求放入排队队列中
 * @param groups: IMGetGroupMsgReq对象的groups属性拷贝
 * @param  sn:APP层请求序列号
 * @return 操作结果
 */
-(BOOL)queueSyncGroupReq:(NSSet*) groups Trace:(int64_t)traceId;

/**
 * 发送取消息的请求队列中的所有请求
 * @param pSn:输出参数，用于指示实际的请求sn
 * @return 是否需要等待
 */
-(BOOL)sendGetMsgsReqWithSn:(int64_t*)pSn;

/**
 * 发送同步群概要信息请求队列中的所有请求
 * @param pSn:输出参数，用于指示实际的请求sn
 * @return 是否需要等待
 */
-(BOOL)sendSyncGroupReqWithSn:(int64_t*) pSn;

/**
 * 处理指定类型的群请求
 * @param type: 需要处理的群请求类型
 * @return 操作结果
 */
-(IMErrorCode)handleGroupRequestWithType:(GroupRequestType) type;
@end

@implementation IMFeatureGroupChat
// const variables
// 允许服务端要求的最大等待时长，秒数
static const int maxMore = 60;
// 单个实际服务器网络需求所合并的数量
static const int maxReqs = 10;

// static variables
static int opCount = 0;
static NSLock* snLock = nil;
static int64_t lastTimeSatmp = 0;

// functions

/**
 *
 * @param config: user config data
 * @returns IMProtoHelper instance
 */
-(id)initWithUser:(IMUser*)user
{
    self = [super initWithUser:user];
    if (self)
    {
        self.taskQueue = [[NSOperationQueue alloc]init];
        self.taskQueue.maxConcurrentOperationCount = 1;
        self.syncCondition = [[NSCondition alloc]init];
        self.sendNextAfter = 0;
        self.getMsgReqQueue = [[NSMutableDictionary alloc]init];
        self.syncGroupQueue = [[NSMutableArray alloc]init];
        self.requestSnDic = [[NSMutableDictionary alloc]init];
    }
    return self;
}

/**
 * 添加一个需要取消息的请求信息
 * @param data: 需要发送的群请求包GroupUpPacket的二进制流
 * @param sn: 需要发送的群请求包对应的sn
 * @return 结果代码
 */
-(IMErrorCode)sendGroupRequest:(NSData*)data Sn:(int64_t)sn
{
    NSData* outData = [self.imUser.protoMessage createServiceRequest:data ServiceID:SERVICE_ID_GROUP SN:sn];

    IMMessage *message = [[IMMessage alloc] init];
    message.IsSend = true;
    message.sn = sn;
    message.requestBody = outData;
    message.featureID = IM_Feature_GroupChat;
    message.isWaitResp = NO;

    return [self sendMessage:message];
}

/**
 * 将取消息的请求放入排队队列中
 * @param groups: IMGetGroupMsgReq对象的groups属性拷贝
 * @param  sn:APP层请求序列号
 * @return 操作结果
 */
-(BOOL)queueGetMsgsReq:(NSArray*) groups
{
    BOOL ret = NO;
    NSMutableSet* traceIds = [[NSMutableSet alloc]init];
    NSMutableArray* logs = [[NSMutableArray alloc]init];
    [self.syncCondition lock];
    for (GetGroupMsgParam* group  in groups)
    {
        char c = '+';
        if (group.offset < 0)
        {
            c = '-';
        }

        NSString* key = [NSString stringWithFormat:@"%@-%@%c", group.groupId, [NSNumber numberWithUnsignedLongLong:group.startId], c];

        // 用于记录日志使用
        [traceIds unionSet:group.traceIds];

        GetGroupMsgParam* old = [self.getMsgReqQueue objectForKey:key];
        if (old != nil)
        {

            [logs addObject:[NSString stringWithFormat:@"get gp msg [%@] covers [%@]", [[group.traceIds allObjects] componentsJoinedByString:@","], [[old.traceIds allObjects] componentsJoinedByString:@","]]];
            [group.traceIds unionSet:old.traceIds];
        }
        [self.getMsgReqQueue setObject:group forKey:key];

        ret = YES;
    }
    [self.syncCondition unlock];

    // 记录客户端请求日志
    [logs addObject:[NSString stringWithFormat:@"queued gp req:[%@]", [[traceIds allObjects] componentsJoinedByString:@", "]]];
    [self.imUser writeLog:[logs componentsJoinedByString:@";"]];
    return ret;
}

/**
 * 将同步群概要的请求放入排队队列中
 * @param groups: IMGetGroupMsgReq对象的groups属性拷贝
 * @param  sn:APP层请求序列号
 * @return 操作结果
 */
-(BOOL)queueSyncGroupReq:(NSSet*) groups Trace:(int64_t)traceId
{
    SyncGroupParam* req = [[SyncGroupParam alloc]init];
    [req.groupIds unionSet:groups];
    [req.traceIds addObject:[NSNumber numberWithLongLong:traceId]];

    [self.syncCondition lock];
    [self.syncGroupQueue addObject:req];
    [self.syncCondition unlock];

    [self.imUser writeLog:[NSString stringWithFormat:@"queued gp sync:[%lld]", traceId]];
    return YES;
}

/**
 * 发送取消息的请求队列中的所有请求
 *  注：无需加锁，本函数调用方负责加锁
 * @param pSn:输出参数，用于指示实际的请求sn
 * @return 是否需要等待
 */
-(BOOL)sendGetMsgsReqWithSn:(int64_t*)pSn
{
    NSMutableSet* traceIds = [[NSMutableSet alloc]init];
    NSMutableArray* groups = [[NSMutableArray alloc]init];
    int count = 0;
    NSArray* allKeys = self.getMsgReqQueue.allKeys;
    for (NSString* key in allKeys)
    {
        if (count >= maxReqs)
        {
            // 我们每个取消息的实际请求仅最多合并指定数量的客户端请求
            break;
        }
        GetGroupMsgParam* group = [self.getMsgReqQueue objectForKey:key];
        [groups addObject:group];
        [traceIds unionSet:group.traceIds];
        [self.getMsgReqQueue removeObjectForKey:key];
        ++count;
    }

    if (groups.count == 0)
    {
        return NO;
    }

    int64_t reqSn = [IMUtil createSN];
    if (NULL != pSn)
    {
        *pSn = reqSn;
    }

    NSData* data = [self.imUser.protoGroupChat createGetGroupMessagePacketFrom:groups];
    if(IM_Success == [self sendGroupRequest:data Sn:reqSn])
    {
        [self.requestSnDic setObject:traceIds forKey:[NSString stringWithFormat:@"%lld", reqSn]];

        [self.imUser writeLog:[NSString stringWithFormat:@"+gp get msg req:%lld, tid:[%@]", reqSn,  [[traceIds allObjects] componentsJoinedByString:@","]]];

        return YES;
    }

    return NO;
}

/**
 * 发送同步群概要信息请求队列中的所有请求
 *  注：无需加锁，本函数调用方负责加锁
 * @param pSn:输出参数，用于指示实际的请求sn
 * @return 是否需要等待
 */
-(BOOL)sendSyncGroupReqWithSn:(int64_t*)pSn
{
    // 预定的任务已经被合并处理了
    if (self.syncGroupQueue.count == 0)
    {
        return NO;
    }

    SyncGroupParam* req = [[SyncGroupParam alloc]init];

    BOOL all = NO;
    for (SyncGroupParam* group in self.syncGroupQueue)
    {
        if (group.groupIds.count == 0)
        {
            all = true;
        }
        else
        {
            [req.groupIds unionSet:group.groupIds];
        }
        [req.traceIds unionSet:group.traceIds];
    }

    [self.syncGroupQueue removeAllObjects];

    if (all)
    {
        [req.groupIds removeAllObjects];
    }

    int64_t reqSn = [IMUtil createSN];
    if (NULL != pSn)
    {
        *pSn = reqSn;
    }

    NSData* data = [self.imUser.protoGroupChat createSyncGroupListPacketFor:req.groupIds];
    if (IM_Success == [self sendGroupRequest:data Sn:reqSn])
    {
        [self.requestSnDic setObject:[req.traceIds copy] forKey:[NSString stringWithFormat:@"%lld", reqSn]];

        [self.imUser writeLog:[NSString stringWithFormat:@"+gp sync req:%lld, tid:[%@]", reqSn, [[req.traceIds allObjects] componentsJoinedByString:@","]]];

        return YES;
    }

    return NO;
}

/**
 * 发送去消息的请求(类似原来的get_info)
 * @param request:请求对象
 * @param sn: 需要发送的群请求包对应的sn
 * @return 操作结果代码
 */
-(IMErrorCode) getMsg:(IMGetGroupMsgReq *)request
{
    if (request == nil || ![self queueGetMsgsReq:request.groups])
    {
        return IM_InvalidParam;
    }

    return [self handleGroupRequestWithType:REQ_GET_GROUP_MSGS];
}

/**
 * 发送同步群概要列表的请求(异步模式)
 * @param ids: 需要同步概要信息的群id列表，nil或者为空表示查询查询所有群的概要信息
 * @return 操作结果代码
 */
-(IMErrorCode)syncGroupSummaryFor:(NSSet*)ids  Trace:(int64_t)traceId
{
    if (![self queueSyncGroupReq:ids Trace:traceId])
    {
        return IM_InvalidParam;
    }

    return [self handleGroupRequestWithType:REQ_SYNC_GROUP];
}

/**
 * 内部回执处理函数
 * @param sn：request sn
 * @param sleep: 下一个请求需要等待的时间
 * @return response所对应的APP层请求sn列表, 如果为空表示非法请求
 */
-(NSSet*)handleRequestReceipt:(int64_t)sn nextAfter:(int)sleep
{
    NSString* snKey = [NSString stringWithFormat:@"%lld", sn];
    NSSet* traceIds = nil;
    NSString* log;

    // 确保下一个请求的等待时间处于一个合法的范围，避免服务端或网络通讯异常导致sleep值过大从而影响无法处理等待队列中的请求
    if (sleep < 0)
    {
        sleep = 0;
    }
    else if (sleep > maxMore)
    {
        sleep = maxMore;
    }

    [self.syncCondition lock];
    traceIds = [self.requestSnDic objectForKey:snKey];
    if (traceIds == nil)
    {
        log =[NSString stringWithFormat:@"got gp resp:%lld but no tid", sn];
    }
    else
    {
        // 通知已经接受到发送任务响应
        [self.syncCondition signal];
        self.sendNextAfter = sleep;
        // 清除掉已经返回的请求对应项
        [self.requestSnDic removeObjectForKey:snKey];
        log = [NSString stringWithFormat:@"got gp resp:%lld", sn];
    }
    [self.syncCondition unlock];

    [self.imUser writeLog:log];
    return traceIds;
}

/**
 * 处理指定类型的群请求
 * @param type: 需要处理的群请求类型
 * @return 操作结果
 */
-(IMErrorCode)handleGroupRequestWithType:(GroupRequestType) type
{
    NSBlockOperation *operation = [NSBlockOperation blockOperationWithBlock:^{
        BOOL wait = NO;
        int64_t reqSn;
        NSString* log = @"";
        // lock first
        [self.syncCondition lock];

        // handle request accroding to request type
        switch (type) {
            case REQ_SYNC_GROUP:
                wait = [self sendSyncGroupReqWithSn:&reqSn];
                break;

            case REQ_GET_GROUP_MSGS:
                wait = [self sendGetMsgsReqWithSn:&reqSn];
                break;

            default:
                break;
        }

        // wait a moment if need
        if (wait)
        {
            NSString* strReqSn = [NSString stringWithFormat:@"%lld", reqSn];
            NSSet* traceIds = [self.requestSnDic objectForKey:strReqSn];
            // 等待服务器请求响应或5秒超时
            if (![self.syncCondition waitUntilDate:[NSDate dateWithTimeIntervalSinceNow:5]])
            {
                // 清除掉超时的请求对应项
                [self.requestSnDic removeObjectForKey:strReqSn];
                log = [NSString stringWithFormat:@"gp req:%lld timeout, rm sn tid map.", reqSn];
                
                NSMutableDictionary *dataDict = [[NSMutableDictionary alloc] init];
                [dataDict setObject:[NSNumber numberWithInt:IM_OperateTimeout] forKey:@"result"];
                [dataDict setObject:log forKey:@"reason"];
                [dataDict setObject:[NSNumber numberWithUnsignedInt:type] forKey:@"payload"];
                [dataDict setObject: traceIds forKey:@"sn"];
                [self.imUser notifyDelegateGroup:dataDict];
            }
            else
            {
                log = [NSString stringWithFormat:@"gp req:%lld resp set next %ds.", reqSn, self.sendNextAfter];
            }

            // 上一个请求的回包要求第二个请求需要暂缓发送到服务器
            if (self.sendNextAfter > 0)
            {
                // self.sendNextAfter单位是秒
                if ([self.syncCondition waitUntilDate:[NSDate dateWithTimeIntervalSinceNow:self.sendNextAfter]])
                {
                    log = [log stringByAppendingString:@"2nd unexpected singal."];
                }
            }
        }

        // more operations if need
        /*
         switch (type) {
         case REQ_SYNC_GROUP:
         break;

         case REQ_GET_GROUP_MSGS:
         break;

         default:
         break;
         }
         */

        // unlock
        [self.syncCondition unlock];
        
        // 记录等待操作日志
        if ([log length] > 0)
        {
            [self.imUser writeLog:log];
        }
    }];
    
    // 同一时间只允许一个任务执行(由NSOperationQueue控制)
    [self.taskQueue addOperation:operation];
    return IM_Success;
}

@end
