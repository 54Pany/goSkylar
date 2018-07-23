package main

import (
	"log"
	"strconv"
	"github.com/go-redis/redis"
	"git.jd.com/wangshuo30/goworker"
	"time"
	"net/url"
	"strings"
	"goSkylar/lib/redispool"
	"fmt"
)

var (
	OuterRedisDriver *redis.Client
)

func init() {

	// 初始化
	settings := goworker.WorkerSettings{
		URI:            "redis://root:a1x06awvaBpD@116.196.96.123:23177/5",
		Connections:    100,
		Queues:         []string{"masscan", "nmap"},
		UseNumber:      true,
		ExitOnComplete: false,
		Concurrency:    2,
		Namespace:      "goskylar:",
		Interval:       5.0,
	}

	goworker.SetSettings(settings)
}

func main() {

	ipRangeList := []string{"45.76.205.0/24"}
	for port := 0; port <= 65535; port++ {
		for _, ipRange := range ipRangeList {

			log.Println("例行扫描Adding：" + ipRange)
			err := goworker.Enqueue(&goworker.Job{
				Queue: "masscan",
				Payload: goworker.Payload{
					Class: "masscan",
					Args:  []interface{}{string(ipRange), "5000", "xxxtest", strconv.Itoa(port)},
				},
			},
				false)
			if err != nil {
				log.Println("例行扫描goworker Enqueue时报错,ip段：" + ipRange + "，端口：" + strconv.Itoa(port))
			}
		}
	}

	u, err := url.Parse("redis://root:a1x06awvaBpD@116.196.96.123:23177/5")
	if err != nil {
		panic(err)
	}

	redisAddr := u.Host
	redisPass, ok := u.User.Password()
	if !ok {
		redisPass = ""
	}
	redisDB := strings.Trim(u.Path, "/")
	RedisPool := redispool.NewRedisPool(redispool.Options{
		RedisAddr:        redisAddr,         //redis链接地址
		RedisPass:        redisPass,         //redis认证密码
		RedisDB:          redisDB,           //redis数据库
		RedisMaxActive:   500,               // 最大的激活连接数，表示同时最多有N个连接
		RedisMaxIdle:     100,               //最大的空闲连接数，表示即使没有redis连接时依然可以保持N个空闲的连接，而不被清除，随时处于待命状态
		RedisIdleTimeout: 180 * time.Second, // 最大的空闲连接等待时间，超过此时间后，空闲连接将被关闭
	})

	conn := RedisPool.Get()
	defer conn.Close()

	go func() {
		for {

			reply, err := conn.Do("LPOP", fmt.Sprintf("masscan_result"))
			if err != nil {

				fmt.Println(err)
				continue
			}

			if reply != nil {
				taskInfo := string(reply.([]byte))
				err := goworker.Enqueue(&goworker.Job{
					Queue: "nmap",
					Payload: goworker.Payload{
						Class: "nmap",
						Args:  []interface{}{taskInfo},
					},
				},
					false)

			fmt.Println(taskInfo)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	select {}

	//
	////定时获取masscan扫描结果，给nmap集群进行扫描
	//go func() {
	//	for {
	//		select {
	//		case <-tickerNmapUrgent.C:
	//
	//
	//
	//
	//			count, err := conn.Do("LLen","masscan_result")
	//			if err != nil {
	//				log.Println("redis LLen失败" + err.Error())
	//				continue
	//			}
	//			if count > 0 {
	//				log.Println("masscan_result 存在Data")
	//				for {
	//					nmapTask, err := lib.RedisOuterDriver.LPop("masscan_result").Result()
	//					if err != nil {
	//						log.Println("redis LPop失败" + err.Error())
	//						break
	//					}
	//
	//					if nmapTask == "" {
	//						break
	//					}
	//
	//					err = goworker.Enqueue(&goworker.Job{
	//						Queue: "nmap",
	//						Payload: goworker.Payload{
	//							Class: "nmap",
	//							Args:  []interface{}{nmapTask},
	//						},
	//					},
	//						false)
	//					if err != nil {
	//						log.Println("nmap 集群获取任务，goworker Enqueue时报错,具体信息：" + err.Error())
	//					}
	//				}
	//
	//			} else {
	//				log.Println("masscan_result 无最新信息")
	//			}
	//		}
	//	}
	//}()
}
