//
//  IMFeatureGroupChat.h
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-25.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMFeatureBase.h"

@interface IMGetGroupMsgReq : NSObject

/**
 * 添加一个需要取消息的请求信息
 * @param groupId: 群ID
 * @param msgId:获取消息请求中的消息起始ID
 * @param offset:请求获取群消息的最大数量
 * @param tid:操作追踪序号
 * @return 操作结果代码
 */
-(IMErrorCode) addRequestGroup:(NSString*) groupId Start:(unsigned long long) msgId Offset: (int) offset Trace:(int64_t)tid;

/**
 * 获取群消息请求信息
 * @return 请求摘要数组
 */
-(NSArray*) requestSummary;
@end

/**
 * 派生于IMFeatureBase，群聊
 * 包括：
 * 1）所有与群聊相关的操作;
 */
@interface IMFeatureGroupChat : IMFeatureBase

/**
 * 发送取消息的请求(类似原来的get_info)
 * @param request:请求对象
 * @return 操作结果代码
 */
-(IMErrorCode) getMsg:(IMGetGroupMsgReq *)request;

/**
 * 发送同步群概要列表的请求(异步模式)
 * @param ids: 需要同步概要信息的群id列表，nil或者为空表示查询查询所有群的概要信息
 * @param sn:操作序号
 * @return 操作结果代码
 */
-(IMErrorCode)syncGroupSummaryFor:(NSSet*)ids Trace:(int64_t)traceId;

/**
 * 内部回执处理函数
 * @param sn: 请求sn
 * @param sleep: 下一个请求需要等待的时间
 * @return response所对应的APP层请求sn列表
 */
-(NSSet*)handleRequestReceipt:(int64_t)sn nextAfter:(int)sleep;
@end
