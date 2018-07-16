package main

import (
	"goSkylar/core"
	"goSkylar/lib"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/bipabo1l/goworker"
	"github.com/go-redis/redis"
	"github.com/levigross/grequests"
)

var (
	RedisDriver *redis.Client
	waitgroup   sync.WaitGroup
	dsnAddr     string
)

// 扫描
func ScanTask(queue string, args ...interface{}) error {
	log.Println("调用队列:" + queue)
	ipRange := args[0].(string)
	rate := args[1].(string)
	taskTime := args[2].(string)
	log.Println(ipRange)
	core.CoreScanEngine(ipRange, rate, taskTime)
	log.Println("From " + queue " " + args)
	return nil
}

var (
	version     = "1.0.0"
	downloadURL = ""
)

//check版本
func VersionValidate(c chan string, versionURL string, linuxDownloadURL string) bool {
	resp, err := grequests.Get(versionURL, nil)

	// You can modify the request by passing an optional RequestOptions struct
	if err != nil {
		log.Println("Validate version error: Unable to make request ")
		return false
	}

	newVersion := resp.String()
	if version != newVersion {
		downloadURL = linuxDownloadURL
		download, _ := DownloadNewAgent(downloadURL)
		if download == true {
			c <- "new"
			log.Println("-----发现新版本-------" + newVersion)
			return true
		}
		c <- "old"
		return false
	}
	log.Println("-----Version:当前版本已是最新-----版本号：" + version)
	return false
}

//downlaod new agent
func DownloadNewAgent(url string) (bool, error) {
	res, err := http.Get(url)
	if err != nil {
		return false, err
	}
	var fileName string

	fileName = "agent"

	cmd := exec.Command("cp", fileName, "/export/Data/agent_bak/" + fileName + version)
	cmd.Run()

	cmd = exec.Command("rm", "-rf", fileName)
	cmd.Run()

	f, err := os.Create(fileName)
	if err != nil {
		log.Println(err.Error())
		return false, err
	}
	_, er := io.Copy(f, res.Body)
	if er != nil {
		log.Println(er.Error())
		return false, er
	}

	log.Println("-----新版本下载成功-------")

	cmdd := exec.Command("chmod", "+x", file_name)
	cmdd.Run()

	res.Body.Close()
	f.Close()
	return true, er
}

//restart process
func RestartProcess() {
	filePath, _ := filepath.Abs(os.Args[0])
	cmd := exec.Command(filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatalf("GracefulRestart: Failed to launch, error: %v", err)
	}
}

func init() {
	RedisDriver = lib.RedisDriver

	dsnAddr = lib.DsnAddr

	// 初始化
	settings := goworker.WorkerSettings{
		URI:            dsnAddr,
		Connections:    100,
		Queues:         []string{"ScanTaskQueue"},
		UseNumber:      true,
		ExitOnComplete: false,
		Concurrency:    1,
		Namespace:      "goskylar:",
		Interval:       5.0,
	}

	goworker.SetSettings(settings)

	goworker.Register("ScanTask", ScanTask)
}

func main() {
	//lib.LogSetting()

	cfg := lib.NewConfigUtil("")
	versionURL, _ := cfg.GetString("web_default", "version_url")
	DownloadURL, _ := cfg.GetString("web_default", "download_url")

	waitgroup.Add(1)

	signals := make(chan string)

	go func() {
		for {
			VersionValidate(signals, versionURL, DownloadURL)
			time.Sleep(1 * time.Minute)
		}
	}()

	go func() {
		for {
			if err := goworker.Work(); err != nil {
				log.Println("Error:", err)
			}
		}
	}()

	for {
		select {
		case signal := <-signals:
			if signal == "new" {
				RestartProcess()
				return
			}
		case <-time.After(30 * time.Second):
			log.Println("your version is the latest, check again after 10 second...")
			continue
		}
	}

}
