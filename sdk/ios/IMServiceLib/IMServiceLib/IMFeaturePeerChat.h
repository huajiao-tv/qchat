//
//  IMFeaturePeerChat.h
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-25.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMFeatureBase.h"

/**
 * 派生于IMFeatureBase，接入服务器
 * 包括：
 * 1）所有与接入和单聊相关的操作;
 */
@interface IMFeaturePeerChat : IMFeatureBase


#pragma mark - presence
//
- (IMErrorCode)queryUserPresenceWithPhoneNumbers:(NSArray*)phonenumbers sessionType:(int)type sessionID:(NSString*)sid sn:(uint64_t)sn;

- (IMErrorCode)sendCreateChannelRequestWithCaller:(NSString*)caller callee:(NSString*)callee sessionID:(NSString*)sid sn:(uint64_t)sn;

/**
 * 发送消息：send incoming call
 */
-(IMErrorCode)sendPeerMessageToReceiver:(NSString*)phoneNumber Body:(NSData*)body BodyType:(int)bodyType sessionID:(NSString*)sid sn:(uint64_t)sn;

@end
