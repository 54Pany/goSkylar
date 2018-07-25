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
	"git.jd.com/wangshuo30/goworker"
	"fmt"

	"goSkylar/server/data"
	"goSkylar/server/conf"
)

var (
	RedisPool        *redis.Pool
	MaxNum           int
	OrdinaryScanRate string
	AgentAliveTime   int64
)

func init() {

	// 最大任务数量,防止任务堆积,一般设置masscan并发执行的任务数量总和
	MaxNum = 1000
	// 常规扫描速率
	OrdinaryScanRate = "50000"

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

func main() {

	//lib.LogSetting()

	// 初始化
	settings := goworker.WorkerSettings{
		URI:            conf.REDIS_URI,
		Connections:    100,
		Queues:         []string{"masscan", "nmap"},
		UseNumber:      true,
		ExitOnComplete: false,
		Namespace:      "goskylar:",
	}

	goworker.SetSettings(settings)

	task := make(chan string)

	// 任务生成
	go func() {
		for {
			startTime := time.Now().Unix()
			ipRangeList := data.FindIpRanges()
			//ipRangeList := []string{"45.76.205.0/24"}

			// 一个IP段扫描2个端口
			for port := 0; port <= 65535; port++ {
				if port%2 == 0 {
					for _, ipRange := range ipRangeList {
						task <- ipRange + "|" + strconv.Itoa(port) + "," + strconv.Itoa(port+1)
					}
				}
				if port == 65535 {
					endTime := time.Now().Unix()
					log.Println("统计扫描完成约耗时:", endTime-startTime, "s, 任务开始时间:", lib.InterfaceToStr(startTime), ", 任务结束时间:", lib.InterfaceToStr(endTime))
				}
			}
		}
	}()

	go func() {
		for {
			connScan := RedisPool.Get()
			//n 剩余任务
			reply, err := connScan.Do("LLEN", "goskylar:queue:masscan")
			if err != nil {
				log.Println(err)
				connScan.Close()
				time.Sleep(time.Second)
				continue
			}
			if reply == nil {
				log.Println("LLEN Empty.")
				connScan.Close()
				time.Sleep(time.Second)
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

					log.Println("例行扫描Adding：", ipRange, port)
					err := goworker.Enqueue(&goworker.Job{
						Queue: "masscan",
						Payload: goworker.Payload{
							Class: "masscan",
							Args:  []interface{}{ipRange, OrdinaryScanRate, port},
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
			connScan.Close()
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
					for port := 1; port <= 65535; port++ {
						for _, ip := range urgentIPs {

							ipRange := lib.Iptransfer(ip)

							log.Println("临时扫描Adding：" + ipRange)
							err := goworker.Enqueue(&goworker.Job{
								Queue: "ScanMasscanTaskQueue",
								Payload: goworker.Payload{
									Class: "ScanMasscanTask",
									Args:  []interface{}{string(ipRange), "50000", strconv.Itoa(port)},
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
	go func() {
		for {
			conn := RedisPool.Get()

			reply, err := conn.Do("LPOP", fmt.Sprintf("masscan_result"))
			if err != nil {
				log.Println(err)
				conn.Close()
				time.Sleep(time.Second)
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
			conn.Close()
		}
	}()

	// 从nmap扫描结果中获取信息入库
	go func() {
		for {
			connPortInfo := RedisPool.Get()

			reply, err := connPortInfo.Do("LPOP", fmt.Sprintf("portinfo"))
			if err != nil {
				log.Println(err)
				connPortInfo.Close()
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
			connPortInfo.Close()
		}
	}()

	// agent存活探测
	go func() {
		for {
			currentTime := time.Now().Unix()
			connAgentAlive := RedisPool.Get()
			reply, err := connAgentAlive.Do("SMEMBERS", fmt.Sprintf("agent:ip"))
			if err != nil {
				log.Println("Server SMEMBERS Error", err.Error())
			}
			if reply != nil {
				agentList := reply.([]interface{})
				for _, v := range agentList {
					agentIp := string(v.([]byte))
					result, err := connAgentAlive.Do("HGET", "agent:ip:time", agentIp)
					if err != nil {
						log.Println("Server SMEMBERS Error", err.Error())
						continue
					}
					resultTime, err := strconv.ParseInt(string(result.([]byte)), 10, 64)
					if err != nil {
						log.Println("string to int64 Error")
						continue
					}
					if currentTime-resultTime > 150 {
						// 主机心跳探测失败
						log.Println("主机：", agentIp, "停止心跳，请核实")
						// 短信、邮件告警
					} else {
						log.Println("主机：", agentIp, "当前存活")
					}
				}
			}
			connAgentAlive.Close()

			// 每隔1分钟Server端探测一次
			time.Sleep(time.Minute * 1)
		}
	}()

	select {}

}
