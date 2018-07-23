package data

import (
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2"
	"log"
	"goSkylar/server/lib"
	"strings"
	"strconv"
	"time"
)

var (
	mExternalScan       = lib.MongoDriver{}
	mExternalScanUrgent = lib.MongoDriver{}
	mPortScanResult     = lib.MongoDriver{}
	portScanResult      *mgo.Collection
	externalcan         *mgo.Collection
	externalcanurgent   *mgo.Collection
)

type NmapResultStruct struct {
	Ip        string
	PortId    int
	Protocol  string
	Service   string
	InputTime string
	TaskTime  string
	MachineIp string
}

func init() {
	mPortScanResult = lib.MongoDriver{TableName: "port_scan_result"}
	err := mPortScanResult.Init()
	if err != nil {
		log.Println("INIT MONGODB ERRPR:" + err.Error())
	}
	portScanResult, err = mPortScanResult.NewTable()
}

func FindInitIpRanges() (allipranges []string, allipsWhite []string, blackIps []string) {
	mExternalScan = lib.MongoDriver{TableName: "external_scan"}
	err := mExternalScan.Init()
	externalcan, err = mExternalScan.NewTable()
	if err != nil {
		log.Println("INIT MONGODB ERRPR:" + err.Error())
	}
	// 初始化数据库连接
	externalcan.Find(bson.M{}).Distinct("iprange", &allipranges)
	externalcan.Find(bson.M{}).Distinct("iprange_white", &allipsWhite)
	externalcan.Find(bson.M{}).Distinct("ip_black", &blackIps)

	return allipranges, allipsWhite, blackIps
}

func FindUrgentIP() (ipsUrgent []string) {
	mExternalScanUrgent = lib.MongoDriver{TableName: "external_scan_urgent"}
	err := mExternalScanUrgent.Init()
	externalcanurgent, err = mExternalScanUrgent.NewTable()
	if err != nil {
		log.Println("INIT MONGODB ERRPR:" + err.Error())
	}
	externalcanurgent.Find(bson.M{"isscaned": false}).Distinct("ip", &ipsUrgent)
	return ipsUrgent
}

func NmapResultToMongo(msg string) error {
	msgList := strings.Split(msg, "§§§§")
	if len(msgList) == 6 {

		var msgStruct NmapResultStruct
		msgStruct.Ip = msgList[0]
		msgStruct.PortId, _ = strconv.Atoi(msgList[1])
		msgStruct.Protocol = msgList[2]
		msgStruct.Service = msgList[3]
		msgStruct.TaskTime = msgList[4]
		msgStruct.MachineIp = msgList[5]
		msgStruct.InputTime = time.Now().Format("2006-01-02 15:04:05")
		//查询数据库中是否存在记录
		count, err := portScanResult.Find(
			bson.M{"ip": msgStruct.Ip, "portid": msgStruct.PortId,
				"protocol": msgStruct.Protocol, "service": msgStruct.Service,
				"tasktime": msgStruct.TaskTime}).Count()
		if err != nil {
			log.Println("----Pipeline数据库查询报错----" + err.Error())
			return err
		}
		log.Println(count)
		if count == 0 {
			log.Println("--------Data Insert----------")
			log.Println(msgList)
			portScanResult.Insert(msgStruct)
		}
	}
	return nil
}
