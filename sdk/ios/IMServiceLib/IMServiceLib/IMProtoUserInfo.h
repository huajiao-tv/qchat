//
//  IMProtoUserInfo.h
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-26.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMProtoHelper.h"

/**
 * 对IMProtoHelper的分类，所有User相关的web接口放到这里
 * 包括：
 * 1) 上传通讯录;
 * 2) 查询用户状态;
 * 3) 更新用户个人信息;
 * 4) 查询用户个人信息；
 */

/*
 message RespEQ1User {
 required string userid      = 1;
 required string user_type   = 2;
 required int32  status      = 3;           //0:not registry;  1:registry, offline, not reachable; 2:registry, offline, reachable; 3:registry, online, reachable
 optional string jid         = 4;
 optional uint32 app_id      = 5;
 optional string platform    = 6;           // web, pc, mobile
 optional string mobile_type = 7;       //android, ios
 optional uint32 client_ver  = 8;
 }
 */
@interface IMProtoUserInfo : IMProtoHelper

@property (atomic, strong) NSString *userId;

@property (atomic, strong) NSString *userType;

@property (atomic, assign) int status;

@property (atomic, strong) NSString *jid;

@property (atomic, strong) NSString *platform;

@property (atomic, strong) NSString *mobileType;

@property (atomic, assign) int appId;

@property (atomic, assign) int clientVersion;

@end
