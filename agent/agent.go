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
)

var (
	RedisPool *redis.Pool
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
		RedisAddr:        redisAddr,         //redis链接地址
		RedisPass:        redisPass,         //redis认证密码
		RedisDB:          redisDB,           //redis数据库
		RedisMaxActive:   500,               // 最大的激活连接数，表示同时最多有N个连接
		RedisMaxIdle:     100,               //最大的空闲连接数，表示即使没有redis连接时依然可以保持N个空闲的连接，而不被清除，随时处于待命状态
		RedisIdleTimeout: 180 * time.Second, // 最大的空闲连接等待时间，超过此时间后，空闲连接将被关闭
	})
}

func MasscanTask(queue string, args ...interface{}) error {
	log.Println("调用队列Masscan:" + queue)

	if len(args) != 3 {
		log.Println("----ScanMasscanTask 参数个数错误-----")
		log.Println(args)
		return nil
	}

	ipRange := args[0].(string)
	rate := args[1].(string)
	port := args[2].(string)

	selfIpList, err := net.IntranetIP()
	selfIp := ""
	if err != nil {
		log.Println("-------Machine IP获取失败--------")
	} else {
		selfIp = selfIpList[0]
	}

	results, err := core.RunMasscan(ipRange, rate, port)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		return nil
	}

	conn := RedisPool.Get()
	defer conn.Close()
	for _, v := range results {
		val := fmt.Sprintf("%s|%s|%s", v.IP, v.Port, selfIp)
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
		return errors.New("nmap消费队列arg错误")
	}

	taskInfo := args[0].(string)
	wList := strings.Split(taskInfo, "|")

	conn := RedisPool.Get()
	defer conn.Close()

	//判断数量匹配
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
	// agent存活信号，每分钟监听
	ticker := time.NewTicker(time.Minute * 1)

	// 初始化
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

	//机器存活心跳
	go func() {
		conn := RedisPool.Get()
		selfIpList, err := net.IntranetIP()
		selfIp := ""
		if err != nil {
			log.Println("-------Machine IP获取失败--------")
		} else {
			selfIp = selfIpList[0]
		}
		select {
		case <-ticker.C:
			currentTime := time.Now().Unix()
			_, err := conn.Do("SADD", "agent:ip", selfIp)
			if err != nil {
				log.Println("Agent SADD Error", err.Error())
			}
			_, err = conn.Do("HSET", "agent:ip:time", selfIp, currentTime)
			if err != nil {
				log.Println("Agent Error", err.Error())
			}
			log.Println("本机:【" + selfIp + "】已经向server发出心跳")
		}
	}()

}
