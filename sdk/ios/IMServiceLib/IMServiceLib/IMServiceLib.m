//
//  IMServiceLib.m
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-7.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMServiceLib.h"
#import "IMConfigData.h"
#import "IMMessage.h"
#import "IMProtoHelper.h"
#import "IMFeatureWebInterface.h"
#import "IMFeatureUserInfo.h"
#import "IMFeaturePeerChat.h"
#import "IMFeatureGroupChat.h"
#import "IMFeatureChatRoom.h"
#import "IMFeatureCircleChat.h"
#import "IMFeatureRelation.h"
#import "MD5Digest.h"
#import "asihttp/ASIReachability.h"
#import "IMUtil.h"
#import <CoreTelephony/CTTelephonyNetworkInfo.h>


@interface IMServiceLib()

//user实例
@property (atomic, strong) IMUser *imUser;

@property (atomic, assign) BOOL isCloseLog;

/**
 * 检车网络是否可达的方法
 */
@property (atomic, strong) ASIReachability *reacher;

/**
 * print log file
 */
@property (atomic, strong) NSString *curLogFile;

@property (atomic, strong) NSFileHandle *logFileHandle;

/**
 * 获取日志文件目录
 */
-(NSString *) getLogsDirectory;
/**
 * 创建日志文件目录
 */
-(BOOL) createLogsDirectory;
/**
 * 获取当前日志文件全路径
 */
-(NSString *) getLogFilePathName;
/**
 * 获取当前日志文件句柄
 */
-(id) getLogFileHandle;
/**
 * 遍历旧的日志文件
 */
-(void) printOldLogFile;
/**
 * 移除旧的日志文件
 */
-(void) removeOldLogFile;
/**
 * 打印当前日志文件内容到输出窗口
 */
-(void) dumpCurrentLogFile;
/**
 * 打开指定日志文件
 * @param filePath: 日志文件全路径
 */
-(NSFileHandle*) openLogFile:(NSString *) filePath;

@end

@implementation IMServiceLib

@synthesize imUser = _imUser;
@synthesize reacher = _reacher;
@synthesize curLogFile = _curLogFile;
@synthesize logFileHandle = _logFileHandle;

/**
 * 静态变量
 */
static IMServiceLib *singleInstance = nil;

// All logging statements are added to the same queue to ensure FIFO operation.
static dispatch_queue_t _loggingQueue;

// Individual loggers are executed concurrently per log statement.
// Each logger has it's own associated queue, and a dispatch group is used for synchrnoization.
static dispatch_group_t _loggingGroup;

#pragma mark - function list


/**
 * 单例方法
 */
+(IMServiceLib*) sharedInstance
{
    @synchronized(self)
    {
        if (!singleInstance)
        {
            singleInstance = [[self alloc] init];
        }
        return singleInstance;
    }
    
    return singleInstance;
}

/**
 * The runtime sends initialize to each class in a program exactly one time just before the class,
 * or any class that inherits from it, is sent its first message from within the program. (Thus the
 * method may never be invoked if the class is not used.) The runtime sends the initialize message to
 * classes in a thread-safe manner. Superclasses receive this message before their subclasses.
 *
 * This method may also be called directly (assumably by accident), hence the safety mechanism.
 **/
+ (void)initialize {
    static dispatch_once_t DDLogOnceToken;
    
    dispatch_once(&DDLogOnceToken, ^{
//        NSLog(@"File Log: Using grand central dispatch");
        
        _loggingQueue = dispatch_queue_create("cocoa.IMServiceLib", NULL);
        _loggingGroup = dispatch_group_create();
        
    });
}


/**
 *重写方法，防止重复创建
 */
+(id) allocWithZone:(struct _NSZone *)zone
{
    @synchronized(self)
    {
        if (singleInstance == nil)
        {
            singleInstance = [super allocWithZone:zone];
            return singleInstance;
        }
    }
    return nil;
}

/**
 * 初始化函数
 */
-(id) init
{
    self = [super init];
    if (self)
    {
        self.imUser = nil;
        self.isCloseLog = NO;
        self.curLogFile = @"";
        self.logFileHandle = nil;
        
        // 打开日志文件句柄
        [self getLogFileHandle];
        
        //[step1] 开启网络监听
        //self.reacher = [Reachability reachabilityWithHostName:@"http://www.apple.com"];
        self.reacher = [ASIReachability reachabilityForInternetConnection];
        [[NSNotificationCenter defaultCenter] addObserver:self
                                                 selector:@selector(reachabilityChanged:)
                                                     name:kASIReachabilityChangedNotification
                                                   object:nil];
        [self.reacher startNotifier];
        
    }
    
    return self;
}

/**
 * 资源回收函数
 */
-(void) dealloc
{
    @synchronized(self)
    {
        if (self.logFileHandle != nil)
        {
            [self.logFileHandle closeFile];
            self.logFileHandle = nil;
        }
        
        self.curLogFile = @"";
    }
}

/**
 * 判断网络是否可达
 */
-(BOOL) isNetworkReachable
{
    if (self.reacher.currentReachabilityStatus != NotReachable)
    {
        return YES;
    }
    else
    {
        return NO;
    }
}

/**
 * 判断网络是否是gprs
 */
-(BOOL) isGPRSNetwork
{
    if (self.reacher.currentReachabilityStatus == ReachableViaWWAN)
    {
        return YES;
    }
    else
    {
        return NO;
    }
}

/**
 * 判断网络是否是gprs
 */
-(BOOL) isWifiNetwork
{
    if (self.reacher.currentReachabilityStatus == ReachableViaWiFi)
    {
        return YES;
    }
    else
    {
        return NO;
    }
}

/**
 * 判断网络类型：0:unkonwn 1:2g 2:3g 3:wifi 4:ethe 5: 4G LTE
 */
-(int) networkType
{
    //创建零地址，0.0.0.0的地址表示查询本机的网络连接状态
    struct sockaddr_storage zeroAddress;
    
    bzero(&zeroAddress, sizeof(zeroAddress));
    zeroAddress.ss_len = sizeof(zeroAddress);
    zeroAddress.ss_family = AF_INET;
    
    // Recover reachability flags
    SCNetworkReachabilityRef defaultRouteReachability = SCNetworkReachabilityCreateWithAddress(NULL, (struct sockaddr *)&zeroAddress);
    SCNetworkReachabilityFlags flags;
    
    //获得连接的标志
    BOOL didRetrieveFlags = SCNetworkReachabilityGetFlags(defaultRouteReachability, &flags);
    CFRelease(defaultRouteReachability);
    
    //如果不能获取连接标志，则不能连接网络，直接返回
    if (!didRetrieveFlags)
    {
        return 0;
    }
    
    int nt = 0;
    if ((flags & kSCNetworkReachabilityFlagsConnectionRequired) == 0)
    {
        // if target host is reachable and no connection is required
        // then we'll assume (for now) that your on Wi-Fi
        nt = 3; // "WIFI";
    }
    
    if (
        ((flags & kSCNetworkReachabilityFlagsConnectionOnDemand ) != 0) ||
        (flags & kSCNetworkReachabilityFlagsConnectionOnTraffic) != 0
        )
    {
        // ... and the connection is on-demand (or on-traffic) if the
        // calling application is using the CFSocketStream or higher APIs
        if ((flags & kSCNetworkReachabilityFlagsInterventionRequired) == 0)
        {
            // ... and no [user] intervention is needed
            nt = 3; // "WIFI";
        }
    }
    
    if ((flags & kSCNetworkReachabilityFlagsIsWWAN) == kSCNetworkReachabilityFlagsIsWWAN)
    {
        nt = 1;
        
        if ([[[UIDevice currentDevice] systemVersion] floatValue] >= 7.0)
        {
            CTTelephonyNetworkInfo * info = [[CTTelephonyNetworkInfo alloc] init];
            NSString *currentRadioAccessTechnology = info.currentRadioAccessTechnology;
            
            if (currentRadioAccessTechnology)
            {
                if ([currentRadioAccessTechnology isEqualToString:CTRadioAccessTechnologyLTE])
                {
                    nt = 5; //  "4G";
                }
                else if ([currentRadioAccessTechnology isEqualToString:CTRadioAccessTechnologyEdge] || [currentRadioAccessTechnology isEqualToString:CTRadioAccessTechnologyGPRS])
                {
                    nt = 1; // "2G";
                }
                else
                {
                    nt = 2; // "3G";
                }
            }
        }
        else
        {
            if((flags & kSCNetworkReachabilityFlagsReachable) == kSCNetworkReachabilityFlagsReachable)
            {
                if ((flags & kSCNetworkReachabilityFlagsTransientConnection) == kSCNetworkReachabilityFlagsTransientConnection)
                {
                    if((flags & kSCNetworkReachabilityFlagsConnectionRequired) == kSCNetworkReachabilityFlagsConnectionRequired)
                    {
                        nt = 1; // "2G";
                    }
                    else
                    {
                        nt = 2; // "3G";
                    }
                }
            }
        }
    }
    
    return nt;
}

/**
 * 系统网络切换回调函数
 */
- (void) reachabilityChanged:(id)reachObj
{
    if (self.reacher != nil)
    {
        if(self.reacher.currentReachabilityStatus == NotReachable)
        {
//            NSLog(@"network power off");
            if (self.imUser != nil)
            {
                [self.imUser reachabilityChanged:0];
            }
        }
        else
        {
            if (self.reacher.currentReachabilityStatus == ReachableViaWiFi)
            {
//                NSLog(@"network power on through wifi");
                if (self.imUser != nil)
                {
                    [self.imUser reachabilityChanged:2];
                }
            }
            else if (self.reacher.currentReachabilityStatus == ReachableViaWWAN)
            {
//                NSLog(@"network power on through mobile");
                if (self.imUser != nil)
                {
                    [self.imUser reachabilityChanged:1];
                }
            }
        }
    }
}

-(IMUser*) createUser:(NSString*)userID token:(NSString*)token deviceID:(NSString*)deviceID appID:(int)appid defalutKey:(NSString*)defalutKey version:(int)version serverList:(NSArray<IMServerAddress*>*)serverList  dispatcherServer:(IMServerAddress*)dispatcherServer withDelegate:(id<IMNotifyDelegate>)delegate
{
    return [self createUser:userID token:token sig:token deviceID:deviceID appID:appid defalutKey:defalutKey version:version serverList:serverList dispatcherServer:dispatcherServer withDelegate:delegate];
}
-(IMUser*) createUser:(NSString*)userID token:(NSString*)token sig:(NSString*)sig deviceID:(NSString*)deviceID appID:(int)appid defalutKey:(NSString*)defalutKey version:(int)version serverList:(NSArray<IMServerAddress*>*)serverList  dispatcherServer:(IMServerAddress*)dispatcherServer withDelegate:(id<IMNotifyDelegate>)delegate
{
    if (self.imUser != nil)
    {
        if ([userID isEqual:self.imUser.configData.jid])
        {
            return self.imUser;
        }
        else
        {
            //如果用户已经存在，则关闭连接
            [self.imUser stopService];
            self.imUser = nil;
        }
    }
    
    if (self.imUser == nil)
    {
        self.imUser = [[IMUser alloc] initWithDelegate:delegate];
        self.imUser.configData.jid = userID;
        self.imUser.configData.password = token;
        self.imUser.configData.sigToken = sig;
        self.imUser.configData.deviceToken = deviceID;
        [self.imUser.configData loadDefaultConf:appid defalutKey:defalutKey version:version serverList:serverList dispatcherServer:dispatcherServer];
    }
    else
    {
//        NSLog(@"user:%@ exist", userID);
    }
    
    return self.imUser;
}

/**
 * 删除tantan当天用户
 */
-(void) shutdownUser
{
    if (self.imUser != nil)
    {
        [self.imUser stopService];
        self.imUser = nil;
    }
}

/**
 * 返回谈谈的当前用户
 */
-(IMUser*) curUser
{
    return self.imUser;
}

/**
 * 获取日志文件目录
 */
-(NSString *) getLogsDirectory
{
    //获得应用程序沙盒的Documents目录，官方推荐数据文件保存在此
    NSArray *path = NSSearchPathForDirectoriesInDomains(NSCachesDirectory, NSUserDomainMask, YES);
    NSString* logs_dir = [path objectAtIndex:0];
    logs_dir = [logs_dir stringByAppendingPathComponent:@"sdkr_logs"];
    return logs_dir;
}

/**
 * 创建日志文件目录
 */
-(BOOL) createLogsDirectory
{
    NSString* logs_dir = [self getLogsDirectory];
    NSFileManager *fileManager = [NSFileManager defaultManager];
    BOOL is_dir = NO;
    if ([fileManager fileExistsAtPath:logs_dir isDirectory:&is_dir] && is_dir != YES)
    {
        // 应用程序沙盒的Documents目录的sdkr_logs为指定目录，不应该是文件
        [fileManager removeItemAtPath:logs_dir error:nil];
//        NSLog(@"removed %@ for it should be a directory but actually it was a file.", logs_dir);
    }
    
    NSError* error;
    BOOL result = [fileManager createDirectoryAtPath:logs_dir withIntermediateDirectories:YES attributes:nil error:&error];
    if (result != YES)
    {
//        NSLog(@"FileManager created logs directory failed, error information: %@.", [error description]);
    }
    else
    {
//        NSLog(@"FileManager created logs directory successfully");
    }
    
    return result;
}

/**
 * 获取当前日志文件全路径
 */
-(NSString*) getLogFilePathName
{
    NSString *logfile = [NSString stringWithFormat:@"log_%@.log", [IMUtil getDateString]];
    
    NSString* logs_dir = [self getLogsDirectory];
    
    NSString *fileFullPath= [logs_dir stringByAppendingPathComponent:logfile];
    //NSLog(@"log file:%@", fileFullPath);
    return fileFullPath;
}

/**
 * 打开指定日志文件
 * @param filePath: 日志文件全路径
 */
-(NSFileHandle*) openLogFile:(NSString *) filePath
{
    NSFileManager *fileManager = [NSFileManager defaultManager];
    if ([fileManager fileExistsAtPath:filePath] != YES)
    {
        [[NSFileManager defaultManager]createFileAtPath:filePath contents:nil attributes:nil];
    }
    return [NSFileHandle fileHandleForUpdatingAtPath:filePath];
}

/**
 * 获取当前日志文件句柄
 */
-(id) getLogFileHandle
{
    // 总是创建日志文件目录
    [self createLogsDirectory];
    
    NSString *logFile = [self getLogFilePathName];
    if ([self.curLogFile length] == 0)
    {
        self.curLogFile = logFile;
        self.logFileHandle = [self openLogFile:self.curLogFile];
        [self.logFileHandle seekToEndOfFile];
    }
    else if ([self.curLogFile isEqualToString:logFile])
    {
        if (self.logFileHandle == nil)
        {
            self.logFileHandle = [self openLogFile:self.curLogFile];
            [self.logFileHandle seekToEndOfFile];
        }
    }
    else
    {
//        NSLog(@"cur log file:%@, logfile:%@", self.curLogFile, logFile);
        if (self.logFileHandle != nil)
        {
            [self.logFileHandle closeFile];
        }
        self.curLogFile = logFile;
        self.logFileHandle = [self openLogFile:self.curLogFile];
        [self.logFileHandle seekToEndOfFile];
    }
    
    [self removeOldLogFile];
    return self.logFileHandle;
}

/**
 * 打印当前日志文件内容到输出窗口
 */
-(void) dumpCurrentLogFile
{
    NSData *content  = [self getCurrentLogFileContent];
//    NSLog(@"dump log file:\n%@", [IMUtil NSDataToNSString:content]);
}

/**
 * 获取当前日志文件内容
 */
-(NSData *) getCurrentLogFileContent;
{
    NSData *content = nil;
    @synchronized(self) {
        [self.logFileHandle seekToFileOffset:0];
        content = [self.logFileHandle readDataToEndOfFile];
        [self.logFileHandle seekToEndOfFile];
    }
    
    return content;
}

-(void) printOldLogFile
{
    NSString* logs_dir = [self getLogsDirectory];
    for (int i=0; i<60; ++i)
    {
        NSString *logfile = [NSString stringWithFormat:@"log_%@.log", [IMUtil getOldDateString:i]];
        NSString *fileFullPath= [logs_dir stringByAppendingPathComponent:logfile];
        NSFileManager *fileManager = [NSFileManager defaultManager];
        if ([fileManager fileExistsAtPath:fileFullPath] == YES)
        {
//            NSLog(@"exist log file: %@", logfile);
        }
        else
        {
            //NSLog(@"log file:%@ not exists, break out!", fileFullPath);
            //break;
        }
    }
}

/**
 * 移除旧的日志文件
 */
-(void) removeOldLogFile
{
    NSString* logs_dir = [self getLogsDirectory];
    NSFileManager* fileManager = [NSFileManager defaultManager];
    // 保留60天的日志文件
    int remainDate = [[IMUtil getOldDateString:60] intValue];
    NSError* error = nil;
    NSArray* filenames = [fileManager contentsOfDirectoryAtPath:logs_dir error:&error];
    
    NSRange range = {4,8};
    for (NSString* fn in filenames)
    {
        //NSLog(@"log file %@ is found.", fn);
        
        if ([fn length] > 12 && [[fn substringFromIndex:([fn length] - 4)]  isEqual: @".log"])
        {
            NSString* dt = [fn substringWithRange:range];
            int fnDt = [dt intValue];
            
            if (fnDt < remainDate)
            {
                NSString *fileFullPath= [logs_dir stringByAppendingPathComponent:fn];
                
                if ([fileManager fileExistsAtPath:fileFullPath] == YES)
                {
                    if ([fileManager isDeletableFileAtPath:fileFullPath] == YES)
                    {
                        NSError *error = nil;
                        if ([fileManager removeItemAtPath:fileFullPath error:&error] == YES)
                        {
//                            NSLog(@"delete log file:%@ successful", fileFullPath);
                        }
                        else
                        {
//                            NSLog(@"delete log file:%@ failed", fileFullPath);
                        }
                    }
                    else
                    {
//                        NSLog(@"can't delete log file:%@", fileFullPath);
                    }
                } // if ([fileManager fileExistsAtPath:fileFullPath] == YES)
            } // if (fnDt < remainDate)
        } // if ([fn length] > 12 && [[fn substringFromIndex:([fn length] - 4)]  isEqual: @".log"])
    } // for (NSString* fn in filenames)
}

-(void)closeLog:(BOOL)isClose
{
    self.isCloseLog = isClose;
}


/**
 * 写日志函数
 * @param data: 日志内容
 */
-(void) writeLog:(NSString*)data
{
#ifdef DEBUG
    @synchronized(self)
    {
        __block NSString *blockData = data;
        
        dispatch_group_async(_loggingGroup, _loggingQueue, ^{ @autoreleasepool {
            
            NSFileHandle *fileHandle = self.logFileHandle;
            if (fileHandle != nil)
            {
                NSString *logData = [NSString stringWithFormat:@"[%@] %@", [IMUtil getSystemTimeStamp], blockData];
                
                if (!self.isCloseLog) {
//                    NSLog(@"%@", logData);
                }
                
                NSData *log = [IMUtil NSStringToNSData:logData];
                
                @try {
                    [fileHandle writeData:log];
                }
                @catch (NSException *exception) {
//                    NSLog(@"writeLog--exception:%@", exception);
                }
                @finally {
                    ;
                }
                
                [fileHandle synchronizeFile];
                
                if (!self.isCloseLog) {
//                    NSLog(@"%@", data);
                }
            }
        } });
    }
#endif
}

NS_INLINE NSException * _Nullable tryBlock(void(^_Nonnull tryBlock)(void)) {
    @try {
        tryBlock();
    }
    @catch (NSException *exception) {
        return exception;
    }
    return nil;
}


/**
 * 获取指定日期日志文件内容
 * @param dateString: 使用 [IMUtil getDateString]或[IMUtil getOldDateString:i]获得的日子格式化字符串
 */
-(NSData*) getLogFileContent:(NSString *)dateString
{
    NSString* filePath = [[self getLogsDirectory] stringByAppendingPathComponent: [NSString stringWithFormat:@"log_%@.log", dateString]];
    NSFileManager *fileManager = [NSFileManager defaultManager];
    BOOL is_dir = NO;
    if ([fileManager fileExistsAtPath:filePath isDirectory:&is_dir] != YES || is_dir == YES)
    {
        // 文件不存在或者指定路径不是文件而是目录
        return nil;
    }
    NSFileHandle *fileHandle = [NSFileHandle fileHandleForUpdatingAtPath:filePath];
    return [fileHandle readDataToEndOfFile];
}

@end
