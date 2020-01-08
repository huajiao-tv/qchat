//
//  IMUser.h
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-7.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import <Foundation/Foundation.h>
#import "IMNotifyDelegate.h"
#import "IMConfigData.h"

//protocol class
@class IMProtoRegistry;
@class IMProtoWeb;
@class IMProtoUserInfo;
@class IMProtoMessage;
@class IMProtoVoiceCallProxy;
@class IMProtoRelation;
@class IMProtoServiceInner;
@class IMProtoGroupChat;
@class IMProtoChatRoom;
@class IMProtoCircleChat;
@class IMProtoPrivateChat;

//feature class
@class IMConfigData;
@class IMProtoHelper;
@class IMFeatureWebInterface;
@class IMFeatureUserInfo;
@class IMFeaturePeerChat;
@class IMFeatureGroupChat;
@class IMFeatureChatRoom;
@class IMFeatureCircleChat;
@class IMFeatureRelation;
@class IMFeaturePrivateChat;

@class IMMessage;
@class IMTask;


/**
 * 每个登录账号就是一个IMUser
 */
@interface IMUser : NSObject

#pragma mark - property list
/**
 * user's config data
 */
@property (atomic, strong) IMConfigData *configData;

/**
 * user data call back
 */
@property (atomic) id<IMNotifyDelegate> notifyReceiver;



//***********************proto buffer helper**********************/
/**
 * registry proto buffer
 */
@property (atomic, strong) IMProtoRegistry *protoRegistry;

/**
 * web proto buffer
 */
@property (atomic, strong) IMProtoWeb *protoWeb;

/**
 * user info proto buffer
 */
@property (atomic, strong) IMProtoUserInfo *protoUserInfo;

/**
 * peer message proto buffer
 */
@property (atomic, strong) IMProtoMessage *protoMessage;

/**
 * voice Call Proxy proto buffer (vcp)
 */
@property (atomic, strong) IMProtoVoiceCallProxy *protoVoiceCallProxy;

/**
 * service inner proto buffer
 */
@property (atomic, strong) IMProtoServiceInner *protoServiceInner;

/**
 * groupchat proto buffer
 */
@property (atomic, strong) IMProtoGroupChat *protoGroupChat;

/**
 * chatroom proto buffer
 */
@property (atomic, strong) IMProtoChatRoom *protoChatRoom;

/**
 * circlechat proto buffer
 */
@property (atomic, strong) IMProtoCircleChat *protoCircleChat;

/**
 * relation proto buffer
 */
@property (atomic, strong) IMProtoRelation *protoRelation;

/**
 * 新单聊模块解码
 */
@property (atomic, strong) IMProtoPrivateChat *protoPrivateChat;


//登录时，服务器时间
@property(atomic,assign) int lastLoginedTimestamp;
//***********************feature list**********************/

/**
 * web interface feature
 */
@property (atomic, strong) IMFeatureWebInterface *webFeature;
/**
 * user info feature
 */
@property (atomic, strong) IMFeatureUserInfo *userInfoFeature;
/**
 * peer chat feature
 */
@property (atomic, strong) IMFeaturePeerChat *peerFeature;
/**
 * group chat feature
 */
@property (atomic, strong) IMFeatureGroupChat *groupChatFeature;
/**
 * chatroom feature
 */
@property (atomic, strong) IMFeatureChatRoom *chatRoomFeatrue;
/**
 * circle chat feature
 */
@property (atomic, strong) IMFeatureCircleChat *circleChatFeature;
/**
 * relation feature
 */
@property (atomic, strong) IMFeatureRelation *relationFeature;

/**
 * private chat proto buffer
 */
@property (atomic, strong) IMFeaturePrivateChat *privatechatFeature;

/**
 * indicates whether pull preferred hosts from http
 */
@property (atomic, assign) BOOL pullPreferredHosts;
#pragma mark - function list

/**
 * init im user object, with config data is default values
 */
-(id) init;

/**
 * 加载当前用户的配置项
 */
-(void) loadUserSetting;

/**
 * init im user instance
 * @param delegate IMNotifyDelegate instance
 * @returns IMUser pointer
 */
-(id) initWithDelegate:(id<IMNotifyDelegate>)delegate;

/**
 * 系统网络切换回调函数
 * 0 -- power off
 * 1 -- power on with gprs
 * 2 -- power on with wifi
 */
- (void) reachabilityChanged:(int)status;

/**
 * startService:启动服务，该函数可以多次被调用。当app从后台切到前台时，需要调用该函数
 * returns: true--启动成功；false--启动失败，一定是长连接服务器地址没有配置才会返回false
 */
-(BOOL) startService;

/**
 * startWithHB:启动服务，该函数可以多次被调用。当app从后台切到前台时，需要调用该函数
 * @param hbInterval  心跳时间间隔(s)
 * returns: true--启动成功；false--启动失败，一定是长连接服务器地址没有配置才会返回false
 */
-(BOOL) startWithHB:(int)hbInterval;

/**
 * 停止服务
 */
-(void) stopService;

/**
 * 判断当前用户是否处于连接状态
 * @returns: true--连接状态，false--未连接
 */
-(BOOL) isConnected;


/**
 * imsdk的chatroom的丢失消息统计数组的元素个数
 *
 */
-(void)setChatRoomLostIdCount:(NSInteger)lostCount;

/**
 * 往任务队列里添加任务，仅供内部调用
 */
-(void) addTask:(IMTask*)task;

/**
 * 往任务队列里添加任务，仅供内部调用
 */
-(void) addTask:(IMTaskType)type Message:(IMMessage*)msg;

/**
 * 往任务队列里添加一个发送任务
 */
-(IMErrorCode) addSendTask:(IMMessage*)msg;

/**
 * 往任务队列里添加一个指定优先级发送任务
 * @param msg  需要发送的消息
 * @param priority  任务优先级
 */
-(IMErrorCode) addSendTask:(IMMessage*)msg With:(NSOperationQueuePriority)priority;

/**
 * 处理任务
 */
- (void) handleTask:(IMTaskType)type Message:(IMMessage*)message Cancelled:(BOOL)isCancel;

/**
 * 日志文件
 */
-(void) writeLog:(NSString*)data;

/**
 * 群回调delegate
 */
- (void) notifyDelegateGroup:(NSMutableDictionary*)data;

@end
