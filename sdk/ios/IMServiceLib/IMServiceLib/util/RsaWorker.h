//
//  RsaEncryptor.h
//  IMServiceLib
//
//  Created by 龙军 on 16/9/5.
//  Copyright © 2016年 qihoo. All rights reserved.
//

#import <Foundation/Foundation.h>

typedef enum _key_file_type{
    CER_TYPE_PEM,
    CER_TYPE_DER,
    CER_TYPE_P12,
}KeyFileType;

@interface RsaWorker : NSObject
#pragma mark - Properties

#pragma mark - Methods
/**
 * 单例方法，单例方法生成的对象默认将使用public_key.der证书，如果证书不存在，则该对象无法正常进行rsa工作
 * 可以在任何时候重新设置实例对象的证书文件名
 */
+(RsaWorker*) sharedInstance;

/**
 * 设置实例对象使用的公钥证书及文件类型，如果证书文件不存在，证书加载不成功则继续使用之前的证书
 * @param file:证书文件名
 * @param type:证书文件类型
 * @return YES:证书加载成功
 */

-(BOOL) setPublicKeyFileName:(NSString*)file withType:(KeyFileType)type;

/**
 * 使用公钥验证服务器下发的签名是否合法
 * @param data:用于校验签名的数据
 * @param sign:已经进行过base64解码的服务器下发的使用私钥签名的数据
 * @return YES:
 */
-(BOOL) rawVerify:(NSData*) data with:(NSData*)sign;


@end
