//
//  RsaEncryptor.m
//  IMServiceLib
//
//  Created by 龙军 on 16/9/5.
//  Copyright © 2016年 qihoo. All rights reserved.
//

#import "RsaWorker.h"
#import <Security/Security.h>
#import "IMUtil.h"
#import "Base64.h"
#import "IMServiceLib.h"

@interface RsaWorker()
#pragma mark - Properties

#pragma mark - Private Methods
/**
 * 获取证书实际路径
 * @param file:证书文件名
 * @return 证书路径或者nil
 */
+(NSString *)publicKeyPath:(NSString*)file;

/**
 * 设置实例对象使用的公钥证书及文件类型，如果证书文件不存在，证书加载不成功则继续使用之前的证书
 * @param file:证书文件名
 * @param type:证书文件类型
 * @return 证书指针或者nil
 */
+(SecKeyRef) loadPublicKeyFromFile:(NSString*) file withType:(KeyFileType)type;

/**
 * 从der证书类型文件内容装载证书
 * @param data:der证书类型文件内容
 * @return 证书指针或者nil
 */
+(SecKeyRef) getPublicKeyRefrenceFromDerData: (NSData*)data;

/**
 * 从pem证书类型文件内容装载证书
 * @param data:pem证书类型文件内容
 * @return 证书指针或者nil
 */
+(SecKeyRef) getPublicKeyRefrenceFromPemData: (NSData*)data;

/**
 * verify that it is in fact a PKCS#1 key and strip the header
 * @param data:pem证书类型文件内容
 * @return 处理后的数据或者nil
 */
+(NSData *) stripPublicKeyHeader:(NSData *)keyData;

@end

#pragma mark - Implementation of RsaWorker
@implementation RsaWorker
{
    SecKeyRef _publicKey;
}

/**
 * 静态变量
 */
static RsaWorker *singleInstance = nil;

-(void) dealloc{
    @synchronized (self) {
        if (_publicKey != nil) {
            CFRelease(_publicKey);
            _publicKey = nil;
        }
    }
}

/**
 * 单例方法
 */
+(RsaWorker*) sharedInstance
{
    @synchronized(self)
    {
        if (!singleInstance)
        {
            singleInstance = [[[self class] alloc] init];
            [singleInstance setPublicKeyFileName:@"public_key.der" withType:CER_TYPE_DER];
        }
        return singleInstance;
    }
}

/**
 * 设置实例对象使用的公钥证书及文件类型，如果证书文件不存在，证书加载不成功则继续使用之前的证书
 * @param file:证书文件名
 * @param type:证书文件类型
 * @return YES:证书加载成功
 */
-(BOOL) setPublicKeyFileName:(NSString*)file withType:(KeyFileType)type
{
    // 获取证书文件实际路径
    NSString* cerPath = [RsaWorker publicKeyPath:file];
    if (cerPath == nil) {
        return NO;
    }
    
    SecKeyRef key = [RsaWorker loadPublicKeyFromFile:cerPath withType:type];
    if (key == nil) {
        return NO;
    }
    
    @synchronized (self) {
        if (_publicKey != nil) {
            CFRelease(_publicKey);
            _publicKey = nil;
        }
        _publicKey = key;
    }
    return YES;
}

/**
 * 使用公钥验证服务器下发的签名是否合法
 * @param data:用于校验签名的数据
 * @param sign:已经进行过base64解码的服务器下发的使用私钥签名的数据
 * @return YES:
 */
-(BOOL) rawVerify:(NSData*) data  with:(NSData*)sign
{
    SecKeyRef key = nil;
    
    @synchronized (self) {
        if (_publicKey == nil) {
            [[IMServiceLib sharedInstance] writeLog:@"SecKeyRawVerify:no public key\n"];
            return NO;
        }
        
        key = CFRetain(_publicKey);
    }
    
    OSStatus status = SecKeyRawVerify(key, kSecPaddingPKCS1, (const uint8_t *)[data bytes], data.length, (const uint8_t *)[sign bytes], sign.length);
    
    CFRelease(key);
    
    if (status != errSecSuccess)
    {
        [[IMServiceLib sharedInstance] writeLog:[NSString stringWithFormat:@"SecKeyRawVerify failed: %d\n", (int)status]];
    }
    return status == errSecSuccess;
}

#pragma mark - Private Methods


/**
 * 获取证书实际路径
 * @param file:证书文件名
 * @return 证书路径或者nil
 */
+(NSString *) publicKeyPath:(NSString*)file
{
    if (file == nil || [file isEqualToString:@""]) {
        return nil;
    }
    
    NSMutableArray * chunks = [[file componentsSeparatedByString:@"."] mutableCopy];
    NSString * extension = chunks[[chunks count] - 1];
    [chunks removeLastObject]; // remove the extension
    NSString * filename = [chunks componentsJoinedByString:@"."]; // reconstruct the filename with no extension
    NSString * keyPath = [[NSBundle mainBundle] pathForResource:filename ofType:extension];
    return keyPath;
}

/**
 * 从指定的证书文件装载public key
 * @param file:证书文件名
 * @param type:证书文件类型
 * @return 证书指针或者nil
 */
+(SecKeyRef) loadPublicKeyFromFile:(NSString*) file withType:(KeyFileType)type
{
    NSData *data = [[NSData alloc] initWithContentsOfFile:file];
    
    SecKeyRef key = nil;
    
    switch (type) {
        case CER_TYPE_DER:
        {
            return [RsaWorker getPublicKeyRefrenceFromDerData:data];
        }
            break;
            
        case CER_TYPE_PEM:
        {
            return [RsaWorker getPublicKeyRefrenceFromPemData:data];
        }
            break;
            
        default:
            [[IMServiceLib sharedInstance] writeLog:[NSString stringWithFormat:@"load public key failed, %d\n", (int)type]];
            break;
    }
    
    return nil;
}

/**
 * 从der证书类型文件内容装载证书
 * @param data:der证书类型文件内容
 * @return 证书指针或者nil
 */
+(SecKeyRef) getPublicKeyRefrenceFromDerData: (NSData*)data
{
    // 装载der文件
    SecKeyRef securityKey = nil;
    SecCertificateRef certificate = SecCertificateCreateWithData(kCFAllocatorDefault, (__bridge CFDataRef)data);
    SecPolicyRef policy = SecPolicyCreateBasicX509();
    SecTrustRef trust;
    OSStatus status = SecTrustCreateWithCertificates(certificate, policy, &trust);
    SecTrustResultType trustResult;
    if (status == noErr) {
        status = SecTrustEvaluate(trust, &trustResult);
    }
    if (status == noErr) {
        securityKey = SecTrustCopyPublicKey(trust);
    }
    
    CFRelease(certificate);
    CFRelease(policy);
    CFRelease(trust);
    
    return securityKey;
}

/**
 * 从pem证书类型文件内容装载证书
 * @param data:pem证书类型文件内容
 * @return 证书指针或者nil
 */
+(SecKeyRef) getPublicKeyRefrenceFromPemData: (NSData*)data
{
    // 直接装载pem格式的public key
    NSString* key = [IMUtil NSDataToNSString:data];
    NSString * s_key = @"";
    NSArray *a_key = [key componentsSeparatedByString:@"\n"];
    BOOL isKey = NO;
    for(NSString * line in a_key) {
        if([line isEqualToString:@"-----BEGIN PUBLIC KEY-----"]) {
            isKey = YES;
        } else if([line isEqualToString:@"-----END PUBLIC KEY-----"]) {
            isKey = NO;
        } else if(isKey) {
            s_key = [s_key stringByAppendingString:line];
        }
    }
    
    if(s_key.length == 0) {
        return nil;
    }
    
    // This will be base64 encoded, decode it.
    NSData *d_key = [Base64 dataWithBase64EncodedString:s_key];
    d_key = [RsaWorker stripPublicKeyHeader:d_key];
    if(d_key == nil) {
        return nil;
    }
    
    const NSString* tag = @"cer.sdk.ios.im.huajiao";
    NSData *d_tag = [NSData dataWithBytes:[tag UTF8String] length:[tag length]];
    
    // Delete any old lingering key with the same tag
    NSMutableDictionary *publicKey = [[NSMutableDictionary alloc] init];
    [publicKey setObject:(id)kSecClassKey forKey:(id)kSecClass];
    [publicKey setObject:(id)kSecAttrKeyTypeRSA forKey:(id)kSecAttrKeyType];
    [publicKey setObject:d_tag forKey:(id)kSecAttrApplicationTag];
    SecItemDelete((CFDictionaryRef)publicKey);
    
    CFTypeRef persistKey = nil;
    
    // Add persistent version of the key to system keychain
    [publicKey setObject:d_key forKey:(id)kSecValueData];
    [publicKey setObject:(id)kSecAttrKeyClassPublic forKey:(id)kSecAttrKeyClass];
    
    [publicKey setObject:[NSNumber numberWithBool:YES] forKey:(id)kSecReturnPersistentRef];
    
    OSStatus secStatus = SecItemAdd((CFDictionaryRef)publicKey, &persistKey);
    if(persistKey != nil) {
        CFRelease(persistKey);
    }
    
    if(secStatus != noErr && secStatus != errSecDuplicateItem) {
        return nil;
    }
    
    // Now fetch the SecKeyRef version of the key
    SecKeyRef keyRef = nil;
    
    [publicKey removeObjectForKey:(id)kSecValueData];
    [publicKey removeObjectForKey:(id)kSecReturnPersistentRef];
    [publicKey setObject:[NSNumber numberWithBool:YES] forKey:(id)kSecReturnRef];
    [publicKey setObject:(id)kSecAttrKeyTypeRSA forKey:(id)kSecAttrKeyType];
    secStatus = SecItemCopyMatching((CFDictionaryRef)publicKey, (CFTypeRef *)&keyRef);
    
    
    if(keyRef != nil) {
        return keyRef;
    }

    return nil;
}

/**
 * verify that it is in fact a PKCS#1 key and strip the header
 * @param data:pem证书类型文件内容
 * @return 处理后的数据或者nil
 */
+(NSData *) stripPublicKeyHeader:(NSData *)keyData
{
    if(keyData == nil || keyData.length == 0) {
        return nil;
    }
    
    unsigned int len = [keyData length];
    unsigned char * cKey = (unsigned char *)[keyData bytes];
    unsigned int idx = 0;
    
    if(cKey[idx++] != 0x30) {
        return nil;
    }
    
    if(cKey[idx] > 0x80) {
        idx += cKey[idx] - 0x80 + 1;
    }
    else {
        ++idx;
    }
    
    // PKCS #1 rsaEncryption szOID_RSA_RSA
    static unsigned char seqiod[] = {
        0x30, 0x0d, 0x06, 0x09, 0x2a, 0x86, 0x48, 0x86,
        0xf7, 0x0d, 0x01, 0x01, 0x01, 0x05, 0x00
    };
    if(memcmp(&cKey[idx], seqiod, 15)) {
        return nil;
    }
    idx += 15;
    
    if(cKey[idx++] != 0x03) {
        return nil;
    }
    if(cKey[idx] > 0x80) {
        idx += cKey[idx] - 0x80 + 1;
    }
    else
    {
        idx++;
    }
    
    if(cKey[idx++] != '\0') {
        return nil;
    }
    
    // Now make a new NSData from this buffer
    return [NSData dataWithBytes:&cKey[idx] length:len-idx];
}

@end
