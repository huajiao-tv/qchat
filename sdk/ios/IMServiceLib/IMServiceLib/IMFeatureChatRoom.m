//
//  IMFeatureChatRoom.m
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-25.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMFeatureChatRoom.h"
#import "IMUtil.h"
#import "IMUser.h"
#import "IMProtoChatRoom.h"
#import "IMProtoMessage.h"


@implementation IMFeatureChatRoom

/**
 * 发送需要优先数据
 * @param message: immessage
 * @returns 成功-IM_Success，失败-other
 */
-(IMErrorCode) sendHighPriorityMessage:(IMMessage*)message
{
    return [self.imUser addSendTask:message With:NSOperationQueuePriorityHigh];
}

/**
 * 加入聊天室函数
 */
- (IMErrorCode) sendJoinChatroom:(NSString*)roomid withProperties:(NSDictionary*)properties
{
    return [self sendJoinChatroom:roomid withData:nil Properties:properties];
}

/**
 * 新增一个带用户数据的加入聊天室函数，允许用户在加入聊天室时上传一个二进制参数，服务器在通知所有成员说某个用户加入聊天室时会带上改参数，
 * 参数的具体含义由上层业务自己定义
 * 该参数用户在加入聊天室时用户可以将自己的头像url，昵称等打包为一个二进制流
 * 放到加入请求里，服务器完成加入操作后下发的通知里将该二进制流通过通知消息透传给所有的聊天室成员，
 * 其他成员收到通知后，解析该二进制数据，就可以直接展示昵称和取头像了。
 */
- (IMErrorCode) sendJoinChatroom:(NSString*)roomid withData:(NSData*)userdata Properties:(NSDictionary*)properties
{
    int64_t sn = [IMUtil createSN];
    NSData* data = [self.imUser.protoChatRoom createJoinRoomRequest:roomid withData:userdata Properties:properties];
    NSData* outData = [self.imUser.protoMessage createServiceRequest:data ServiceID:SERVICE_ID_CHATROOM SN:sn];

    IMMessage *message = [[IMMessage alloc] init];
    message.IsSend = true;
    message.sn = sn;
    message.requestBody = outData;
    message.featureID = IM_Feature_ChatRoom;
    message.payload = PAYLOAD_JOIN_CHATROOM;

    [self.imUser writeLog:[NSString stringWithFormat:@"+cr:%@,sn:%lld", roomid, sn]];
    return [self sendHighPriorityMessage:message];
}


/**
 * 发送查询聊天室详情请求
 */
- (IMErrorCode) sendQueryChatroom:(NSString*)roomid
{
    return [self sendQueryChatroom:roomid From:0 Count:0];
}
- (IMErrorCode) sendQueryChatroom:(NSString*)roomid From:(int)from Count:(int)count
{
    int64_t sn = [IMUtil createSN];
    NSData* data = [self.imUser.protoChatRoom createQueryRoomRequest:roomid From:from Count:count];
    NSData* outData = [self.imUser.protoMessage createServiceRequest:data ServiceID:SERVICE_ID_CHATROOM SN:sn];

    IMMessage *message = [[IMMessage alloc] init];
    message.IsSend = true;
    message.sn = sn;
    message.requestBody = outData;
    message.featureID = IM_Feature_ChatRoom;
    message.payload = PAYLOAD_QUERY_CHATROOM;

    [self.imUser writeLog:[NSString stringWithFormat:@"q cr:%@,sn:%lld", roomid, sn]];
    return [self sendMessage:message];
}

/**
 * 发送退出聊天室请求
 */
- (IMErrorCode) sendQuitChatroom:(NSString*)roomid
{
    int64_t sn = [IMUtil createSN];
    NSData* data = [self.imUser.protoChatRoom createQuitRoomRequest:roomid];
    NSData* outData = [self.imUser.protoMessage createServiceRequest:data ServiceID:SERVICE_ID_CHATROOM SN:sn];

    IMMessage *message = [[IMMessage alloc] init];
    message.IsSend = true;
    message.sn = sn;
    message.requestBody = outData;
    message.featureID = IM_Feature_ChatRoom;
    message.payload = PAYLOAD_QUIT_CHATROOM;

    [self.imUser writeLog:[NSString stringWithFormat:@"-cr:%@,sn:%lld", roomid, sn]];
    return [self sendHighPriorityMessage:message];
}

/**
 * 给聊天室发消息
 */
- (IMErrorCode) sendChatroom:(NSString*)roomid Data:(NSData*)content
{
    int64_t sn = [IMUtil createSN];
    NSData* data = [self.imUser.protoChatRoom createChatroomMessageRequest:roomid Message:content];
    NSData* outData = [self.imUser.protoMessage createServiceRequest:data ServiceID:SERVICE_ID_CHATROOM SN:sn];

    IMMessage *message = [[IMMessage alloc] init];
    message.IsSend = true;
    message.sn = sn;
    message.requestBody = outData;
    message.featureID = IM_Feature_ChatRoom;
    message.isWaitResp = NO;

    [self.imUser writeLog:[NSString stringWithFormat:@"cr:%@ msg,sn:%lld", roomid, sn]];
    return [self sendMessage:message];
}

- (IMErrorCode) sendChatroom:(NSString*)roomid String:(NSString*)content
{
    return [self sendChatroom:roomid Data:[IMUtil NSStringToNSData:content]];
}

/**
 * 发送订阅/取消定于聊天室消息请求
 */
- (IMErrorCode) subscribe:(BOOL)sub MessageOfChatroom:(NSString*)roomid
{
    int64_t sn = [IMUtil createSN];
    NSData* data = [self.imUser.protoChatRoom createSubscribe:sub Request:roomid];
    NSData* outData = [self.imUser.protoMessage createServiceRequest:data ServiceID:SERVICE_ID_CHATROOM SN:sn];
    
    IMMessage *message = [[IMMessage alloc] init];
    message.IsSend = true;
    message.sn = sn;
    message.requestBody = outData;
    message.featureID = IM_Feature_ChatRoom;
    message.payload = PAYLOAD_SUBSCRIBE_CHATROOM;
    message.isWaitResp = NO;
    
    [self.imUser writeLog:[NSString stringWithFormat:@"sub cr:%@,sn:%lld", roomid, sn]];
    return [self sendMessage:message];
}

@end
