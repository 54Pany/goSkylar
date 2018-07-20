package core

import (
	"log"
	"strconv"
	"goSkylar/lib"
	"goSkylar/core/masscan"
	"goSkylar/core/nmap"
	"github.com/toolkits/net"
	"strings"
)

func CoreScanEngine(ipRange string, rate string, taskTime string) error {
	//lib.LogSetting()
	selfIpList, err := net.IntranetIP()
	selfIp := ""
	if err != nil {
		log.Println("-------Machine IP获取失败--------")
	} else {
		selfIp = selfIpList[0]
	}
	masscanResultStruct, err := RunMasscan(ipRange, rate)
	for _, v := range masscanResultStruct {
		err := lib.RedisDriver.RPush("masscan_result", v.IP+"§§§§"+strconv.Itoa(v.Port)+"§§§§"+selfIp).Err()
		if err != nil {
			log.Println("-----masscan_result push to redis error----" + err.Error())
		}
	}
	return err
}

func CoreScanNmapEngine(masscanTask string) error {
	wList := strings.Split(masscanTask, "§§§§")
	//判断数量匹配
	if len(wList) == 3 {
		engineResult, _ := RunNmap(wList[0], wList[1])
		for _, v := range engineResult {
			log.Println("--------------")
			log.Println(v)
			err := lib.PushPortInfoToRedis(ScannerResultTransfer(v), "", wList[2])
			return err
		}
	}
	return nil
}

func RunMasscan(ip_range string, rate string) ([]masscan.MasscanResultStruct, error) {
	m := masscan.New()
	//
	//// masscan可执行文件路径,默认不需要设置
	//m.SetSystemPath("/usr/local/masscan/bin/masscan")

	// 扫描端口范围
	m.SetPorts("1-65535")

	//m.SetInclude("1.txt")

	// 扫描IP范围
	m.SetRanges(ip_range)

	// 扫描速率
	m.SetRate(rate)

	//m.SetFileName()

	// 隔离扫描名单
	m.SetExclude("exclude.txt")

	// 开始扫描
	err := m.Run()
	if err != nil {
		log.Println("masscan scanner failed:", err)
		return nil, err
	}

	// 解析扫描结果
	results, err := m.Parse()
	if err != nil {
		log.Println("Parse scanner result:", err)
		return nil, err
	}

	var masscanResultStruct []masscan.MasscanResultStruct

	for _, result := range results {
		for _, v := range result.Ports {
			var masscanResult masscan.MasscanResultStruct
			masscanResult.IP = result.Address.Addr
			port, _ := strconv.Atoi(v.Portid)
			masscanResult.Port = port
			masscanResultStruct = append(masscanResultStruct, masscanResult)
		}
	}
	return masscanResultStruct, err
}

func RunNmap(ip string, port string) ([]nmap.NmapResultStruct, error) {
	m := nmap.New()
	m.SetIP(ip)
	m.SetHostTimeOut("1800000ms")
	m.SetMaxRttTimeOut("10000ms")
	m.SetPorts(port)
	err := m.Run()
	if err != nil {
		log.Println("nmap scanner failed:", err)
		return nil, err
	}

	results, err := m.Parse()
	return results, err
}

func ScannerResultTransfer(resultStruct nmap.NmapResultStruct) string {
	return resultStruct.Ip + "§§§§" + strconv.Itoa(resultStruct.PortId) + "§§§§" + resultStruct.Protocol + "§§§§" + resultStruct.Service
}
