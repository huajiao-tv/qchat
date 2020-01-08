//
//  IMProtoHelper.h
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-24.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import <Foundation/Foundation.h>
#import "IMConstant.h"
#import "IMConfigData.h"

@class IMUser;

/**
 * proto buf协议编解码基类
 */
@interface IMProtoHelper : NSObject

/**
 * property list
 */
#pragma mark - property list
@property (atomic, strong) IMConfigData *userConfig;

@property (readonly, atomic, strong) IMUser *imUser;

/**
 * 保存协议里请求命令名称与id的对应关系
 */
@property (atomic, strong) NSMutableDictionary *cmdReqNameToIDDict;

/**
 * 保存协议里请求命令id与名称的对应关系
 */
@property (atomic, strong) NSMutableDictionary *cmdReqIDToNameDict;

/**
 * 保存协议里应答命令名称与id的对应关系
 */
@property (atomic, strong) NSMutableDictionary *cmdResNameToIDDict;

/**
 * 保存协议里应答命令id与名称的对应关系
 */
@property (atomic, strong) NSMutableDictionary *cmdResIDToNameDict;


/**
 * function list
 */
#pragma mark - function list

/**
 *
 * @param config: user config data
 * @returns IMProtoHelper instance
 */
-(id)initWithUser:(IMUser*)imUser;

/**
 * 初始化数据，子类可以重载该函数并实现自己的初始化工作
 */
-(void) initData;

/**
 * 添加命令名称与id的对应关系
 * @param name: 命令名称
 * @param id: 命令id
 */
-(void) addRequestMap:(NSString*)name ID:(int)code;

/**
 * 根据id获得名称
 * @param id: 命令id
 * @returns 命令名称
 */
-(NSString*) getRequestNameByID:(int)code;

/**
 * 根据名称获得id
 * @param name:proto文件中的命令名称，不区分大小写
 * @returns 命令id
 */
-(int) getRequestIDByName:(NSString*)name;

/**
 * 添加命令名称与id的对应关系
 * @param name: 命令名称
 * @param id: 命令id
 */
-(void) addResponseMap:(NSString*)name ID:(int)code;

/**
 * 根据id获得名称
 * @param id: 命令id
 * @returns 命令名称
 */
-(NSString*) getResponseNameByID:(int)code;

/**
 * 根据名称获得id
 * @param name:proto文件中的命令名称，不区分大小写
 * @returns 命令id
 */
-(int) getResponseIDByName:(NSString*)name;

/**
 * 创建一个时间戳
 */
-(UInt64) createSN;

/**
 * 讲一个二进制打包为Flag+Length+Body格式,因为只有initLogin才会使用default key加密
 * 里面会自动加上Flag
 * @param data: 需要发出去的二进制流(std::string *);
 */
-(NSData*) createDefaultKeyOutData:(void*)data;

/**
 * 讲一个二进制打包为Length+Body格式
 * @param data: 需要发出去的二进制流(std::string *);
 */
-(NSData*) createEncryptData:(void*)data Key:(NSString*)key;

/**
 * 讲一个二进制打包为Length+Body格式
 * @param data: 需要发出去的二进制流(std::string *);
 */
//-(NSData*) createSessionKeyOutData:(NSData*)data;
/**
 * 讲一个二进制打包为Length+Body格式
 * @param data: 需要发出去的二进制流(std::string *) 转换为的NSdata;
 */
-(NSData*) createSessionKeyOutDataWithStrData:(NSData*)data;

@end
