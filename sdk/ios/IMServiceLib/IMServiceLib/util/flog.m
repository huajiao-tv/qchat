//
//  flog.c
//  360safe
//
//  Created by Jiyun Liu on 12-2-22.
//  Copyright (c) 2012年 Qihoo360. All rights reserved.
//


#include "flog.h"
#include <stdio.h>
#include <string.h>
#include <stdarg.h>

#define _FILE_LOG_ENABLE_

void logFile(const char *log_file_path,const char *str)
{
#ifdef _FILE_LOG_ENABLE_
    FILE *f = fopen(log_file_path, "r+");
    if(!f)
    {
        f = fopen(log_file_path, "wb");
    }
    
    fseek(f, 0, SEEK_END);
    fwrite(str, strlen(str), 1, f);
    fwrite("\n", 1, 1, f);
    fclose(f);
#endif
}


void flog(const char *log_file_path,const char*format, ...)
{
#ifdef _FILE_LOG_ENABLE_
    va_list vl;
    va_start(vl, format);
    
    char buf[4096] = {0};
    vsprintf(buf, format, vl);
    
    va_end(vl);
    
    logFile(log_file_path,buf);
#endif
}


void debugLogToFile(NSString *logFilePath, NSString *format, ...)
{
    
//#ifdef DEBUG
    va_list args;
    va_start(args, format);
    NSString *string = [[NSString alloc] initWithFormat:format arguments:args] ;
    va_end(args);
    
    if (logFilePath && logFilePath.length && string && string.length) {
        logFile([logFilePath UTF8String], [string UTF8String]);
//        CPLog(@"%@", string);
    }else{
//         CPLog(@"debugLog format wrong");
    }
//#else
//
//#endif

}


void WriteLog(int ulErrorLevel, const char *func, int lineNumber, NSString *format, ...)
{
    va_list args;
    va_start(args, format);
    NSString *string = [[NSString alloc] initWithFormat:format arguments:args];
    va_end(args);
    
    NSString *strFormat = [NSString stringWithFormat:@"%@%s, %@%i, %@%@",@"Function: ",func,@"Line: ",lineNumber, @"Format: ",string];
    
    NSString * strModelName = @"WriteLogTest"; //模块名
    
    NSString *strErrorLevel = [[NSString alloc] init];
    switch (ulErrorLevel) {
        case ERR_LOG:
            strErrorLevel = @"Error";
            break;
        case WARN_LOG:
            strErrorLevel = @"Warning";
            break;
        case NOTICE_LOG:
            strErrorLevel = @"Notice";
            break;
        case DEBUG_LOG:
            strErrorLevel = @"Debug";
            break;
        default:
            break;
    }
//    CPLog(@"ModalName: %@, ErrorLevel: %@, %@.",strModelName, strErrorLevel, strFormat);
}


