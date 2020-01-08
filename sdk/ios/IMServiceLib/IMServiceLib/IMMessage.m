//
//  IMMessage.m
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-24.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMMessage.h"

@implementation IMMessage

/**
 * 属性列表
 */
#pragma mark - property list

@synthesize IsSend = _IsSend;
@synthesize errorCode = _errorCode;
@synthesize errorReason = _errorReason;
@synthesize sessionID = _sessionID;
@synthesize sessionType = _sessionType;
@synthesize featureID = _featureID;
@synthesize senderID = _senderID;
@synthesize convID = _convID;
@synthesize msgType = _msgType;
@synthesize msgID = _msgID;
@synthesize sn = _sn;
@synthesize sendTime = _sendTime;
@synthesize resultBody = _resultBody;
@synthesize requestBody = requestBody;
@synthesize infoType = _infoType;
@synthesize lastInfoID = _lastInfoID;
@synthesize curID = _curID;
@synthesize isWaitResp = _isWaitResp;



/**
 * 重载初始函数
 */
-(id) init
{
    self = [super init];
    if (self)
    {
        self.IsSend = false;
        self.errorCode = IM_Success;
        self.errorReason = @"success";
        self.featureID = IM_Feature_Ctrl;
        self.senderID = @"";
        self.convID = @"";
        self.msgType = 0;
        self.msgID = 0;
        self.sn = 0;
        self.sendTime = 0;
        self.resultBody = nil;
        self.requestBody = nil;
        self.infoType = @"";
        self.lastInfoID = -1;
        self.curID = -1;
        self.isWaitResp = YES;
        
        
        self.roomID = nil;
        self.lostInfoIds = nil;
    }
    return self;
}


/**
 * 返回当前message属于什么功能的名称
 * returns Feature name
 */
-(NSString*) featureName
{
    NSString *name = nil;
    switch (self.featureID)
    {
        case IM_Feature_Ctrl:
        {
            name = @"control";
        }
        break;
            
        case IM_Feature_Peer:
        {
            name = @"peer";
        }
        break;
            
        case IM_Feature_IM:
        {
            name = @"im";
        }
            break;
            
        case IM_Feature_Notify:
        {
            name = @"notify";
        }
        break;
            
        case IM_Feature_GroupChat:
        {
            name = @"group";
        }
        break;

        case IM_Feature_ChatRoom:
        {
            name = @"chatroom";
        }
        break;
            
        case IM_Feature_Circle:
        {
            name = @"circle";
        }
        break;
            
        default:
        break;
    }
    
    return name;
}

- (void)dealloc
{
}
@end
