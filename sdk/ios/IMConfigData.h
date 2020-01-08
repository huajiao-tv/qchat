//
//  ConfigData.h
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-7.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import <Foundation/Foundation.h>

@class IMServerAddress;

/**
 * @class ConfigData
 * @brief store im server needs all config data
 */
@interface IMConfigData : NSObject

#pragma mark - property list

/**
 * store user phone number
 */
@property (atomic, strong) NSString *phone;

/**
 * store user jid
 */
@property (atomic, strong) NSString *jid;

/**
 * store user password
 */
@property (atomic, strong) NSString *password;

/**
 * product identify
 */
@property (atomic, assign) int appid;

/**
 * store proto buff protocol version, it's not sdk version
 */
@property (atomic, assign) int version;

/**
 * qihoo id
 */
@property (atomic, strong) NSString *qid;

/**
 * default address
 */
@property (atomic, strong) NSMutableArray *msgIPList;
@property (atomic, strong) NSMutableArray *msgPortList;

@property (atomic, assign) int curIPIndex;

/**
 * login server's host
 */
@property (atomic, strong) NSString* loginServer;
/**
 * login server's port
 */
@property (atomic, assign) int loginPort;

/**
 * http dispatcher server address
 */
@property (atomic, strong) NSString *dispatcherServer;

/**
 * http register server address
 */
@property (atomic, strong) NSString *registerServer;

/**
 * upload picture server address
 */
@property (atomic, strong) NSString *uploadServer;

/**
 * download picture server address
 */
@property (atomic, strong) NSString *downloadServer;

/**
 * web interface serve address
 */
@property (atomic, strong) NSString *webServer;

/**
 * default key
 */
@property (atomic, strong) NSString *defaultKey;

/**
 * random key which is from server
 */
@property (atomic, strong) NSString *randomKey;

/**
 * session key
 */
@property (atomic, strong) NSString *sessionKey;

/**
 * ios device token
 */
@property (atomic, strong) NSString *deviceToken;

/**
 * heart beat interval(second)
 */
@property (atomic, assign) int hbInterval;

/**
 * msg router protocol version
 */
@property (atomic, assign) Byte protocolVersion;

/**
 * in down registry, server send verify code through sms
 */
@property (atomic, strong) NSString *verifyCode;

/**
 * 是否连接线上服务器
 */
@property (nonatomic, assign) BOOL isOnline;

/**
 * 非游客用户的token
 */
@property (nonatomic, strong) NSString *sigToken;

#pragma mark - function list

/**
 * init config data instance
 */
-(id) init;

/**
 * 检查是否可以执行上行注册了,只要有phone，registry server，default key，appid和version就可以发起上行注册
 * returns true--yes, false--no
 */
-(BOOL) isReadyForUpRegistry;

/**
 * 检查是否可以执行下行注册了,只要有phone，registry server，default key，appid和version就可以发起上行注册
 * returns true--yes, false--no
 */
-(BOOL) isReadyForDownRegistry;

/**
 * 检查是否可以执行qid注册了,只要有qid, registry server，default key，appid和version就可以发起上行注册
 * returns true--yes, false--no
 */
-(BOOL) isReadyForQidRegistry;

/*
 * 根据appid设置默认配置
 */
-(void) loadDefaultConf:(int)appid defalutKey:(NSString*)defalutKey version:(int)version serverList:(NSArray<IMServerAddress*>*)serverList dispatcherServer:(IMServerAddress*)dispatcherServer;

/*
 * 循环获得长连接server IP数量
 */
-(NSUInteger) totalServerIpCount;

/*
 * 增加一个优先服务器地址
 * @param host:优先服务器地址
 * @param port:优先服务器端口
 */
-(void) addPreferredHost:(NSString*)host Port:(NSNumber*)port;

/*
 * 是否有优先服务器地址
 */
-(BOOL) hasPreferredHost;

/*
 * 重置服务器设置
 */
-(void) resetServerIP;

/*
 * 循环获得长连接server IP
 */
-(NSString*) getServerIP;

@end
