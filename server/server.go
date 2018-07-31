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
	MessageNum       map[string]int
)

func init() {

	// 最大任务数量,防止任务堆积,一般设置masscan并发执行的任务数量总和
	MaxNum = 20

	// 常规扫描速率
	OrdinaryScanRate = "30000"

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
		RedisMaxActive:   0,                 // 最大的激活连接数，表示同时最多有N个连接
		RedisMaxIdle:     100,               //最大的空闲连接数，表示即使没有redis连接时依然可以保持N个空闲的连接，而不被清除，随时处于待命状态
		RedisIdleTimeout: 180 * time.Second, // 最大的空闲连接等待时间，超过此时间后，空闲连接将被关闭
	})

	MessageNum = map[string]int{}

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
			counter := 0
			startTime := time.Now().Unix()
			ipRangeList := data.FindIpRanges()

			// 一个IP段扫描2个端口

			//for port := 0; port <= 65535; port += 10 {
			//
			//	if port >= 65535 {
			//		endTime := time.Now().Unix()
			//		msg := fmt.Sprintf("统计扫描完成约耗时:%d s, 任务开始时间: %s, 任务结束时间:%s", endTime-startTime, lib.InterfaceToStr(startTime), lib.InterfaceToStr(endTime))
			//		log.Println(msg)
			//		lib.SendSMessage(msg)
			//	}
			//
			//	if port > 65535 {
			//		break
			//	}
			//
			//	portEnd := port + 10
			//	if portEnd > 65535 {
			//		portEnd = 65535
			//	}
			//
			//	portStart := port + 1
			//
			//	tmpSlice := []string{}
			//	for _, ipRange := range ipRangeList {
			//		if len(tmpSlice) == 10 {
			//			tmp := strings.Join(tmpSlice, " ")
			//			task <- tmp + "|" + strconv.Itoa(portStart) + "-" + strconv.Itoa(portEnd)
			//			tmpSlice = []string{}
			//		}
			//		tmpSlice = append(tmpSlice, ipRange)
			//	}
			//	if len(tmpSlice) > 0 {
			//		tmp := strings.Join(tmpSlice, " ")
			//		task <- tmp + "|" + strconv.Itoa(portStart) + "-" + strconv.Itoa(portEnd)
			//	}
			//}

			for _, ipRange := range ipRangeList {
				task <- ipRange
				counter++
				log.Println("Add Counter:", counter)
			}

			// 记录任务添加完成时间
			endTime := time.Now().Unix()
			msg := fmt.Sprintf("一轮扫描任务已经添加完成,约耗时:%d s, 开始时间: %s, 结束时间:%s", endTime-startTime, lib.InterfaceToStr(startTime), lib.InterfaceToStr(endTime))
			log.Println(msg)
			lib.SendSMessage(msg)
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
					//infoList := strings.Split(info, "|")

					//开始处理任务
					//ipRange := infoList[0]
					//port := infoList[1]

					port := "0-65535"
					err := goworker.Enqueue(&goworker.Job{
						Queue: "masscan",
						Payload: goworker.Payload{
							Class: "masscan",
							Args:  []interface{}{info, OrdinaryScanRate, port},
						},
					},
						false)
					if err != nil {
						log.Println("例行扫描goworker Enqueue时报错:", err)
					} else {
						log.Println("成功添加任务:", info, port)
					}
				}
			} else {
			//	log.Println("当前剩余任务数量:", n, ",队列最大任务数量", MaxNum)
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
				log.Println("save nmap:", string(taskInfo) )
				err = data.NmapResultToMongo(taskInfo)
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
				continue
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
					if currentTime-resultTime > 30 {
						// 超过五分钟，主机心跳探测失败
						log.Println("主机：", agentIp, "停止心跳，请核实")
						// 短信、邮件告警
						if _, ok := MessageNum[agentIp]; ok {
							// 存在
							if MessageNum[agentIp] <= 3 {
								// 发送告警短信，主机停止心跳
								go lib.SendAlarmMessage(agentIp)
								MessageNum[agentIp] += 1
							}
						} else {
							MessageNum[agentIp] = 1
							// 发送告警短信，主机停止心跳
							go lib.SendAlarmMessage(agentIp)
						}
					} else {
						if _, ok := MessageNum[agentIp]; ok {
							if MessageNum[agentIp] > 0 {
								// 发送重启短信，主机已经重连
								go lib.SendRebootMessage(agentIp)
							}
						}
						MessageNum[agentIp] = 0
						//log.Println("主机：", agentIp, "当前存活")
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
