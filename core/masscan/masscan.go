package masscan

import (
	"os/exec"
	"os"
	"bytes"
	"log"
	"io"
	"encoding/xml"
)

type Masscan struct {
	SystemPath  string
	Args        []string
	Ports       string
	Ranges      string
	Rate        string
	Include     string
	ExcludeFile string
	//FileName   string
	Result []byte
}

type Address struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
}
type State struct {
	State     string `xml:"state,attr"`
	Reason    string `xml:"reason,attr"`
	ReasonTTL string `xml:"reason_ttl,attr"`
}

type Host struct {
	XMLName xml.Name `xml:"host"`
	Endtime string   `xml:"endtime,attr"`
	Address Address  `xml:"address"`
	Ports   Ports    `xml:"ports>port"`
}
type Ports []struct {
	Protocol string  `xml:"protocol,attr"`
	Portid   string  `xml:"portid,attr"`
	State    State   `xml:"state"`
	Service  Service `xml:"service"`
}
type Service struct {
	Name   string `xml:"name,attr"`
	Banner string `xml:"banner,attr"`
}

type MasscanStruct struct {
	IP string `json:"ip"`
	Ports []struct {
		Port   int    `json:"port"`
		Proto  string `json:"proto"`
		Status string `json:"status"`
		Reason string `json:"reason"`
		TTL    int    `json:"ttl"`
	} `json:"ports"`
}

type MasscanResultStruct struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

func deleteFile(filePath string) error {
	return os.Remove(filePath)
}

func (m *Masscan) SetSystemPath(systemPath string) {
	if systemPath != "" {
		//test 将来不会写死
		m.SystemPath = "/export/Data/1.txt"
	}
}
func (m *Masscan) SetArgs(arg ...string) {
	m.Args = arg
}

func (m *Masscan) SetRanges(ranges string) {
	m.Ranges = ranges
}

func (m *Masscan) SetPorts(ports string) {
	m.Ports = ports
}

func (m *Masscan) SetInclude(include string) {
	m.Include = include
}

func (m *Masscan) SetRate(rate string) {
	m.Rate = rate
}

func (m *Masscan) SetExclude(excludefile string) {
	m.ExcludeFile = excludefile
}

// Start scanning
func (m *Masscan) Run() error {
	var (
		cmd  *exec.Cmd
		outb bytes.Buffer
	)
	if m.Rate != "" {
		m.Args = append(m.Args, "--rate")
		m.Args = append(m.Args, m.Rate)
	}
	if m.Ranges != "" {
		m.Args = append(m.Args, "--range")
		m.Args = append(m.Args, m.Ranges)
	}
	if m.Ports != "" {
		m.Args = append(m.Args, "-p")
		m.Args = append(m.Args, m.Ports)
	}
	if m.ExcludeFile != "" {
		m.Args = append(m.Args, "--excludefile")
		m.Args = append(m.Args, m.ExcludeFile)
	}
	if m.Include != "" {
		m.Args = append(m.Args, "--include")
		m.Args = append(m.Args, m.Include)
	}
	m.Args = append(m.Args, "-oX")
	m.Args = append(m.Args, "-")
	cmd = exec.Command("masscan", m.Args...)
	log.Println(cmd.Args)
	cmd.Stdout = &outb
	err := cmd.Run()
	if err != nil {
		log.Println("----Masscan 执行错误-----\n" + string(outb.Bytes()))
		return err
	}
	m.Result = outb.Bytes()
	return err
}

// Parse scans result.
func (m *Masscan) Parse() ([]Host, error) {
	var hosts []Host
	log.Println("-----Masscan------")
	log.Println(string(m.Result))
	decoder := xml.NewDecoder(bytes.NewReader(m.Result))
	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "host" {
				var host Host
				err := decoder.DecodeElement(&host, &se)
				if err == io.EOF {
					break
				}
				if err != nil {
					return nil, err
				}
				hosts = append(hosts, host)
			}
		default:
		}
	}
	return hosts, nil
}

func New() *Masscan {
	return &Masscan{
		SystemPath: "masscan",
	}
}
