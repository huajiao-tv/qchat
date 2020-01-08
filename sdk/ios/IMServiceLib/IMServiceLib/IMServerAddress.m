//
//  IMServerAddress.m
//  IMServiceLib
//
//  Created by lupengyan on 2019/8/27.
//  Copyright © 2019 qihoo. All rights reserved.
//

#import "IMServerAddress.h"

@implementation IMServerAddress

+ (instancetype) serverAddressWithAddress:(NSString*)address ports:(NSArray<NSString*>*)ports
{
    IMServerAddress *server = [IMServerAddress new];
    server.address = address;
    server.ports = ports;
    
    return server;
}

@end
