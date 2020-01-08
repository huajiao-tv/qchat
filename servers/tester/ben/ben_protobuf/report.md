# goprotobuf压测报告

## 定义

    F5服务器：24cpus，64G mem， 1T disk

## 测试环境
- 压测程序部署在一台F5服务器，i
- 压测程序采用1.6版本编译

## proto文件

### 简单proto文件

- 采用client.proto里的getmultiinfosreq协议
- 协议嵌套结构：`Message -> Request ->GetMultiInfosReq -> infotype,sparameters,ids`
    每个**ids**包含10个int64的id


### 复杂proto文件 getinforesp

- 采用client.proto里的getmultiinfosresp协议
- 协议嵌套结构：`Message -> Response ->GetMultiInfosResp -> infotype, infos, lastinfoid, sparameter`,
**infos** -> info -> pair -> key,value
- 每个infos包含10个info，每个info包含4个pair，包含数据的pair内含1k字节的[]byte


## 简单proto文件压测结果

- 初始条件：压测程序goroutinu数目24，每个goroutinue发起60000次请求
- 编码压测结果：处理速度 1,850,000 req/s，cpu占用1850%，提高goroutinue到48后cpu占用到1900%，处理速度下降到1,800,000 req/s
- 解码压测结果：处理速度 950,000 req/s，cpu占用 1650%，提高goroutinue到48后cpu占用到1750%，处理速度下降到930,000 req/s

---
- 初步分析：


## 复杂proto文件压测结果

- 初始条件：压测程序goroutinu数目24，每个goroutinue发起60000次请求
- 编码压测结果：cpu占用1100%，处理速度达到 45700 req/s，
    提高goroutinue到48后cpu占用到1300%，处理速度达到 44800 req/s
    提高goroutinue到96后cpu占用到1400%，处理速度达到 46800 req/s
    提高goroutinue到192后cpu占用到1500%，处理速度达到 48800 req/s
    提高goroutinue到288后cpu占用到1550%，处理速度达到 46300 req/s
    提高goroutinue到384后cpu占用到1650%，处理速度达到 40500 req/s
- 初步分析：
goroutinue在48有个异常下降，之后上升在192达到最高，之后下降。
- pprof分析：

---   

- 解码压测结果：cpu占用1400%，处理速度达到 68700 req/s，
    提高goroutinue到48后cpu占用到1680%，处理速度达到 65300 req/s
    提高goroutinue到96后cpu占用到1800%，处理速度达到 60100 req/s
    提高goroutinue到192后cpu占用到1880%，处理速度达到 47850 req/s
    提高goroutinue到288后cpu占用到1900%，处理速度达到 40700 req/s
    提高goroutinue到384后cpu占用到1900%，处理速度达到 41450 req/s
- 初步分析：
随着goroutinue增加处理能力持续下降
