package main

import (
	"goSkylar/lib"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/satori/go.uuid"
	"git.jd.com/wangshuo30/goworker"
)

var (
	OuterRedisDriver  *redis.Client
	ordinaryScanRate  string
)

func init() {
	var cfg = lib.NewConfigUtil("")

	ordinaryScanRate, _ = cfg.GetString("masscan_rate", "ordinary_scan_rate")

	OuterRedisDriver = lib.RedisOuterDriver
	dsnAddr := lib.DsnOuterAddr

	// 初始化
	settings := goworker.WorkerSettings{
		URI:            dsnAddr,
		Connections:    100,
		Queues:         []string{"ScanMasscanTaskQueue", "ScanNmapTaskQueue"},
		UseNumber:      true,
		ExitOnComplete: false,
		Concurrency:    2,
		Namespace:      "goskylar:",
		Interval:       5.0,
	}

	goworker.SetSettings(settings)
}

func main() {
	lib.LogSetting()
	var whiteIpsIprange []string
	ipRangeList, whiteIps, _ := lib.FindInitIpRanges()
	log.Println("例行IP段数量:" + strconv.Itoa(len(ipRangeList)))
	log.Println(OuterRedisDriver.Ping())

	for _, v := range whiteIps {
		whiteIpsIprange = append(whiteIpsIprange, lib.Iptransfer(v))
	}

	//例行任务，每7小时一次
	ticker := time.NewTicker(time.Hour * 7)
	//获取Masscan扫描结果，每1分钟监听一次
	tickerNmapUrgent := time.NewTicker(time.Minute * 1)

	log.Println("开始例行扫描任务")

	//例行扫描：非白名单IP，扫描rate：50000
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("ticked at: " + lib.DateToStr(time.Now().Unix()))
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
							Queue: "ScanMasscanTaskQueue",
							Payload: goworker.Payload{
								Class: "ScanMasscanTask",
								Args:  []interface{}{string(ipRange), ordinaryScanRate, taskid, strconv.Itoa(port)},
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

	//例行扫描：白名单IP，扫描rate：20

	//添加临时扫描任务
	//go func() {
	//	for {
	//		select {
	//		case <-tickerUrgent.C:
	//			log.Println("ticked at: " + lib.DateToStr(time.Now().Unix()))
	//			urgentIPs := lib.FindUrgentIP()
	//			if len(urgentIPs) > 0 {
	//				u, err := uuid.NewV4()
	//				if err != nil {
	//					log.Println(err)
	//				}
	//				taskid := u.String()
	//				//临时任务id
	//				log.Println(taskid)
	//				for port := 1; port <= 65535; port++ {
	//					for _, ip := range urgentIPs {
	//
	//						ipRange := lib.Iptransfer(ip)
	//
	//						log.Println("临时扫描Adding：" + ipRange)
	//						err := goworker.Enqueue(&goworker.Job{
	//							Queue: "ScanMasscanTaskQueue",
	//							Payload: goworker.Payload{
	//								Class: "ScanMasscanTask",
	//								Args:  []interface{}{string(ipRange), ordinaryScanRate, taskid, strconv.Itoa(port)},
	//							},
	//						},
	//							true)
	//						if err != nil {
	//							log.Println("临时扫描goworker Enqueue时报错,ip段：" + ipRange + "，端口：" + strconv.Itoa(port))
	//						}
	//					}
	//				}
	//				lib.UpdateUrgentScanStatus()
	//			} else {
	//				log.Println("无最新临时任务: " + lib.DateToStr(time.Now().Unix()))
	//			}
	//		}
	//	}
	//}()

	//定时获取masscan扫描结果，给nmap集群进行扫描
	go func() {
		for {
			select {
			case <-tickerNmapUrgent.C:
				count, err := lib.RedisOuterDriver.LLen("masscan_result").Result()
				if err != nil {
					log.Println("redis LLen失败" + err.Error())
					continue
				}
				if count > 0 {
					for {
						nmapTask, err := lib.RedisOuterDriver.LPop("masscan_result").Result()
						if err != nil {
							log.Println("redis LPop失败" + err.Error())
							break
						}

						if nmapTask == "" {
							break
						}

						err = goworker.Enqueue(&goworker.Job{
							Queue: "ScanNmapTaskQueue",
							Payload: goworker.Payload{
								Class: "ScanNmapTask",
								Args:  []interface{}{nmapTask},
							},
						},
							false)
						if err != nil {
							log.Println("nmap 集群获取任务，goworker Enqueue时报错,具体信息：" + err.Error())
						}
					}

				}
			}
		}
	}()

	select {}

}
