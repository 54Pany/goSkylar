package main

import (
	"goSkylar/lib"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/satori/go.uuid"
	"goworker"
)

var (
	OuterRedisDriver  *redis.Client
	waitgroup         sync.WaitGroup
	ordinaryScanRate  string
	whitelistScanRate string
)

func init() {
	var cfg = lib.NewConfigUtil("")

	ordinaryScanRate, _ = cfg.GetString("masscan_rate", "ordinary_scan_rate")
	whitelistScanRate, _ = cfg.GetString("masscan_rate", "whitelist_scan_rate")

	OuterRedisDriver = lib.RedisOuterDriver

	var dsnAddr string
	dsnAddr = lib.DsnOuterAddr

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
	ipRangeList, whiteIps, _ := lib.FindInitIpRanges()
	log.Println("例行IP段数量:" + strconv.Itoa(len(ipRangeList)))
	log.Println(OuterRedisDriver.Ping())

	whiteIpsIprange := []string{}
	for _, v := range whiteIps {
		whiteIpsIprange = append(whiteIpsIprange, lib.Iptransfer(v))
	}

	//aChan := make(chan int, 1)
	waitgroup.Add(4)
	ticker := time.NewTicker(time.Hour * 7)
	tickerWhite := time.NewTicker(time.Hour * 20)
	tickerUrgent := time.NewTicker(time.Minute * 1)
	tickerNmapUrgent := time.NewTicker(time.Minute * 1)

	log.Println("开始例行扫描任务")

	//例行扫描：非白名单IP，扫描rate：50000
	go func() {
		defer waitgroup.Done()
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

				for port := 1; port <= 65535; port++ {
					for _, ipRange := range ipRangeList {

						log.Println("例行扫描Adding：" + ipRange)
						err := goworker.Enqueue(&goworker.Job{
							Queue: "ScanMasscanTaskQueue",
							Payload: goworker.Payload{
								Class: "ScanMasscanTask",
								Args:  []interface{}{string(ipRange), ordinaryScanRate, taskid, strconv.Itoa(port)},
							},
						},
							true)
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
	go func() {
		defer waitgroup.Done()
		for {
			select {
			case <-tickerWhite.C:
				log.Println("ticked at: " + lib.DateToStr(time.Now().Unix()))
				u, err := uuid.NewV4()
				if err != nil {
					log.Println(err)
				}
				taskid := u.String()
				log.Println(taskid)
				log.Println("开始例行白名单扫描任务")

				for port := 1; port <= 65535; port++ {
					for _, ipRange := range whiteIpsIprange {

						log.Println("例行扫描Adding：" + ipRange)
						err := goworker.Enqueue(&goworker.Job{
							Queue: "ScanMasscanTaskQueue",
							Payload: goworker.Payload{
								Class: "ScanMasscanTask",
								Args:  []interface{}{string(ipRange), whitelistScanRate, taskid, strconv.Itoa(port)},
							},
						},
							true)
						if err != nil {
							log.Println("白名单扫描goworker Enqueue时报错,ip段：" + ipRange + "，端口：" + strconv.Itoa(port))
						}
					}
				}
			}
			log.Println("例行白名单扫描任务加入结束")
		}
	}()
	//添加临时扫描任务
	go func() {
		defer waitgroup.Done()
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
									Args:  []interface{}{string(ipRange), ordinaryScanRate, taskid, strconv.Itoa(port)},
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
		defer waitgroup.Done()
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
							true)
						if err != nil {
							log.Println("nmap 集群获取任务，goworker Enqueue时报错,具体信息：" + w)
						}
					}

				}
			}
		}
	}()

	waitgroup.Wait()

}
