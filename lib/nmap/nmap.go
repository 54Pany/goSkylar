package nmap

import (
	"os/exec"
	"bytes"
	"github.com/lair-framework/go-nmap"
	"strconv"
	"log"
)

type Nmap struct {
	SystemPath    string
	Args          []string
	Ports         string
	HostTimeOut   string
	MaxRttTimeOut string
	IP            string
	Result        []byte
}

type NmapResult struct {
	Ip       string
	PortId   int
	Protocol string
	Service  string
}

type PortsDesc struct {
	PortId   int
	Protocol string
	Service  string
}

type OsDesc struct {
	Name     string
	Accuracy int
}

func (m *Nmap) SetArgs(arg ...string) {
	m.Args = arg
}

func (m *Nmap) SetPorts(ports string) {
	m.Ports = ports
}

func (m *Nmap) SetHostTimeOut(hostTimeOut string) {
	m.HostTimeOut = hostTimeOut
}

func (m *Nmap) SetMaxRttTimeOut(maxRttTimeOut string) {
	m.MaxRttTimeOut = maxRttTimeOut
}

func (m *Nmap) SetIP(ip string) {
	m.IP = ip
}

// Start scanning
func (m *Nmap) Run() error {
	var (
		cmd  *exec.Cmd
		outb bytes.Buffer
	)

	m.Args = append(m.Args, "-sS")
	m.Args = append(m.Args, "-n")
	m.Args = append(m.Args, "-T4")
	m.Args = append(m.Args, "-Pn")
	m.Args = append(m.Args, "-open")

	if m.Ports != "" {
		m.Args = append(m.Args, "-p")
		m.Args = append(m.Args, m.Ports)
	}

	if m.HostTimeOut != "" {
		m.Args = append(m.Args, "--host-timeout")
		m.Args = append(m.Args, m.HostTimeOut)
	}

	if m.MaxRttTimeOut != "" {
		m.Args = append(m.Args, "--max-rtt-timeout")
		m.Args = append(m.Args, m.MaxRttTimeOut)
	}

	m.Args = append(m.Args, "-oX")

	if m.IP != "" {
		m.Args = append(m.Args, "-")
		m.Args = append(m.Args, m.IP)
	}
	cmd = exec.Command("nmap", m.Args...)
	log.Println(cmd.Args)
	cmd.Stdout = &outb
	err := cmd.Run()
	if err != nil {
		return err
	}
	m.Result = outb.Bytes()
	return nil
}

// Parse scans result.
func (m *Nmap) Parse() ([]NmapResult, error) {

	var nmapResults []NmapResult

	log.Println("-----Nmap------")

	x, err := nmap.Parse(m.Result)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(x.Hosts); i++ {
		if x.Hosts[i].Status.State == "up" {
			var (
				mPortDesc PortsDesc
				PortList  []PortsDesc
				OsInfo    OsDesc
			)
			IP := x.Hosts[i].Addresses[0].Addr
			for t := 0; t < len(x.Hosts[i].Ports); t++ {
				mPortDesc.PortId = x.Hosts[i].Ports[t].PortId
				mPortDesc.Protocol = x.Hosts[i].Ports[t].Protocol
				mPortDesc.Service = x.Hosts[i].Ports[t].Service.Name
				PortList = append(PortList, mPortDesc)
			}
			for y := 0; y < len(x.Hosts[i].Os.OsMatches); y++ {
				tmp, _ := strconv.Atoi(x.Hosts[i].Os.OsMatches[y].Accuracy)
				if tmp > OsInfo.Accuracy {
					OsInfo.Name = x.Hosts[i].Os.OsMatches[y].Name
					OsInfo.Accuracy = tmp
				}

			}
			if len(PortList) != 0 {

				for _, v := range PortList {
					nmapResults = append(nmapResults, NmapResult{
						Ip: IP,
						Service:v.Service,
						PortId:v.PortId,
						Protocol:v.Protocol,
					})
				}
			}
		}
	}
	return nmapResults, err
}

func New() *Nmap {
	return &Nmap{
		SystemPath: "nmap",
	}
}
