//
//  IMProtoMessage.h
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-26.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMProtoHelper.h"

/**
 * 对IMProtoHelper的分类，与msg通信的协议
 * 包括：
 * 1）登录;
 * 2) 发消息、收消息;
 * 3) 收通知；
 */
@interface IMProtoMessage : IMProtoHelper

/**
 * 创建initLogin请求
 * @param sn:上层传下来的时间戳，如果没有，则sdk内部产生一个
 */
-(NSData*) createInitLoginRequest;

/**
 * 解析InitLogin应答
 * @param data:收到的数据
 * @returns 返回的参数
 */
-(NSMutableDictionary*) parseInitLoginResponse:(NSData*)data;


/**
 * 创建login请求
 * @param netType: 网络类型， "wifi", "gprs"
 */
-(NSData*) createLoginRequest:(int)netType;

-(NSData*) createWeimiLoginRequest:(int)netType;

/**
 * 解析注册应答
 * @param data: 收到的数据
 * @returns 返回解析参数
 */
-(NSMutableDictionary*) parseLoginResponse:(NSData*)data;

/**
 * 解析服务器数据，数据可以是通知，getinfo的应答，发送数据的ack等
 * @param data: 收到的数据
 * @returns 返回解析参数
 */
-(NSMutableDictionary*) parseServerData:(NSData*)data;

/**
 * 根据解析出来的字典构造getInfo请求
 * @param infoType: 消息盒子名称;
 * @param infoID: 其实消息id;
 * @param offset: 消息数量;
 * @param sn: 请求序号
 * @returns 打包后的数据
 */
-(NSData*) createGetInfoRequest:(NSString*)infoType StartID:(int64_t)infoID Offset:(int)offset Sn:(int64_t)sn;

-(NSData*) createGetInfoRequest:(NSString*)infoType StartID:(int64_t)infoID Offset:(int)offset RoomID:(NSString*)roomid Sn:(int64_t)sn;

-(NSData*) createGetMultiInfosRequest:(NSString*)infoType InfoIds:(NSArray*) infoIds RoomID:(NSString*)roomid Sn:(int64_t)sn;

/**
 * 为了得到控制命令的发送者，需要从body部分解码
 */
-(NSString*) parseSenderID:(NSData*)data;

/**
 * 为了得到channel的信息，需要从body部分解码
 */
-(NSDictionary*) parseChannelData:(NSData*)data;


-(NSData*) createChatRequest:(NSString*)receiver Data:(NSData*)data;

/**
 * 创建是否在线请求
 */
-(NSData*) createEx1QueryUserStatusRequest:(uint64_t)sn UserType:(NSString*)userType UserIds:(NSArray*) userIds;

/**
 * 创建单聊请求
 */
-(NSData*) createPeerChatRequest:(uint64_t)sn Receiver:(NSString*)receiver RecvType:(NSString*)recvType Body:(NSData*)body BodyType:(int)bodyType ExpireTime:(uint32_t)expireTime;

/**
 * 创建service 请求
 */
-(NSData*) createServiceRequest:(NSData*)serviceData ServiceID:(int)serviceid SN:(int64_t)sn;
-(NSData*) createServiceRequest:(NSData*)serviceData ServiceID:(int)serviceid;

-(NSData*) createServiceMsgRequest:(NSData*)serviceData ServiceID:(int)serviceid;

@end
