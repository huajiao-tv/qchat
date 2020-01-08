//
//  IMRC4.h
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-27.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//
#include <string>
using namespace std;

typedef struct rc4_key
{
	unsigned char state[256];
	unsigned char x;
	unsigned char y;
}rc4_key;

class RC4
{
public:
	RC4(const unsigned char *key_data_ptr,int nLen);//类初始化时接受字串，初始化key
	RC4(const string & szKey);//类初始化时接受字串，初始化key
	void rc4_encode(string & szOrig);
	void rc4_encode(unsigned char *buffer_ptr, int buffer_len);//明文与暗文使用同一个函数转换
    
private:
	//const unsigned char * m_p_szKey; // key
	string m_szKey;
	int m_nKeyLen; // key 长度
	rc4_key key;//加密与解密用到的key，初始化时就需要赋值
    
	void prepare_key(const unsigned char *key_data_ptr, int key_data_len);//初始化key
	void swap_byte(unsigned char *a, unsigned char *b);//交换
    
};
