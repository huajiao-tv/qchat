//
//  NSString+MyAdditions.m
//  IMServiceLib
//
//  Created by longjun on 2014-08-20.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#import "MD5Digest.h"
#import <CommonCrypto/CommonDigest.h>

@implementation MD5Digest
+ (NSString *)md5: (NSString*) input
{
    const char* cStr = [input UTF8String];
    unsigned char result[CC_MD5_DIGEST_LENGTH];
    CC_MD5(cStr, strlen(cStr), result);
    
    static const char HexEncodeChars[] = { '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f' };
    char *resultData = malloc(CC_MD5_DIGEST_LENGTH * 2 + 1);
    
    for (uint index = 0; index < CC_MD5_DIGEST_LENGTH; index++) {
        resultData[index * 2] = HexEncodeChars[(result[index] >> 4)];
        resultData[index * 2 + 1] = HexEncodeChars[(result[index] % 0x10)];
    }
    resultData[CC_MD5_DIGEST_LENGTH * 2] = 0;
    
    NSString *resultString = [NSString stringWithCString:resultData encoding:NSASCIIStringEncoding];
    free(resultData);
    
    return resultString;
}

+ (NSString*)md5NSData:(NSData*) data
{
    unsigned char result[CC_MD5_DIGEST_LENGTH];
    CC_MD5( data.bytes, data.length, result ); // This is the md5 call
 
    static const char HexEncodeChars[] = { '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f' };
    char *resultData = malloc(CC_MD5_DIGEST_LENGTH * 2 + 1);
    
    for (uint index = 0; index < CC_MD5_DIGEST_LENGTH; index++) {
        resultData[index * 2] = HexEncodeChars[(result[index] >> 4)];
        resultData[index * 2 + 1] = HexEncodeChars[(result[index] % 0x10)];
    }
    resultData[CC_MD5_DIGEST_LENGTH * 2] = 0;
    
    NSString *resultString = [NSString stringWithCString:resultData encoding:NSASCIIStringEncoding];
    free(resultData);
    
    return resultString;
}

+ (NSData*)bytesMd5With:(NSData*)data
{
    unsigned char result[CC_MD5_DIGEST_LENGTH];
    CC_MD5(data.bytes, data.length, result ); // This is the md5 call
    return [NSData dataWithBytes:result length:CC_MD5_DIGEST_LENGTH];
    
}

+ (NSData*)bytesMd5:(NSString*)input
{
    const char* cStr = [input UTF8String];
    unsigned char result[CC_MD5_DIGEST_LENGTH];
    CC_MD5(cStr, strlen(cStr), result);
    return [NSData dataWithBytes:result length:CC_MD5_DIGEST_LENGTH];
}
@end