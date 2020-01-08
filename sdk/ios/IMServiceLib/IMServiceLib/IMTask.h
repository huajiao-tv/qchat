//
//  IMTask.h
//  IMServiceLib
//
//  Created by guanjianjun on 14-7-3.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import <Foundation/Foundation.h>
#import <Foundation/NSOperation.h>
#import "IMConstant.h"

@class IMMessage;
@class IMUser;

/**
 * 将每个操作抽象为一个task，使用NSOperation的原因是想让task有优先级.
 *
 */
@interface IMTask : NSOperation

/**
 * 任务类型
 */
@property (atomic, assign) IMTaskType taskType;

/**
 * message
 */
@property (atomic, strong) IMMessage *message;

/**
 * IMUser
 */
@property (atomic, strong) IMUser *user;


/**
 * 初始化函数
 */
-(id) initTask:(IMTaskType)type Message:(IMMessage*)msg User:(IMUser*)user;

/**
 * 初始化任务，默认是normal的类型
 */
-(id) initTask:(IMMessage*)msg User:(IMUser*)user;

@end
