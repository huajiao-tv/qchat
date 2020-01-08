//
//  IMFeaturePrivateChat.h
//  IMServiceLib
//
//  Created by guanjianjun on 15/6/18.
//  Copyright (c) 2015年 qihoo. All rights reserved.
//

#import "IMFeatureBase.h"

@interface IMFeaturePrivateChat : IMFeatureBase

/**
 * 创建发送消息二进制流
 * destid: 接收者的userid；
 * appid: 为了各个app间打通，所以提供了这个参数;
 * type: data类型;
 * data: 真正的消息体;
 * seconds: 消息过期时间;
 * sn: 上层如果关心本次发送出去的消息最后到服务器写成功后被服务器编的消息id的话，就应该自己传入sn‘参数；如果不关心就可以用下面的函数；
 */
-(IMErrorCode) sendMsg:(NSString*)destid Appid:(int)appid Type:(int)type Data:(NSData*)data Expire:(int)seconds SN:(int64_t)sn;
-(IMErrorCode) sendMsg:(NSString*)destid Type:(int)type Data:(NSData*)data Expire:(int)seconds SN:(int64_t)sn;
/**
 * 创建发送消息二进制流,消息过期时间采用服务器默认值(7天)
 * destid: 接收者的userid；
 * appid: 为了各个app间打通，所以提供了这个参数;
 * type: data类型;
 * data: 真正的消息体;
 */
-(IMErrorCode) sendMsg:(NSString*)destid Appid:(int)appid Type:(int)type Data:(NSData*)data;

-(IMErrorCode) sendMsg:(NSString*)destid Type:(int)type Data:(NSData*)data SN:(int64_t)sn;
/**
 * 创建发送消息二进制流,消息过期时间采用服务器默认值(7天),且接收者的appid采用与当前登录用户相同
 * destid: 接收者的userid；
 * type: data类型;
 * data: 真正的消息体;
 */
-(IMErrorCode) sendMsg:(NSString*)destid Type:(int)type Data:(NSData*)data;

/**
 * 发送去消息的请求(类似原来的get_info)
 * start:起始消息id;
 * count:一次取多少条消息
 */
-(IMErrorCode) getMsg:(int64_t)start Count:(int)count;

@end
