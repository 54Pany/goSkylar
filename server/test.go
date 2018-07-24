package main

import (
	"github.com/go-redis/redis"
	"strconv"
	"log"
	"git.jd.com/wangshuo30/goworker"
)

var (
	OuterRedisDriver *redis.Client
)

func init() {
}

func main() {

	ipRangeList := []string{"45.76.205.0/24"}
	for port := 0; port <= 100; port++ {
		for _, ipRange := range ipRangeList {

			log.Println("例行扫描Adding：" + ipRange + "  " + strconv.Itoa(port))
			err := goworker.Enqueue(&goworker.Job{
				Queue: "masscan",
				Payload: goworker.Payload{
					Class: "masscan",
					Args:  []interface{}{string(ipRange), "5000", "xxxtest", strconv.Itoa(port)},
				},
			}, false)
			if err != nil {
				log.Println("例行扫描goworker Enqueue时报错,ip段：" + ipRange + "，端口：" + strconv.Itoa(port))
			}
		}
	}

	select {}

}
