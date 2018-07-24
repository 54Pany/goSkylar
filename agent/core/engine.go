package core

import (
	"log"
	"goSkylar/lib/masscan"
	"goSkylar/lib/nmap"
)

func RunMasscan(ipRange string, rate string, port string) ([]masscan.MasscanResult, error) {

	var masscanResults []masscan.MasscanResult

	m := masscan.New()

	// 扫描端口范围
	m.SetPorts(port)

	// 扫描IP范围
	m.SetRanges(ipRange)

	// 扫描速率
	m.SetRate(rate)

	// 隔离扫描名单
	m.SetExclude("exclude.txt")

	// 设置等待时间
	m.SetWaitTime("5")

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

	for _, result := range results {
		for _, v := range result.Ports {
			masscanResults = append(masscanResults, masscan.MasscanResult{
				IP:   result.Address.Addr,
				Port: v.Portid,
			})
		}
	}

	return masscanResults, err
}

func RunNmap(ip string, port string) ([]nmap.NmapResult, error) {
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
