//
//  IMProtoHelper.m
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-24.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMProtoHelper.h"
#import <string>
#import "IMUtil.h"
#import "IMUser.h"

using namespace std;


@implementation IMProtoHelper

/**
 * property list
 */
#pragma mark - property list
@synthesize userConfig = _userConfig;
@synthesize imUser = _imUser;
@synthesize cmdReqIDToNameDict = _cmdReqIDToNameDict;
@synthesize cmdReqNameToIDDict = _cmdReqNameToIDDict;
@synthesize cmdResIDToNameDict = _cmdResIDToNameDict;
@synthesize cmdResNameToIDDict = _cmdResNameToIDDict;


/**
 * function list
 */
#pragma mark - function list

/**
 *
 * @param config: user config data
 * @returns IMProtoHelper instance
 */
-(id)initWithUser:(IMUser*)imUser
{
    self = [super init];
    if (self)
    {
        _imUser = imUser;
        _userConfig = imUser.configData;
        
        self.cmdReqIDToNameDict = [[NSMutableDictionary alloc] init];
        self.cmdReqNameToIDDict = [[NSMutableDictionary alloc] init];
        self.cmdResIDToNameDict = [[NSMutableDictionary alloc] init];
        self.cmdResNameToIDDict = [[NSMutableDictionary alloc] init];
        
        [self initData];
    }
    return self;
}

/**
 * 初始化数据，子类可以重载该函数并实现自己的初始化工作
 */
-(void) initData
{
    
}

/**
 * 添加命令名称与id的对应关系
 * @param name: 命令名称
 * @param code: 命令id
 */
-(void) addRequestMap:(NSString*)name ID:(int)code
{
    NSNumber *key = [NSNumber numberWithInt:code];
    NSString *value = [self.cmdReqIDToNameDict objectForKey:key];
    if (value != nil) //exists, remove it
    {
        [self.cmdReqIDToNameDict removeObjectForKey:key];
        [self.cmdReqNameToIDDict removeObjectForKey:value];
    }
    
    [self.cmdReqIDToNameDict setObject:name forKey:key];
    [self.cmdReqNameToIDDict setObject:key forKey:name];
}

/**
 * 根据id获得名称
 * @param code: 命令id
 * @returns 命令名称
 */
-(NSString*) getRequestNameByID:(int)code
{
    NSNumber *key = [NSNumber numberWithInt:code];
    NSString *value = [self.cmdReqIDToNameDict objectForKey:key];
    if (value == nil)
    {
        return @"";
    }
    else
    {
        return value;
    }
}

/**
 * 根据名称获得id
 * @param name:proto文件中的命令名称，不区分大小写
 * @returns 命令id
 */
-(int) getRequestIDByName:(NSString*)name
{
    NSNumber *value = [self.cmdReqNameToIDDict objectForKey:name];
    if (value == nil)
    {
        return 0;
    }
    else
    {
        return [value intValue];
    }
}

/**
 * 添加命令名称与id的对应关系
 * @param name: 命令名称
 * @param code: 命令id
 */
-(void) addResponseMap:(NSString*)name ID:(int)code
{
    NSNumber *key = [NSNumber numberWithInt:code];
    NSString *value = [self.cmdResIDToNameDict objectForKey:key];
    if (value != nil) //exists, remove it
    {
        [self.cmdResIDToNameDict removeObjectForKey:key];
        [self.cmdResNameToIDDict removeObjectForKey:value];
    }
    
    [self.cmdResIDToNameDict setObject:name forKey:key];
    [self.cmdResNameToIDDict setObject:key forKey:name];
}

/**
 * 根据id获得名称
 * @param code: 命令id
 * @returns 命令名称
 */
-(NSString*) getResponseNameByID:(int)code
{
    NSNumber *key = [NSNumber numberWithInt:code];
    NSString *value = [self.cmdResIDToNameDict objectForKey:key];
    if (value == nil)
    {
        return @"";
    }
    else
    {
        return value;
    }
}

/**
 * 根据名称获得id
 * @param name:proto文件中的命令名称，不区分大小写
 * @returns 命令id
 */
-(int) getResponseIDByName:(NSString*)name
{
    NSNumber *value = [self.cmdResNameToIDDict objectForKey:name];
    if (value == nil)
    {
        return 0;
    }
    else
    {
        return [value intValue];
    }
}

/**
 * 讲一个二进制打包为Flag+Length+Body格式,因为只有initLogin才会使用default key加密
 * 里面会自动加上Flag
 * @param data: 需要发出去的二进制流(std::string *);
 */
-(NSData*) createDefaultKeyOutData:(void*)data
{
    /*
     C ⇒ S, 第一个包的格式如下:
     magic(12bytes) + len(4bytes) + Message(protobuf)
     magic = flag(2bytes) + protocol_version(4bits) + client_version(12bits) + appid(2bytes) + reserved(6bytes)
     flag = "qh"
     protocol_version: 协议版本号， 填1
     client_version: 客户端版本号，填102
     appid: 应用ID， 填2010 (悄悄)
     len = magic_len + length of ‘len’ + message_len
     */
    NSString *flag = @"qh";
    std::string *input = (std::string*)data;
    std::string result;
    NSString *key = self.userConfig.defaultKey;
    [IMUtil rc4EncodeFromString:data Key:key Return:&result];
    
    
    
    NSMutableData *stream = [[NSMutableData alloc] initWithLength:0];
    //[step1]: flag(2bytes)
    [stream appendData:[IMUtil NSStringToNSData:flag]];
    //[step2]: protocol_version(4bits) + client version(12bits)
    Byte protoVer = self.userConfig.protocolVersion;
    int clientVer = self.userConfig.version;
    Byte version[2];
    version[0] = ((protoVer & 0xF) << 4) | ((clientVer & 0xF00) >> 8);
    version[1] = clientVer & 0xFF;
    NSData *protoData=[NSData dataWithBytes:version length:2];
    [stream appendData:protoData];
    //[step4]: appid(2bytes)
    short appid = self.userConfig.appid;
    Byte appID[2];
    appID[0] = ((appid & 0xFF00) >> 8);
    appID[1] = appid & 0xFF;
    NSData *appidData = [NSData dataWithBytes:appID length:2];
    [stream appendData:appidData];
    //[step5]: reserve(6bytes)
    NSString *random = [IMUtil createRandomString:6];
    NSData *randomData = [IMUtil NSStringToNSData:random];
    [stream appendData:randomData];
    //[step6]: length
    int dataLength = ntohl(16 + input->size());
    NSData *lengthData=[NSData dataWithBytes:&dataLength length:sizeof(int)];
    [stream appendData:lengthData];
    //[step7]: body
    NSData *body = [NSData dataWithBytes:result.c_str() length:result.size()];
    [stream appendData:body];
    
    return stream;

}

/**
 * 讲一个二进制打包为Length+Body格式
 * @param data: 需要发出去的二进制流(std::string *);
 */
-(NSData*) createEncryptData:(void*)data Key:(NSString*)key
{
    std::string *input = (std::string*)data;
    std::string result;
    
    //NSString *key = self.userConfig.password;
    [IMUtil rc4EncodeFromString:data Key:key Return:&result];
    
    //[step1]: add length
    int dataLength = ntohl(4 + input->size());
    NSData *lengthData=[NSData dataWithBytes:&dataLength length:sizeof(int)];
    NSMutableData *stream = [[NSMutableData alloc] initWithLength:0];
    [stream appendData:lengthData];
    //[step2]:
    NSData *body = [NSData dataWithBytes:result.c_str() length:result.size()];
    [stream appendData:body];
    
    return stream;
}

/**
 * 讲一个二进制打包为Length+Body格式
 * @param data: 需要发出去的二进制流(std::string *);
 */
-(NSData*) createSessionKeyOutData:(void*)data
{
    std::string *input = (std::string*)data;
    std::string result;
    
    NSString *key = self.userConfig.sessionKey;
    if (key != nil && key.length > 0) {
        [IMUtil rc4EncodeFromString:data Key:key Return:&result];
    }
    else {// 服务器降级的时候登录时会返回session key为空字符串，不加密
        result = *input;
    }
    
    //[step1]: add length
    int dataLen = 4 + input->size();
    int dataLength = ntohl(dataLen);
    //CPLog(@"src len:%d, net len:%d", dataLen, dataLength);
    NSData *lengthData=[NSData dataWithBytes:&dataLength length:sizeof(int)];
    NSMutableData *stream = [[NSMutableData alloc] initWithLength:0];
    [stream appendData:lengthData];
    //[step2]: 
    NSData *body = [NSData dataWithBytes:result.c_str() length:result.size()];
    [stream appendData:body];
    
    return stream;
}

-(NSData*) createSessionKeyOutDataWithStrData:(NSData*)data
{
    if (data == nil) {
        return [NSData data];
    }
    
    std::string unzipData;
    [IMUtil NSDataToStlString:data Return:&unzipData];
    
    std::string *input = &unzipData;
    std::string result;
    
    NSString *key = self.userConfig.sessionKey;
    if (key != nil && key.length > 0) {
        [IMUtil rc4EncodeFromString:input Key:key Return:&result];
    }
    else {// 服务器降级的时候登录时会返回session key为空字符串，不加密
        result = *input;
    }
    
    
    //[step1]: add length
    int dataLen = 4 + input->size();
    int dataLength = ntohl(dataLen);
    //CPLog(@"src len:%d, net len:%d", dataLen, dataLength);
    NSData *lengthData=[NSData dataWithBytes:&dataLength length:sizeof(int)];
    NSMutableData *stream = [[NSMutableData alloc] initWithLength:0];
    [stream appendData:lengthData];
    //[step2]:
    NSData *body = [NSData dataWithBytes:result.c_str() length:result.size()];
    [stream appendData:body];
    
    return stream;
}




@end
