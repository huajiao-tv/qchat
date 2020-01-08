//
//  IMRC4.m
//  IMServiceLib
//
//  Created by guanjianjun on 14-6-27.
//  Copyright (c) 2014年 qihoo. All rights reserved.
//

#include "IMRC4.h"


RC4::RC4(const unsigned char *key_data_ptr,int nLen)
{
    m_szKey = (char*)key_data_ptr;
    m_nKeyLen = nLen;
}

RC4::RC4(const string & szKey)
{
    m_szKey = szKey;
    m_nKeyLen = szKey.length();
}

void RC4::prepare_key(const unsigned char *key_data_ptr, int key_data_len)
{
    unsigned char index1;
    unsigned char index2;
    unsigned char* state;
    int counter;
    key.x = key.y = 0;
    state = &key.state[0];
    
    for(counter = 0; counter < 256; counter++)
        state[counter] = (unsigned char)counter;
    
    key.x = 0;
    key.y = 0;
    index1 = 0;
    index2 = 0;
    
    for(counter = 0; counter < 256; counter++)
    {
        index2 = (key_data_ptr[index1] + state[counter] + index2) % 256;
        swap_byte(&state[counter], &state[index2]);
        index1 = (index1 + 1) % key_data_len;
    }
}

void RC4::swap_byte(unsigned char *a, unsigned char *b)
{
    unsigned char x;
    x=*a;*a=*b;*b=x;
}

void RC4::rc4_encode(string & szOrig)
{
    int buffer_len = szOrig.length();
    if(buffer_len <= 0)
    {
        return;
    }
    
    unsigned char * buffer_ptr = new unsigned char[buffer_len + 1];
    memset(buffer_ptr, 0, (buffer_len + 1) * sizeof(unsigned char));
    //szOrig.copy((char*)buffer_ptr, buffer_len, 0);
    memcpy(buffer_ptr, szOrig.c_str(), buffer_len * sizeof(unsigned char));
    rc4_encode(buffer_ptr, buffer_len);
    
    szOrig.assign((char*)buffer_ptr, buffer_len);

    delete [] buffer_ptr;
    buffer_ptr = NULL;
}

void RC4::rc4_encode(unsigned char *buffer_ptr, int buffer_len)
{
    prepare_key((unsigned char *)m_szKey.c_str(), m_nKeyLen);
    
    unsigned char x;
    unsigned char y;
    unsigned char* state;
    unsigned char xorIndex;
    int counter;
    
    x = key.x;
    y = key.y;
    state = &key.state[0];
    
    for(counter = 0; counter < buffer_len; counter++)
    {
        x = (x + 1) % 256;
        y = (state[x] + y) % 256;
        swap_byte(&state[x], &state[y]);
        xorIndex = (state[x] + state[y]) % 256;
        buffer_ptr[counter] ^= state[xorIndex];
    }
    
    key.x = x;
    key.y = y;
}

