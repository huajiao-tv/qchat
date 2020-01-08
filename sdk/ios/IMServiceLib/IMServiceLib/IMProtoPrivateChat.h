//
//  IMProtoPrivateChat.h
//  IMServiceLib
//
//  Created by guanjianjun on 15/6/18.
//  Copyright (c) 2015年 qihoo. All rights reserved.
//

#import "IMProtoHelper.h"

@interface IMProtoPrivateChat : IMProtoHelper

/**
 * 根据接收到的二进制流，解析
 * @param data: 二进制流(std::string *)
 * @returns 字典
 */
-(NSMutableDictionary *)parseMessage:(NSData*)data;

/**
 * 创建发送消息二进制流
 */
-(NSData*) createSendMsgRequest:(NSString*)destid Appid:(int)appid Type:(int)type Data:(NSData*)data Expire:(int)seconds;
-(NSData*) createSendMsgRequest:(NSString*)destid Appid:(int)appid Type:(int)type Data:(NSData*)data;

/**
 * 创建取消息二进制流
 */
-(NSData*) createGetMsgRequest:(long long)start Count:(int)count;

@end
