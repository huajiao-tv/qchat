//
//  IMMessage.h
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-24.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import <Foundation/Foundation.h>
#import "IMConstant.h"

/**
 * 用户收到的message对象
 */
@interface IMMessage : NSObject

#pragma mark - property list

/**
 * 表示当前message是上行还是下行
 */
@property (atomic, assign) BOOL IsSend;

//errorCode和reason仅在http接口的返回对象里使用
/**
 * send result,消息结果 请参照 Error.MESSAGE_* 值
 */
@property (atomic, assign) IMErrorCode errorCode;

/**
 * fail reason
 */
@property (atomic, strong) NSString *errorReason;

/**
 * 功能分类ID：
 * 0) notify;
 * 1）peer；
 * 2）notify；
 * 3）groupchat；
 * 4）chatroom；
 * 5）circlechat;
 * 6) relation;
 * 7） channel;
 */
@property (atomic, assign) IMFeatureCode featureID;

/**
 * 某个功能下的某个payload
 */
@property (nonatomic, assign) int payload;

/**
 * 发送者的phone
 */
@property (atomic, strong) NSString* senderID;

/**
 * sessionID， 区分通话
 */
@property (atomic, strong) NSString* sessionID;

/**
 * sessionType， 通话类型
 */
@property (atomic, assign) int sessionType;

/**
 * 会话ID， 可以填主题ID区分会话
 */
@property (atomic, strong) NSString* convID;

/**
 * 消息类型
 */
@property (atomic, assign) int msgType;

/**
 * 消息类型
 */
@property (atomic, assign) int64_t msgID;

/**
 * 序列号, 由发送者确定， 最好不重复
 */
@property (atomic, assign) int64_t sn;

/**
 * 消息的发送时间
 */
@property (atomic, assign) int64_t sendTime;

/**
 * 返回消息的内容
 */
@property (atomic, strong) NSData* resultBody;

/**
 * 请求消息的内容, 正常发送消息时，等到connected成功后，在发送时再加密
 */
@property (atomic, strong) NSData* requestBody;

/**
 * 消息盒子名称
 */
@property (atomic, strong) NSString *infoType;


#pragma mark - public function list

/**
 * 返回当前message属于什么功能的名称
 * returns featrue name
 */
-(NSString*) featureName;



/**
 * 需要读取的消息的最大ID
 */
@property (atomic, assign) int64_t lastInfoID;

/**
 * 以读取的消息最大ID,也即当前消息的readID
 */
@property (atomic, assign) int64_t curID;


/**
 * 是否等待服务器应答
 */
@property (atomic, assign) BOOL isWaitResp;

/**
 * chatroom丢失的消息，重新申请服务器
 */
@property (atomic, strong) NSString* roomID;
@property (atomic, strong) NSArray* lostInfoIds;

@end
