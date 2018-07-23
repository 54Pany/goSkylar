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
	"goSkylar/lib"
	"fmt"
)

var (
	version = "1.0.12"
)

func MasscanTask(queue string, args ...interface{}) error {
	log.Println("调用队列Masscan:" + queue)

	if len(args) != 4 {
		log.Println("----ScanMasscanTask 参数个数错误-----")
		log.Println(args)
		return nil
	}

	ipRange := args[0].(string)
	rate := args[1].(string)
	port := args[3].(string)

	selfIpList, err := net.IntranetIP()
	selfIp := ""
	if err != nil {
		log.Println("-------Machine IP获取失败--------")
	} else {
		selfIp = selfIpList[0]
	}

	results, err := core.RunMasscan(ipRange, rate, port)

	for _, v := range results {
		err := lib.RedisDriver.RPush("masscan_result", fmt.Sprintf("%s|%s|%s", v.IP, v.Port, selfIp)).Err()
		if err != nil {
			log.Println("-----masscan_result push to redis error----" + err.Error())
		}
	}

	log.Println("From " + queue + " " + args[2].(string))
	return err
}

func NmapTask(queue string, args ...interface{}) error {

	log.Println("调用队列Nmap:" + queue)

	if len(args) < 1 {
		return errors.New("nmap消费队列arg错误")
	}

	taskInfo := args[0].(string)
	wList := strings.Split(taskInfo, "|")

	//判断数量匹配
	if len(wList) >= 2 {
		machineIp := ""
		if len(wList) == 3 {
			machineIp = wList[2]
		}

		results, _ := core.RunNmap(wList[0], wList[1])
		for _, v := range results {
			log.Println(v)
			err := lib.PushPortInfoToRedis(core.ScannerResultTransfer(v), "", machineIp)
			return err
		}
	}

	return nil
}

func main() {

	signals := make(chan string)

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
			time.Sleep(1 * time.Minute)
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

}
