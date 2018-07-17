package main

import (
	"goSkylar/lib"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/bipabo1l/goworker"
	"github.com/go-redis/redis"
	"github.com/satori/go.uuid"
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
		Queues:         []string{"ScanTaskQueue"},
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
	waitgroup.Add(1)
	ticker := time.NewTicker(time.Hour * 10)
	tickerWhite := time.NewTicker(time.Hour * 20)
	tickerUrgent := time.NewTicker(time.Minute * 1)

	// 首次运行

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
				property := "RoutineScan"
				log.Println("开始例行扫描任务")

				for _, ipRange := range ipRangeList {

					log.Println("例行扫描Adding：" + ipRange)
					goworker.Enqueue(&goworker.Job{
						Queue: "ScanTaskQueue",
						Payload: goworker.Payload{
							Class: "ScanTask",
							Args:  []interface{}{string(ipRange), ordinaryScanRate, taskid},
						},
					},
						true, taskid, property)
				}
			}
			log.Println("例行扫描任务加入结束")
		}
	}()
	//例行扫描：白名单IP，扫描rate：20
	go func() {
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
				property := "RoutineWhiteListScan"
				log.Println("开始例行白名单扫描任务")

				for _, ipRange := range whiteIpsIprange {

					log.Println("例行扫描Adding：" + ipRange)
					goworker.Enqueue(&goworker.Job{
						Queue: "ScanTaskQueue",
						Payload: goworker.Payload{
							Class: "ScanTask",
							Args:  []interface{}{string(ipRange), whitelistScanRate, taskid},
						},
					},
						true, taskid, property)
				}
			}
			log.Println("例行白名单扫描任务加入结束")
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
					property := "UrgentScan"
					for _, ip := range urgentIPs {

						ipRange := lib.Iptransfer(ip)

						log.Println("临时扫描Adding：" + ipRange)
						goworker.Enqueue(&goworker.Job{
							Queue: "ScanTaskQueue",
							Payload: goworker.Payload{
								Class: "ScanTask",
								Args:  []interface{}{string(ipRange), ordinaryScanRate, taskid},
							},
						},
							true, taskid, property)
					}
					lib.UpdateUrgentScanStatus()
				} else {
					log.Println("无最新临时任务: " + lib.DateToStr(time.Now().Unix()))
				}
			}
		}
	}()

	waitgroup.Wait()

}
