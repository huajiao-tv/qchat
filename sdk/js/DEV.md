## 开发
```
yarn run dev
```
## 编译
```
yarn run build
```


## 主要流程图
![主要流程](https://p4.ssl.qhimg.com/t019ff588334107130a.png)
## 主要流程包组成
![主要流程组成](https://p1.ssl.qhimg.com/t011d2a677984ab2f5f.png)
> 除 HandShakePack 外, 业务与长连都是通过 [ length ( 4bytes ) ] + [ MessagesMessage ( protobuf ) ] 进行通讯

## 心跳包
心跳包为 [ length ( 4bytes ) ] 空包, 无业务时 5000ms 发一次

## 业务包处理
处理具体业务逻辑部分在 ```processMessagePack``` 方法中通过 msgid 寻找合适的处理方法
方法命名规范为
```javascript
`process${MESSAGE_NAME}Message`
```
MESSAGE_NAME 为 ```constants.js``` 中 信息 Id 常量部分 ( MESSAGE_ID ) 部分对应的属性名称。

例如 peer 消息解析函数名为 ```processGetInfoRespMessage```
