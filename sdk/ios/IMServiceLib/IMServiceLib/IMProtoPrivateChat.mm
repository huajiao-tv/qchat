//
//  IMProtoPrivateChat.m
//  IMServiceLib
//
//  Created by guanjianjun on 15/6/18.
//  Copyright (c) 2015年 qihoo. All rights reserved.
//

#import "IMProtoPrivateChat.h"
#import "IMUtil.h"
#import "privatechat.pb.h"
using namespace qihoo::protocol::privatechat;

@interface IMProtoPrivateChat()

/**
 * 根据命令的类名称，生成Msg Message
 */
-(std::string) createPacket:(google::protobuf::Message*)command;


@end

@implementation IMProtoPrivateChat

-(std::string) createPacket:(google::protobuf::Message*)command
{
    @try
    {
        PChatRequest *pRequest = new PChatRequest();
        PChatRequest &upRequest = *pRequest;
        std::string tmp = typeid(*command).name();
        NSString *className = [IMUtil CharsToNSString:tmp.c_str()];

            //根据command生成Message的字符串
        if ([IMUtil hasSubString:@"PChatSendMsgRequest" Data:className])
        {
            upRequest.set_payload(1001);
            upRequest.set_allocated_sendreq((::qihoo::protocol::privatechat::PChatSendMsgRequest *)command);
        }
        else if ([IMUtil hasSubString:@"PChatGetMsgRequest" Data:className])
        {
            upRequest.set_payload(1002);
            upRequest.set_allocated_getreq((::qihoo::protocol::privatechat::PChatGetMsgRequest *)command);
        }
        else
        {
//            CPLog(@"not support command type:%@", className);
        }

        PChatPacket packet;
        packet.set_allocated_request(pRequest);
        packet.set_uuid([IMUtil NSStringToChars:[IMUtil createRandomString:8]]);

        return packet.SerializeAsString();
    }
    @catch (NSException *exception)
    {
//        CPLog(@"createMessageString exception, name:%@, reason:%@", exception.name, exception.reason);
    }
    
    return "";
}

-(NSMutableDictionary *)parseMessage:(NSData*)data
{
    NSMutableDictionary *dataDict = [[NSMutableDictionary alloc] init];
    [dataDict setObject:[NSNumber numberWithInt:0] forKey:@"result"];

    @try
    {
        std::string strData;
        [IMUtil NSDataToStlString:data Return:&strData];

        PChatPacket packet;
        packet.ParseFromString(strData);
        const PChatResponse &respData = packet.response();
        [dataDict setObject:[NSNumber numberWithInt:respData.payload()] forKey:@"payload"];

        if (respData.result() != 0)
        {
            [dataDict setObject:[NSNumber numberWithInt:respData.result()] forKey:@"result"];
            if (respData.has_reason())
            {
                [dataDict setObject:[IMUtil CharsToNSData:respData.reason().c_str() withLength:respData.reason().size()] forKey:@"reason"];
            }
        }
        else
        {
            if (respData.payload() == 1001) //send msg response
            {
                [dataDict setObject:[NSNumber numberWithLongLong:respData.sendres().msgid()] forKey:@"msgid"];
                [dataDict setObject:[NSNumber numberWithLongLong:respData.sendres().code()] forKey:@"code"];
                [dataDict setObject:[IMUtil CharsToNSString:respData.sendres().reason().c_str()] forKey:@"reason"];
            }
            else if (respData.payload() == 1002) //get msg response
            {
                const PChatGetMsgResponse &getResp = respData.getres();
                if (getResp.has_maxid())
                {
                    [dataDict setObject:[NSNumber numberWithLongLong:getResp.maxid()] forKey:@"maxmsgid"];
                }
                NSMutableArray *msgList = [[NSMutableArray alloc] init];
                for (int i=0; i<getResp.msglist_size(); ++i)
                {
                    [msgList addObject:[self PChatMsgToDict:getResp.msglist(i)]];
                }
                [dataDict setObject:msgList forKey:@"msglist"];
            }
            else if (respData.payload() == 10001) //quit room
            {
                const PChatNewMsgNotify &notify = respData.msgnotify();
                if (notify.has_maxid())
                {
                    [dataDict setObject:[NSNumber numberWithLongLong:notify.maxid()] forKey:@"maxmsgid"];
                }
                if (notify.has_msg())
                {
                    [dataDict setObject:[self PChatMsgToDict:notify.msg()] forKey:@"msg"];
                }
            }
        }
    }
    @catch (NSException *exception)
    {
//        CPLog(@"createMessageString exception, name:%@, reason:%@", exception.name, exception.reason);
    }

    return dataDict;
}

    //将一条消息转换为字典
-(NSMutableDictionary*) PChatMsgToDict:(const PChatMsg &)chatMsg
{
    NSMutableDictionary *msgDict = [[NSMutableDictionary alloc] init];

    for (int k=0; k<chatMsg.msgprops_size(); ++k)
    {
        const PChatPair &pair = chatMsg.msgprops(k);
        NSString *strKey = [IMUtil CharsToNSString:pair.key().c_str()];
        id thisValue = nil;
        if ([strKey isEqualToString:@"msgid"] || [strKey isEqualToString:@"sendtime"])
        {
            thisValue = [NSNumber numberWithLongLong:[[IMUtil CharsToNSString:pair.value().c_str()] longLongValue]];
        }
        else if ([strKey isEqualToString:@"type"] || [strKey isEqualToString:@"expire"])
        {
            thisValue = [NSNumber numberWithInt:[[IMUtil CharsToNSString:pair.value().c_str()] intValue]];
        }
        else if ([strKey isEqualToString:@"srcid"] || [strKey isEqualToString:@"destid"])
        {
            thisValue = [IMUtil CharsToNSString:pair.value().c_str()];
        }
        else if ([strKey isEqualToString:@"data"])
        {
            thisValue = [IMUtil CharsToNSData:pair.value().c_str() withLength:pair.value().size()];
        }
        if (thisValue != nil)
        {
            [msgDict setObject:thisValue forKey:strKey];
        }
    }

    return msgDict;
}

/**
 * 创建发送消息二进制流
 */
-(NSData*) createSendMsgRequest:(NSString*)destid Appid:(int)appid Type:(int)type Data:(NSData*)data Expire:(int)seconds
{
    PChatSendMsgRequest *sendReq = new PChatSendMsgRequest();
    sendReq->set_destid([IMUtil NSStringToChars:destid]);
    sendReq->set_destappid(appid);
    sendReq->set_bodytype(type);
    std::string result;
    [IMUtil NSDataToStlString:data Return:&result];
    sendReq->set_bodydata(result);
    if (seconds > 0)
    {
        sendReq->set_expiresec(seconds);
    }

    std::string strPacket = [self createPacket:sendReq];
    return [IMUtil CharsToNSData:strPacket.c_str() withLength:strPacket.size()];
}

-(NSData*) createSendMsgRequest:(NSString*)destid Appid:(int)appid Type:(int)type Data:(NSData*)data
{
    return [self createSendMsgRequest:destid Appid:appid Type:type Data:data Expire:0];
}

/**
 * 创建取消息二进制流
 */
-(NSData*) createGetMsgRequest:(long long)start Count:(int)count
{
    PChatGetMsgRequest *getReq = new PChatGetMsgRequest();
    getReq->set_start(start);
    getReq->set_count(count);
    std::string strPacket = [self createPacket:getReq];
    return [IMUtil CharsToNSData:strPacket.c_str() withLength:strPacket.size()];
}

@end
