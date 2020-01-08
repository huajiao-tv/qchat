//
//  NSString+MyAdditions.h
//  IMServiceLib
//
//  Created by longjun on 2014-08-20.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

@interface MD5Digest : NSObject
+ (NSString *)md5: (NSString*) input;

+ (NSString*)md5NSData:(NSData*) data;

+ (NSData*)bytesMd5With:(NSData*)data;
+ (NSData*)bytesMd5:(NSString*)input;
@end