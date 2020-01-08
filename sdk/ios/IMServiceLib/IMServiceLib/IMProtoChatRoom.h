//
//  IMProtoChatRoom.h
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-26.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMProtoHelper.h"

/**
 * 对IMProtoHelper的分类，聊天室
 * 包括：
 * 1）所有与圈子相关的;
 */
@interface IMProtoChatRoom : IMProtoHelper

/**
 * 根据接收到的二进制流，解析
 * @param data: 二进制流(std::string *)
 * @param msgID: 该Message的id，在函数内部改变这个值，调用后调用者应该根据这个msgID的值来判断返回的Message到底是什么类型的
 * @returns Message引用
 */
-(NSMutableDictionary *)parseMessage:(NSData*)data;


/**
 * 根据接收到的二进制流，解析
 * @param data: 二进制流(std::string *)
 * @returns ChatRoomNewMsg引用
 */
-(NSMutableDictionary *)parseNewMessage:(NSData*)data;

/**
 * 加入聊天室
 */
-(NSData*) createJoinRoomRequest:(NSString*)roomid withProperties:(NSDictionary*)properties;
-(NSData*) createJoinRoomRequest:(NSString*)roomid withData:(NSData*)userdata Properties:(NSDictionary*)properties;
-(NSMutableDictionary*) parseJoinRoomResponse:(NSData*)data;


/**
 * 取聊天室详情
 */
-(NSData*) createQueryRoomRequest:(NSString*)roomid From:(int)from Count:(int)count;
-(NSMutableDictionary*) parseQueryRoomResponse:(NSData*)data;

/**
 * 退出聊天室
 */
-(NSData*) createQuitRoomRequest:(NSString*)roomid;
-(NSMutableDictionary*) parseQuitRoomResponse:(NSData*)data;

-(NSData*) createChatroomMessageRequest:(NSString*)roomid Message:(NSData*)content;

/**
 * 取消/恢复订阅指定聊天室消息
 * @param sub: 是否订阅， YES订阅；NO不订阅
 * @param roomid: 指定聊天室
 */
-(NSData*) createSubscribe:(BOOL)sub Request:(NSString*)roomid;
-(NSMutableDictionary*) parseSubscribeResponse:(NSData*)data;

@end


@interface IMChatRoomMsgLost : NSObject

@property (nonatomic, assign)NSInteger msgLostTime;
@property (nonatomic, assign)NSInteger msgReloadTime;

@property (nonatomic, assign)NSInteger msgLostId;


@end
