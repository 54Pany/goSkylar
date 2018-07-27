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

    部署Masscan：
        114.67.230.216（公）  192.168.0.8（内）

        114.67.231.9（公）  192.168.0.9（内）

        114.67.231.22（公）  192.168.0.10（内）

    部署Nmap：
        114.67.230.108（公）  192.168.0.5（内）

        114.67.231.13（公）  192.168.0.11（内）

### Reids 机器

    116.196.96.123（公）  192.168.0.41（内）

### Mongo机器

    172.20.222.88:80
    sea ﻿port_scan_result

### 公有云跳板机

    116.196.121.126（公）  192.168.0.32（内)

### 升级机器

    116.196.96.123（公）  192.168.0.41（内）

