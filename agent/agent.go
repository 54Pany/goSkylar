package main

import (
	"log"
	"time"

	"git.jd.com/wangshuo30/goworker"
	"goSkylar/agent/conf"
	"errors"
	"github.com/toolkits/net"
	"strings"
	"goSkylar/agent/core"
	"fmt"
	"goSkylar/lib/redispool"
	"net/url"
	"github.com/garyburd/redigo/redis"
	"os"
)

var (
	RedisPool *redis.Pool
	localIP   string
)

func init() {
	u, err := url.Parse(conf.REDIS_URI)
	if err != nil {
		panic(err)
	}

	redisAddr := u.Host
	redisPass, ok := u.User.Password()
	if !ok {
		redisPass = ""
	}
	redisDB := strings.Trim(u.Path, "/")
	RedisPool = redispool.NewRedisPool(redispool.Options{
		RedisAddr:        redisAddr,        // redis链接地址
		RedisPass:        redisPass,        // redis认证密码
		RedisDB:          redisDB,          // redis数据库
		RedisMaxActive:   100,              // 最大的激活连接数，表示同时最多有N个连接
		RedisMaxIdle:     100,              // 最大的空闲连接数，表示即使没有redis连接时依然可以保持N个空闲的连接，而不被清除，随时处于待命状态
		RedisIdleTimeout: 20 * time.Second, // 最大的空闲连接等待时间，超过此时间后，空闲连接将被关闭
	})

	selfIpList, err := net.IntranetIP()
	if err != nil {
		log.Println("-------Machine IP获取失败--------")
	} else {
		localIP = selfIpList[0]
	}
}

func MasscanTask(queue string, args ...interface{}) error {
	log.Println("调用队列Masscan:" + queue)

	if len(args) != 3 {
		log.Println("----ScanMasscanTask 参数个数错误-----", len(args))
		return errors.New(fmt.Sprintf("masscan任务参数个数错误:%d, args:%s", len(args), args))
	}

	ipRange := args[0].(string)
	rate := args[1].(string)
	port := args[2].(string)

	results, err := core.RunMasscan(ipRange, rate, port)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		return nil
	}

	log.Println(args)
	conn := RedisPool.Get()
	defer conn.Close()

	// TODO 可以一次push进去
	for _, v := range results {
		val := fmt.Sprintf("%s|%s|%s", v.IP, v.Port, localIP)
		log.Println("Insert a scan result of masscan to redis:" + val)
		_, err := conn.Do("RPUSH", "masscan_result", val)
		if err != nil {
			log.Println("-----masscan_result push to redis error----" + err.Error())
		}
	}

	return err

}

func NmapTask(queue string, args ...interface{}) error {

	log.Println("调用队列Nmap:" + queue)

	if len(args) < 1 {
		log.Println("----NmapTask 参数个数错误-----", len(args))
		return errors.New(fmt.Sprintf("nmap任务参数个数错误:%d, args:%s", len(args), args))
	}

	taskInfo := args[0].(string)
	wList := strings.Split(taskInfo, "|")

	conn := RedisPool.Get()
	defer conn.Close()

	// 判断数量匹配
	if len(wList) >= 2 {
		machineIp := ""
		if len(wList) == 3 {
			machineIp = wList[2]
		}

		results, _ := core.RunNmap(wList[0], wList[1])
		for _, v := range results {
			val := fmt.Sprintf("%s|%d|%s|%s|%s", v.Ip, v.PortId, v.Protocol, v.Service, machineIp)
			log.Println("Insert a scan result of nmap to redis:" + val)
			_, err := conn.Do("RPUSH", "portinfo", val)
			if err != nil {
				log.Println("----- portinfo push to redis error----" + err.Error())
			}
		}
	}

	return nil
}

func main() {

	signals := make(chan string)

	// 初始化  TODO 连接数待优化
	settings := goworker.WorkerSettings{
		URI:            conf.REDIS_URI,
		UseNumber:      true,
		ExitOnComplete: false,
		Namespace:      "goskylar:",
	}

	goworker.SetSettings(settings)
	goworker.Register("masscan", MasscanTask)
	goworker.Register("nmap", NmapTask)

	// 检查升级
	go func() {
		for {
			VersionValidate(signals, conf.VERSION_URL, conf.DOWNLOAD_URL, conf.BACK_FILE_PATH)
			time.Sleep(10 * time.Second)
		}
	}()

	// 机器存活心跳
	go func() {
		for {
			conn := RedisPool.Get()
			currentTime := time.Now().Unix()
			_, err := conn.Do("SADD", "agent:ip", localIP)
			if err != nil {
				log.Println("Agent SADD Error", err.Error())
				continue
			}
			_, err = conn.Do("HSET", "agent:ip:time", localIP, currentTime)
			if err != nil {
				log.Println("Agent HSET Error", err.Error())
				continue
			}
			log.Println("本机:【" + localIP + "】已经向server发出心跳")
			time.Sleep(time.Second)
			conn.Close()
		}
	}()

	// 获取服务端下发指令,目前仅支持shutdown
	go func() {
		for {
			conn := RedisPool.Get()
			command, err := conn.Do("HGET", "agent:command", localIP)
			if err != nil {
				log.Println("Agent HKEYS Error", err.Error())
				continue
			}

			if command == nil {
				log.Println("无最新指令")
				conn.Close()
				time.Sleep(time.Second)
				continue
			}

			if string(command.([]byte)) == "shutdown" {
				_, err := conn.Do("HDEL", "agent:command", localIP)
				if err != nil {
					log.Println("HDEL", err.Error())
				}
				os.Exit(0)
			}

			time.Sleep(time.Second)
			conn.Close()
		}
	}()

	// loop处理任务
	go func() {
		for {
			if err := goworker.Work(); err != nil {
				log.Println("Error:", err)
			}
			time.Sleep(time.Second * 15)
		}
	}()

	// 根据信号处理任务
	for {
		select {
		case signal := <-signals:
			if signal == "new" {
				RestartProcess()
			}
		case <-time.After(time.Second * 10):
			continue
		}
	}

}
