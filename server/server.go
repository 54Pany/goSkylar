package main

import (
	"strconv"
	"time"

	"net/url"
	"strings"
	"github.com/garyburd/redigo/redis"
	"goSkylar/lib/redispool"
	"goSkylar/server/lib"
	"log"
	"github.com/satori/go.uuid"
	"git.jd.com/wangshuo30/goworker"
	"fmt"

	"goSkylar/server/data"
	"goSkylar/server/conf"
)

var (
	RedisPool          *redis.Pool
	MaxNum             int
	ORDINARY_SCAN_RATE string
)

func init() {
	MaxNum = 50
	ORDINARY_SCAN_RATE = "50000"

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
	RedisPool = redispool.NewRedisPool(redispool.Options{
		RedisAddr:        redisAddr,         //redis链接地址
		RedisPass:        redisPass,         //redis认证密码
		RedisDB:          redisDB,           //redis数据库
		RedisMaxActive:   500,               // 最大的激活连接数，表示同时最多有N个连接
		RedisMaxIdle:     100,               //最大的空闲连接数，表示即使没有redis连接时依然可以保持N个空闲的连接，而不被清除，随时处于待命状态
		RedisIdleTimeout: 180 * time.Second, // 最大的空闲连接等待时间，超过此时间后，空闲连接将被关闭
	})

}

func main() {

	//lib.LogSetting()

	// 初始化
	settings := goworker.WorkerSettings{
		URI:            conf.REDIS_URI,
		Connections:    100,
		Queues:         []string{"masscan","nmap"},
		UseNumber:      true,
		ExitOnComplete: false,
		Namespace:      "goskylar:",
	}

	goworker.SetSettings(settings)

	task := make(chan string)

	go func() {
		for {
			ipRangeList := data.FindIpRanges()
			for port := 80; port <= 81; port++ {
				for _, ipRange := range ipRangeList {
					task <- ipRange + "|" + strconv.Itoa(port)
				}
			}
		}
	}()

	connScan := RedisPool.Get()
	defer connScan.Close()

	go func() {
		for {
			//n 剩余任务
			reply, err := connScan.Do("LLEN", "goskylar:queue:masscan")
			if err != nil {

				log.Println(err)
				continue
			}
			if reply == nil {
				log.Println("LLEN Empty.")
				continue
			}

			n := int(reply.(int64))

			if n < MaxNum {
				taskNum := MaxNum - n
				for i := 1; i <= taskNum; i++ {
					info := <-task
					infoList := strings.Split(info, "|")

					//开始处理任务
					ipRange := infoList[0]
					port := infoList[1]
					taskid := lib.TimeToStr(time.Now().Unix())

					log.Println("例行扫描Adding：", ipRange, port)
					err := goworker.Enqueue(&goworker.Job{
						Queue: "masscan",
						Payload: goworker.Payload{
							Class: "masscan",
							Args:  []interface{}{ipRange, ORDINARY_SCAN_RATE, taskid, port},
						},
					},
						false)
					if err != nil {
						log.Println("例行扫描goworker Enqueue时报错:", err)
					} else {
						log.Println("成功添加任务:", ipRange, port)
					}
				}
			} else {
				log.Println("当前剩余任务数量:", n, ",队列最大任务数量", MaxNum)
				time.Sleep(time.Second)
			}
		}
	}()

	// 临时任务，每分钟监听
	tickerUrgent := time.NewTicker(time.Minute * 1)

	// 添加临时扫描任务
	go func() {
		for {
			select {
			case <-tickerUrgent.C:
				log.Println("ticked at: " + lib.DateToStr(time.Now().Unix()))
				urgentIPs := data.FindUrgentIP()
				if len(urgentIPs) > 0 {
					u, err := uuid.NewV4()
					if err != nil {
						log.Println(err)
					}
					taskid := u.String()
					// 临时任务id
					log.Println(taskid)
					for port := 1; port <= 65535; port++ {
						for _, ip := range urgentIPs {

							ipRange := lib.Iptransfer(ip)

							log.Println("临时扫描Adding：" + ipRange)
							err := goworker.Enqueue(&goworker.Job{
								Queue: "ScanMasscanTaskQueue",
								Payload: goworker.Payload{
									Class: "ScanMasscanTask",
									Args:  []interface{}{string(ipRange), "50000", taskid, strconv.Itoa(port)},
								},
							},
								true)
							if err != nil {
								log.Println("临时扫描goworker Enqueue时报错,ip段：" + ipRange + "，端口：" + strconv.Itoa(port))
							}
						}
					}
					data.UpdateUrgentScanStatus()
				} else {
					log.Println("无最新临时任务: " + lib.DateToStr(time.Now().Unix()))
				}
			}
		}
	}()

	// 定时获取masscan扫描结果，给nmap集群进行扫描
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

	connPortInfo := RedisPool.Get()
	defer connPortInfo.Close()

	// 从nmap扫描结果中获取信息入库
	go func() {
		for {

			reply, err := connPortInfo.Do("LPOP", fmt.Sprintf("portinfo"))
			if err != nil {
				log.Println(err)
				continue
			}

			if reply != nil {
				taskInfo := string(reply.([]byte))
				err = data.NmapResultToMongo(taskInfo)

				log.Println(taskInfo)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	select {}

}
