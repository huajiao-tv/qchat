//
//  IMProtoRegistry.h
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-26.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMProtoHelper.h"

/**
 * 对IMProtoHelper的分类，将注册相关的协议封装到该分类中
 * 包括：
 * 1）上行注册;
 * 2) 下行注册;
 * 3) QID注册；
 */
@interface IMProtoRegistry : IMProtoHelper

/**
 * property list
 */
#pragma mark - property list


/**
 * 创建下行注册时获取验证码的请求
 * @returns request的二进制流
 */
-(NSData*) createDownRegistryGetVerifyCodeRequest:(NSString*)password;

/**
 * 解析下行注册时获取验证码的应答
 * @returns true--表示获取成功，false--表示获取失败
 */
-(BOOL) parseDownRegistryGetVerifyCodeRespone:(NSData*)data;


/**
 * 创建下行注册的请求
 * @returns request的二进制流
 */
-(NSData*) createDownRegistryRequest:(NSString*)verifyCode Token:(NSString*)devToken;

/**
 * 解析下行注册的应答
 * @returns 参数都都放到字典里了
 * dict keys:
 * code:
 * jid:
 * password
 */
-(NSMutableDictionary*) parseDownRegistryResponse:(NSData*)data verifyCode:(NSString*)code;


@end
