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

	"strconv"

	"goworker"
	"github.com/go-redis/redis"
	"github.com/levigross/grequests"
)

var (
	RedisDriver *redis.Client
	Waitgroup   sync.WaitGroup
	DsnAddr     string
)

// 扫描
func ScanMasscanTask(queue string, args ...interface{}) error {
	log.Println("调用队列Masscan:" + queue)
	ipRange := ""
	rate := "20"
	taskTime := "get_queue_error"
	if len(args) != 4 {
		log.Println("----ScanMasscanTask 参数个数错误-----")
		return nil
	}

	ipRange = args[0].(string)
	rate = args[1].(string)
	taskTime = args[2].(string)
	port := args[3].(string)
	log.Println(ipRange)
	err := core.CoreScanEngine(ipRange, rate, taskTime, port)

	log.Println("From " + queue + " " + args[2].(string))
	return err
}

func ScanNmapTask(queue string, args ...interface{}) error {
	log.Println("调用队列Nmap:" + queue)
	if len(args) == 1 {
		err := core.CoreScanNmapEngine(args[0].(string))
		if err != nil {
			log.Println("CoreScanNmapEngine Error:" + err.Error())
			return err
		}
	}
	return nil
}

var (
	version     = "1.0.11"
	downloadURL = ""
)

//判断路径是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//check版本
func VersionValidate(c chan string, versionURL string,
	linuxDownloadURL string, bakFile string) bool {

	resp, err := grequests.Get(versionURL, nil)

	// You can modify the request by passing an optional RequestOptions struct
	if err != nil {
		log.Println("Validate version error: Unable to make request " + err.Error())
		return false
	}

	newVersion := resp.String()
	if version != newVersion {
		t := strconv.FormatInt(time.Now().Unix(), 10)
		timestampStr := lib.InterfaceToStr(t)
		authkey := "gPv94qxP"
		sign := lib.Md5Str(timestampStr + authkey)

		downloadURL = linuxDownloadURL + "?timestamp=" + timestampStr + "&sign=" + sign
		log.Println(downloadURL)
		download, _ := DownloadNewAgent(downloadURL, bakFile)
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
func DownloadNewAgent(url, bakFile string) (bool, error) {
	res, err := http.Get(url)
	if err != nil {
		return false, err
	}

	existPath, err := PathExists(bakFile)
	if err != nil {
		log.Println("get dir error:" + err.Error())
		return false, err
	}
	if existPath == false {
		err := os.Mkdir(bakFile, os.ModePerm)
		if err != nil {
			log.Printf("mkdir bak_file failed:" + err.Error())
			return false, err
		}
	}

	fileName := "agent"
	cmd := exec.Command("cp", fileName, bakFile+fileName+"."+version)
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

	cmdd := exec.Command("chmod", "+x", fileName)
	cmdd.Run()

	res.Body.Close()
	f.Close()
	return true, er
}

//restart process
func RestartProcess() {
	filePath, _ := filepath.Abs(os.Args[0])
	cmd := exec.Command(filePath)
	log.Println("FilePath:")
	log.Println(filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatalf("GracefulRestart: Failed to launch, error: %v", err)
	}
}

func init() {
	RedisDriver = lib.RedisDriver

	DsnAddr = lib.DsnAddr

	// 初始化
	settings := goworker.WorkerSettings{
		URI:            DsnAddr,
		Connections:    100,
		Queues:         []string{"ScanMasscanTaskQueue", "ScanNmapTaskQueue"},
		UseNumber:      true,
		ExitOnComplete: false,
		Concurrency:    1,
		Namespace:      "goskylar:",
		Interval:       5.0,
	}

	goworker.SetSettings(settings)

	goworker.Register("ScanMasscanTask", ScanMasscanTask)

	goworker.Register("ScanNmapTask", ScanNmapTask)

}

func main() {
	//lib.LogSetting()

	cfg := lib.NewConfigUtil("")
	versionURL, _ := cfg.GetString("web_default", "version_url")
	DownloadURL, _ := cfg.GetString("web_default", "download_url")
	bakFile, _ := cfg.GetString("bak_file", "bak_path")

	signals := make(chan string)

	Waitgroup.Add(2)

	go func() {
		defer Waitgroup.Done()
		for {
			VersionValidate(signals, versionURL, DownloadURL, bakFile)
			time.Sleep(1 * time.Minute)
		}
	}()

	go func() {
		defer Waitgroup.Done()

		if err := goworker.Work(); err != nil {
			log.Println("Error:", err)
		}

	}()

	for {
		select {
		case signal := <-signals:
			if signal == "new" {
				RestartProcess()
				return
			}
		case <-time.After(time.Second * 10):
			continue
		}
	}

	Waitgroup.Wait()
}
