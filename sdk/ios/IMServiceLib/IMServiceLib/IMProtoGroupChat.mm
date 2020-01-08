//
//  IMProtoGroupChat.m
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-26.
//  Modified by longjun on 2016-07-15.
//  Copyright © 2016年 qihoo. All rights reserved.
//

#import "IMProtoGroupChat.h"
#import "IMUtil.h"
#import "groupchat.pb.h"
#import "IMConstant.h"
using namespace qihoo::protocol::group;

@implementation GetGroupMsgParam
/**
 * 重载初始函数
 */
-(id) init
{
    self = [super init];
    if (self)
    {
        self.groupId = @"";
        self.startId = 0;
        self.offset = 20;
        self.traceIds =  [[NSMutableSet alloc]init];
    }
    return self;
}
@end

@implementation SyncGroupParam
/**
 * 重载初始函数
 */
-(id) init
{
    self = [super init];
    if (self)
    {
        self.groupIds = [[NSMutableSet alloc]init];
        self.traceIds = [[NSMutableSet alloc]init];
    }
    return self;
}

@end

@interface IMProtoGroupChat()
/**
 * 根据群下行协议包指针读取群聊消息通知到字典
 * @param packet: 解析完成的群下行协议包对象（C++原生引用对象）
 * @param dict: 存放通知结果的字典
 */
-(void) getNewMsgNotifyFromPacket: (const  GroupDownPacket&) packet toDic:(NSMutableDictionary*) dict;

/**
 * 根据群下行协议包指针读取群聊消息到字典
 * @param packet: 解析完成的群下行协议包对象（C++原生引用对象）
 * @param dict: 存放通知结果的字典
 */
-(void) getMsgsFromPacket: (const  GroupDownPacket&) packet toDic:(NSMutableDictionary*) dict;

/**
 * 转换群聊消息C++对象到字典
 * @param resp: 待转换的群聊消息C++对象
 * @return 存放转换结果的字典
 */
-(NSMutableDictionary*) convertGroupMessageResp: (const GroupMessageResp&) resp;

/**
 * 根据群下行协议包指针读取群信息列表到字典
 * @param packet: 解析完成的群下行协议包对象（C++原生引用对象）
 * @param dict: 存放通知结果的字典
 */
-(void) getGroupListFromPacket: (const  GroupDownPacket&) packet toDic:(NSMutableDictionary*) dict;

@end

@implementation IMProtoGroupChat

-(void) getNewMsgNotifyFromPacket: (const  GroupDownPacket&) packet toDic:(NSMutableDictionary*) dict
{
        NSMutableArray *notifies = [[NSMutableArray alloc] init];
        for (int i = 0 ; i < packet.newmsgnotify_size(); ++i) {
            NSMutableDictionary *notifyDict = [[NSMutableDictionary alloc] init];
            const GroupNotify& gpNotify = packet.newmsgnotify(i);
            [notifyDict setObject:[IMUtil CharsToNSString:gpNotify.groupid().c_str()] forKey:@"groupid"];
            [notifyDict setObject:[NSNumber numberWithUnsignedLongLong:gpNotify.msgid()] forKey:@"msgid"];
            if (gpNotify.has_summary())
            {
                [notifyDict setObject:[IMUtil CharsToNSString:gpNotify.summary().c_str()] forKey:@"summary"];
            }

            // add notification object to array
            [notifies addObject:notifyDict];
        }
        [dict setObject:notifies forKey:@"newmsgnotify"];

}

-(void) getMsgsFromPacket: (const  GroupDownPacket&) packet toDic:(NSMutableDictionary*) dict
{
    NSMutableArray *msgs = [[NSMutableArray alloc] init];
    for (int i = 0; i < packet.getmsgresp_size(); ++i) {
        [msgs addObject:[self convertGroupMessageResp:packet.getmsgresp(i)]];
    }
    [dict setObject:msgs forKey:@"groupmsgs"];
}

-(NSMutableDictionary*) convertGroupMessageResp: (const GroupMessageResp&) resp
{
    NSMutableDictionary* results = [[NSMutableDictionary alloc] init];

    // get group id first
    [results setObject:[IMUtil CharsToNSString:resp.groupid().c_str()] forKey:@"groupid"];
    NSString* traceIds = [IMUtil CharsToNSString:resp.traceid().c_str()];
    NSArray* strTids = [traceIds componentsSeparatedByString:@","];
    NSMutableSet* tids = [[NSMutableSet alloc]init];
    for (NSString* strTid in strTids)
    {
        [tids addObject:[NSNumber numberWithLongLong:[strTid longLongValue]]];
    }
    [results setObject:tids forKey:@"traceid"];

    // get group messages list
    NSMutableArray *msgs = [[NSMutableArray alloc] init];
    for (int i = 0; i < resp.msglist_size(); ++i) {
        NSMutableDictionary* msg = [[NSMutableDictionary alloc] init];
        const GroupMessage& gpMsg = resp.msglist(i);

        [msg setObject:[NSNumber numberWithUnsignedLongLong:gpMsg.msgid()] forKey:@"msgid"];
        [msg setObject:[IMUtil CharsToNSString:gpMsg.content().c_str()] forKey:@"content"];
        
        if (gpMsg.has_sendtime()) {
            [msg setObject:[NSNumber numberWithLongLong:gpMsg.sendtime()] forKey:@"sendtime"];
        }
        if (gpMsg.has_sender()) {
            [msg setObject:[IMUtil CharsToNSString:gpMsg.sender().c_str()] forKey:@"sender"];
        }

        [msgs addObject:msg];
    }
    [results setObject:msgs forKey:@"messages"];

    // get group version and message max id if there is
    if (resp.has_maxid()) {
        [results setObject:[NSNumber numberWithUnsignedLongLong:resp.maxid()] forKey:@"maxid"];
    }
    if (resp.has_version()) {
        [results setObject:[NSNumber numberWithLongLong:resp.version()] forKey:@"version"];
    }

    return results;
}

/**
 * 根据接收到的二进制流，解析群下行协议包
 * @param data: 二进制流
 * @return 解析完成的数据字典
 */
-(NSMutableDictionary *)parseGroupDownPacket:(NSData*)data
{
    NSMutableDictionary *dataDict = [[NSMutableDictionary alloc] init];
    [dataDict setObject:[NSNumber numberWithInt:0] forKey:@"result"];

    @try
    {
        std::string strData;
        [IMUtil NSDataToStlString:data Return:&strData];

        GroupDownPacket packet;
        if (!packet.ParseFromString(strData)) {
            [dataDict setObject:[NSNumber numberWithInt:IM_InvalidNetData] forKey:@"result"];
            [dataDict setObject:@"parsed group down packet failed!" forKey:@"reason"];
            return dataDict;
        }

        // always set payload value if parsed data OK
        [dataDict setObject:[NSNumber numberWithUnsignedInt:packet.payload()] forKey:@"payload"];

        if (packet.result() != 0)
        {
            [dataDict setObject:[NSNumber numberWithInt:packet.result()] forKey:@"result"];
            if (packet.has_reason())
            {
                [dataDict setObject:[IMUtil CharsToNSData:packet.reason().c_str() withLength:packet.reason().size()] forKey:@"reason"];
            }

            return dataDict;
        }

        if (packet.has_sleep())
        {
            [dataDict setObject:[NSNumber numberWithInt:packet.sleep()] forKey:@"sendNextAfter"];
        }

        switch (packet.payload()) {
            case PAYLOAD_REQ_SYNC_GROUP_LIST:
                [self getGroupListFromPacket:packet toDic:dataDict];
                break;

            case PAYLOAD_RESP_GET_GROUP_MSGS:
                [self getMsgsFromPacket:packet toDic:dataDict];
                break;

            case PAYLOAD_RESP_NEW_GROUP_MSG_NOTIFY:
                [self getNewMsgNotifyFromPacket:packet toDic:dataDict];
                break;

            default:
                [dataDict setObject:[NSNumber numberWithInt:IM_InvalidNetData] forKey:@"result"];
                [dataDict setObject:@"unrecognized server response payload!" forKey:@"reason"];
                break;
        }
    }
    @catch (NSException *exception)
    {
        [dataDict setObject:exception.reason forKey:@"reason"];
//        CPLog(@"parseGroupDownPacket exception, name:%@, reason:%@", exception.name, exception.reason);
    }

    return dataDict;
}

/**
 * 根据请求信息生成
 * @param request: 需要收取消息的群信息列表
 * @return 群上行包二进制流
 */
-(NSData*) createGetGroupMessagePacketFrom:(NSArray*) request
{
    GroupUpPacket packet;
    packet.set_payload(PAYLOAD_REQ_GET_GROUP_MSGS);

    for (GetGroupMsgParam* group in request)
    {
        GroupMessageReq* pReq = packet.add_getmsgreq();
        pReq->set_groupid([IMUtil NSStringToChars:group.groupId]);
        pReq->set_traceid([IMUtil NSStringToChars:[[group.traceIds allObjects] componentsJoinedByString:@","]]);
        pReq->set_startid(group.startId);
        pReq->set_offset(group.offset);
    }

    std::string strPacket = packet.SerializeAsString();
    return [IMUtil CharsToNSData:strPacket.c_str() withLength:strPacket.size()];
}

/**
 * 根据群下行协议包指针读取群信息列表到字典
 * @param packet: 解析完成的群下行协议包对象（C++原生引用对象）
 * @param dict: 存放通知结果的字典
 */
-(void) getGroupListFromPacket: (const  GroupDownPacket&) packet toDic:(NSMutableDictionary*) dict
{
    NSMutableArray *groupList = [[NSMutableArray alloc] init];
    for (int i = 0; i < packet.syncresp_size(); ++i) {
        NSMutableDictionary * summary = [[NSMutableDictionary alloc] init];
        const GroupInfo& group = packet.syncresp(i);
        [summary setObject:[IMUtil CharsToNSString:group.groupid().c_str()] forKey:@"groupid"];
        if (group.has_maxid())
        {
            [summary setObject:[NSNumber numberWithUnsignedLongLong:group.maxid()] forKey:@"maxid"];
        }
        if (group.has_version())
        {
            [summary setObject:[NSNumber numberWithLongLong:group.version()] forKey:@"version"];
        }
        if (group.has_startid())
        {
            [summary setObject:[NSNumber numberWithUnsignedLongLong:group.startid()] forKey:@"startid"];
        }

        // add notification object to array
        [groupList addObject:summary];
    }
    [dict setObject:groupList forKey:@"grouplists"];
}

/**
 * 生成同步群信息上行包二进制流
 * @param groups: 存放了需要同步的群信息，nil对象或空集合同步所有群概要
 * @return 群上行包二进制流
 */
-(NSData*) createSyncGroupListPacketFor:(NSSet*)groups
{
    GroupUpPacket packet;
    packet.set_payload(PAYLOAD_REQ_SYNC_GROUP_LIST);
    
    if (groups != nil && groups.count > 0)
    {
        for (NSString* groupId in groups)
        {
            GroupSyncReq* pReq =packet.add_syncreq();
            pReq->set_groupid([IMUtil NSStringToChars:groupId]);
        }
    }

    std::string strPacket = packet.SerializeAsString();
    return [IMUtil CharsToNSData:strPacket.c_str() withLength:strPacket.size()];
}
@end
