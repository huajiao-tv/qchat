//
//  IMProtoGroupChat.h
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-26.
//  Modified by longjun on 2016-07-15.
//  Copyright © 2016年 qihoo. All rights reserved.
//

#import "IMProtoHelper.h"

/**
 * 对IMProtoHelper的分类，群聊
 * 包括：
 * 1）所有与群聊相关的;
 */
@interface IMProtoGroupChat : IMProtoHelper
/**
 * 根据接收到的二进制流，解析群下行协议包
 * @param data: 二进制流
 * @return 解析完成的数据字典
 */
-(NSMutableDictionary *)parseGroupDownPacket:(NSData*)data;

/**
 * 根据收取群消息请求信息生成群上行包二进制流
 * @param request: 需要收取消息的群信息列表
 * @return 群上行包二进制流
 */
-(NSData*) createGetGroupMessagePacketFrom:(NSArray*) request;

/**
 * 生成同步群信息上行包二进制流
 * @param groups: 存放了需要同步的群信息，nil对象或空集合同步所有群概要
 * @return 群上行包二进制流
 */
-(NSData*) createSyncGroupListPacketFor:(NSSet*)groups;
@end

@interface GetGroupMsgParam : NSObject
@property(atomic, strong) NSString* groupId;
@property(atomic, assign) unsigned long long startId;
@property(atomic, assign) int offset;
@property(atomic, strong) NSMutableSet* traceIds;
@end

@interface SyncGroupParam  : NSObject
@property(atomic, strong) NSMutableSet* groupIds;
@property(atomic, strong) NSMutableSet* traceIds;
@end