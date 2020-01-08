//
//  IMProtoRegistry.m
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-26.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "IMProtoRegistry.h"
#import "IMUtil.h"
#import "registry.pb.h"

using namespace qihoo::protocol::registry;

@implementation IMProtoRegistry

/**
 * 创建下行注册时获取验证码的请求
 * @returns request的二进制流
 */
-(NSData*) createDownRegistryGetVerifyCodeRequest:(NSString*)password
{
    /*
     message Down_Request_Get_Verifi_Code{
     required int32 appid = 1;
     required string cliver = 2;
     required string pn = 3;//encrypt by default key;
     optional string pwd = 4; //encrypt by default key;
     }*/
    Down_Request_Get_Verifi_Code request;
    std::string tt = typeid(request).name();
    request.set_appid(self.userConfig.appid);
    request.set_cliver([IMUtil NSStringToChars:[IMUtil IntToNSString:self.userConfig.version]]);
    std::string encryptPhone;
    [IMUtil rc4Encode:self.userConfig.phone Key:self.userConfig.defaultKey Return:&encryptPhone];
    request.set_pn(encryptPhone);
    request.set_pwd([password UTF8String]);
    
//    CPLog(@"down registry get vc request:%s", request.DebugString().c_str());
    std::string result = request.SerializeAsString();
    return [IMUtil CharsToNSData:result.c_str() withLength:(int)(result.size())];
}

/**
 * 解析下行注册时获取验证码的应答
 * @returns true--表示获取成功，false--表示获取失败
 */
-(BOOL) parseDownRegistryGetVerifyCodeRespone:(NSData*)data
{
    /*
     message Down_Response_Get_Verifi_Code{
     required int32 errorcode = 1; //1--means success, 2--means failed
     }*/
    
    int len = [data length];
    char raw[len];
    [data getBytes:raw length:len];
    Down_Response_Get_Verifi_Code response;
    response.ParseFromArray(raw, len);
    
    if (response.has_errorcode() && response.errorcode() == 1)
        return true;
    else
        return false;
}

/**
 * 创建下行注册的请求
 * @returns request的二进制流
 */
-(NSData*) createDownRegistryRequest:(NSString*)verifyCode Token:(NSString*)devToken
{
    /*
     message Down_Request_Register{
     required int32 appid = 1;
     required string cliver = 2;
     required string pn = 3;//encryot by default key
     required string rvc = 4;//encrypt by  vc
     optional string app_uuid = 5; //encrypt by vc;
     repeated Pair info = 6;//encrypt by vc,only on value
     }*/
    Down_Request_Register request;
    
    request.set_appid(self.userConfig.appid);
    request.set_cliver([[IMUtil IntToNSString:self.userConfig.version] UTF8String]);
    std::string encryptPhone;
    [IMUtil rc4Encode:self.userConfig.phone Key:self.userConfig.defaultKey Return:&encryptPhone];
    request.set_pn(encryptPhone);
    
    //按照约定，需要在验证码前加上6个字符长的随机串
    std::string encryptVC;
    NSString *random = [IMUtil createRandomString:6];
    [IMUtil rc4Encode:[random stringByAppendingString:verifyCode] Key:verifyCode Return:&encryptVC];
    request.set_rvc(encryptVC);
    
    std::string encryptDevToken;
    [IMUtil rc4Encode:devToken Key:verifyCode Return:&encryptDevToken];
    request.set_app_uuid(encryptDevToken);
    
//    CPLog(@"down registry request:%s", request.DebugString().c_str());
    std::string result = request.SerializeAsString();
    return [IMUtil CharsToNSData:result.c_str() withLength:result.size()];
    
}

/**
 * 解析下行注册的应答
 * @returns 参数都都放到字典里了
 * dict keys:
 * errorcode:
 * jid:
 * password
 */
-(NSMutableDictionary*) parseDownRegistryResponse:(NSData*)data verifyCode:(NSString*)code
{
    /*
     message Down_Response_Register{
     required int32 code = 1;
     optional string jid = 2;//encrypt by vc;
     optional string password = 3;//encrypt by vc
     }*/
    NSMutableDictionary *dataDict = [[NSMutableDictionary alloc] init];
    int len = [data length];
    char *raw = new char[len+1];
    [data getBytes:raw length:len];
    raw[len] = '\0';
    Down_Response_Register response;
    response.ParseFromArray(raw, len);
//    CPLog(@"down registry response:%s", response.DebugString().c_str());
    delete [] raw;
    
    if (response.has_errorcode() && response.errorcode() == 1)
    {
        [dataDict setObject:[NSNumber numberWithInt:IM_Success] forKey:@"code"];
        
        if (response.has_jid())
        {
            NSString *jid = [IMUtil rc4DecodeToString:&(response.jid()) Key:code];
//            CPLog(@"jid:%@", jid);
            [dataDict setObject:jid forKey:@"jid"];
            self.userConfig.jid = jid;
        }
        
        if (response.has_password())
        {
            NSString *password = [IMUtil rc4DecodeToString:&(response.password()) Key:code];
//            CPLog(@"password:%@", password);
            [dataDict setObject:password forKey:@"password"];
            self.userConfig.password = password;
        }
    }
    else
    {
        [dataDict setObject:[NSNumber numberWithInt:IM_InvalidParam] forKey:@"code"];
    }
    
    return dataDict;
}


@end
