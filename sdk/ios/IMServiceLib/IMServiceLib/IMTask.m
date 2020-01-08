//
//  IMTask.m
//  IMServiceLib
//
//  Created by guanjianjun on 14-7-3.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMTask.h"
#import "IMUser.h"
#import "IMMessage.h"

@implementation IMTask

/**
 * 任务类型
 */
@synthesize taskType = _taskType;

/**
 * message
 */
@synthesize message = _message;

/**
 * IMUser
 */
@synthesize user = _user;


/**
 * 初始化函数
 */
-(id) initTask:(IMTaskType)type Message:(IMMessage*)msg User:(IMUser*)user
{
    self = [super init];
    if (self)
    {
        self.taskType = type;
        self.message = msg;
        self.user = user;
    }
    return self;

}

/**
 * 初始化任务，默认是normal的类型
 */
-(id) initTask:(IMMessage*)msg User:(IMUser*)user
{
    self = [super init];
    if (self)
    {
        self.taskType = IM_TaskType_Normal;
        self.message = msg;
        self.user = user;
    }
    return self;
}

/**
 * 重载main函数，类似于java里的runnable
 */
-(void)main
{
    @autoreleasepool
    {
        if (self.user != nil)
        {
            [self.user handleTask:self.taskType Message:self.message Cancelled:[self isCancelled]];
        }
    }
}
    
@end
