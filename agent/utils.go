package main

import (
	"path/filepath"
	"os"
	"os/exec"
	"log"
	"io"
	"net/http"
	"strconv"
	"time"
	"github.com/levigross/grequests"
	"goSkylar/lib"
	"strings"
	"fmt"
	"github.com/onsi/ginkgo/config"
)

// restart process
func RestartProcess() {
	filePath, _ := filepath.Abs(os.Args[0])
	var args []string
	for i := 1; i < len(os.Args); i++ {
		for _, v := range strings.Split(os.Args[i], "=") {
			args = append(args, v)
		}
	}
	cmd := exec.Command(filePath, args...)
	log.Println("FilePath:")
	log.Println(filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatalf("GracefulRestart: Failed to launch, error: %v", err)
	}
}

// downlaod new agent
func DownloadNewAgent(url, bakPath string) (bool, error) {

	res, err := http.Get(url)
	if err != nil {
		return false, err
	}

	bakPath = fmt.Sprintf("%s/%s/", bakPath, config.VERSION)
	existPath, err := PathExists(bakPath)
	if err != nil {
		log.Println("get dir error:" + err.Error())
		return false, err
	}
	if existPath == false {
		err := os.Mkdir(bakPath, os.ModePerm)
		if err != nil {
			log.Printf("mkdir bak_file failed:" + err.Error())
			return false, err
		}
	}

	fileName := "agent"
	cmd := exec.Command("cp", fileName, bakPath)
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
	cmd = exec.Command("chmod", "+x", fileName)
	cmd.Run()

	res.Body.Close()
	f.Close()

	return true, er
}

// 判断路径是否存在
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

// check版本
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

		downloadURL := linuxDownloadURL + "?timestamp=" + timestampStr + "&sign=" + sign
		log.Println(downloadURL)

		// 下载远程新版本文件
		download, err := DownloadNewAgent(downloadURL, bakFile)
		if err != nil {
			log.Println("Validate version error: Unable to make request " + err.Error())
			return false
		}

		if download == true {
			c <- "new"
			log.Println("-----升级成功,等待重启-------" + newVersion)
			return true
		}

		c <- "old"

		return false
	}

	log.Println("-----Version:当前版本已是最新-----版本号：" + version)
	return false
}
