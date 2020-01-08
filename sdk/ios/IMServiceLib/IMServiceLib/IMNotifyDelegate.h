//
//  IMUserDelegate.h
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-24.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import <Foundation/Foundation.h>
#import "IMConstant.h"

@class IMUser;
@class IMMessage;

/**
 * 定义IMUser代理接口，数据到达，发送数据的结果，以及进度都是有这个delegeate来通知上层
 */
@protocol IMNotifyDelegate <NSObject>

//required interface
@required

//optional interface
@optional

/**
 * 状态改变， 用于通知是否登录成功
 * @param fromState: 旧状态
 * @param curState: 新状态(目前的状态 )
 * */
-(void) onStateChange:(IMUserState)curState From:(IMUserState)fromState User:(IMUser*)user;

/**
 * 状态改变通知
 */
-(void) onStateChange:(NSString *)changeInfo;

/**
 * 返回发送消息的结果
 * */
-(void) onSendResult:(IMMessage*)result User:(IMUser*)user;

/**
 * 收到了新消息
 */
-(void) onMessage:(IMMessage*)message User:(IMUser*)user;


/**
 *  查询的状态返回, 用户手机号和状态一一对应
 *
 *  @param sn       关联之前的查询动作
 *  @param result   查询的结果0代表成功
 *  @param users    对应的用户手机号码
 *  @param statuses 请查阅 Constant USER_PRESENCE_UNREGISTERED ,USER_PRESENCE_REGISTERED
 */
- (void)onPresenceWithSid:(NSString*)sid sn:(int64_t)sn result:(int)result users:(NSArray*)users statuses:(NSArray*)statuses;

/**
 *  创建频道返回，
 *
 *  @param sn   关联之前的动作
 *  @param result   查询的结果0代表成功
 *  @param data 频道参数
 */
- (void)onChannelWithSid:(NSString*)sid sn:(int64_t)sn result:(int)result channelId:(NSString*)channelId channelInfo:(NSData*)data;

/**
 * 加入，退出，查询聊天室的应答事件
 * eventTyp: 101 -- 查询聊天室，102--加入聊天室， 103--退出聊天室
 * success: YES --成功， NO -- 失败，如果失败，roominfo为nil
 * roominfo: 聊天室详情字典包括如下key(:
 * roomid[NSString]:聊天室id
 * version[NSNumber(longlong):版本号
 * memcount[NSNumber(int)]:成员数量(包括qid用户和非qid用户)
 * regmemcount[NSNumber(int)]:非qid用户数量
 * members[NSArray]:成员的userid
 */
- (void)onChatroomEvent:(int)eventType IsSuccessful:(BOOL)success RoomInfo:(NSDictionary*)roominfo;

/**
 * roomid:聊天室id
 * data: 消息(谈谈服务器发过来的，是个json)
 */
- (void)onChatroomData:(NSString*)roomid Sender:(NSString*)userid Data:(NSData*)data;
/**
 * roomid：聊天室id;
 * userid: 发送消息的用户id;
 * data: 用户发送的消息内容;
 * memcount: 聊天室里的总人数;
 * regcount: 聊天室里的注册用户数;
 */
- (void)onChatroomData:(NSString*)roomid Sender:(NSString*)userid Data:(NSData*)data MemCount:(int)memcount RegCount:(int)regcount;

/**
 * roomid: 聊天室ID
 * eventType: 1001 -- 加入聊天室, 1002 -- 退出聊天室
 * userid: 成员id，例如eventType为1001时，表示该成员加入了聊天室，为1002时表示该成员退出了聊天室
 * memcount: 聊天室总成员数
 */
- (void)onChatroom:(NSString*)roomid Change:(int)eventType Member:(NSString*)userid MemCount:(int)memcount;
/**
 * roomid: 聊天室ID
 * eventType: 1001 -- 加入聊天室, 1002 -- 退出聊天室
 * userid: 成员id，例如eventType为1001时，表示该成员加入了聊天室，为1002时表示该成员退出了聊天室
 * memcount: 聊天室总成员数
 * userdata: 只有eventype为1001时有效，表示加入者的个人信息(来自花椒服务器)
 */
- (void)onChatroom:(NSString*)roomid Change:(int)eventType Member:(NSString*)userid MemCount:(int)memcount withData:(NSData*)userdata;
/**
 * roomid: 聊天室ID
 * eventType: 1001 -- 加入聊天室, 1002 -- 退出聊天室
 * userid: 成员id，例如eventType为1001时，表示该成员加入了聊天室，为1002时表示该成员退出了聊天室
 * memcount: 聊天室总成员数
 * regcount: 聊天室中的注册成员数
 * userdata: 只有eventype为1001时有效，表示加入者的个人信息(来自花椒服务器)
 */
- (void)onChatroom:(NSString*)roomid Change:(int)eventType Member:(NSString*)userid MemCount:(int)memcount RegCount:(int)regcount withData:(NSData*)userdata;

/**
 * 从聊天室系统推送下来的给单个人的消息
 */
- (void)onPeerchat:(int64_t)msgid Data:(NSData*)message;

/**
 * 从聊天室系统推送下来的给 私聊
 */
- (void)onIMchat:(int64_t)msgid Data:(NSData*)message;

/**
 * 公共收件箱消息
 */
- (void)onPublic:(int64_t)msgid Data:(NSData*)message;

/**
 * 私聊消息发送回调
 * sn:发送消息时上层传下来的sn，服务器会原值返回，供上层界面做异步关联;
 * msgid: 如果发送消息成功，这个值应该大于0，如果等于0表明发送消息失败;
 * code: 发送失败时，这个表示错误码;
 * reason: 发送失败时，这个表示错误原因;
 */
- (void) onPrivateChat:(int64_t)sn Msgid:(int64_t)msgid Code:(int)code Reason:(NSString*)reason;

/**
 * 收到私聊消息
 * from: 发送者id;
 * micsecond: 相对于1970-01-01 00:00:00的发送时间(毫秒)
 * msgid: 消息id;
 * message: 消息内容;
 * sn: 发送方为消息打的序号(主要用于追踪消息),如果登录后取的离线消息，这sn永远是0；
 */
- (void)onPrivateChat:(NSString*)from SendTime:(int64_t)micsecond Msgid:(int64_t)msgid Type:(int)type Data:(NSData*)message SN:(int64_t)sn;

/**
 * 群组收件箱通知/消息
 */
- (void)onGroup:(NSMutableDictionary*)data;

@end
