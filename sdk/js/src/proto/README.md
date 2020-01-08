## 重新生成文件

不要直接更改 ```bundle.js``` 文件

使用以下命令生成

```bash
pbjs -t static-module -w es6 -o bundle.js --keep-case chatroom.proto messages.proto
```
> -t static-module 生成静态模块类型

> -w es6 生成 ES6 风格供导入

> --keep-case 生成文件不修改字段为驼峰