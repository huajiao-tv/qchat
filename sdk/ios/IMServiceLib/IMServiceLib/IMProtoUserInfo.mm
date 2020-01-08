//
//  IMProtoUserInfo.m
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-26.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMProtoUserInfo.h"
#import "IMUtil.h"
@implementation IMProtoUserInfo

/**
 * 属性列表
 */
#pragma mark - property list

@synthesize userId = _userId;
@synthesize userType = _userType;
@synthesize status = _status;
@synthesize jid = _jid;
@synthesize platform = _platform;
@synthesize mobileType = _mobileType;
@synthesize appId = _appId;
@synthesize clientVersion = _clientVersion;

/**
 * 重载初始函数
 */
-(id) init
{
    self = [super init];
    if (self)
    {
        self.userId = @"";
        self.userType = @"";
        self.status = 0;
        self.jid = @"";
        self.platform = @"";
        self.mobileType = @"iOS";
        self.appId = 0;
        self.clientVersion = 0;
    }
    return self;
}

-(NSString *)description{
    return [NSString stringWithFormat:@"IMProtoUserInfo = %p: User ID= [%@], User Type= [%@], Status = [%d], jid = [%@], Platform = [%@], Mobile Type = [%@], App ID= [%d], Client Version = [%d] \n",self, self.userId, self.userType, self.status, self.jid, self.platform, self.mobileType, self.appId, self.clientVersion];
}

@end
