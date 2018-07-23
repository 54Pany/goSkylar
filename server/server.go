package main

import (
	"strconv"
	"time"

	"net/url"
	"goSkylar/server/conf"
	"strings"
	"github.com/garyburd/redigo/redis"
	"goSkylar/lib/redispool"
	"goSkylar/server/lib"
	"log"
	"github.com/satori/go.uuid"
	"git.jd.com/wangshuo30/goworker"
	"fmt"

	"goSkylar/server/data"
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

func main() {
	var whiteIpsIprange []string

	// TODO BUG
	ipRangeList, whiteIps, _ := lib.FindInitIpRanges()
	log.Println("例行IP段数量:" + strconv.Itoa(len(ipRangeList)))
	OrdinaryScanRate := conf.ORDINARY_SCAN_RATE

	for _, v := range whiteIps {
		whiteIpsIprange = append(whiteIpsIprange, lib.Iptransfer(v))
	}

	//例行任务，每7小时一次
	tickerLib := time.NewTicker(time.Hour * 24)
	//例行任务，每7小时一次
	ticker := time.NewTicker(time.Hour * 7)
	//临时任务，每分钟监听
	tickerUrgent := time.NewTicker(time.Minute * 1)

	conn := RedisPool.Get()
	defer conn.Close()

	//例行扫描：非白名单IP，扫描rate：50000
	go func() {
		for {
			select {
			case <-ticker.C:
				u, err := uuid.NewV4()
				if err != nil {
					log.Println(err)
				}
				taskid := u.String()
				log.Println(taskid)
				log.Println("开始例行扫描任务")

				for port := 0; port <= 65535; port++ {
					for _, ipRange := range ipRangeList {

						log.Println("例行扫描Adding：" + ipRange)
						err := goworker.Enqueue(&goworker.Job{
							Queue: "masscan",
							Payload: goworker.Payload{
								Class: "masscan",
								Args:  []interface{}{string(ipRange), OrdinaryScanRate, taskid, strconv.Itoa(port)},
							},
						},
							false)
						if err != nil {
							log.Println("例行扫描goworker Enqueue时报错,ip段：" + ipRange + "，端口：" + strconv.Itoa(port))
						}
					}
				}
			}
			log.Println("例行扫描任务加入结束")
		}
	}()

	//添加临时扫描任务
	go func() {
		for {
			select {
			case <-tickerUrgent.C:
				log.Println("ticked at: " + lib.DateToStr(time.Now().Unix()))
				urgentIPs := lib.FindUrgentIP()
				if len(urgentIPs) > 0 {
					u, err := uuid.NewV4()
					if err != nil {
						log.Println(err)
					}
					taskid := u.String()
					//临时任务id
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
					lib.UpdateUrgentScanStatus()
				} else {
					log.Println("无最新临时任务: " + lib.DateToStr(time.Now().Unix()))
				}
			}
		}
	}()

	//定时获取masscan扫描结果，给nmap集群进行扫描

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

				log.Println(taskInfo)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-tickerLib.C:
				ipRangeList, whiteIps, _ = lib.FindInitIpRanges()
				log.Println("更新数据库ip段")
			}
		}
	}()


	connPortInfo := RedisPool.Get()
	defer connPortInfo.Close()

	go func() {
		for {

			reply, err := connPortInfo.Do("LPOP", fmt.Sprintf("portinfo"))
			if err != nil {

				log.Println(err)
				continue
			}

			if reply != nil {
				taskInfo := string(reply.([]byte))
				data.NmapResultToMongo(taskInfo)

				log.Println(taskInfo)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	select {}

}
