//
//  ConfigData.m
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-7.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMConfigData.h"
#import "IMServerAddress.h"

@interface IMConfigData()

/**
 * preferred address
 */
@property (atomic, strong) NSMutableArray *preferredIPList;
@property (atomic, strong) NSMutableArray *preferredPortList;
@property (atomic, assign) int preferredIPIndex;

@end

/**
 * store im server needs all config data
 */
@implementation IMConfigData

#pragma mark - property list

/**
 * synthesize property list
 */
@synthesize phone = _phone;
@synthesize jid = _jid;
@synthesize password = _password;
@synthesize appid = _appid;
@synthesize version = _version;
@synthesize qid = _qid;
@synthesize preferredIPList = _preferredIPList;
@synthesize preferredPortList = _preferredPortList;
@synthesize preferredIPIndex = _preferredIPIndex;
@synthesize msgIPList = _msgIPList;
@synthesize msgPortList = _msgPortList;
@synthesize curIPIndex = _curIPIndex;
@synthesize registerServer = _registerServer;
@synthesize uploadServer = _uploadServer;
@synthesize downloadServer = _downloadServer;
@synthesize webServer = _webServer;
@synthesize defaultKey = _defaultKey;
@synthesize randomKey = _randomKey;
@synthesize sessionKey = _sessionKey;
@synthesize deviceToken = _deviceToken;
@synthesize hbInterval = _hbInterval;
@synthesize protocolVersion = _protocolVersion;
@synthesize verifyCode = _verifyCode;
@synthesize isOnline = _isOnline;
@synthesize sigToken = _sigToken;

#pragma mark - function list

/**
 * init function
 */
-(id) init
{
    self = [super init];
    if (self)
    {
        self.isOnline = YES;
        self.preferredIPIndex = 0;
        self.preferredIPList = [[NSMutableArray alloc] init];
        self.preferredPortList = [[NSMutableArray alloc] init];
        self.msgIPList = [[NSMutableArray alloc] init];
        self.msgPortList = [[NSMutableArray alloc]init];
        self.curIPIndex = 0;
    }
    return self;

}

/**
 * 检查是否可以执行上行注册了,只要有phone，registry server，default key，appid和version就可以发起上行注册
 * returns true--yes, false--no
 */
-(BOOL) isReadyForUpRegistry
{
    if ([self.phone length] > 0 &&
        [self.registerServer length] > 0 &&
        [self.defaultKey length] > 0 &&
        self.appid > 0 &&
        self.version > 0)
    {
        return true;
    }
    return false;
}

/*
 * 循环获得长连接server IP数量
 */
-(NSUInteger) totalServerIpCount
{
    @synchronized (self) {
        return self.preferredIPList.count + self.msgIPList.count;
    }
}

/*
 * 增加一个优先服务器地址
 * @param host:优先服务器地址
 * @param port:优先服务器端口
 */
-(void) addPreferredHost:(NSString*)host Port:(NSNumber*)port
{
    if (host == nil || port == nil) {
        return;
    }
    
    @synchronized (self) {
        [self.preferredIPList addObject:host];
        [self.preferredPortList addObject:port];
    }
}

/*
 * 是否有优先服务器地址
 */
-(BOOL) hasPreferredHost
{
    @synchronized (self) {
        return self.preferredIPList.count > 0;
    }
}

/*
 * 重置服务器设置
 */
-(void) resetServerIP
{
    @synchronized (self) {
        [self.preferredIPList removeAllObjects];
        [self.preferredPortList removeAllObjects];
        self.preferredIPIndex = 0;
        self.curIPIndex = 0;
    }
}

/*
 * 循环获得长连接server IP
 */
-(NSString*) getServerIP
{
    int i = self.curIPIndex % [self.msgIPList count] ;
    
    @synchronized (self) {
        // 如果存在优先地址，先尝试连接优先地址
        if (self.preferredIPIndex < self.preferredIPList.count) {
            int i = self.preferredIPIndex;
            self.preferredIPIndex = self.preferredIPIndex + 1;
            
            self.loginPort = [[self.preferredPortList objectAtIndex:i] intValue];
            return [self.preferredIPList objectAtIndex:i];
        }
        
        self.curIPIndex = self.curIPIndex + 1;
        self.loginPort = [[self.msgPortList objectAtIndex:i]intValue];
        return [self.msgIPList objectAtIndex:i];
    }
}

/**
 * 检查是否可以执行下行注册了,只要有phone，registry server，default key，appid和version就可以发起上行注册
 * returns true--yes, false--no
 */
-(BOOL) isReadyForDownRegistry
{
    return [self isReadyForUpRegistry];
}

/**
 * 检查是否可以执行qid注册了,只要有 registry server，default key，appid和version就可以发起上行注册
 * returns true--yes, false--no
 */
-(BOOL) isReadyForQidRegistry
{
    if ([self.qid length] > 0 &&
        [self.registerServer length] > 0 &&
        [self.defaultKey length] > 0 &&
        self.appid > 0 &&
        self.version > 0)
    {
        return true;
    }

    return false;
}

/*
 * 根据appid设置默认配置
 */
-(void) loadDefaultConf:(int)appid defalutKey:(NSString*)defalutKey version:(int)version serverList:(NSArray<IMServerAddress*>*)serverList dispatcherServer:(IMServerAddress*)dispatcherServer
{
    self.protocolVersion = 1;
    self.defaultKey = defalutKey; //102
    self.hbInterval = 300; //5 minutes
    self.appid = appid;
    self.version = version;
    self.qid = @"";
    
    for (IMServerAddress *server in serverList) {
        for (NSString *aPort in server.ports) {
            [self.msgIPList addObject:server.address];
            [self.msgPortList addObject:aPort];
        }
    }
    
    self.dispatcherServer = dispatcherServer.address;
}

@end
