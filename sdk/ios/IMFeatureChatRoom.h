//
//  IMFeatureChatRoom.h
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-25.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMFeatureBase.h"

/**
 * 派生于IMFeatureBase，聊天室
 * 包括：
 * 1）所有与聊天室相关的操作;
 */
@interface IMFeatureChatRoom : IMFeatureBase

/**
 * 发送加入聊天室请求
 */
- (IMErrorCode) sendJoinChatroom:(NSString*)roomid withProperties:(NSDictionary*)properties;
/**
 * 新增一个函数，允许用户在加入聊天室时上传一个二进制参数，服务器在通知所有成员说某个用户加入聊天室时会带上改参数，
 * 参数的具体含义由上层业务自己定义
 */
- (IMErrorCode) sendJoinChatroom:(NSString*)roomid withData:(NSData*)userdata Properties:(NSDictionary*)properties;

/**
 * 发送查询聊天室详情请求
 */
/**
 * 发送查询聊天室详情请求
 */
- (IMErrorCode) sendQueryChatroom:(NSString*)roomid;
- (IMErrorCode) sendQueryChatroom:(NSString*)roomid From:(int)from Count:(int)count;

/**
 * 发送退出聊天室请求
 */
- (IMErrorCode) sendQuitChatroom:(NSString*)roomid;

/**
 * 给聊天室发消息
 */
- (IMErrorCode) sendChatroom:(NSString*)roomid Data:(NSData*)content;
- (IMErrorCode) sendChatroom:(NSString*)roomid String:(NSString*)content;

/**
 * 发送订阅/取消定于聊天室消息请求
 */
- (IMErrorCode) subscribe:(BOOL)sub MessageOfChatroom:(NSString*)roomid;

@end
