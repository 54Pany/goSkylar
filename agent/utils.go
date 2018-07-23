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
)

// restart process
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

// downlaod new agent
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
