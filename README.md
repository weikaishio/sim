# sim - simple im 
```text
   简单即时通讯，功能待一步步完善，做过多个长链类的项目，最后都越来越冗余复杂。其实我们需要的只是建立连接收发消息，在这个简单
需求上面加支持消息多种编解码方式，服务分层，高可用，分布式微服务还是躲不过。
   这里暂时用的消息结构体按顺序和长度编码成字节流，所以每个消息协议都需要实现encode和decode，这里我做了个根据定义的结构体，
使用语法树自动生成编解码的工具。当然这种编解码比pb灵活度小了，准备再强化下编解码功能，以及做个抽象工厂兼容多种编解码。先简单实
现通讯和编解码，todo: 网关层、路由层(一个区域一个路由对应N个gate)、业务服务层（消息入队存储、上行和业务系统使用），支持负载后
面多网关，路由按区域，并且每个冗余主从(有哨兵支撑)部署。
``` 

## 构建
1. shell命令（GOOS=linux go build -o xxx）构建提交一体
2. jenkins自动触发构建
3. bazel构建，赶脚也不错


## 考虑实现逻辑
### login
1. 建立连接：client 直接访问lvs负载，lvs负载后面是gate
2、与gate连接成功，tsl握手交换密钥，auth登录

### rev msg
```text
新建立连接考虑通过http取历史消息，或者拆包组包通过gate直接取消息
```
1. 新消息过来(new msg notify)，im_service->route->gate->client
2. 发送取新消息请求，带上之前取到消息的最大消息id client->gate->route->im_service
3. im_service根据最大id，置之前消息已取，然后把大于该Id的消息发下去

### send msg


## 时序图
```text
简单实现 
消息经过User -> Gate -> route -> IMService -> route -> Gate
已读未读状态User -> Gate -> route -> Gate -> User

```
```sequence
title im时序图

participant Client
participant Gate
participant Route
participant IMService

Client->Gate:Login, SendMsg, GetMsg
Gate->Client:RecvMsg, ReadMsg, RevReadMsg, NewMsg
Gate->Route:transfer signal
Route->IMService:handle signal
```

* 压测看连接数
```text
netstat -n | awk '/^tcp/ {++State[$NF]} END {for(i in State) print i, State[i]}'
```