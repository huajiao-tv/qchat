//
//  IMServerAddress.h
//  IMServiceLib
//
//  Created by lupengyan on 2019/8/27.
//  Copyright © 2019 qihoo. All rights reserved.
//

#import <Foundation/Foundation.h>

NS_ASSUME_NONNULL_BEGIN

@interface IMServerAddress : NSObject

@property (atomic, strong) NSString *address;

@property (atomic, strong) NSArray<NSString*> *ports;

+ (instancetype) serverAddressWithAddress:(NSString*)address ports:(NSArray<NSString*>*)ports;

@end

NS_ASSUME_NONNULL_END
