//
//  IMProtoChatRoom.m
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-26.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMProtoChatRoom.h"
#import "IMUtil.h"
#import "chatroom.pb.h"
using namespace qihoo::protocol::chatroom;

@interface IMProtoChatRoom()

/**
 * 根据命令的类名称，生成Msg Message
 */
-(std::string) createChatrromPacket:(google::protobuf::Message*)command;


@end


@implementation IMProtoChatRoom

-(std::string) createChatrromPacket:(google::protobuf::Message*)command RoomID:(NSString*)roomid;
{
    @try
    {
        qihoo::protocol::chatroom::ChatRoomUpToServer *pRequest = new qihoo::protocol::chatroom::ChatRoomUpToServer();
        qihoo::protocol::chatroom::ChatRoomUpToServer &upRequest = *pRequest;
        std::string tmp = typeid(*command).name();
        NSString *className = [IMUtil CharsToNSString:tmp.c_str()];

            //根据command生成Message的字符串
        if ([IMUtil hasSubString:@"GetChatRoomDetailRequest" Data:className])
        {
            upRequest.set_payloadtype(PAYLOAD_QUERY_CHATROOM);
            upRequest.set_allocated_getchatroominforeq((::qihoo::protocol::chatroom::GetChatRoomDetailRequest *)command);
        }
        else if ([IMUtil hasSubString:@"ApplyJoinChatRoomRequest" Data:className])
        {
            upRequest.set_payloadtype(PAYLOAD_JOIN_CHATROOM);
            upRequest.set_allocated_applyjoinchatroomreq((::qihoo::protocol::chatroom::ApplyJoinChatRoomRequest *)command);
        }
        else if ([IMUtil hasSubString:@"QuitChatRoomRequest" Data:className])
        {
            upRequest.set_payloadtype(PAYLOAD_QUIT_CHATROOM);
            upRequest.set_allocated_quitchatroomreq((::qihoo::protocol::chatroom::QuitChatRoomRequest *)command);
        }
        else if ([IMUtil hasSubString:@"SubscribeRequest" Data:className])
        {
            upRequest.set_payloadtype(PAYLOAD_SUBSCRIBE_CHATROOM);
            upRequest.set_allocated_subreq((::qihoo::protocol::chatroom::SubscribeRequest *)command);
        }
        else if ([IMUtil hasSubString:@"ChatRoomMessageRequest" Data:className])
        {
            upRequest.set_payloadtype(PAYLOAD_MESSAGE_CHATROOM);
            upRequest.set_allocated_chatroommessagereq((::qihoo::protocol::chatroom::ChatRoomMessageRequest *)command);
        }
        else
        {
//            CPLog(@"not support command type:%@", className);
            return [@"" UTF8String];
        }

        /*
         message ChatRoomUpToServer{
         required uint32                             payloadtype =1;

         optional CreateChatRoomRequest              createchatroomreq = 2;
         optional GetChatRoomDetailRequest           getchatroominforeq = 3;
         optional ApplyJoinChatRoomRequest           applyjoinchatroomreq = 4;
         optional QuitChatRoomRequest                quitchatroomreq = 5;
         optional UpdateChatRoomRequest              updatechatroomreq = 6;
         optional KickChatRoomMemberRequest          kickmemberreq = 7;

         optional QueryChatRoomIDRequest             querychatroomidreq = 8;
         optional UpdateRoomIDRequest                updategameidreq = 9;
         optional QueryAllGameRoomRequest            queryallgameroomreq = 10;


         optional ChatRoomMessageRequest             chatroommessagereq = 11;
         optional CreateMultiChatRoomRequest         createrooms = 12;
         
         
         optional SyncRoomToDBRequest                syncroomtodba = 13;
         optional SubscribeRequest                   subreq = 14;
         }
         */

        ChatRoomPacket packet;
        packet.set_allocated_to_server_data(&upRequest);
        packet.set_roomid([roomid UTF8String]);
        packet.set_appid(self.userConfig.appid);

        return packet.SerializeAsString();
    }
    @catch (NSException *exception)
    {
//        CPLog(@"createMessageString exception, name:%@, reason:%@", exception.name, exception.reason);
    }
    
    return "";
}

-(NSMutableDictionary*) roomToDict:(const ::qihoo::protocol::chatroom::ChatRoom &)room AddMyself:(BOOL)add
{
    NSMutableDictionary *roomDict = [[NSMutableDictionary alloc] init];
    [roomDict setObject:[IMUtil CharsToNSString:room.roomid().c_str()] forKey:@"roomid"];
    if (room.has_version())
    {
        [roomDict setObject:[NSNumber numberWithLongLong:room.version()] forKey:@"version"];
    }
    if (room.has_partnerdata())
    {
        [roomDict setObject:[IMUtil CharsToNSData:room.partnerdata().c_str() withLength:room.partnerdata().size()] forKey:@"partnerdata"];
    }
    BOOL hasMyself = false;
    NSMutableArray *userArray = [[NSMutableArray alloc] init];
    NSMutableDictionary *udataDict = [[NSMutableDictionary alloc] init];
    for (int i=0; i<room.members().size(); ++i)
    {
        const ::qihoo::protocol::chatroom::CRUser &user = room.members(i);
        NSString *userID = [IMUtil CharsToNSString:user.userid().c_str()];
        if ([userID isEqualToString:self.userConfig.jid])
        {
            hasMyself = true;
        }

        if (user.has_userdata())
        {
            [udataDict setObject:[IMUtil CharsToNSData:user.userdata().c_str() withLength:user.userdata().size()] forKey:userID];
        }
        [userArray addObject:userID];
    }

    if (add == YES && hasMyself == NO)
    {
        [userArray addObject:self.userConfig.jid];
    }

    if ([userArray count] > 0)
    {
        [roomDict setObject:userArray forKey:@"members"];
    }
    [roomDict setObject:udataDict forKey:@"udatadict"];

    for (int i=0; i<room.properties().size(); ++i)
    {
        const ::qihoo::protocol::chatroom::CRPair &pair = room.properties(i);
        NSString *key = [IMUtil CharsToNSString:pair.key().c_str()];
        NSString *value = [IMUtil CharsToNSString:pair.value().c_str()];

        if ([key isEqualToString:@"memcount"])
        {
            int count = [IMUtil NSStringToInt:value];
            /*
            if (add == YES && hasMyself == NO)
            {
                ++count;
            }
            */
            [roomDict setObject:[NSNumber numberWithInt:count] forKey:@"memcount"];
        }
        else if ([key isEqualToString:@"regmemcount"])
        {
            [roomDict setObject:[NSNumber numberWithInt:[IMUtil NSStringToInt:value]] forKey:@"regmemcount"];
        }
    }
    if (room.has_maxmsgid())
    {
        [roomDict setObject:[NSNumber numberWithLongLong:room.maxmsgid()] forKey:@"maxmsgid"];
    }
    else
    {
        [roomDict setObject:[NSNumber numberWithLongLong:0] forKey:@"maxmsgid"];
    }
        return roomDict;
}


/**
 * 根据接收到的二进制流，解析
 * @param data: 二进制流(std::string *)
 * @param msgID: 该Message的id，在函数内部改变这个值，调用后调用者应该根据这个msgID的值来判断返回的Message到底是什么类型的
 * @returns Message引用
 */
-(NSMutableDictionary *)parseMessage:(NSData*)data
{
    NSMutableDictionary *dataDict = [[NSMutableDictionary alloc] init];
    [dataDict setObject:[NSNumber numberWithInt:0] forKey:@"result"];

    @try
    {
        std::string strData;
        [IMUtil NSDataToStlString:data Return:&strData];

        ChatRoomPacket packet;
        packet.ParseFromString(strData);
        const ChatRoomDownToUser &downData = packet.to_user_data();
        [dataDict setObject:[NSNumber numberWithInt:downData.payloadtype()] forKey:@"payload"];

        if (downData.result() != 0)
        {
            [dataDict setObject:[NSNumber numberWithInt:downData.result()] forKey:@"result"];
            if (downData.has_reason())
            {
                [dataDict setObject:[IMUtil CharsToNSData:downData.reason().c_str() withLength:downData.reason().size()] forKey:@"reason"];
            }
        }
        else
        {
            if (downData.payloadtype() == PAYLOAD_QUERY_CHATROOM) //query room
            {
                [dataDict setObject:[self roomToDict:downData.getchatroominforesp().room() AddMyself:NO] forKey:@"room"];
            }
            else if (downData.payloadtype() == PAYLOAD_JOIN_CHATROOM) //join room
            {
                const ApplyJoinChatRoomResponse & resp = downData.applyjoinchatroomresp();
                [dataDict setObject:[self roomToDict:downData.applyjoinchatroomresp().room() AddMyself:YES] forKey:@"room"];
                [dataDict setObject:[NSNumber numberWithBool:resp.pull_lost() ? YES : NO] forKey:@"pull_lost"];
            }
            else if (downData.payloadtype() == PAYLOAD_QUIT_CHATROOM) //quit room
            {
                [dataDict setObject:[self roomToDict:downData.quitchatroomresp().room() AddMyself:NO] forKey:@"room"];
            }
            else if (downData.payloadtype() == PAYLOAD_SUBSCRIBE_CHATROOM) // subscribe room message result
            {
                [dataDict setObject:[IMUtil CharsToNSString:downData.subresp().roomid().c_str()] forKey:@"roomid"];
                [dataDict setObject:(downData.subresp().sub() ? [NSNumber numberWithBool:YES] : [NSNumber numberWithBool:NO]) forKey:@"subscribe"];
            }
            else if (downData.payloadtype() == PAYLOAD_NEW_MSG_NTF_CHATROOM) //new message notify
            {
                const ::qihoo::protocol::chatroom::ChatRoomNewMsg &roomMsg = downData.newmsgnotify();
                [dataDict setObject:[IMUtil CharsToNSString:roomMsg.roomid().c_str()] forKey:@"roomid"];
                [dataDict setObject:[IMUtil CharsToNSString:roomMsg.sender().userid().c_str()] forKey:@"sender"];
                [dataDict setObject:[NSNumber numberWithInt:roomMsg.msgtype()] forKey:@"msgtype"];
                [dataDict setObject:[IMUtil CharsToNSData:roomMsg.msgcontent().c_str() withLength:roomMsg.msgcontent().size()] forKey:@"msgbody"];
                if (roomMsg.has_memcount())
                {
                    [dataDict setObject:[NSNumber numberWithInt:roomMsg.memcount()] forKey:@"memcount"];
                }
                if (roomMsg.has_regmemcount())
                {
                    [dataDict setObject:[NSNumber numberWithInt:roomMsg.regmemcount()] forKey:@"regmemcount"];
                }
                
                if (roomMsg.has_msgid())
                {
                    [dataDict setObject:[NSNumber numberWithInt:roomMsg.msgid()] forKey:@"msgid"];
                }
                else
                {
                    [dataDict setObject:[NSNumber numberWithInt:0] forKey:@"msgid"];
                }
                
                if (roomMsg.has_maxid())
                {
                    [dataDict setObject:[NSNumber numberWithInt:roomMsg.maxid()] forKey:@"maxid"];
                }
                else
                {
                    [dataDict setObject:[NSNumber numberWithInt:0] forKey:@"maxid"];
                }
                
                if (roomMsg.has_timestamp())
                {
                    [dataDict setObject:[NSNumber numberWithInteger:roomMsg.timestamp()] forKey:@"timestamp"];
                }

            }
            else if (downData.payloadtype() == PAYLOAD_JOIN_NTF_CHATROOM) //member join notify
            {
                [dataDict setObject:[self roomToDict:downData.memberjoinnotify().room() AddMyself:NO] forKey:@"room"];
            }
            else if (downData.payloadtype() == PAYLOAD_QUIT_NTF_CHATROOM) //member quit notify
            {
                [dataDict setObject:[self roomToDict:downData.memberquitnotify().room() AddMyself:NO] forKey:@"room"];
            }
            else if (downData.payloadtype() == PAYLOAD_MIX_NTF_CHATROOM) //multiple notify
            {
                NSMutableArray *notifies = [[NSMutableArray alloc] init];
                for (int i=0; i<downData.multinotify_size(); ++i)
                {
                    NSMutableDictionary *roomDict = [[NSMutableDictionary alloc] init];
                    const ::qihoo::protocol::chatroom::ChatRoomMNotify &notify = downData.multinotify(i);
                    NSData *tmp1 = [IMUtil CharsToNSData:notify.data().c_str() withLength:notify.data().size()];
                    NSData *tmp2 = [IMUtil unzipData:tmp1];
                    std::string unzipData;
                    [IMUtil NSDataToStlString:tmp2 Return:&unzipData];
                    if (notify.type() == PAYLOAD_NEW_MSG_NTF_CHATROOM) //new msg notify
                    {
                        ::qihoo::protocol::chatroom::ChatRoomNewMsg msg;
                        msg.ParseFromString(unzipData);

                        [roomDict setObject:[IMUtil CharsToNSString:msg.roomid().c_str()] forKey:@"roomid"];
                        [roomDict setObject:[IMUtil CharsToNSString:msg.sender().userid().c_str()] forKey:@"sender"];
                        [roomDict setObject:[NSNumber numberWithInt:msg.msgtype()] forKey:@"msgtype"];
                        [roomDict setObject:[IMUtil CharsToNSData:msg.msgcontent().c_str() withLength:msg.msgcontent().size()] forKey:@"msgbody"];
                        
                        if (msg.has_msgid())
                        {
                            [roomDict setObject:[NSNumber numberWithInt:msg.msgid()] forKey:@"msgid"];
                        }
                        else
                        {
                            [roomDict setObject:[NSNumber numberWithInt:0] forKey:@"msgid"];
                        }
                        
                        if (msg.has_maxid())
                        {
                            [roomDict setObject:[NSNumber numberWithInt:msg.maxid()] forKey:@"maxid"];
                        }
                        else
                        {
                            [roomDict setObject:[NSNumber numberWithInt:0] forKey:@"maxid"];
                        }
                        
                        if (msg.has_timestamp())
                        {
                            [roomDict setObject:[NSNumber numberWithInteger:msg.timestamp()] forKey:@"timestamp"];
                        }
                        
                    }
                    else if (notify.type() == PAYLOAD_JOIN_NTF_CHATROOM) //join notify
                    {
                        ::qihoo::protocol::chatroom::MemberJoinChatRoomNotify msg;
                        msg.ParseFromString(unzipData);
                        roomDict = [self roomToDict:msg.room() AddMyself:NO];
                    }
                    else if (notify.type() == PAYLOAD_QUIT_NTF_CHATROOM) //quit notify
                    {
                        ::qihoo::protocol::chatroom::MemberQuitChatRoomNotify msg;
                        msg.ParseFromString(unzipData);
                        roomDict = [self roomToDict:msg.room() AddMyself:NO];
                    }
                    [roomDict setObject:[NSNumber numberWithInt:notify.type()] forKey:@"payload"];
                    if (notify.has_memcount())
                    {
                        [roomDict setObject:[NSNumber numberWithInt:notify.memcount()] forKeyedSubscript:@"memcount"];
                    }
                    if (notify.has_regmemcount())
                    {
                        [roomDict setObject:[NSNumber numberWithInt:notify.regmemcount()] forKeyedSubscript:@"regmemcount"];
                    }

                    [notifies addObject:roomDict];
                }
                [dataDict setObject:notifies forKey:@"roomlist"];
            }
        }
    }
    @catch (NSException *exception)
    {
//        CPLog(@"createMessageString exception, name:%@, reason:%@", exception.name, exception.reason);
    }
    
    return dataDict;
}


-(NSMutableDictionary *)parseNewMessage:(NSData*)data
{
    NSMutableDictionary *dataDict = [[NSMutableDictionary alloc] init];
    [dataDict setObject:[NSNumber numberWithInt:0] forKey:@"result"];
    
    @try
    {
        std::string strData;
        [IMUtil NSDataToStlString:data Return:&strData];
        
        ChatRoomNewMsg roomMsg;
        roomMsg.ParseFromString(strData);
        
        [dataDict setObject:[IMUtil CharsToNSString:roomMsg.roomid().c_str()] forKey:@"roomid"];
        [dataDict setObject:[IMUtil CharsToNSString:roomMsg.sender().userid().c_str()] forKey:@"sender"];
        [dataDict setObject:[NSNumber numberWithInt:roomMsg.msgtype()] forKey:@"msgtype"];
        [dataDict setObject:[IMUtil CharsToNSData:roomMsg.msgcontent().c_str() withLength:roomMsg.msgcontent().size()] forKey:@"msgbody"];
        
        if (roomMsg.has_memcount())
        {
            [dataDict setObject:[NSNumber numberWithInt:roomMsg.memcount()] forKeyedSubscript:@"memcount"];
        }
        if (roomMsg.has_regmemcount())
        {
            [dataDict setObject:[NSNumber numberWithInt:roomMsg.regmemcount()] forKeyedSubscript:@"regmemcount"];
        }
        
        if (roomMsg.has_msgid())
        {
            [dataDict setObject:[NSNumber numberWithInt:roomMsg.msgid()] forKey:@"msgid"];
        }
        else
        {
            [dataDict setObject:[NSNumber numberWithInt:0] forKey:@"msgid"];
        }
        
        if (roomMsg.has_maxid())
        {
            [dataDict setObject:[NSNumber numberWithInt:roomMsg.maxid()] forKey:@"maxid"];
        }
        else
        {
            [dataDict setObject:[NSNumber numberWithInt:0] forKey:@"maxid"];
        }
        if (roomMsg.has_timestamp())
        {
            [dataDict setObject:[NSNumber numberWithInteger:roomMsg.timestamp()] forKey:@"timestamp"];
        }

    }
    @catch (NSException *exception)
    {
//        CPLog(@"createMessageString exception, name:%@, reason:%@", exception.name, exception.reason);
    }
    
    return dataDict;
}




/**
 * 加入聊天室
 */
-(NSData*) createJoinRoomRequest:(NSString*)roomid withProperties:(NSDictionary*)properties
{
    return [self createJoinRoomRequest:roomid withData:nil Properties:properties];
}

-(NSData*) createJoinRoomRequest:(NSString*)roomid withData:(NSData*)userdata Properties:(NSDictionary*)properties
{
    ApplyJoinChatRoomRequest *command = new ApplyJoinChatRoomRequest();
    command->set_roomid([roomid UTF8String]);
    if (userdata != nil)
    {
        std::string strUserData;
        [IMUtil NSDataToStlString:userdata Return:&strUserData];
        command->set_userdata(strUserData);
    }

    // 设置加入聊天室属性，透传至服务器，因此无需检查内容
    ChatRoom* pRoom = command->mutable_room();
    if (pRoom != NULL) {
        pRoom->set_roomid([roomid UTF8String]);
        if (properties != nil) {
            for (id key in properties) {
                NSString* val = [IMUtil idToString:properties[key]];
                CRPair* pPair = pRoom->add_properties();
                if (pPair != NULL) {
                    pPair->set_key([[IMUtil idToString:key] UTF8String]);
                    pPair->set_value([val UTF8String]);
                } // if (pPair != NULL)
            } // for (id key in properties)
        } // if (properties != nil)
    } // if (pRoom != NULL)

    // 用户列表客户端自行从http接口获取，2016-08-26
    command->set_no_userlist(true);
    std::string strPacket = [self createChatrromPacket:command RoomID:roomid];
    return [IMUtil CharsToNSData:strPacket.c_str() withLength:strPacket.size()];
}

-(NSMutableDictionary*) parseJoinRoomResponse:(NSData*)data
{
    NSMutableDictionary *dataDict = [self parseMessage:data];
    if ([((NSNumber*)[dataDict objectForKey:@"payload"]) intValue] != 102)
    {
//        CPLog(@"this is not join response data");
    }

    return dataDict;
}


/**
 * 取聊天室详情
 */
-(NSData*) createQueryRoomRequest:(NSString*)roomid From:(int)from Count:(int)count
{
    GetChatRoomDetailRequest *command = new GetChatRoomDetailRequest();
    command->set_roomid([roomid UTF8String]);
    command->set_index(from);
    command->set_offset(count);
    std::string strPacket = [self createChatrromPacket:command RoomID:roomid];
    return [IMUtil CharsToNSData:strPacket.c_str() withLength:strPacket.size()];
}

-(NSMutableDictionary*) parseQueryRoomResponse:(NSData*)data
{
    NSMutableDictionary *dataDict = [self parseMessage:data];
    if ([((NSNumber*)[dataDict objectForKey:@"payload"]) intValue] != 101)
    {
//        CPLog(@"this is not query response data");
    }

    return dataDict;
}

/**
 * 退出聊天室
 */
-(NSData*) createQuitRoomRequest:(NSString*)roomid
{
    QuitChatRoomRequest *command = new QuitChatRoomRequest();
    command->set_roomid([roomid UTF8String]);
    std::string strPacket = [self createChatrromPacket:command RoomID:roomid];
    return [IMUtil CharsToNSData:strPacket.c_str() withLength:strPacket.size()];
}
-(NSMutableDictionary*) parseQuitRoomResponse:(NSData*)data
{
    NSMutableDictionary *dataDict = [self parseMessage:data];
    if ([((NSNumber*)[dataDict objectForKey:@"payload"]) intValue] != 103)
    {
//        CPLog(@"this is not quit response data");
    }

    return dataDict;
}

-(NSData*) createChatroomMessageRequest:(NSString*)roomid Message:(NSData*)content
{
    ChatRoomMessageRequest *command = new ChatRoomMessageRequest();
    command->set_roomid([roomid UTF8String]);
    command->set_msgtype(0);
    std::string strContent;
    [IMUtil NSDataToStlString:content Return:&strContent];
    command->set_msgcontent(strContent);

    std::string strPacket = [self createChatrromPacket:command RoomID:roomid];
    return [IMUtil CharsToNSData:strPacket.c_str() withLength:strPacket.size()];
}

/**
 * 取消/恢复订阅指定聊天室消息
 * @param sub: 是否订阅， YES订阅；NO不订阅
 * @param roomid: 指定聊天室
 */
-(NSData*) createSubscribe:(BOOL)sub Request:(NSString*)roomid
{
    SubscribeRequest * command = new SubscribeRequest();
    command->set_roomid([roomid UTF8String]);
    if (sub) {
        command->set_sub(true);
    }
    else {
        command->set_sub(false);
    }
    
    std::string strPacket = [self createChatrromPacket:command RoomID:roomid];
    return [IMUtil CharsToNSData:strPacket.c_str() withLength:strPacket.size()];
}

-(NSMutableDictionary*) parseSubscribeResponse:(NSData*)data
{
    NSMutableDictionary *dataDict = [self parseMessage:data];
    if ([((NSNumber*)[dataDict objectForKey:@"payload"]) intValue] != PAYLOAD_SUBSCRIBE_CHATROOM) {
//        CPLog(@"this is not subscribe response data");
    }
    
    return dataDict;
}

@end


@implementation IMChatRoomMsgLost

-(instancetype)init
{
    self = [super init];
    if (self) {
        self.msgLostTime = -1;
        self.msgReloadTime = -1;
    }
    return self;
}



@end
