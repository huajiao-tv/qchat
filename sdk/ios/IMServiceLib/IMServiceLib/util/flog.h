//
//  flog.h
//  360safe
//
//  Created by Jiyun Liu on 12-2-22.
//  Copyright (c) 2012年 Qihoo360. All rights reserved.
//


#ifndef __FILE_LOG__
#define __FILE_LOG__

#define LOGFILEPATH [[NSString stringWithFormat:@"%@/Documents/log.txt",NSHomeDirectory()] UTF8String]
#ifdef __cplusplus
extern "C" {
#endif
    
#define ERR_LOG 1 /* 应用程序无法正常完成操作，比如网络断开，内存分配失败等 */
#define WARN_LOG 2 /* 进入一个异常分支，但并不会引起程序错误 */
#define NOTICE_LOG 3 /* 日常运行提示信息，比如登录、退出日志 */
#define DEBUG_LOG 4 /* 调试信息，打印比较频繁，打印内容较多的日志 */
    
#define LOGERR(format,...) WriteLog(ERR_LOG,__FUNCTION__,__LINE__,format,##__VA_ARGS__)
#define LOGWARN(format,...) WriteLog(WARN_LOG,__FUNCTION__,__LINE__,format,##__VA_ARGS__)
#define LOGNOTICE(format,...) WriteLog(NOTICE_LOG,__FUNCTION__,__LINE__,format,##__VA_ARGS__)
#define LOGDEBUG(format,...) WriteLog(DEBUG_LOG,__FUNCTION__,__LINE__,format,##__VA_ARGS__)
    


//不常用，易产生溢出
void flog(const char *log_file_path,const char*format, ...);

//打印日志，并将日志写入的指定文件中
void debugLogToFile(NSString *logFilePath, NSString *format, ...);

//按等级，打印日志
void WriteLog(int ulErrorLevel, const char *func, int lineNumber, NSString *format, ...);
    
#ifdef __cplusplus
}
#endif

#endif /* __FILE_LOG__ */
