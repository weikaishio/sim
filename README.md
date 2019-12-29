# sim - simple im 
```text
    简单即时通讯，功能待一步步完善，做过多个长链类的项目，但是最后都越来越冗余繁杂加复杂。其实我们
需要的只是建立连接收发消息，在这个简单需求上面加支持消息多种编解码方式，这里暂时用的消息结构体按顺序
和长度编码成字节流，所以每个消息协议都需要实现encode和decode，这里我做了个根据定义的结构体，使用语
法树自动生成编解码的工具。当然这种编解码比pb灵活度小了，准备再强化下编解码功能，以及做个抽象工厂兼容
多种编解码。先简单实现通讯和编解码，todo: 网关层、路由层、业务层、消息入队多消费（存储、上行和业务系
统使用），支持负载后面多网关，路由冗余部署。
``` 

* 构建
```text
之前使用1、shell命令（GOOS=linux go build -o xxx）构建提交一体
2、jenkins自动触发构建
这边尝试下bazel构建，赶脚也不错
```
* 设计考虑时序图
```text
https://www.jianshu.com/p/8f8e7fd20054 https://blog.csdn.net/zhishengqianjun/article/details/74065232 时序图
https://www.jianshu.com/p/a9ff5a9cdb25 UML流程图
- 代表实线 ， 主动发送消息，比如 request请求
> 代表实心箭头 ， 同步消息，比如 AJAX 的同步请求
-- 代表虚线，表示返回消息，spring Controller return
>> 代表非实心箭头 ，异步消息，比如AJAX请求 
```
```sequence
title im时序图

participant User
participant Gate
participant Route
participant Server

User->Gate:Login, SendMsg
Gate->User:RecvMsg
Gate->Route:UploadMsg
Route->Server:UploadMsg
```

* 压测看连接数
```text
netstat -n | awk '/^tcp/ {++State[$NF]} END {for(i in State) print i, State[i]}'
```