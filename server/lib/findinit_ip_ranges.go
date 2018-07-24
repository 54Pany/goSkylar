package lib

import (
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2"
	"log"
)

var (
	mExternalScan       = MongoDriver{}
	mExternalScanUrgent = MongoDriver{}
	externalcan         *mgo.Collection
	externalcanurgent   *mgo.Collection
)


func FindInitIpRanges() (allipranges []string, allipsWhite []string, blackIps []string) {
	mExternalScan = MongoDriver{TableName: "external_scan"}
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
	mExternalScanUrgent = MongoDriver{TableName: "external_scan_urgent"}
	err := mExternalScanUrgent.Init()
	externalcanurgent, err = mExternalScanUrgent.NewTable()
	if err != nil {
		log.Println("INIT MONGODB ERRPR:" + err.Error())
	}
	externalcanurgent.Find(bson.M{"isscaned": false}).Distinct("ip", &ipsUrgent)
	return ipsUrgent
}
