# sim - simple im 
```text
    简单即时通讯，功能待一步步完善，做过多个长链类的项目，但是最后都越来越冗余繁杂加复杂。其实我们
需要的只是建立连接收发消息，在这个简单需求上面加支持消息多种编解码方式，这里暂时用的消息结构体按顺序
和长度编码成字节流，所以每个消息协议都需要实现encode和decode，这里我做了个根据定义的结构体，使用语
法树自动生成编解码的工具。当然这种编解码比pb灵活度小了，准备再强化下编解码功能，以及做个抽象工厂兼容
多种编解码。先简单实现通讯和编解码，todo: 网关层、路由层(一个区域一个路由对应N个gate)、业务服务层
（消息入队存储、上行和业务系统使用），支持负载后面多网关，路由按区域，并且每个冗余主从(有哨兵支撑)部署。
``` 

* 构建
```text
之前使用1、shell命令（GOOS=linux go build -o xxx）构建提交一体
2、jenkins自动触发构建
这边尝试下bazel构建，赶脚也不错
```
* 设计考虑时序图
```text
简单实现 
消息经过User -> Gate -> route -> IMService -> route -> Gate
已读未读状态User -> Gate -> route -> Gate -> User
```
```sequence
title im时序图

participant User
participant Gate
participant Route
participant IMService

User->Gate:Login, SendMsg, ReadMsg
Gate->User:RecvMsg, ReadMsg, RevReadMsg
Gate->Route:UploadMsg
Route->IMService:UploadMsg
```

* 压测看连接数
```text
netstat -n | awk '/^tcp/ {++State[$NF]} END {for(i in State) print i, State[i]}'
```