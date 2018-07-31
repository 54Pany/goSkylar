package data

import (
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2"
	"log"
	"strings"
	"time"
	"goSkylar/lib/mongo"
	"github.com/pkg/errors"
	"goSkylar/server/lib"
)

var (
	mPortScanResult     = mongo.MongoDriver{}
	portScanResult      *mgo.Collection
	mExternalScan       = mongo.MongoDriver{}
	mExternalScanUrgent = mongo.MongoDriver{}
	externalcan         *mgo.Collection
	externalcanurgent   *mgo.Collection
)

type NmapResult struct {
	Ip        string
	PortId    string
	Protocol  string
	Service   string
	InputTime string
	MachineIp string
	Timestamp int64
	Date      string
}

func init() {
	mPortScanResult = mongo.MongoDriver{TableName: "port_scan_result"}
	err := mPortScanResult.Init()
	if err != nil {
		log.Println("INIT MONGODB ERRPR:" + err.Error())
	}
	portScanResult, err = mPortScanResult.NewTable()
	if err != nil {
		log.Println("MONGODB NewTable portScanResult ERRPR:" + err.Error())
	}
}

func NmapResultToMongo(msg string) error {
	var result NmapResult
	msgList := strings.Split(msg, "|")
	if len(msgList) != 5 {
		log.Println("Nmap 结果数量错误")
		return errors.New("Nmap 结果数量错误")
	}
	timestamp := time.Now().Unix()

	result.Ip = msgList[0]
	result.PortId = msgList[1]
	result.Protocol = msgList[2]
	result.Service = msgList[3]
	result.MachineIp = msgList[4]
	result.InputTime = lib.TimeToStr(timestamp)
	result.Timestamp = timestamp
	result.Date = lib.TimeToData(timestamp)

	//查询数据库中是否存在记录
	count, err := portScanResult.Find(bson.M{"ip": result.Ip, "portid": result.PortId, "protocol": result.Protocol,
		"service": result.Service, "date": result.Date}).Count()
	if err != nil {
		log.Println("----Pipeline数据库查询报错----" + err.Error())
		return err
	}
	if count == 0 {
		//log.Println(result)
		err = portScanResult.Insert(result)
		return err
	}
	return nil
}

func FindIpRanges() []string {
	var allIpRanges []string
	mExternalScan = mongo.MongoDriver{TableName: "external_scan"}
	err := mExternalScan.Init()
	externalcan, err = mExternalScan.NewTable()
	if err != nil {
		log.Println("INIT MONGODB ERRPR:" + err.Error())
	}
	// 初始化数据库连接
	externalcan.Find(bson.M{}).Distinct("iprange", &allIpRanges)

	return allIpRanges
}

func FindUrgentIP() (ipsUrgent []string) {
	mExternalScanUrgent = mongo.MongoDriver{TableName: "external_scan_urgent"}
	err := mExternalScanUrgent.Init()
	externalcanurgent, err = mExternalScanUrgent.NewTable()
	if err != nil {
		log.Println("INIT MONGODB ERRPR:" + err.Error())
	}
	externalcanurgent.Find(bson.M{"isscaned": false}).Distinct("ip", &ipsUrgent)
	return ipsUrgent
}

func UpdateUrgentScanStatus() {
	externalcanurgent.UpdateAll(bson.M{"isscaned": false}, bson.M{"$set": bson.M{"isscaned": true}})
}
