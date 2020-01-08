//
//  IMFeatureBase.m
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-24.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMFeatureBase.h"
#import "IMUtil.h"
#import "IMUser.h"
#import "IMProtoHelper.h"

@implementation IMFeatureBase

/**
 * property list
 */
#pragma mark - property list
@synthesize imUser = _imUser;
@synthesize featureUriDict = _featureUriDict;

/**
 * function list
 */
#pragma mark - function list

/**
 *
 * @param config: user config data
 * @returns IMProtoHelper instance
 */
-(id)initWithUser:(IMUser*)user
{
    self = [super init];
    if (self)
    {
        self.imUser = user;
        [self initUriDict];
    }
    return self;
}

-(void)initUriDict
{
    _featureUriDict = [[NSMutableDictionary alloc] init];
    [_featureUriDict setObject:@"Downreg/getvc" forKey:@"down_reg_vc"];
    [_featureUriDict setObject:@"Downreg/get" forKey:@"down_reg_get"];
    [_featureUriDict setObject:@"Reg/sms_cb" forKey:@"up_reg_sms"];
    [_featureUriDict setObject:@"Reg/get" forKey:@"up_reg_get"];
    [_featureUriDict setObject:@"Download/download" forKey:@"download_pic"];
    [_featureUriDict setObject:@"Upload/up" forKey:@"upload_pic"];
    [_featureUriDict setObject:@"Qidreg/r" forKey:@"qid_reg"];
}

/**
 * 发送http请求
 * @param url: url address
 * @param data: request data which is put into body
 * @param method: 1:get, 2:post
 * @param timeout: time out
 * @returns immessage pointer
 */
-(IMMessage*) sendHttp:(NSString *)uri requestData:(NSData *)data method:(int)method timeout:(NSTimeInterval)timeout
{
    NSURL *url = [NSURL URLWithString:uri];
    //创建http请求
    NSMutableURLRequest *request = [NSMutableURLRequest requestWithURL:url cachePolicy:NSURLRequestReloadIgnoringCacheData timeoutInterval:timeout];
    //设置请求方式 默认是GET，2：post
    if (method == 2)
    {
        [request setHTTPMethod:@"POST"];
        //设置请求体
        [request setHTTPBody:data];
    }
    //发出HTTP请求并且得到服务器的返回数据
    NSData *resultData = [NSURLConnection sendSynchronousRequest:request returningResponse:Nil error:Nil];
    NSString *resultXMLString = [[NSString alloc]initWithData:resultData encoding:NSUTF8StringEncoding];
//    CPLog(@"server response:%@", resultXMLString);
    //把返回数据转成字符串
    IMMessage *msgData = [[IMMessage alloc] init];
    if (resultData != nil)
    {
        msgData.errorCode = IM_Success;
        msgData.resultBody = resultData;
    }
    else
    {
        msgData.errorCode = IM_OperateTimeout;
    }
    
    return msgData;
}


/**
 * 发送http post请求
 * @param uri: url address
 * @param data: request data which is put into body
 * @returns immessage pointer
 */
-(IMMessage*) postHttp:(NSString*)uri requestData:(NSData*)data
{
    return [self sendHttp:uri requestData:data method:2 timeout:30];
}

/**
 * 发送http post请求
 * @param uri: url address
 * @param data: request data which is put into body
 * @param timeout: time out
 * @returns immessage pointer
 */
-(IMMessage*) postHttp:(NSString*)uri requestData:(NSData*)data timeout:(NSTimeInterval)timeout
{
    return [self sendHttp:uri requestData:data method:2 timeout:timeout];
}


/**
 * 发送http get请求
 * @param uri: url address
 * @returns immessage pointer
 */
-(IMMessage*) getHttp:(NSString*)uri
{
    return [self sendHttp:uri requestData:nil method:1 timeout:5];
}

/**
 * 发送http get请求
 * @param uri: url address
 * @param timeout: time out
 * @returns immessage pointer
 */
-(IMMessage*) getHttp:(NSString*)uri timeout:(NSTimeInterval)timeout
{
    return [self sendHttp:uri requestData:nil method:1 timeout:timeout];
}

/**
 * 检查用户是否处于连接状态
 */
-(BOOL) isUserConnected
{
    return [self.imUser isConnected];
}

/**
 * 给msgrouter发送数据
 * @param message: immessage
 * @returns 成功-IM_Success，失败-other
 */
-(IMErrorCode) sendMessage:(IMMessage*)message
{
    return [self.imUser addSendTask:message];
}



@end
