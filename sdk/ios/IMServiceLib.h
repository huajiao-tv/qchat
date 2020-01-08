//
//  IMServiceLib.h
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-7.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import <Foundation/Foundation.h>
#import "IMUser.h"
#import "IMNotifyDelegate.h"
#import "IMServerAddress.h"

/**
 * @class IMServiceLib
 * @brief IMServiceLib is enter class which is used to create user.
 */
@interface IMServiceLib : NSObject

#pragma mark - function list

/**
 * 单例方法
 */
+(IMServiceLib*) sharedInstance;

/**
 *  关闭日志
 *
 *  @param isClose default:No，开启日志
 */
-(void)closeLog:(BOOL)isClose;

/**
 *重写方法，防止重复创建
 */
+(id) allocWithZone:(struct _NSZone *)zone;

/**
 * 判断网络是否可达
 */
-(BOOL) isNetworkReachable;

/**
 * 判断网络是否是gprs
 */
-(BOOL) isGPRSNetwork;

/**
 * 判断网络是否是gprs
 */
-(BOOL) isWifiNetwork;

/**
 * 判断网络类型：0:unkonwn 1:2g 2:3g 3:wifi 4:ethe 5: 4G LTE
 */
-(int) networkType;

/**
 * 创建user
 * @param userID 用户id;
 * @param token 用户密码
 * @param deviceID 设备id，用于支持同一个帐号在不同的设备上同时登录的；
 * @param delegate 接收通知的协议;
 * @param isOnline true--sdk将连接线上服务器；false--sdk将连接测试服务器;(为了支持切换服务器环境时不用再改sdk)
 * @returns 创建成功返回一个IMUser的指针，失败返回nil
 */
-(IMUser*) createUser:(NSString*)userID token:(NSString*)token deviceID:(NSString*)deviceID appID:(int)appid defalutKey:(NSString*)defalutKey version:(int)version serverList:(NSArray<IMServerAddress*>*)serverList  dispatcherServer:(IMServerAddress*)dispatcherServer withDelegate:(id<IMNotifyDelegate>)delegate;

-(IMUser*) createUser:(NSString*)userID token:(NSString*)token sig:(NSString*)sig deviceID:(NSString*)deviceID appID:(int)appid defalutKey:(NSString*)defalutKey version:(int)version serverList:(NSArray<IMServerAddress*>*)serverList  dispatcherServer:(IMServerAddress*)dispatcherServer withDelegate:(id<IMNotifyDelegate>)delegate;

/**
 * 删除tantan当天用户
 */
-(void) shutdownUser;

/**
 * 返回谈谈的当前用户
 */
-(IMUser*) curUser;

/**
 * 写日志函数
 * @param data 日志内容
 */
-(void) writeLog:(NSString*)data;

/**
 * 获取指定日期日志文件内容
 * @param dateString 使用 [IMUtil getDateString]或[IMUtil getOldDateString:i]获得的日子格式化字符串
 */
-(NSData*) getLogFileContent:(NSString *)dateString;

/**
 * 获取当前日志文件内容
 */
-(NSData *) getCurrentLogFileContent;

@end
