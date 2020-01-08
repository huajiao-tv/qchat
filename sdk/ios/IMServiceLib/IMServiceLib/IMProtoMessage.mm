//
//  IMProtoMessage.m
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-26.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMProtoMessage.h"
#import "IMProtoUserInfo.h"
#import "IMProtoChatRoom.h"
#import "IMProtoGroupChat.h"
#import "IMUtil.h"
#import "IMUser.h"
#import "address_book.pb.h"
#import "vcproxy.pb.h"

#import "IMServiceLib.h"

using namespace qihoo::protocol::messages;
using namespace qihoo::protocol::vcproxy;

@interface IMProtoMessage()

/**
 * 根据命令的类名称，生成Msg Message
 */
-(std::string) createMessageString:(google::protobuf::Message*)command Sn:(UInt64)sn RecvType:(NSString*)type RecvID:(NSString*)recvID UserReserve:(UInt64)reserveID;

/**
 * 根据接收到的二进制流，解析
 * @param data: 二进制流(std::string *)
 * @param msgID: 该Message的id，在函数内部改变这个值，调用后调用者应该根据这个msgID的值来判断返回的Message到底是什么类型的
 * @returns Message引用
 */
-(const google::protobuf::Message *)parseMessage:(std::string &)data MsgMessage:(qihoo::protocol::messages::Message &)msg MsgID:(int *)msgID SN:(int64_t *)sn;

@end

/**
 * address book协议的实现文件
 */
@implementation IMProtoMessage

-(void) initData
{
    //必须先调用父类的initData
    [super initData];

    //添加请求关联关系
    /*
     message Request {
     optional LoginReq login                          = 2;    //msgid = 100001
     optional ChatReq chat                            = 3;    //msgid = 100002
     optional GetInfoReq get_info                     = 5;    //msgid = 100004
     optional LogoutReq logout                        = 6;    //msgid = 100005
     //HeartBeatReq ,msgid = 100006
     optional InitLoginReq  init_login_req            = 9;    //msgid = 100009
     optional Service_Req service_req                 = 11;   //msgid = 100011
     optional Ex1QueryUserStatusReq e1_query_user     = 12;   //msgid = 100012
     optional QueryQuotaReq		query_quota			 = 15;	 //msgid = 100015
     optional UpdateConvSummaryReq update_conv_summary = 16;  //msgid = 100016

     optional GetMultiInfosReq get_multi_infos        = 100;  //msgid = 100100

     }
     */
    [self addRequestMap:@"LoginReq" ID:MSG_ID_REQ_LOGIN];
    [self addRequestMap:@"ChatReq" ID:MSG_ID_REQ_CHAT];
    [self addRequestMap:@"GetInfoReq" ID:MSG_ID_REQ_GET_INFO];
    [self addRequestMap:@"LogoutReq" ID:MSG_ID_REQ_LOGOUT];
    [self addRequestMap:@"InitLoginReq" ID:MSG_ID_REQ_INIT_LOGIN];
    [self addRequestMap:@"Service_Req" ID:MSG_ID_REQ_SERVICE_CONTROL];
    [self addRequestMap:@"Ex1QueryUserStatusReq" ID:MSG_ID_REQ_EX1_QUERY_USER_STATUS];
    [self addRequestMap:@"QueryQuotaReq" ID:MSG_ID_REQ_QUERY_QUOTA];
    [self addRequestMap:@"UpdateConvSummaryReq" ID:MSG_ID_REQ_UPDATE_CONVSUMMARY];
    [self addRequestMap:@"GetMultiInfosReq" ID:MSG_ID_GET_MULTI_INFOS_REQ];

    //添加应答关联关系
    /*
     message Response {
     optional Error error                             = 1;
     optional RestoreSessionResp restore              = 2;    //msgid = 200000
     optional LoginResp login                         = 3;    //msgid = 200001
     optional ChatResp chat                           = 4;    //msgid = 200002
     optional GetInfoResp get_info                    = 6;    //msgid = 200004
     optional LogoutResp logout                       = 7;    //msgid = 200005
     //reserve msgid = 200006 for HeartBeatResp
     optional QueryUserStatusResp query_user          = 8;    //msgid = 200007
     optional QueryUserRegResp  query_reg             = 9;    //msgid = 200008
     optional InitLoginResp  init_login_resp          = 10;   //msgid = 200009
     optional ExQueryUserStatusResp e_query_user      = 11;   //msgid = 200010
     optional Service_Resp service_resp               = 12;   //msgid = 200011
     optional Ex1QueryUserStatusResp e1_query_user    = 13;   //msgid = 200012
     optional QueryConvSummaryResp query_conv_summary = 15;   //msgid = 200014
     optional QueryQuotaResp		query_quota			 = 16;	 //msgid = 200015
     optional UpdateConvSummaryResp update_conv_summary = 17; //msgid = 200016
     optional GetMultiInfosResp get_multi_infos       = 100;  //msgid = 200100
     }
     optional NewMessageNotify newinfo_ntf   = 1;    //msgid = 300000
     optional ReLoginNotify    relogin_ntf   = 2;    //msgid = 300001
     optional ReConnectNotify  reconnect_ntf = 3;    //msgid = 300002
     */
    [self addResponseMap:@"LoginResp" ID:MSG_ID_RESP_LOGIN];
    [self addResponseMap:@"ChatResp" ID:MSG_ID_RESP_CHAT];
    [self addResponseMap:@"GetInfoResp" ID:MSG_ID_RESP_GET_INFO];
    [self addResponseMap:@"LogoutResp" ID:MSG_ID_RESP_LOGOUT];
    [self addResponseMap:@"QueryUserStatusResp" ID:MSG_ID_RESP_QUERY_USER_STATUS];
    [self addResponseMap:@"QueryUserRegResp" ID:MSG_ID_RESP_QUERY_USER_REG];
    [self addResponseMap:@"InitLoginResp" ID:MSG_ID_RESP_INIT_LOGIN];
    [self addResponseMap:@"ExQueryUserStatusResp" ID:MSG_ID_RESP_EX_QUERY_USER_STATUS];
    [self addResponseMap:@"Service_Resp" ID:MSG_ID_RESP_SERVICE_CONTROL];
    [self addResponseMap:@"Ex1QueryUserStatusResp" ID:MSG_ID_RESP_EX1_QUERY_USER_STATUS];
    [self addResponseMap:@"GetMultiInfosResp" ID:MSG_ID_RESP_GET_MULTI_INFOS];
    [self addResponseMap:@"QueryConvSummaryResp" ID:MSG_ID_RESP_QUERY_CONVSUMMARY];
    [self addResponseMap:@"QueryQuotaResp" ID:MSG_ID_RESP_QUERY_QUOTA];
    [self addResponseMap:@"UpdateConvSummaryResp" ID:MSG_ID_RESP_UPDATE_CONVSUMMARY];
    [self addResponseMap:@"NewMessageNotify" ID:MSG_ID_NTF_NEW_MESSAGE];
    [self addResponseMap:@"ReLoginNotify" ID:MSG_ID_NTF_RELOGIN];
    [self addResponseMap:@"ReConnectNotify" ID:MSG_ID_NTF_RECONNECT];
}

-(std::string) createMessageString:(google::protobuf::Message *)command Sn:(UInt64)sn RecvType:(NSString*)type
RecvID:(NSString*)recvID UserReserve:(UInt64)reserveID
{
    @try
    {
        qihoo::protocol::messages::Request *pRequest = new qihoo::protocol::messages::Request();
        qihoo::protocol::messages::Request &msgRequest = *pRequest;
        std::string tmp = typeid(*command).name();
        NSString *className = [IMUtil CharsToNSString:tmp.c_str()];

        NSString *commandName = @"";
        //根据command生成Message的字符串
        if ([IMUtil hasSubString:@"InitLoginReq" Data:className])
        {
            commandName = @"InitLoginReq";
            msgRequest.set_allocated_init_login_req((qihoo::protocol::messages::InitLoginReq*)(command));
        }
        else if ([IMUtil hasSubString:@"LoginReq" Data:className])
        {
            commandName = @"LoginReq";
            msgRequest.set_allocated_login((qihoo::protocol::messages::LoginReq*)(command));
        }
        else if ([IMUtil hasSubString:@"GetInfoReq" Data:className])
        {
            commandName = @"GetInfoReq";
            msgRequest.set_allocated_get_info((qihoo::protocol::messages::GetInfoReq*)(command));
        }
        else if ([IMUtil hasSubString:@"ChatReq" Data:className])
        {
            commandName = @"ChatReq";
            msgRequest.set_allocated_chat((qihoo::protocol::messages::ChatReq*)(command));
        }
        else if ([IMUtil hasSubString:@"Ex1QueryUserStatusReq" Data:className])
        {
            commandName = @"Ex1QueryUserStatusReq";
            msgRequest.set_allocated_e1_query_user((qihoo::protocol::messages::Ex1QueryUserStatusReq*)(command));
        }
        else if ([IMUtil hasSubString:@"GetMultiInfosReq" Data:className])
        {
            commandName = @"GetMultiInfosReq";
            msgRequest.set_allocated_get_multi_infos((qihoo::protocol::messages::GetMultiInfosReq*)(command));
        }
        else
        {
//            CPLog(@"not support command type:%@", className);
        }

        /*
         message Message {
         required uint32 msgid         = 1;  //message type
         required uint64 sn            = 2;  //Response's sn equal Request's sn;Ack's sn equal Notify's sn
         optional string sender        = 3;  //jid or qid
         optional string receiver      = 4;  //phonenumber or qid
         optional string receiver_type = 5;  //default:phone, qid,  other service: null
         optional Request req          = 6;
         optional Response resp        = 7;
         optional Notify notify        = 8;
         optional Ack ack              = 9;
         optional Proxy proxy_mesg     = 10;
         optional uint64 client_data   = 11; //reverse for user_data
         optional string sender_type   = 12; //default:jid, qid
         optional string sender_jid    = 13;
         }
         */

        qihoo::protocol::messages::Message msgMessage;
        //[step1] msgid
        msgMessage.set_msgid([self getRequestIDByName:commandName]);
        //[step2]: sn
        if (sn == 0)
        {
            msgMessage.set_sn([IMUtil createSN]);
        }else{
            msgMessage.set_sn(sn);
        }

        //[step3]: sender(jid or qid), sender_type, sender_jid
        //        if (self.userConfig.jid != nil && [self.userConfig.jid length] > 0)
        //        {
        //            msgMessage.set_sender_type("jid");
        //            msgMessage.set_sender([self.userConfig.jid UTF8String], [self.userConfig.jid length]);
        //            msgMessage.set_sender_jid([self.userConfig.jid UTF8String], [self.userConfig.jid length]);
        //        }
        //        else if (self.userConfig.qid != nil && [self.userConfig.qid length] > 0)
        //        {
        //            msgMessage.set_sender_type("qid");
        //            msgMessage.set_sender([self.userConfig.qid UTF8String], [self.userConfig.qid length]);
        //        }

        //谈谈仅支持jid登录
        if ([IMUtil hasSubString:@"InitLoginReq" Data:className] || [IMUtil hasSubString:@"LoginReq" Data:className]) {
            if (self.userConfig.jid !=nil && [self.userConfig.jid length] > 0) {
                msgMessage.set_sender_type("jid");
                msgMessage.set_sender([self.userConfig.jid UTF8String]);
            }
        }

        //[step4]: receiver
        if (recvID != nil && [recvID length] > 0)
        {
            msgMessage.set_receiver([recvID UTF8String], [recvID length]);
        }
        //[step5]: receiver_type
        if (type != nil && [type length] > 0)
        {
            msgMessage.set_receiver_type([type UTF8String], [type length]);
        }
        //[step6]: req
        msgMessage.set_allocated_req(&msgRequest);
        //[step7]: client_data
        msgMessage.set_client_data(reserveID);

        return msgMessage.SerializeAsString();
    }
    @catch (NSException *exception)
    {
//        CPLog(@"createMessageString exception, name:%@, reason:%@", exception.name, exception.reason);
    }

    return "";

}

/**
 * 根据接收到的二进制流，解析
 * @param data: 二进制流
 * @param msgID: 该Message的id，在函数内部改变这个值，调用后调用者应该根据这个msgID的值来判断返回的Message到底是什么类型的
 * @returns Message引用
 */
-(const google::protobuf::Message *)parseMessage:(std::string &)data MsgMessage:(qihoo::protocol::messages::Message &)msg MsgID:(int *)msgID SN:(int64_t *)sn
{
    qihoo::protocol::messages::Message &resMessage = msg;
    resMessage.ParseFromString(data);

    //赋值msgid
    int payloadType = resMessage.msgid();
    if (resMessage.has_msgid())
    {
        *msgID = payloadType;
    }
    else
    {
        *msgID = 0;
    }

    if (resMessage.has_sn())
    {
        *sn = resMessage.sn();
    }
    else
    {
        *sn = 0;
    }

    /*
     [self addResponseMap:@"RestoreSessionResp" ID:200000];
     [self addResponseMap:@"LoginResp" ID:200001];
     [self addResponseMap:@"ChatResp" ID:200002];
     [self addResponseMap:@"GetInfoResp" ID:200004];
     [self addResponseMap:@"LogoutResp" ID:200005];
     [self addResponseMap:@"QueryUserStatusResp" ID:200007];
     [self addResponseMap:@"QueryUserRegResp" ID:200008];
     [self addResponseMap:@"InitLoginResp" ID:200009];
     [self addResponseMap:@"ExQueryUserStatusResp" ID:2000010];
     [self addResponseMap:@"Service_Resp" ID:200011];
     [self addResponseMap:@"Ex1QueryUserStatusResp" ID:200012];
     [self addResponseMap:@"QueryConvSummaryResp" ID:200014];
     [self addResponseMap:@"NewMessageNotify" ID:300000];
     [self addResponseMap:@"ReLoginNotify" ID:300001];
     [self addResponseMap:@"ReConnectNotify" ID:300002];
     */
    const qihoo::protocol::messages::Response &response = resMessage.resp();
    if (response.has_error())
    {
        //出错了
        const qihoo::protocol::messages::Error error = response.error();
//        CPLog(@"has error, error code:%d, reason:%s", error.id(), error.description().c_str());
        return &resMessage;
    }

    switch (payloadType)
    {
        case MSG_ID_RESP_LOGIN: //LoginResp
        {
            if (response.has_login())
            {
                return &(response.login());
            }
            else
            {
//                CPLog(@"init login resp not exist");
            }
        }
            break;
        case MSG_ID_RESP_CHAT: //ChatResp
        {
            if (response.has_chat())
            {
                return &(response.chat());
            }
            else
            {
//                CPLog(@"chat response not exists");
            }
        }
            break;

        case MSG_ID_RESP_GET_INFO: //GetInfoResp
        {
            if (response.has_get_info())
            {
                return &(response.get_info());
            }
            else
            {
//                CPLog(@"GetInfoResp not exist");
            }
        }
            break;
        case MSG_ID_RESP_LOGOUT: //LogoutResp
        {
        }
            break;
        case MSG_ID_RESP_QUERY_USER_STATUS: //QueryUserStatusResp
        {
        }
            break;
        case MSG_ID_RESP_QUERY_USER_REG: //QueryUserRegResp
        {
        }
            break;
        case MSG_ID_RESP_INIT_LOGIN: //InitLoginResp
        {
            if (response.has_init_login_resp())
            {
                return &(response.init_login_resp());
            }
            else
            {
//                CPLog(@"init login resp not exist");
            }
        }
            break;
        case MSG_ID_RESP_EX_QUERY_USER_STATUS: //ExQueryUserStatusResp
        {
        }
            break;
        case MSG_ID_RESP_SERVICE_CONTROL: //Service_Resp
        {
            if (response.has_service_resp())
            {
                return &(response.service_resp());
            }
            else
            {
//                CPLog(@"Service_Resp not exist");
            }
        }
            break;
        case MSG_ID_RESP_EX1_QUERY_USER_STATUS: //Ex1QueryUserStatusResp
        {
            if (response.has_e1_query_user())
            {
                return &(response.e1_query_user());
            }
            else
            {
//                CPLog(@"Ex1QueryUserStatusResp not exist");
            }
        }
            break;

        case MSG_ID_RESP_GET_MULTI_INFOS: //GetMultiInfosResp
        {

            if (response.has_get_multi_infos()) {

                return &(response.get_multi_infos());
            }
            else
            {
//                CPLog(@"GetMultiInfosResp not exist");
            }

        }
            break;

        case MSG_ID_NTF_NEW_MESSAGE: //NewMessageNotify
        {
            if (resMessage.has_notify() && resMessage.notify().has_newinfo_ntf())
            {
                return &(resMessage.notify().newinfo_ntf());
            }
            else
            {
//                CPLog(@"NewMessageNotify is empty");
            }
        }
            break;
        case MSG_ID_NTF_RELOGIN: //ReLoginNotify
        {
            if (resMessage.has_notify() && resMessage.notify().has_relogin_ntf())
            {
                return &(resMessage.notify().relogin_ntf());
            }
            else
            {
//                CPLog(@"ReLoginNotify is empty");
            }
        }
            break;
        case MSG_ID_NTF_RECONNECT: //ReConnectNotify
        {
            if (resMessage.has_notify() && resMessage.notify().has_reconnect_ntf())
            {
                return &(resMessage.notify().reconnect_ntf());
            }
            else
            {
//                CPLog(@"ReConnectNotify is empty");
            }
        }
            break;
        default:
//            CPLog(@"not support response%d", payloadType);
            break;
    }

    return NULL;
}

/**
 * 根据接收到的二进制流，解析
 * @param data: 二进制流
 * @param msgID: 该Message的id，在函数内部改变这个值，调用后调用者应该根据这个msgID的值来判断返回的Message到底是什么类型的
 * @returns Message引用
 */
-(BOOL)tryParseMessage:(std::string &)data
{
    try
    {
        qihoo::protocol::messages::Message resMessage;
        resMessage.ParseFromString(data);

        if (!resMessage.has_msgid() || resMessage.msgid() == 0)
        {
            return false;
        }
        else
        {
            return true;
        }
    }
    catch (std::exception &ex) {
        return false;
    }
}

/**
 * 创建initLogin请求
 */
-(NSData*) createInitLoginRequest
{
    /*
     message InitLoginReq {         //msgid = 100009
     required string client_ram      = 1;
     optional string sig             = 2; // signature of token
     }
     */
    InitLoginReq *command = new InitLoginReq();

    NSString *random = [IMUtil createRandomString:10];
    command->set_client_ram([random UTF8String], [random length]);

    //如果token不为nil，则放入请求里
    if (self.userConfig.sigToken != nil)
    {
        command->set_sig([IMUtil NSStringToChars:self.userConfig.sigToken]);
    }

    std::string result = [self createMessageString:command Sn:0 RecvType:@"" RecvID:@"" UserReserve:0];
    return [self createDefaultKeyOutData:&result];
}

/**
 * 解析InitLogin应答
 * @param data:收到的数据
 * @returns 返回的参数
 */
-(NSMutableDictionary*) parseInitLoginResponse:(NSData*)data
{
    /*
     message InitLoginResp {        //msgid = 200009
     required string client_ram      = 1;
     required string server_ram      = 2;
     }
     */
    std::string result;
    [IMUtil rc4Decode:data Key:self.userConfig.defaultKey Return:&result];
    NSMutableDictionary *dataDict = [[NSMutableDictionary alloc] init];
    [dataDict setObject:[NSNumber numberWithInt:IM_InvalidParam] forKey:@"code"];
    qihoo::protocol::messages::Message msgMessage;
    int msgID;
    int64_t sn;
    const google::protobuf::Message &parsedMsg = *([self parseMessage:result MsgMessage:(msgMessage) MsgID:(&msgID) SN:&sn]);

    if (msgID != 200009)
    {
        return dataDict;
    }
    qihoo::protocol::messages::InitLoginResp &response = (qihoo::protocol::messages::InitLoginResp &)parsedMsg;
    //直接改写config data
    if (response.has_server_ram())
    {
        self.userConfig.randomKey = [IMUtil CharsToNSString:response.server_ram().c_str()];
        [dataDict setObject:[NSNumber numberWithInt:IM_Success] forKey:@"code"];
    }

    return dataDict;
}

/**
 * 创建login请求
 * @param netType: 网络类型， "wifi", "gprs"
 */
-(NSData*) createLoginRequest:(int)netType
{
    /*
     //plain txt
     message LoginReq {    //msgid = 100001
     required string mobile_type = 1;                   //android, ios, pc
     required uint32 net_type    = 2;                    //net_type: 0:unkonwn 1:2g 2:3g 3:wifi 4:ethe
     required string server_ram  = 3;
     optional bytes  secret_ram  = 4;                    // secret_server + 8 char ram, the field crypt use user's password
     optional uint32 app_id      = 5 [default = 2000];   // application id , default value is 2000
     optional uint32 heart_feq   = 6 [default = 300];    // heart frequency of second, default value is 300
     optional string deviceid    = 7 [default = ""];
     optional string platform    = 8;                    // web, pc, mobile
     optional string verf_code   = 9;                    // verification code
     optional bool not_encrypt = 10;                  // indicates whether client supports communicates without encrypt, true means support
     }
     */
    qihoo::protocol::messages::LoginReq *command = new qihoo::protocol::messages::LoginReq();

    command->set_mobile_type("ios");
    command->set_net_type(netType);
    command->set_server_ram([self.userConfig.randomKey UTF8String]);
    NSString *secretRam = [self.userConfig.randomKey stringByAppendingString:[IMUtil createRandomString:8]];

    std::string strSecretRam;
    [IMUtil rc4Encode:secretRam Key:self.userConfig.password Return:&strSecretRam];
    command->set_secret_ram(strSecretRam);
    command->set_app_id(self.userConfig.appid);
    command->set_heart_feq(self.userConfig.hbInterval);
    if (self.userConfig.deviceToken != nil) {
        command->set_deviceid([self.userConfig.deviceToken UTF8String]);
    }
    command->set_platform("mobile");
    //verify code 是加盐的jid: "jid360tantan@1408$";
    NSString *tmp = [NSString stringWithFormat:@"%@360tantan@1408$", self.userConfig.jid];
    NSString *saltVerCode = [IMUtil getMd5_32Bit_String:tmp];
    //    CPLog(@"salt code:%@", [saltVerCode substringFromIndex:24]);
    if (self.userConfig.sigToken == nil)
    {
        command->set_verf_code([[saltVerCode substringFromIndex:24] UTF8String]);
    }
    // 本版本支持不使用session key加密通信
    command->set_not_encrypt(TRUE);

    std::string result = [self createMessageString:command Sn:0 RecvType:@"" RecvID:@"" UserReserve:0];
    return [self createEncryptData:&result Key:self.userConfig.defaultKey];

}

/**
 * 创建login请求
 * @param netType: 网络类型， "wifi", "gprs"
 */
-(NSData*) createWeimiLoginRequest:(int)netType
{
    /*
     //plain txt
     message LoginReq {    //msgid = 100001
     required string mobile_type = 1;                   //android, ios, pc
     required uint32 net_type    = 2;                    //net_type: 0:unkonwn 1:2g 2:3g 3:wifi 4:ethe
     required string server_ram  = 3;
     optional bytes  secret_ram  = 4;                    // secret_server + 8 char ram, the field crypt use user's password
     optional uint32 app_id      = 5 [default = 2000];   // application id , default value is 2000
     optional uint32 heart_feq   = 6 [default = 300];    // heart frequency of second, default value is 300
     optional string deviceid    = 7 [default = ""];
     optional string platform    = 8;                    // web, pc, mobile
     }
     */
    qihoo::protocol::messages::LoginReq *command = new qihoo::protocol::messages::LoginReq();

    command->set_mobile_type("ios");
    command->set_net_type(netType);
    command->set_server_ram([self.userConfig.randomKey UTF8String]);
    NSString *secretRam = [self.userConfig.randomKey stringByAppendingString:[IMUtil createRandomString:8]];

    std::string strSecretRam;
    [IMUtil rc4Encode:secretRam Key:self.userConfig.password Return:&strSecretRam];
    command->set_secret_ram(strSecretRam);
    command->set_app_id(self.userConfig.appid);
    command->set_heart_feq(self.userConfig.hbInterval);
    command->set_deviceid([self.userConfig.deviceToken UTF8String]);
    command->set_platform("mobile");

    std::string result = [self createMessageString:command Sn:0 RecvType:@"" RecvID:@"" UserReserve:0];
    return [self createEncryptData:&result Key:self.userConfig.defaultKey];

}


/**
 * 解析登录应答
 */
-(NSMutableDictionary *) parseLoginResponse:(NSData*)data
{
    /*
     //crypt txt, key is md5(password)
     message LoginResp {   //msgid = 200001
     required uint32 timestamp       = 1;
     required string session_id      = 2;
     required string session_key     = 3;
     optional string client_login_ip = 4;
     optional string serverip        = 5;//for debug
     }
     */

    std::string result;
    NSMutableDictionary *dataDict = [[NSMutableDictionary alloc] init];
    [dataDict setObject:[NSNumber numberWithInt:IM_InvalidParam] forKey:@"code"];
    qihoo::protocol::messages::Message msgMessage;
    int payloadType;
    int64_t sn;
    //这个地方比较奇葩，登录成功的应答用user password加密，登录失败用default_key加密的，所以这里需要先尝试用password解密，如果失败再用default key解密
    [IMUtil rc4Decode:data Key:self.userConfig.password Return:&result];
    //CPLog(@"parseLoginResponse, input data length:%d, after decrypt:%lu", [data length], result.size());

    const google::protobuf::Message *pMessage = NULL;
    if ([self tryParseMessage:result])
    {
        pMessage = [self parseMessage:result MsgMessage:(msgMessage) MsgID:(&payloadType) SN:&sn];
    }
    else
    {
        [IMUtil rc4Decode:data Key:self.userConfig.defaultKey Return:&result];
        pMessage = [self parseMessage:result MsgMessage:(msgMessage) MsgID:(&payloadType) SN:&sn];
    }

    //防止类型错误以及pMessage为NULL
    if (payloadType == MSG_ID_RESP_LOGIN && pMessage != NULL) {
        //如果出错了，解析error.id
        if (typeid (*pMessage) == typeid (qihoo::protocol::messages::Message))
        {
            const qihoo::protocol::messages::Response &response = ((qihoo::protocol::messages::Message*)pMessage)->resp();
            if (response.has_error())
            {
                //出错了
                const qihoo::protocol::messages::Error error = response.error();

                if (error.has_id()) {
                    [dataDict setObject:[NSNumber numberWithInt:error.id()] forKey:@"error"];
                }

//                CPLog(@"has error, error code:%d, reason:%s", error.id(), error.description().c_str());
            }
        }

        if (msgMessage.has_sender() && [IMUtil isStringEmpty:self.userConfig.jid])
        {
            self.userConfig.jid =  [IMUtil CharsToNSString:msgMessage.sender().c_str()];
            IMUser* user = self.imUser;
        }

        [dataDict setObject:[NSNumber numberWithLongLong:sn] forKey:@"sn"];
        qihoo::protocol::messages::LoginResp &response = (qihoo::protocol::messages::LoginResp &)(*pMessage);

        [dataDict setObject:[NSNumber numberWithInt:response.timestamp()] forKey:@"timestamp"];

        if (response.has_session_key())
        {
            double time = [[NSDate date]timeIntervalSince1970];
//            CPLog(@"%d,%f,%f",response.timestamp(),time, time - response.timestamp());
            self.userConfig.sessionKey = [IMUtil CharsToNSString:response.session_key().c_str()];
            [dataDict setObject:[NSNumber numberWithInt:IM_Success] forKey:@"code"];

            [self.imUser writeLog:[NSString stringWithFormat:@"login success,skey:%@", self.userConfig.sessionKey]];
        }

        if (response.has_serverip())
        {
            //CPLog(@"msgrouter ip: %s", response.serverip().c_str());
        }
    } // if (payloadType == MSG_ID_RESP_LOGIN && pMessage != NULL)
    else
    {
        if (payloadType == MSG_ID_RESP_LOGIN)
        {
            [self.imUser writeLog:[NSString stringWithFormat:@"login fail,data:%@,pwd:%@", data, self.userConfig.password]];
        }
        else
        {
            [self.imUser writeLog:[NSString stringWithFormat:@"login fail,pt:%d", payloadType]];
        }
    }

    return dataDict;
}


/**
 * 解析服务器数据，数据可以是通知，getinfo的应答，发送数据的ack等
 * @param data: 收到的数据
 * @returns 返回解析参数
 */
-(NSMutableDictionary*) parseServerData:(NSData*)data
{
    std::string result;
    NSMutableDictionary *dataDict = [[NSMutableDictionary alloc] init];
    [dataDict setObject:[NSNumber numberWithInt:IM_Success] forKey:@"code"];
    [dataDict setObject:[NSNumber numberWithInt:0] forKey:@"sn"];

    qihoo::protocol::messages::Message msgMessage;
    int payloadType;
    int64_t sn;
    NSString *key = self.userConfig.sessionKey;
    if (key != nil && key.length > 0) {
        [IMUtil rc4Decode:data Key:self.userConfig.sessionKey Return:&result];
    }
    else { // 服务器降级的时候登录时会返回session key为空字符串，不加密
        [IMUtil NSDataToStlString:data Return:&result];
    }
    //CPLog(@"input data lenght:%d, after decrypt length:%lu", [data length], result.size());

    const google::protobuf::Message *pMessage = NULL;
    pMessage = [self parseMessage:result MsgMessage:(msgMessage) MsgID:(&payloadType) SN:&sn];
    if (msgMessage.has_sn())
    {
        [dataDict setObject:[NSNumber numberWithUnsignedLongLong:msgMessage.sn()] forKey:@"sn"];
    }

    if (msgMessage.resp().has_error())
    {
        [dataDict setObject:[NSNumber numberWithInt:msgMessage.resp().error().id()] forKey:@"code"];
        return dataDict;
    }

    if (msgMessage.has_msgid())
    {
        [dataDict setObject:[NSNumber numberWithInt:payloadType] forKey:@"msgid"];
    }

    if (pMessage == NULL) {
        return dataDict;
    }

    switch (payloadType)
    {
        case MSG_ID_NTF_NEW_MESSAGE: //NewMessageNotify
        {
            /*
             message NewMessageNotify {    //msgid = 300000
             required string info_type    = 1; //peer, push, group, status, chatroom ...
             optional bytes  info_content = 2;
             optional int64  info_id      = 3; //the message's id.
             optional uint32 query_after_seconds = 4; //client send query after this seconds
             }
             */
            qihoo::protocol::messages::NewMessageNotify &resp = (qihoo::protocol::messages::NewMessageNotify &)(*pMessage);

            NSString *strInfoType = nil;
            if (resp.has_info_type())
            {
                strInfoType = [IMUtil CharsToNSString:resp.info_type().c_str()];
            }
            if (strInfoType != nil)
            {
                [dataDict setObject:strInfoType forKey:@"info_type"];
                NSData *infoContent = nil;
                if (resp.has_info_content())
                {
                    infoContent = [IMUtil CharsToNSData:resp.info_content().c_str() withLength:resp.info_content().size()];
                }
                if (infoContent != nil)
                {
                    if ([strInfoType isEqualToString:@"chatroom"])
                    {
                        NSMutableDictionary *chatroomDict = [self.imUser.protoChatRoom parseMessage:infoContent];
                        [dataDict setObject:[self.imUser.protoChatRoom parseMessage:infoContent] forKey:@"chatroom"];
                    }
                    else if  ([strInfoType isEqualToString:@"group"])
                    {
                        [dataDict setObject:[self.imUser.protoGroupChat parseGroupDownPacket:infoContent] forKey:@"group"];
                    }
                    else
                    {
                        [dataDict setObject:infoContent  forKey:@"info_content"];
                    }
                }

                if (resp.has_info_id())
                {
                    [dataDict setObject:[NSNumber numberWithInteger:resp.info_id()] forKey:@"info_id"];
                }

                if (resp.has_query_after_seconds())
                {
                    [dataDict setObject:[NSNumber numberWithInt:resp.query_after_seconds()] forKey:@"query_after_seconds"];
                }
            }
        }
            break;
            
        case MSG_ID_NTF_RECONNECT:
        {
            /*
            message ReConnectNotify {     //msgid = 300002
                optional string ip       = 1; // new dispatcher's ip
                optional uint32 port     = 2; // new dispatcher's port
                repeated string more_ips = 3; // more candidate ips
            }
             */
            qihoo::protocol::messages::ReConnectNotify &resp = (qihoo::protocol::messages::ReConnectNotify &)(*pMessage);
            NSMutableArray* ips = [[NSMutableArray alloc]init];
            std::string ip = resp.ip();
            if (resp.has_ip() && ip.length() > 0) {
                [ips addObject:[IMUtil CharsToNSString:ip.c_str()]];
            }
            for (int i = 0; i < resp.more_ips_size(); ++i) {
                ip = resp.more_ips(i);
                if (ip.length() > 0) {
                    [ips addObject:[IMUtil CharsToNSString:ip.c_str()]];
                }
            }
            [dataDict setObject:ips forKey:@"ips"];
            [dataDict setObject:[NSNumber numberWithUnsignedInt:resp.port()] forKey:@"port"];
        }
            break;

        case MSG_ID_RESP_GET_INFO: //GetInfoResp
        {
            /*
             message Pair {
             required bytes key   = 1;
             optional bytes value = 2;
             }

             message Info {
             repeated Pair property_pairs = 1;    //key=info_id,sender,content,time
             }

             message GetInfoResp {    //msgid = 200004
             required string info_type      = 1; //peer, push, group, mobile, chatroom ...
             repeated Info   infos          = 2; //Info may include info_id,body_type,timestamp,sender,original sn,...
             optional int64  last_info_id   = 3; //server
             optional bytes  s_parameter    = 4;
             }
             */
            qihoo::protocol::messages::GetInfoResp &resp = (qihoo::protocol::messages::GetInfoResp &)(*pMessage);
            if (resp.has_info_type())
            {
                [dataDict setObject:[NSString stringWithUTF8String:resp.info_type().c_str()] forKey:@"info_type"];
            }
            if (resp.has_last_info_id())
            {
                [dataDict setObject:[NSNumber numberWithLong:resp.last_info_id()] forKey:@"last_info_id"];
            }
            if (resp.has_s_parameter())
            {
                [dataDict setObject:[IMUtil CharsToNSData:resp.s_parameter().c_str() withLength:resp.s_parameter().size()] forKey:@"s_parameter"];
            }

            //解析infos
            //CPLog(@"resp has msg:%d, detail:%s", resp.infos().size(), resp.DebugString().c_str());
            //CPLog(@"resp has msg:%d", resp.infos().size());
            NSMutableArray *infoList = [[NSMutableArray alloc] init];
            for(int i=0; i<resp.infos().size(); ++i)
            {
                const qihoo::protocol::messages::Info &info = resp.infos(i);
                NSMutableDictionary *oneMsg = [[NSMutableDictionary alloc] init];
                for (int k=0; k<info.property_pairs().size(); ++k)
                {
                    const qihoo::protocol::messages::Pair &pair = info.property_pairs(k);
                    /*
                     key options:
                     "info_id": msg id;
                     "chat_body": chat content;
                     "time_sent": send timestamp;
                     "msg_type": only weimi has this field(peer里的100表示字符串)
                     */
                    NSString *key = [IMUtil CharsToNSString:pair.key().c_str()];
                    NSData *value = [[NSData alloc] initWithBytes:pair.value().c_str() length:pair.value().size()];
                    if (pair.has_value())
                    {
                        if ([key isEqualToString:@"info_id"])
                        {
                            [oneMsg setObject:[NSNumber numberWithLong:[IMUtil getInt64FromNetData:value]] forKey:key];
                        }
                        else if ([key isEqualToString:@"msg_type"])
                        {
                            [oneMsg setObject:[NSNumber numberWithInt:[IMUtil getInt32FromNetData:value]] forKey:key];
                        }
                        else if ([key isEqualToString:@"time_sent"])
                        {
                            int64_t seconds = [IMUtil getInt64FromNetData:value]/1000;
                            NSDate *sendTime = [NSDate dateWithTimeIntervalSince1970:seconds];
                            [oneMsg setObject:sendTime forKey:key];
                        }
                        else if ([key isEqualToString:@"chat_body"])
                        {
                            [oneMsg setObject:value forKey:key];

                            //如果是chatroom，解析chat_body，获取msgID和maxID
                            if ([[dataDict objectForKey:@"info_type"]isEqualToString:@"chatroom"]) {

                                NSMutableDictionary *chatroomDict = [self.imUser.protoChatRoom parseNewMessage:[IMUtil unzipData:value]];
                                [oneMsg setObject:chatroomDict forKey:@"chatroomnewmsg"];
                            }
                        }
                        else if ([key isEqualToString:@"msg_valid"]) //0--表示消息已经删除了，1-－表示消息有效
                        {
                            [oneMsg setObject:[NSNumber numberWithInt:[IMUtil getInt32FromNetData:value]] forKey:key];
                        }
                        else
                        {
                            //CPLog(@"no usage key:%@", key);
                        }
                    }
                    else
                    {
//                        CPLog(@"pair no value >> %s : null", pair.key().c_str());
                    }
                }

                [infoList addObject:oneMsg];
            }
            [dataDict setObject:infoList forKey:@"infos"];
        }
            break;

        case MSG_ID_RESP_GET_MULTI_INFOS: //GetMultiInfosResp
        {
            /*
             message Pair {
             required bytes key   = 1;
             optional bytes value = 2;
             }

             message Info {
             repeated Pair property_pairs = 1;    //key=info_id,sender,content,time
             }

             message GetMultiInfosResp {    //msgid = 200100
             required string info_type      = 1; //peer, push, group, mobile, chatroom ...
             repeated Info   infos          = 2; //Info may include info_id,body_type,timestamp,sender,original sn,...
             optional int64  last_info_id   = 3; //server
             optional bytes  s_parameter    = 4;
             }
             */
            qihoo::protocol::messages::GetMultiInfosResp &resp = (qihoo::protocol::messages::GetMultiInfosResp &)(*pMessage);
            if (resp.has_info_type())
            {
                [dataDict setObject:[NSString stringWithUTF8String:resp.info_type().c_str()] forKey:@"info_type"];
            }
            if (resp.has_last_info_id())
            {
                [dataDict setObject:[NSNumber numberWithLong:resp.last_info_id()] forKey:@"last_info_id"];
            }
            if (resp.has_s_parameter())
            {
                [dataDict setObject:[IMUtil CharsToNSData:resp.s_parameter().c_str() withLength:resp.s_parameter().size()] forKey:@"s_parameter"];
            }

            //解析infos
            //CPLog(@"resp has msg:%d, detail:%s", resp.infos().size(), resp.DebugString().c_str());
            //CPLog(@"resp has msg:%d", resp.infos().size());
            NSMutableArray *infoList = [[NSMutableArray alloc] init];
            for(int i=0; i<resp.infos().size(); ++i)
            {
                const qihoo::protocol::messages::Info &info = resp.infos(i);
                NSMutableDictionary *oneMsg = [[NSMutableDictionary alloc] init];
                for (int k=0; k<info.property_pairs().size(); ++k)
                {
                    const qihoo::protocol::messages::Pair &pair = info.property_pairs(k);
                    /*
                     key options:
                     "info_id": msg id;
                     "chat_body": chat content;
                     "time_sent": send timestamp;
                     "msg_type": only weimi has this field(peer里的100表示字符串)
                     */
                    NSString *key = [IMUtil CharsToNSString:pair.key().c_str()];
                    NSData *value = [[NSData alloc] initWithBytes:pair.value().c_str() length:pair.value().size()];
                    if (pair.has_value())
                    {
                        if ([key isEqualToString:@"info_id"])
                        {
                            [oneMsg setObject:[NSNumber numberWithLong:[IMUtil getInt64FromNetData:value]] forKey:key];
                        }
                        else if ([key isEqualToString:@"msg_type"])
                        {
                            [oneMsg setObject:[NSNumber numberWithInt:[IMUtil getInt32FromNetData:value]] forKey:key];
                        }
                        else if ([key isEqualToString:@"time_sent"])
                        {
                            int64_t seconds = [IMUtil getInt64FromNetData:value]/1000;
                            NSDate *sendTime = [NSDate dateWithTimeIntervalSince1970:seconds];
                            [oneMsg setObject:sendTime forKey:key];
                        }
                        else if ([key isEqualToString:@"chat_body"])
                        {
                            [oneMsg setObject:value forKey:key];

                            //如果是chatroom，解析chat_body，获取msgID和maxID
                            if ([[dataDict objectForKey:@"info_type"]isEqualToString:@"chatroom"]) {

                                NSMutableDictionary *chatroomDict = [self.imUser.protoChatRoom parseNewMessage:[IMUtil unzipData:value]];
                                [oneMsg setObject:chatroomDict forKey:@"chatroomnewmsg"];
                            }
                        }
                        else if ([key isEqualToString:@"msg_valid"]) //0--表示消息已经删除了，1-－表示消息有效
                        {
                            [oneMsg setObject:[NSNumber numberWithInt:[IMUtil getInt32FromNetData:value]] forKey:key];
                        }
                        else
                        {
                            //CPLog(@"no usage key:%@", key);
                        }
                    }
                    else
                    {
//                        CPLog(@"pair no value >> %s : null", pair.key().c_str());
                    }
                }

                [infoList addObject:oneMsg];
            }
            [dataDict setObject:infoList forKey:@"infos"];
        }
            break;

        case MSG_ID_RESP_EX1_QUERY_USER_STATUS:
        {
            /*
             message RespEQ1User {
             required string userid      = 1;
             required string user_type   = 2;
             required int32  status      = 3;           //0:not registry;  1:registry, offline, not reachable; 2:registry, offline, reachable; 3:registry, online, reachable
             optional string jid         = 4;
             optional uint32 app_id      = 5;
             optional string platform    = 6;           // web, pc, mobile
             optional string mobile_type = 7;       //android, ios
             optional uint32 client_ver  = 8;
             }

             message Ex1QueryUserStatusResp {  //msgid = 200012
             repeated RespEQ1User user_list = 1;
             }
             */
            qihoo::protocol::messages::Ex1QueryUserStatusResp &resp = (qihoo::protocol::messages::Ex1QueryUserStatusResp &)(*pMessage);
            NSMutableArray* users = [NSMutableArray arrayWithCapacity:100];
            for (int k=0 ; k < resp.user_list().size(); k++) {
                const qihoo::protocol::messages::RespEQ1User &user = resp.user_list(k);
                IMProtoUserInfo* userInfo = [[IMProtoUserInfo alloc]init];
                userInfo.userId = [NSString stringWithUTF8String:user.userid().c_str()];
                userInfo.userType = [NSString stringWithUTF8String:user.user_type().c_str()];
                userInfo.status = user.status();
                if (user.has_jid()) {
                    if (user.has_jid())
                    {
                        userInfo.jid = [NSString stringWithUTF8String:user.jid().c_str()];
                    }
                    if (user.has_app_id())
                    {
                        userInfo.appId = user.app_id();
                    }
                    if (user.has_platform())
                    {
                        userInfo.platform = [NSString stringWithUTF8String:user.platform().c_str()];
                    }
                    if (user.has_mobile_type())
                    {
                        userInfo.mobileType = [NSString stringWithUTF8String:user.mobile_type().c_str()];
                    }
                    if (user.has_client_ver())
                    {
                        userInfo.clientVersion = user.client_ver();
                    }

                    [users addObject:userInfo];
                }
            }
            [dataDict setObject:users forKey:@"users"];
        }
            break;

        case MSG_ID_RESP_CHAT:
        {
            /*
             message ChatResp {   //msgid = 200002 server -> peer1
             required uint32 result     = 1;  //success : 0, failed = 1
             optional uint32 body_id    = 2;
             }*/
            qihoo::protocol::messages::ChatResp &resp = (qihoo::protocol::messages::ChatResp &)(*pMessage);
            if (resp.has_body_id())
            {
                [dataDict setObject:[NSNumber numberWithInt:resp.body_id()] forKey:@"body_id"];
            }

            if (resp.has_result() && resp.result() == 0)
            {
                [dataDict setObject:[NSNumber numberWithInt:IM_Success] forKey:@"result"];
                [dataDict setObject:[NSNumber numberWithInt:IM_Success] forKey:@"code"];
            }
            else
            {
                [dataDict setObject:[NSNumber numberWithInt:resp.result()] forKey:@"result"];
                [dataDict setObject:[NSNumber numberWithInt:resp.result()] forKey:@"code"];
            }

        }
            break;

        case MSG_ID_RESP_SERVICE_CONTROL:
        {
            qihoo::protocol::messages::Service_Resp & resp = (qihoo::protocol::messages::Service_Resp &)(*pMessage);
            if (resp.has_service_id()) {
                [dataDict setObject:[NSNumber numberWithInt:resp.service_id()] forKey:@"service_id"];
            }

            if (resp.has_response()) {

                if (resp.service_id() == SERVICE_ID_CHATROOM)
                {
                    NSMutableDictionary *chatroomDict = [self.imUser.protoChatRoom parseMessage:[IMUtil CharsToNSData:resp.response().c_str() withLength:resp.response().size()]];
                    [dataDict setObject:chatroomDict forKey:@"chatroom"];
                }
                else if (resp.service_id() == SERVICE_ID_GROUP)
                {
                    // parse group down packet, parsed value is a NSMutableDictionary object includes group messages
                     [dataDict setObject:[self.imUser.protoGroupChat parseGroupDownPacket:[IMUtil CharsToNSData:resp.response().c_str() withLength:resp.response().size()]] forKey:@"group"];
                }
                else
                {
                    qihoo::protocol::vcproxy::VCProxyPacket  vcPacket;
                    vcPacket.ParseFromString(resp.response());

                    qihoo::protocol::vcproxy::VCProxyServerToUser serverTo = vcPacket.server_data();
                    qihoo::protocol::vcproxy::CreateChannelResponse respsonse = serverTo.create_channel_resp();

                    [dataDict setObject:[IMUtil CharsToNSString:respsonse.channel_id().c_str()] forKey:@"channel_id"];
                    [dataDict setObject:[IMUtil CharsToNSData:respsonse.channel_info().c_str() withLength:respsonse.channel_info().size()] forKey:@"channel_info"];
                }
            }
        }
            break;


        default:
            break;
    }

    return dataDict;
}

/**
 * 根据解析出来的字典构造getInfo请求
 * @param infoType: 消息盒子名称;
 * @param infoID: 其实消息id;
 * @param offset: 消息数量;
 * @param sn: 请求序号
 * @returns 打包后的数据
 */
-(NSData*) createGetInfoRequest:(NSString*)infoType StartID:(int64_t)infoID Offset:(int)offset Sn:(int64_t)sn
{
    /*
     message GetInfoReq {     //msgid = 100004
     required string info_type      = 1; //peer, push, group, mobile, chatroom ...
     required int64 get_info_id     = 2;
     optional int32 get_info_offset = 3;//1,2,3 or -1,-2,-3
     optional bytes s_parameter     = 4;
     }
     */
    //CPLog(@"info_type:%@, start msgid:%lld, offset:%d", infoType, infoID, offset);
    qihoo::protocol::messages::GetInfoReq *command = new qihoo::protocol::messages::GetInfoReq();
    command->set_info_type([IMUtil NSStringToChars:infoType]);
    command->set_get_info_id(infoID);
    command->set_get_info_offset(offset);

    std::string result = [self createMessageString:command Sn:sn RecvType:@"" RecvID:@"" UserReserve:0];
    return [NSData dataWithBytes:result.c_str() length:result.size()];
}


-(NSData*) createGetInfoRequest:(NSString*)infoType StartID:(int64_t)infoID Offset:(int)offset RoomID:(NSString*)roomid Sn:(int64_t)sn
{
    /*
     message GetInfoReq {     //msgid = 100004
     required string info_type      = 1; //peer, push, group, mobile, chatroom ...
     required int64 get_info_id     = 2;
     optional int32 get_info_offset = 3;//1,2,3 or -1,-2,-3
     optional bytes s_parameter     = 4;
     }
     */
    //CPLog(@"info_type:%@, start msgid:%lld, offset:%d", infoType, infoID, offset);
    qihoo::protocol::messages::GetInfoReq *command = new qihoo::protocol::messages::GetInfoReq();
    command->set_info_type([IMUtil NSStringToChars:infoType]);
    command->set_get_info_id(infoID);
    command->set_get_info_offset(offset);
    command->set_s_parameter([IMUtil NSStringToChars:roomid]);

    std::string result = [self createMessageString:command Sn:sn RecvType:@"" RecvID:@"" UserReserve:0];
    return [NSData dataWithBytes:result.c_str() length:result.size()];
}

/**
 * 拉取多条消息
 */
-(NSData*) createGetMultiInfosRequest:(NSString*)infoType InfoIds:(NSArray*) infoIds RoomID:(NSString*)roomid Sn:(int64_t)sn
{
    /**
     required string info_type      = 1; //peer, push, group, mobile, chatroom ...
     repeated int64 get_info_ids    = 2;
     optional bytes s_parameter     = 3;
     */
    qihoo::protocol::messages::GetMultiInfosReq *command = new qihoo::protocol::messages::GetMultiInfosReq;
    command->set_info_type([IMUtil NSStringToChars:infoType]);

    for (NSNumber* infoId in infoIds) {
        command->add_get_info_ids([infoId integerValue]);
    }

    if (roomid != nil)
    {
        command->set_s_parameter([IMUtil NSStringToChars:roomid]);
    }

    std::string result = [self createMessageString:command Sn:sn RecvType:@"" RecvID:@"" UserReserve:0];
    return [NSData dataWithBytes:result.c_str() length:result.size()];
}


/**
 * 为了得到控制命令的发送者，需要从body部分解码
 */
-(NSString*) parseSenderID:(NSData*)data
{
    std::string result;
    [IMUtil NSDataToStlString:data Return:&result];

    qihoo::protocol::messages::Message msgMessage;
    msgMessage.ParseFromString(result);
    //sender_jid_	string *	"13693692095#2040#mobile-44aa8f1fe66ae1b9a1055c3dc19ce109"	0x1656e1b0
    if (msgMessage.has_sender_jid())
    {
        NSString *senderUUID = [IMUtil CharsToNSString:msgMessage.sender_jid().c_str()];
        NSArray *items = [senderUUID componentsSeparatedByString:@"#"];
        if ([items count] > 0)
        {
            NSString *sender = (NSString*)items[0];
            return sender;
        }
        else
        {
            return nil;
        }
    }
    return nil;
}


/**
 * 为了得到channel的信息，需要从body部分解码
 */
-(NSDictionary*) parseChannelData:(NSData*)data
{
    std::string result;
    [IMUtil NSDataToStlString:data Return:&result];

    qihoo::protocol::messages::Message msgMessage;
    msgMessage.ParseFromString(result);

    NSMutableDictionary* dic = [NSMutableDictionary dictionary];

    if (msgMessage.has_sender()) {
        [dic setObject:[IMUtil CharsToNSString:msgMessage.sender().c_str()] forKey:@"sender"];
    }
    if (msgMessage.has_sender_jid()) {
        [dic setObject:[IMUtil CharsToNSString:msgMessage.sender_jid().c_str()] forKey:@"sender_jid"];
    }

    if (msgMessage.has_receiver()) {
        [dic setObject:[IMUtil CharsToNSString:msgMessage.receiver().c_str()] forKey:@"receiver"];
    }

    if (msgMessage.has_receiver_type()) {
        [dic setObject:[IMUtil CharsToNSString:msgMessage.receiver_type().c_str()] forKey:@"receiver_type"];
    }

    if (msgMessage.has_msgid()) {
        [dic setObject:[NSString stringWithFormat:@"%d",msgMessage.msgid()] forKey:@"msgid"];
    }
    if (msgMessage.has_client_data()) {
        [dic setObject:[NSString stringWithFormat:@"%lld",msgMessage.client_data()] forKey:@"client_data"];
    }

    if (msgMessage.has_sn()) {
        [dic setObject:[NSString stringWithFormat:@"%lld",msgMessage.sn()] forKey:@"sn"];
    }

    if (msgMessage.has_req()) {
        qihoo::protocol::messages::Request request = msgMessage.req();
        if (request.has_chat()) {
            qihoo::protocol::messages::ChatReq chatReq = request.chat();
            NSData* body = [IMUtil CharsToNSData:chatReq.body().c_str() withLength:chatReq.body().length()];
            [dic setObject:body forKey:@"channelData"];

            return dic;
        }

    }
    return nil;
}


-(NSData*) createChatRequest:(NSString*)receiver Data:(NSData*)data
{
    /*
     message ChatReq {    //msgid = 100002  peer1 -> server -> peer2
     required bytes  body         = 1;
     optional uint32 body_id      = 2; //from 1 begin

     more_flag: 0:end 1:continue
     optional uint32 more_flag    = 3; //have more packet

     body_type: 0:text 1:audio 2:pic_url, 3:audio_and_pic

     required uint32 body_type    = 4; //text,audio,pic_url,audio_and_pic,...
     optional bool   store        = 5; //yes:need store, no:dont store, default yes
     optional bytes  m_parameter  = 6;
     optional uint32 service_id   = 7;
     optional bytes  s_parameter  = 8;
     optional string  conv_id     = 9; // conversation id, patched in every message
     optional bool   is_new_conv  = 10[default = false]; // whether this is the first message of a new conversation
     }
     }
     */
    //CPLog(@"info_type:%@, start msgid:%lld, offset:%d", infoType, infoID, offset);
    qihoo::protocol::messages::ChatReq *command = new qihoo::protocol::messages::ChatReq();
    command->set_body_id(1);
    command->set_more_flag(0);
    command->set_body_type(201);
    command->set_store(true);
    std::string strData;
    [IMUtil NSDataToStlString:data Return:&strData];
    command->set_body(strData);

    std::string result = [self createMessageString:command Sn:0 RecvType:@"jid" RecvID:receiver UserReserve:0];

    NSData *tmp1 = [IMUtil CharsToNSData:result.c_str() withLength:result.size()];
    std::string unzipData;
    [IMUtil NSDataToStlString:tmp1 Return:&unzipData];
    return [NSData dataWithBytes:result.c_str() length:result.size()];
}


/**
 * 创建是否在线请求
 */
-(NSData*) createEx1QueryUserStatusRequest:(uint64_t)sn UserType:(NSString*)userType UserIds:(NSArray*) userIds
{
    qihoo::protocol::messages::Ex1QueryUserStatusReq * req = new qihoo::protocol::messages::Ex1QueryUserStatusReq();
    for (NSString* userId in userIds) {
        qihoo::protocol::messages::ReqEQ1User * user = req->add_user_list();
        user->set_user_type([IMUtil NSStringToChars:userType]);
        user->set_userid([IMUtil NSStringToChars:userId]);
        user->set_app_id(2030);
    }

    std::string result = [self createMessageString:req Sn:sn RecvType:@"phone" RecvID:@"" UserReserve:0];
    return [NSData dataWithBytes:result.c_str() length:result.size()];
}


/**
 * 创建单聊请求
 */
-(NSData*) createPeerChatRequest:(uint64_t)sn Receiver:(NSString*)receiver RecvType:(NSString*)recvType Body:(NSData*)body BodyType:(int)bodyType ExpireTime:(uint32_t)expireTime
{
    return [self createChatRequest:sn Receiver:receiver RecvType:recvType Body:body BodyType:bodyType ServiceID:SERVICE_ID_MSGROUTER MParameter:nil SParameter:nil ExpireTime:expireTime];
}

/**
 * 创建chatrequest二进制流
 * @param receiver: 接受者唯一标识;
 * @param recvType: 接收者类型（phonenumber， qid）
 * @param body: 聊天内容
 * @param bodytype: 聊天内容类型
 * @param serviceid: srm分配的serviceid, msgrouter : 10000000; group : 10000001; distribute : 10000002; circle : 10000003; rm : 10000004; apns : 10000005; chatroom : 10000006; vcp : 10000007
 * @param mparameter: 留给客户端保存数据的参数;
 * @param sparameter: 留给各个业务模块保存数据的参数;
 * @param convid: 会话id，随机值,客户端保证唯一;
 * @param isnew: 是否是会话的第一条消息；
 * @returns request 二进制流;
 */
- (NSData*) createChatRequest:(uint64_t)sn Receiver:(NSString*)receiver RecvType:(NSString*)recvType Body:(NSData*)body BodyType:(int)bodyType ServiceID:(int)serviceID MParameter:(NSData*)mParameter SParameter:(NSData*)sParameter ExpireTime:(UInt32)expireTime
{
    /*
     //crypt txt, key is session_key
     message ChatReq {    //msgid = 100002  peer1 -> server -> peer2
     required bytes  body         = 1;
     optional uint32 body_id      = 2; //from 1 begin
     optional uint32 more_flag    = 3; //have more packet more_flag: 0:end 1:continue
     required uint32 body_type    = 4; //text,audio,pic_url,audio_and_pic,...body_type: 0:text 1:audio 2:pic_url, 3:audio_and_pic
     optional bool   store        = 5; //yes:need store, no:dont store, default yes
     optional bytes  m_parameter  = 6;
     optional uint32 service_id   = 7;
     optional bytes  s_parameter  = 8;
     optional string  conv_id     = 9; // conversation id, patched in every message
     optional bool   is_new_conv  = 10[default = false]; // whether this is the first message of a new conversation
     optional ExtraInfo extra_info	   = 11; // server allocates session, client need not set the field

     //to distinguish the way users chat, the following values are valid:
     //1: symmetric public chat (both chat with public id)
     //2: message sent by operation assistant account
     //10: asymmetric chat (one is with public id, the other is anonymous)
     //11: transit from '10' to '1' (asymmetric ==> symmetric public chat)
     //21: anonymous chat (both users are anonymous, ids are covered)
     optional uint32 chat_type          = 12;
     }
     */
    qihoo::protocol::messages::ChatReq *command = new qihoo::protocol::messages::ChatReq();
    std::string strBody;
    [IMUtil convertNSData:body toStlString:&strBody];
    command->set_body(strBody);
    command->set_more_flag(0);
    command->set_body_type(bodyType);
    command->set_store(true);
    command->set_expire_time(expireTime);

    if (mParameter != nil)
    {
        std::string strMP;
        [IMUtil convertNSData:mParameter toStlString:&strMP];
        command->set_m_parameter(strMP);
    }

    if (sParameter != nil)
    {
        std::string strSP;
        [IMUtil convertNSData:sParameter toStlString:&strSP];
        command->set_s_parameter(strSP);
    }

    std::string result = [self createMessageString:command Sn:sn RecvType:recvType RecvID:receiver UserReserve:0];
    return [NSData dataWithBytes:result.c_str() length:result.size()];
}




/**
 * 创建service 请求
 */
-(NSData*) createServiceRequest:(NSData*)serviceData ServiceID:(int)serviceid
{
    return  [self createServiceRequest:serviceData ServiceID:serviceid SN:[IMUtil createSN]];
}
-(NSData*) createServiceRequest:(NSData*)serviceData ServiceID:(int)serviceid SN:(int64_t)sn
{
    std::string strData;
    [IMUtil NSDataToStlString:serviceData Return:&strData];

    ::qihoo::protocol::messages::Service_Req *serReq = new ::qihoo::protocol::messages::Service_Req();
    serReq->set_service_id(serviceid);
    serReq->set_request(strData);

    // [step 2] fill Message
    ::qihoo::protocol::messages::Message msgMessage;
    // [step 2.1] msgid
    msgMessage.set_msgid(MSG_ID_REQ_SERVICE_CONTROL);
    // [step 2.2]: sn
    msgMessage.set_sn(sn);
    // [step 2.3]: receiver_type
    msgMessage.set_receiver_type("null");
    //    msgMessage.set_sender([IMUtil NSStringToChars:caller]);

    // [step 2.4]: fill Request and ChatReq
    ::qihoo::protocol::messages::Request* pReq = msgMessage.mutable_req();
    pReq->set_allocated_service_req(serReq);

    // [step 3]: serialize Message as std::string
    std::string result = msgMessage.SerializeAsString();
    return [NSData dataWithBytes:result.c_str() length:result.size()];


}


-(NSData*) createServiceMsgRequest:(NSData *)data ServiceID:(int)serviceid
{
    /*
     message ChatReq {    //msgid = 100002  peer1 -> server -> peer2
     required bytes  body         = 1;
     optional uint32 body_id      = 2; //from 1 begin

     more_flag: 0:end 1:continue
     optional uint32 more_flag    = 3; //have more packet

     body_type: 0:text 1:audio 2:pic_url, 3:audio_and_pic

     required uint32 body_type    = 4; //text,audio,pic_url,audio_and_pic,...
     optional bool   store        = 5; //yes:need store, no:dont store, default yes
     optional bytes  m_parameter  = 6;
     optional uint32 service_id   = 7;
     optional bytes  s_parameter  = 8;
     optional string  conv_id     = 9; // conversation id, patched in every message
     optional bool   is_new_conv  = 10[default = false]; // whether this is the first message of a new conversation
     }
     }
     */
    //CPLog(@"info_type:%@, start msgid:%lld, offset:%d", infoType, infoID, offset);
    qihoo::protocol::messages::ChatReq *command = new qihoo::protocol::messages::ChatReq();
    command->set_body_id(1);
    command->set_more_flag(0);
    command->set_body_type(0);
    command->set_store(true);
    std::string strData;
    [IMUtil NSDataToStlString:data Return:&strData];
    command->set_body(strData);
    command->set_service_id(serviceid);
    
    std::string result = [self createMessageString:command Sn:0 RecvType:@"jid" RecvID:@"110" UserReserve:0];
    return [NSData dataWithBytes:result.c_str() length:result.size()];
}


@end
