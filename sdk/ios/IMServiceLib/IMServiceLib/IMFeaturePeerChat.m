//
//  IMFeaturePeerChat.m
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-25.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMFeaturePeerChat.h"
#import "IMUtil.h"
#import "IMUser.h"
#import "IMProtoMessage.h"
#import "IMProtoVoiceCallProxy.h"
#import "IMProtoChatRoom.h"

@implementation IMFeaturePeerChat

#pragma mark - presence
- (IMErrorCode)queryUserPresenceWithPhoneNumbers:(NSArray*)phonenumbers sessionType:(int)type sessionID:(NSString*)sid sn:(uint64_t)sn
{
//     CPLog(@"%s",__FUNCTION__);
    if (phonenumbers == nil || phonenumbers.count == 0) {
//        CPLog(@"query_presence: invalid arguments!!!");
        return IM_InvalidParam;
    }
    if ([self isUserConnected]) {
        NSData* data = [self.imUser.protoMessage createEx1QueryUserStatusRequest:sn UserType:@"phone" UserIds:phonenumbers];
        
        IMMessage *message = [[IMMessage alloc] init];
        message.IsSend = true;
        message.sn = sn ;
        message.requestBody = data;
        message.sessionID = sid;
        message.sessionType = type;
        
        return [self sendMessage:message];
    }else{
        return IM_NoConnection;
    }
}

- (IMErrorCode)sendCreateChannelRequestWithCaller:(NSString*)caller callee:(NSString*)callee sessionID:(NSString*)sid sn:(uint64_t)sn
{
//     CPLog(@"%s",__FUNCTION__);
    if (caller == nil || caller.length == 0 || callee == nil || callee.length == 0) {
//        CPLog(@"CreateChannelRequest: invalid arguments!!!");
        return IM_InvalidParam;
    }
    if ([self isUserConnected]) {
        
        NSData* data = [self.imUser.protoVoiceCallProxy createChannelRequestWithCaller:caller callee:callee sn:sn];
        
        IMMessage *message = [[IMMessage alloc] init];
        message.IsSend = true;
        message.sn = sn ;
        message.requestBody = data;
        message.sessionID = sid;
        
        return [self sendMessage:message];
    }else{
        return IM_NoConnection;
    }
}

-(IMErrorCode)sendPeerMessageToReceiver:(NSString*)phoneNumber Body:(NSData*)body BodyType:(int)bodyType sessionID:(NSString*)sid sn:(uint64_t)sn
{
//    CPLog(@"%s",__FUNCTION__);
    if (phoneNumber == nil || body == nil) {
//        CPLog(@"sendPeerMessage: invalid arguments!!!");
        return IM_InvalidParam;
    }
    
    if ([self isUserConnected])
    {
        NSData* data = [self.imUser.protoMessage createPeerChatRequest:sn Receiver:phoneNumber RecvType:@"phone" Body:body BodyType:bodyType ExpireTime:300];
        
        IMMessage *message = [[IMMessage alloc] init];
        message.IsSend = true;
        message.featureID = IM_Feature_Channel;
        message.sn = sn;
        message.requestBody = data;
        message.sessionID = sid;
        
        return  [self sendMessage:message];
    }
    else
    {
        return IM_NoConnection;
    }
}
@end
