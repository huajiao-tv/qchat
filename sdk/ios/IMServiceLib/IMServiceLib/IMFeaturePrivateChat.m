//
//  IMFeaturePrivateChat.m
//  IMServiceLib
//
//  Created by guanjianjun on 15/6/18.
//  Copyright (c) 2015年 qihoo. All rights reserved.
//

#import "IMFeaturePrivateChat.h"
#import "IMUtil.h"
#import "IMUser.h"
#import "IMProtoPrivateChat.h"
#import "IMProtoMessage.h"

@implementation IMFeaturePrivateChat

-(IMErrorCode) sendMsg:(NSString*)destid Appid:(int)appid Type:(int)type Data:(NSData*)data Expire:(int)seconds  SN:(int64_t)sn
{
    NSData* pchatData = [self.imUser.protoPrivateChat createSendMsgRequest:destid Appid:appid Type:type Data:data Expire:seconds];
        //servicemsgreques是将业务数据包装到address_book的chatreq里
    //NSData* msgData = [self.imUser.protoMessage createServiceMsgRequest:pchatData ServiceID:10000013];
    NSData *msgData = [self.imUser.protoMessage createServiceRequest:pchatData ServiceID:10000013 SN:sn];

    IMMessage *message = [[IMMessage alloc] init];
    message.IsSend = true;
    message.sn = sn;
    message.requestBody = msgData;
    message.featureID = IM_Feature_PrivateChat;
    message.payload = 1001;

    return [self sendMessage:message];

}
-(IMErrorCode) sendMsg:(NSString*)destid Type:(int)type Data:(NSData*)data Expire:(int)seconds SN:(int64_t)sn
{
    return [self sendMsg:destid Appid:self.imUser.configData.appid Type:0 Data:data Expire:seconds SN:sn];
}
-(IMErrorCode) sendMsg:(NSString*)destid Type:(int)type Data:(NSData*)data SN:(int64_t)sn
{
    return [self sendMsg:destid Appid:self.imUser.configData.appid Type:0 Data:data Expire:0 SN:sn];
}
-(IMErrorCode) sendMsg:(NSString*)destid Appid:(int)appid Type:(int)type Data:(NSData*)data
{
    return [self sendMsg:destid Appid:appid Type:type Data:data Expire:0 SN:[IMUtil createSN]];
}
-(IMErrorCode) sendMsg:(NSString*)destid Type:(int)type Data:(NSData*)data
{
    return [self sendMsg:destid Appid:self.imUser.configData.appid Type:type Data:data Expire:0 SN:[IMUtil createSN]];
}


/**
 * 发送去消息的请求(类似原来的get_info)
 * start:起始消息id;
 * count:一次取多少条消息
 */
-(IMErrorCode) getMsg:(int64_t)start Count:(int)count
{
    int64_t sn = [IMUtil createSN];
    NSData* pchatData = [self.imUser.protoPrivateChat createGetMsgRequest:start Count:count];
    NSData* msgData = [self.imUser.protoMessage createServiceRequest:pchatData ServiceID:10000013 SN:sn];

    IMMessage *message = [[IMMessage alloc] init];
    message.IsSend = true;
    message.sn = sn;
    message.requestBody = msgData;
    message.featureID = IM_Feature_PrivateChat;

    return [self sendMessage:message];
}

@end
