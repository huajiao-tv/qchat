//
//  IMUtil.h
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-25.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import <Foundation/Foundation.h>
#import "IMConstant.h"


/**
 * 封装用到的工具类
 */
@interface IMUtil : NSObject

/**
 * rc4加/解密函数
 * @param key: 密匙
 * @param input: 数据,是std::string的指针
 * @param result:必须传入一个std:string的指针
 * @returns 加/解密后的字符串
 */
//+(NSData*) rc4Data:(void*)input Key:(NSString*)key Return:(void*)result;

/**
 * rc4加密函数
 * @param key: 加密密匙
 * @param data: 被加密的数据
 * @param result:必须传入一个std:string的指针
 * @returns 加密后的字符串
 */
+(NSData*) rc4Encode:(NSString *)data Key:(NSString *)key Return:(void*)result;

/**
 * rc4加密函数
 * @param key: 加密密匙
 * @param data: 被加密的数据(std::string)指针
 * @param result:必须传入一个std:string的指针
 * @returns 加密后的字符串
 */
+(NSData*) rc4EncodeFromString:(void *)data Key:(NSString *)key Return:(void*)result;

/**
 * rc4加密函数
 * @param key: 加密密匙
 * @param data: 被加密的数据
 * @returns 加密后的字符串
 */
+(NSData*) rc4EncodeFromNSData:(NSData *)data Key:(NSString *)key;

/**
 * rc4解密函数
 * @param key: 解密密匙
 * @param data: 被解密的数据,是std::string的指针
 * @param result: 解密后的数据放入std::string里
 * @returns 解密后的字符串
 */
+(NSData*) rc4DecodeFromStdString:(const void *)data Key:(NSString *)key Return:(void*)result;

/**
 * rc4解密函数
 * @param key: 解密密匙
 * @param data: 被解密的数据,是std::string的指针
 * @param result: 解密后的数据放入std::string里
 * @returns 解密后的字符串
 */
+(NSData*) rc4Decode:(NSData *)data Key:(NSString *)key Return:(void*)result;


/**
 * rc4解密函数,解密为NSString
 * @param key: 解密密匙
 * @param data: 被解密的数据(std::string)
 * @returns 解密后的字符串
 */
+(NSString*) rc4DecodeToString:(const void *)data Key:(NSString *)key;

/**
 * 将NSData转化为NSString
 * @param data: 输入参数
 * @returns 返回值
 */
+(NSString*) NSDataToNSString:(NSData*)data;

/**
 * 将NSData转化为stl string
 */
+(void)NSDataToStlString:(NSData*)data Return:(void*)result;

/**
 * 将NSString转化为NSData
 * @param data: 输入参数
 * @returns 返回值
 */
+(NSData*) NSStringToNSData:(NSString*)data;

/**
 * 将data转为NSString
 * @param data: 输入参数
 * @returns 返回值
 */
+(NSString*) IntToNSString:(int)data;

/**
 * 将data转为int
 * @param data: 输入参数
 * @returns 返回值
 */
+(int) NSStringToInt:(NSString*)data;

/**
 * 将data转为c++字符串
 * @param data: 输入参数
 * @returns 返回值
 */
+(const char*) NSStringToChars:(NSString*)data;

/**
 * 将data转为NSString
 * @param data: 输入参数
 * @returns 返回值
 */
+(NSString*) CharsToNSString:(const char *)data;


/**
 * 将const char*转为NSData
 * @param data: 输入参数
 * @returns 返回值
 */
+(NSData*) CharsToNSData:(const char *)data withLength:(NSUInteger)length;

/**
 * createRandomString:按照长度产生一个随机字符串
 * @param length: 长度
 * @returns 随机字符串
 */
+(NSString*) createRandomString:(int)length;

/**
 * hasSuccessFlag 检查字典里是否有key为'code'的item，且value为IM_SUCCESS
 * @param dataDict:数据字典
 * @returns true--有，false--没有
 */
+(BOOL) hasSuccessFlag:(NSMutableDictionary*)dataDict;

/**
 * 从字典里获得msg id值
 */
+(int) getPayloadType:(NSMutableDictionary*)dataDict;

/**
 * 查找pattern是否在字符串data中，忽略大小写的
 * @param pattern:想查找的字符串
 * @param data: 字符串
 * @returns true--存在, false--不存在
 */
+(BOOL) hasSubString:(NSString*)pattern Data:(NSString*)data;

/**
 * 从结果字典里查找整数值
 * @param dataDict: 数据字典
 * @param key: 数据在字典里的名称;
 * @returns 整数
 */
+(int) getInt32FromDict:(NSMutableDictionary*)dataDict Key:(NSString*)key;

/**
 * 从结果字典里查找整数值
 * @param dataDict: 数据字典
 * @param key: 数据在字典里的名称;
 * @returns 整数
 */
+(int64_t) getInt64FromDict:(NSMutableDictionary*)dataDict Key:(NSString*)key;

/**
 * 从结果字典里查找整数值
 * @param dataDict: 数据字典
 * @param key: 数据在字典里的名称;
 * @returns 整数
 */
+(NSString*) getStringFromDict:(NSMutableDictionary*)dataDict Key:(NSString*)key;

/**
 * 从结果字典里查找整数值
 * @param dataDict: 数据字典
 * @param key: 数据在字典里的名称;
 * @returns 整数
 */
+(NSData*) getDataFromDict:(NSMutableDictionary*)dataDict Key:(NSString*)key;

/**
 * 从结果字典里查找整数值
 * @param dataDict: 数据字典
 * @param key: 数据在字典里的名称;
 * @returns 整数
 */
+(BOOL) getBoolFromDict:(NSMutableDictionary*)dataDict Key:(NSString*)key;

/**
 * 将网络上接收到的二进制转化为long
 * @param data: 高位在前的二进制
 * @returns long型数据
 */
+(int64_t) getInt64FromNetData:(NSData*)data;

/**
 * 将网络上接收到的二进制转化为int32
 * @param data: 高位在前的二进制
 * @returns int32型数据
 */
+(int) getInt32FromNetData:(NSData*)data;

/**
 * 将状态码转化为状态名称
 * @param state: 状态码;
 * @returns 状态名称
 */
+(NSString*) getStateName:(IMUserState)state;

/**
 * 将NSData转化为stl的string
 * @param src: NSData
 * @param dest: 返回值，即：std::string地址
 */
+(void) convertNSData:(NSData*)src toStlString:(void*)dest;

/**
 * 将一个字典的数据拷贝到另一个字典
 */
+(void) copyDictionary:(NSMutableDictionary*)dest From:(NSDictionary*)src;

/**
 * 获取当前系统时间的秒值
 */
+(int) getCurSystemSecond;

/**
 * 获得当前系统时间的字符串形式
 */
+(NSString*) getCurTimeString;

/**
 * json编码
 */
+(NSData*) jsonEncode:(NSDictionary*)dict;

/**
 * json解码
 */
+(NSDictionary*) jsonDecode:(NSData*)data;

/**
 * 检查字符串是否为空
 * @param str: NSString
 * @returns str为nil或长度为0时返回YES
 */
+(BOOL) isStringEmpty:(NSString*)str;

/**
 * 检查字符串是否为空
 * @param str: NSString
 * @param ignore: BOOL，是否忽略空白字符组成的字符串
 * @returns str为nil或长度为0或全有空白字符组成时返回YES
 */
+(BOOL) isStringEmpty:(NSString*)str ignoreWhiteapce:(BOOL)ignore;

+ (NSString *)URLEncodedString:(NSString*)str;

+ (NSString*)URLDecodedString:(NSString*)str;


/**
 * 将状态码转化为状态名称
 * @param state: 状态码;
 * @returns 状态名称
 */
+(NSString*) getDateString;

+(NSString*) getOldDateString:(int)dayDiff;

+(NSString*) getSystemTimeStamp;

+(NSData*) zipData:(NSData*) input;
+(NSData*) unzipData:(NSData*) input;

/**
 * md5加密
 */
+ (NSString *) getMd5_32Bit_String:(NSString *)srcString;

/**
 * 创建sn
 */
+(int64_t) createSN;

/**
 * 解析id类型，总是获得合法的NSString对象
 */
+ (NSString *)idToString:(id)parse;

/**
 * 解析id类型，总是获得合法的NSData对象
 */
+ (NSData *)idToData:(id)parse;


@end
