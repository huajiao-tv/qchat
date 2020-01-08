//
//  IMProtoVoiceCallProxy.h
//  IMServiceLib
//
//  Created by 360 on 12/10/14.
//  Copyright (c) 2014 qihoo. All rights reserved.
//

#import "IMProtoHelper.h"

/**
 * 辅助生成解析VCP协议
 * */


@interface IMProtoVoiceCallProxy : IMProtoHelper

/**
 * 创建频道申请的包
 *
 * @param caller
 *            主叫
 * @param callee
 *            被叫
 * */

- (NSData*)createChannelRequestWithCaller:(NSString*)caller callee:(NSString*)callee sn:(NSInteger)sn;

@end
