# goSkylar

# 简介
分布式资产扫描引擎，帮助企业实时扫描外网ip端口开放和banner信息。扫描内核基于masscan+nmap，调度模块基于改造版goworker，支持Agent在线升级，对任何异常有报警功能。

# 架构设计
![avatar](https://i.loli.net/2018/08/07/5b696289102d2.jpeg)

# Setup

1.配置Agent和Server端的conf文件，包括redis地址和Agent版本号获取地址等。

2.执行sh build_linux.sh获取最新Agent文件，上传到Agent服务器上。

3.Agent端执行（masscan和nmap分别搭建在不同机器）

    ./agent -queues="masscan" -interval=5.0 -connections=100 -concurrency=1
    
    ./agent -queues="nmap" -interval=5.0 -connections=100 -concurrency=1
    
4.部署Server

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

上传Server二进制文件

./Server

5.本地可以通过Console查看当前机器个数，masscan和nmap任务数量。

# 致谢
感谢Dean2021大佬
