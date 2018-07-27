# goSkylar

## 部署

    设置文件打开数
    vim /etc/security/limits.conf

    * hard nofile 65536
    * soft nofile 65536

    reboot

    ip段隔离文件更新

    上传agent

    ./agent -queues="masscan" -interval=5.0 -connections=100 -concurrency=1
    ./agent -queues="nmap" -interval=5.0 -connections=100 -concurrency=1


### Server manager

Manager console HTTP : http://172.20.222.93/
Manager username : maixiang
Manager password : mx123


### Agent 机器

1. 192.168.
2.


### Reids 机器

192.168.0.41

### Mongo机器


### 公有云跳板机



