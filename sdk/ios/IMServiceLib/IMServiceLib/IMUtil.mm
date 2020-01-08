//
//  IMUtil.m
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-25.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMUtil.h"
#import "zlib.h"
#import "IMRC4.h"
#import <CommonCrypto/CommonDigest.h>
#import "GZIP.h"
#include <string>
#include <sys/sysctl.h>

using namespace std;

@implementation IMUtil
// static variables
static int opCount = 0;
static NSLock* snLock = nil;
static int64_t lastTimeSatmp = 0;

//加解密函数
+(NSData*) rc4Data:(void*)input Key:(NSString*)key Return:(std::string &)result
{
    if (key == nil)
    {
        return nil;
    }
    else
    {
        RC4 rc4([key UTF8String]);
        std::string &srcData = *((std::string*)input);
        result.assign(srcData.c_str(), srcData.size());
        rc4.rc4_encode(result);
        return [IMUtil CharsToNSData:result.c_str() withLength:result.size()];
    }
}

//加密函数
+(NSData*) rc4Encode:(NSString *)data Key:(NSString *)key Return:(void*)result
{
    std::string input([data UTF8String]);
    return [IMUtil rc4Data:&input Key:key Return:*((std::string *)result)];
}

/**
 * rc4加密函数
 * @param key: 加密密匙
 * @param data: 被加密的数据(std::string)指针
 * @param result:必须传入一个std:string的指针
 * @returns 加密后的字符串
 */
+(NSData*) rc4EncodeFromString:(void *)data Key:(NSString *)key Return:(void*)result
{
    return [IMUtil rc4Data:data Key:key Return:*((std::string *)result)];
}

+(NSData*) rc4EncodeFromNSData:(NSData *)data Key:(NSString *)key
{
    RC4 rc4([key UTF8String]);
    std::string srcData;
    std::string result;
    [IMUtil NSDataToStlString:data Return:&srcData];
    result.assign(srcData.c_str(), srcData.size());
    rc4.rc4_encode(result);
    return [IMUtil CharsToNSData:result.c_str() withLength:result.size()];
}

//解密函数
+(NSData*) rc4DecodeFromStdString:(const void *)data Key:(NSString *)key Return:(void*)result
{
    std::string &tmp = *((std::string*)data);
    return [IMUtil rc4Data:&tmp Key:key Return:*((std::string *)result)];
}

/**
 * rc4解密函数
 * @param key: 解密密匙
 * @param data: 被解密的数据,是std::string的指针
 * @param result: 解密后的数据放入std::string里
 * @returns 解密后的字符串
 */
+(NSData*) rc4Decode:(NSData *)data Key:(NSString *)key Return:(void*)result
{
    NSUInteger len = [data length];
    char *raw = new char[len];
    [data getBytes:raw length:len];
    std::string input;
    input.assign(raw, len);
    delete [] raw;
    raw = NULL;
    
    std::string &resultData = *((std::string *)result);
    NSData *retData = [IMUtil rc4Data:&input Key:key Return:resultData];
    return retData;
}

//解密函数
+(NSString*) rc4DecodeToString:(const void *)data Key:(NSString *)key
{
    std::string result;
    std::string &tmp = *((std::string*)data);
    return [IMUtil NSDataToNSString:[IMUtil rc4Data:&tmp Key:key Return:result]];
}

/**
 * 将NSData转化为NSString
 * @param data: 输入参数
 * @returns 返回值
 */
+(NSString*) NSDataToNSString:(NSData*)data
{
    return [[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
}

/**
 * 将NSString转化为NSData
 * @param data: 输入参数
 * @returns 返回值
 */
+(NSData*) NSStringToNSData:(NSString*)data
{
    return [data dataUsingEncoding: NSUTF8StringEncoding];
}

+(NSString*) IntToNSString:(int)data
{
    return [NSString stringWithFormat:@"%d",data];
}

+(int) NSStringToInt:(NSString*)data
{
    return [data intValue];
}

/**
 * 将data转为c++字符串
 * @param data: 输入参数
 * @returns 返回值
 */
+(const char*) NSStringToChars:(NSString*)data
{
    return [data UTF8String];
}

/**
 * 将data转为NSString
 * @param data: 输入参数
 * @returns 返回值
 */
+(NSString*) CharsToNSString:(const char *)data
{
    return [NSString stringWithUTF8String:data];
}

/**
 * 将const char*转为NSData
 * @param data: 输入参数
 * @returns 返回值
 */
+(NSData*) CharsToNSData:(const char *)data withLength:(NSUInteger)length
{
    return [NSData dataWithBytes:data length:length];
}

/**
 * createRandomString:按照长度产生一个随机字符串
 * @param length: 长度
 * @returns 随机字符串
 */
+(NSString*) createRandomString:(int)length
{
    char *data = new char[length];
    for (int x=0; x<length; data[x++] = (char)('A' + (arc4random_uniform(26))));
    NSString *randomStr = [[NSString alloc] initWithBytes:data length:length encoding:NSUTF8StringEncoding];
    free(data);
    
    return randomStr;
}

/**
 * hasSuccessFlag 检查字典里是否有key为'code'的item，且value为IM_SUCCESS
 * @param dataDict:数据字典
 * @returns true--有，false--没有
 */
+(BOOL) hasSuccessFlag:(NSMutableDictionary*)dataDict
{
    int value = [IMUtil getInt32FromDict:dataDict Key:@"code"];
    
    if (value == IM_Success)
    {
        return true;
    }
    return false;
}

+(int) getPayloadType:(NSMutableDictionary*)dataDict
{
    return [IMUtil getInt32FromDict:dataDict Key:@"msgid"];
}

/**
 * 查找pattern是否在字符串data中,忽略大小写的
 * @param pattern:想查找的字符串
 * @param data: 字符串
 * @returns true--存在, false--不存在
 */
+(BOOL) hasSubString:(NSString*)pattern Data:(NSString*)data
{
    NSString *lowerData = [data lowercaseString];
    NSString *lowerPattern = [pattern lowercaseString];
    NSRange range = [lowerData rangeOfString:lowerPattern];//判断字符串是否包含
    
    if (range.location == NSNotFound)
    {
        return false;
    }
    else
    {
        return true;
    }
}

/**
 * 从结果字典里查找整数值
 * @param dataDict: 数据字典
 * @param key: 数据在字典里的名称;
 * @returns 整数
 */
+(int) getInt32FromDict:(NSMutableDictionary*)dataDict Key:(NSString*)key
{
    NSObject *object = [dataDict objectForKey:key];
    if (object != nil && [object isKindOfClass:[NSNumber class]])
    {
        return [(NSNumber*)object intValue];
    }
    return -1;
}

/**
 * 从结果字典里查找整数值
 * @param dataDict: 数据字典
 * @param key: 数据在字典里的名称;
 * @returns 整数
 */
+(int64_t) getInt64FromDict:(NSMutableDictionary*)dataDict Key:(NSString*)key
{
    NSObject *object = [dataDict objectForKey:key];
    if (object != nil && [object isKindOfClass:[NSNumber class]])
    {
        return [(NSNumber*)object longLongValue];
    }
    return -1;
}

/**
 * 从结果字典里查找整数值
 * @param dataDict: 数据字典
 * @param key: 数据在字典里的名称;
 * @returns 整数
 */
+(NSString *) getStringFromDict:(NSMutableDictionary*)dataDict Key:(NSString*)key
{
    NSObject *object = [dataDict objectForKey:key];
    if (object != nil && [object isKindOfClass:[NSString class]])
    {
        return (NSString*)object;
    }
    return @"";
}

/**
 * 从结果字典里查找整数值
 * @param dataDict: 数据字典
 * @param key: 数据在字典里的名称;
 * @returns 整数
 */
+(NSData*) getDataFromDict:(NSMutableDictionary*)dataDict Key:(NSString*)key
{
    NSObject *object = [dataDict objectForKey:key];
    if (object != nil && [object isKindOfClass:[NSData class]])
    {
        return (NSData*)object;
    }
    return nil;
}

/**
 * 从结果字典里查找整数值
 * @param dataDict: 数据字典
 * @param key: 数据在字典里的名称;
 * @returns 整数
 */
+(BOOL) getBoolFromDict:(NSMutableDictionary*)dataDict Key:(NSString*)key
{
    NSObject *object = [dataDict objectForKey:key];
    if (object != nil && [object isKindOfClass:[NSNumber class]])
    {
        return [(NSNumber*)object boolValue];
    }
    return false;
}

/**
 * 将网络上接收到的二进制转化为long
 * @param data: 高位在前的二进制
 * @returns long型数据
 */
+(int64_t) getInt64FromNetData:(NSData*)data
{
    Byte buf[8];
    [data getBytes:buf length:8];
    int64_t result = buf[0];
    result = (result << 8) + buf[1];
    result = (result << 8) + buf[2];
    result = (result << 8) + buf[3];
    result = (result << 8) + buf[4];
    result = (result << 8) + buf[5];
    result = (result << 8) + buf[6];
    result = (result << 8) + buf[7];
    return result;
}

/**
 * 将网络上接收到的二进制转化为int32
 * @param data: 高位在前的二进制
 * @returns int32型数据
 */
+(int) getInt32FromNetData:(NSData*)data
{
    int dataLength = 0;
    [data getBytes:&dataLength length:4];
    return htonl(dataLength);
}

/**
 * 将状态码转化为状态名称
 * @param state: 状态码;
 * @returns 状态名称
 */
+(NSString*) getStateName:(IMUserState)state
{
    switch (state)
    {
        case IM_State_Init:
            return @"init";
            break;
        case IM_State_Connecting:
            return @"connecting";
            break;
        case IM_State_Connected:
            return @"connected";
            break;
        case IM_State_Disconnected:
            return @"disconnected";
            break;
        default:
            return @"unknown";
            break;
    }
    
    return @"";
}

/**
 * 将NSData转化为stl的string
 * @param src: NSData
 * @param dest: 返回值，即：std::string地址
 */
+(void) convertNSData:(NSData*)src toStlString:(void*)dest
{
    char *buf = new char[[src length]];
    [src getBytes:buf length:[src length]];
    std::string &strData = *((std::string*)dest);
    strData.assign(buf, [src length]);
    
    delete [] buf;
    buf = NULL;
}

/**
 * 将一个字典的数据拷贝到另一个字典
 */
+(void) copyDictionary:(NSMutableDictionary*)dest From:(NSDictionary*)src
{
    NSArray *keys = [src allKeys];// 所有key
    for(int i=0;i<[keys count];i++)
    {
        NSString *key = [keys objectAtIndex:i];
        [dest setObject:[src objectForKey:key] forKey:key];
    }
}

/**
 * 获取当前系统时间的秒值
 */
+(int) getCurSystemSecond
{
    NSInteger unitFlags = NSYearCalendarUnit | NSMonthCalendarUnit | NSDayCalendarUnit | NSWeekdayCalendarUnit | NSHourCalendarUnit | NSMinuteCalendarUnit | NSSecondCalendarUnit;
    NSCalendar *calendar = [NSCalendar currentCalendar];
    NSDateComponents *components = [calendar components:unitFlags fromDate:[NSDate date]];
    return (int)[components second];
}

/**
 * json编码
 */
+(NSData*) jsonEncode:(NSDictionary*)dict
{
    if ([NSJSONSerialization isValidJSONObject:dict])
    {
        NSError *error;
        return [NSJSONSerialization dataWithJSONObject:dict options:NSJSONWritingPrettyPrinted error:&error];
    }
    else
    {
        return nil;
    }
}

/**
 * json解码
 */
+(NSDictionary*) jsonDecode:(NSData*)data
{
    NSError *error;
    NSDictionary *json = [NSJSONSerialization JSONObjectWithData:data options:kNilOptions error:&error];
    if (json == nil || ![json isKindOfClass: [NSDictionary class]])
    {
//        CPLog(@"json parse failed \r\n");
        return [[NSDictionary alloc]init];
    }
    else
    {
        return json;
    }
}

/**
 * 检查字符串是否为空
 * @param str: NSString
 * @returns str为nil或长度为0时返回YES
 */
+(BOOL) isStringEmpty:(NSString*)str
{
    if (str == nil || [str length] == 0)
    {
        return YES;
    }
    
    return NO;
}

/**
 * 检查字符串是否为空
 * @param str: NSString
 * @param ignore: BOOL，是否忽略空白字符组成的字符串
 * @returns str为nil或长度为0或全有空白字符组成时返回YES
 */
+(BOOL) isStringEmpty:(NSString*)str ignoreWhiteapce:(BOOL)ignore
{
    if (str == nil || [str length] == 0)
    {
        return YES;
    }
    
    if (ignore != NO)
    {
        NSString* ts = [str stringByTrimmingCharactersInSet:[NSCharacterSet whitespaceAndNewlineCharacterSet]];
        if ([ts length] == 0)
        {
            return YES;
        }
    }
    
    return NO;
}

+ (NSString *)URLEncodedString:(NSString*)str
{
     return (NSString *)CFBridgingRelease(CFURLCreateStringByAddingPercentEscapes(kCFAllocatorDefault,(CFStringRef)str,NULL,CFSTR("!*'();:@&=+$,/?%#[]"),kCFStringEncodingUTF8));
}

+ (NSString*)URLDecodedString:(NSString*)str
{
    NSString *result = (NSString *)CFBridgingRelease(CFURLCreateStringByReplacingPercentEscapesUsingEncoding(kCFAllocatorDefault,(CFStringRef)str, CFSTR(""),kCFStringEncodingUTF8));
    return result;
}



/**
 * 将状态码转化为状态名称
 * @param state: 状态码;
 * @returns 状态名称
 */
+(NSString*) getDateString
{
    //获得系统时间
    NSDate * senddate=[NSDate date];
    NSDateFormatter *dateformatter=[[NSDateFormatter alloc] init];
    //[dateformatter setDateFormat:@"HH:mm"];
    [dateformatter setDateFormat:@"YYYYMMdd"];
    [dateformatter setLocale:[NSLocale currentLocale]];
    return [dateformatter stringFromDate:senddate];

}

+(NSString*) getOldDateString:(int)dayDiff
{
    int secOfDay = 3600*24;
    int diff = 0 - secOfDay*dayDiff;
    NSDate *senddate = [NSDate dateWithTimeIntervalSinceNow:diff];
    NSDateFormatter *dateformatter=[[NSDateFormatter alloc] init];
    //[dateformatter setDateFormat:@"HH:mm"];
    [dateformatter setDateFormat:@"YYYYMMdd"];
    [dateformatter setLocale:[NSLocale currentLocale]];
    return [dateformatter stringFromDate:senddate];
}

+(NSString*) getSystemTimeStamp
{
    //获得系统时间
    NSDate * senddate=[NSDate date];
    NSDateFormatter *dateformatter=[[NSDateFormatter alloc] init];
    //[dateformatter setDateFormat:@"HH:mm"];
    [dateformatter setDateFormat:@"YYYY-MM-dd HH:mm:ss.SSS"];
    [dateformatter setLocale:[NSLocale currentLocale]];
    return [dateformatter stringFromDate:senddate];
}


/**
 * 将NSData转化为stl string
 */
+(void)NSDataToStlString:(NSData*)data Return:(void*)result
{
    std::string *p1 = (std::string*)result;
    std::string &p2 = *p1;
    NSUInteger len = [data length];
    char *raw = new char[len];
    [data getBytes:raw length:len];
    
    p2.assign(raw, len);
    delete [] raw;
    raw = NULL;
}

+(NSData*) zipData:(NSData*) data
{
    return [data gzippedData];
}

+(NSData*) unzipData:(NSData*) data
{
    return [data gunzippedData];
}

/**
 * md5加密
 */
+ (NSString *) getMd5_32Bit_String:(NSString *)srcString
{
    const char *cStr = [srcString UTF8String ];
    unsigned char digest[ CC_MD5_DIGEST_LENGTH ];
    CC_MD5 ( cStr, strlen (cStr), digest );
    NSMutableString *result = [ NSMutableString stringWithCapacity : CC_MD5_DIGEST_LENGTH * 2 ];

    for ( int i = 0 ; i < CC_MD5_DIGEST_LENGTH ; i++)
        [result appendFormat : @"%02x" , digest[i]];

    return result;
}

/**
 * 生成操作序列号。一秒钟内可以生成100000个不重复的序列号
 * @return 生成的操作序列号
 */
+(int64_t) createSN
{
    if (snLock == nil) {
        snLock = [[NSLock alloc]init];
    }

    int64_t ret;
    @synchronized (snLock)
    {
        ret = (int64_t)[[NSDate date] timeIntervalSince1970] * 100000;
        if (lastTimeSatmp != ret)
        {
            lastTimeSatmp = ret;
            opCount = 0;
        }
        ++opCount;

        ret += opCount;
    }
    return ret;
}

+(NSString *)getCurTimeString
{
    //获得系统时间
    NSDate * senddate=[NSDate date];
    NSDateFormatter *dateformatter=[[NSDateFormatter alloc] init];
        //[dateformatter setDateFormat:@"HH:mm"];
    [dateformatter setDateFormat:@"YYYY-MM-dd HH:mm:ss.SSS"];
    [dateformatter setLocale:[NSLocale currentLocale]];
    return [dateformatter stringFromDate:senddate];
}

/**
 * 解析id类型，总是获得合法的NSString对象
 */
+ (NSString *)idToString:(id)parse
{
    if (!parse) {
        return @"";
    }
    
    if ([parse isKindOfClass:[NSString class]]) {
        return parse;
    }
    else if ([parse isKindOfClass:[NSData class]]) {
        return [[NSString alloc] initWithData:parse encoding:NSUTF8StringEncoding];
    }

    return  [NSString stringWithFormat:@"%@",parse];
}

/**
 * 解析id类型，总是获得合法的NSData对象
 */
+ (NSData *)idToData:(id)parse
{
    if (!parse) {
        return [NSData data];
    }

    if ([parse isKindOfClass:[NSData class]]) {
        return parse;
    }

    return  [NSData data];
}


@end
