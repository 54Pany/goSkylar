package main

import (
	"time"
	"goSkylar/lib"
	"github.com/go-redis/redis"
	"sync"
	"fmt"
	"github.com/bipabo1l/goworker"
	"goSkylar/core"
	"github.com/levigross/grequests"
	"os/exec"
	"os"
	"io"
	"net/http"
	"path/filepath"
	"log"
)

var (
	RedisDriver *redis.Client
	waitgroup   sync.WaitGroup
	dsnAddr     string
)

// 扫描
func ScanTask(queue string, args ...interface{}) error {
	fmt.Println("调用队列:" + queue)
	ipRange := args[0].(string)
	rate := args[1].(string)
	taskTime := args[2].(string)
	fmt.Println(ipRange)
	core.CoreScanEngine(ipRange, rate, taskTime)
	fmt.Printf("From %s, %v\n", queue, args)
	return nil
}

var (
	version     = "1.0.0"
	downloadUrl = ""
)

func Version_validate(c chan string, version_url string, linux_download_url string) (bool) {
	resp, err := grequests.Get(version_url, nil)

	// You can modify the request by passing an optional RequestOptions struct
	if err != nil {
		fmt.Println("Validate version error: Unable to make request ")
		return false
	} else {
		new_version := resp.String()
		if version != new_version {
			downloadUrl = linux_download_url

			download, _ := Download_new_agent(downloadUrl)
			if download == true {
				c <- "new"
				fmt.Println("-----发现新版本-------" + new_version)
				return true
			} else {
				c <- "old"
				return false
			}
		} else {
			log.Println("-----Version:当前版本已是最新-----版本号：" + version)
			return false
		}
	}
}

func Download_new_agent(url string) (bool, error) {
	res, err := http.Get(url)
	if err != nil {
		return false, err
	}
	var file_name string

	file_name = "agent"

	cmd := exec.Command("cp", file_name, "/export/Data/agent_bak/"+file_name)
	cmd.Run()

	cmd = exec.Command("rm", "-rf", file_name)
	cmd.Run()

	f, err := os.Create(file_name)
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

func Restart_process() {
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
	versionUrl, _ := cfg.GetString("web_default", "version_url")
	DownloadUrl, _ := cfg.GetString("web_default", "download_url")

	waitgroup.Add(1)

	signals := make(chan string)

	go func() {
		for {
			Version_validate(signals, versionUrl, DownloadUrl)
			time.Sleep(1 * time.Minute)
		}
	}()

	go func() {
		for {
			if err := goworker.Work(); err != nil {
				fmt.Println("Error:", err)
			}
		}
	}()

	for {
		select {
		case signal := <-signals:
			if signal == "new" {
				Restart_process()
				return
			}
		case <-time.After(30 * time.Second):
			fmt.Println("your version is the latest, check again after 10 second...")
			continue
		}
	}

	waitgroup.Wait()
}
