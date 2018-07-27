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