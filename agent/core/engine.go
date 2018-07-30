package core

import (
	"log"
	"goSkylar/lib/nmap"
	"github.com/dean2021/go-masscan"
	"strings"
)

type MasscanResult struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

func RunMasscan(ipRange string, rate string, port string) ([]MasscanResult, error) {

	var masscanResults []MasscanResult

	m := masscan.New()

	// 扫描端口范围
	m.SetPorts(port)

	// 扫描IP范围
	m.SetRanges(ipRange)

	// 扫描速率
	m.SetRate(rate)

	args := []string{
		//"--wait", "5",
		"--exclude-file", "exclude.txt",
	}

	m.SetArgs(args ...)

	// 开始扫描
	err := m.Run()
	if err != nil {
		// 由于masscan原因，扫描目标在排除ip段内，也会认为成一个错误，所以需要排除掉这种错误
		if !strings.Contains(err.Error(), "ranges overlapped something in an excludefile range") {
			log.Println("scanner failed:", err)
			return nil, err
		}
	}

	// 解析扫描结果
	results, err := m.Parse()
	if err != nil {
		log.Println("Parse scanner result:", err)
		return nil, err
	}

	for _, result := range results {
		for _, v := range result.Ports {
			masscanResults = append(masscanResults, MasscanResult{
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
