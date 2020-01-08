//
//  IMFeatureBase.h
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-24.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import <Foundation/Foundation.h>
#import "IMMessage.h"

@class IMUser;

/**
 * IM server业务模块的基类，其他各个业务模块都需要基于该类派生
 */
@interface IMFeatureBase : NSObject

/**
 * property list
 */
#pragma mark - property list

/**
 * user pointer
 */
@property (atomic, strong) IMUser *imUser;

/**
 * uri path and param key-value
 */
@property (atomic, strong) NSMutableDictionary *featureUriDict;



/**
 * function list
 */
#pragma mark - function list

/**
 *
 * @param config: user config data
 * @returns IMProtoHelper instance
 */
-(id)initWithUser:(IMUser*)user;

/**
 * 发送http post请求
 * @param uri: url address
 * @param data: request data which is put into body
 * @returns immessage pointer
 */
-(IMMessage*) postHttp:(NSString*)uri requestData:(NSData*)data;

/**
 * 发送http post请求
 * @param uri: url address
 * @param data: request data which is put into body
 * @param timeout: time out
 * @returns immessage pointer
 */
-(IMMessage*) postHttp:(NSString*)uri requestData:(NSData*)data timeout:(NSTimeInterval)timeout;

/**
 * 发送http get请求
 * @param uri: url address
 * @returns immessage pointer
 */
-(IMMessage*) getHttp:(NSString*)uri;

/**
 * 发送http get请求
 * @param uri: url address
 * @param timeout: time out
 * @returns immessage pointer
 */
-(IMMessage*) getHttp:(NSString*)uri timeout:(NSTimeInterval)timeout;

/**
 * 检查用户是否处于连接状态
 */
-(BOOL) isUserConnected;

/**
 * 给msgrouter发送数据
 * @param message: immessage
 * @returns 成功-IM_Success，失败-other
 */
-(IMErrorCode) sendMessage:(IMMessage*)message;


@end
