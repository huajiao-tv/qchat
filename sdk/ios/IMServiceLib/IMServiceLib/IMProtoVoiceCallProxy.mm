//
//  IMProtoVoiceCallProxy.m
//  IMServiceLib
//
//  Created by 360 on 12/10/14.
//  Copyright (c) 2014 qihoo. All rights reserved.
//

#import "IMProtoVoiceCallProxy.h"
#import "IMUtil.h"
#import "address_book.pb.h"
#import "vcproxy.pb.h"

using namespace qihoo::protocol;
using namespace qihoo::protocol::vcproxy;

@implementation IMProtoVoiceCallProxy

/**
 * 创建频道申请的包
 *
 * @param caller
 *            主叫
 * @param callee
 *            被叫
 * */
- (NSData*)createChannelRequestWithCaller:(NSString*)caller callee:(NSString*)callee sn:(NSInteger)sn
{
    // [step 1] create vcProxyPacket
    /*
     required uint32               payload_type    = 1;
     optional VCProxyUserToServer  user_data       = 2;
     optional VCProxyServerToUser  server_data     = 3;
     optional VCProxyNotify        notify_data     = 4;           //vcp message notify
     */
    qihoo::protocol::vcproxy::VCProxyPacket vcProxyPacket;
    vcProxyPacket.set_payload_type(PAYLOAD_REQ_CREATE_CHANNEL);
    // [step 1.1] userToServer
    VCProxyUserToServer* userToServer = vcProxyPacket.mutable_user_data();
    // [step 1.1] fill createReq
    CreateChannelRequest* createReq = userToServer->mutable_create_channel_req();
    createReq->set_requester([IMUtil NSStringToChars:caller]);
    user_info* member1 = createReq->add_member_list();
    member1->set_user_id([IMUtil NSStringToChars:callee]);
    user_info* member2 = createReq->add_member_list();
    member2->set_user_id([IMUtil NSStringToChars:caller]);
    
    // [step 2] fill Message
    messages::Message msgMessage;
    // [step 2.1] msgid
    msgMessage.set_msgid(MSG_ID_REQ_SERVICE_CONTROL);
    // [step 2.2]: sn
    msgMessage.set_sn(sn);
    // [step 2.3]: receiver_type
    msgMessage.set_receiver_type("null");
    //    msgMessage.set_sender([IMUtil NSStringToChars:caller]);
    
    // [step 2.4]: fill Request and ChatReq
    messages::Request* pReq = msgMessage.mutable_req();
    messages::Service_Req* pServiceReq = pReq->mutable_service_req();
    pServiceReq->set_service_id(SERVICE_ID_VCP);
    pServiceReq->set_request(vcProxyPacket.SerializeAsString());
    
    // [step 3]: serialize Message as std::string
    std::string result = msgMessage.SerializeAsString();
    return [NSData dataWithBytes:result.c_str() length:result.size()];;
}


@end
