package lib

import (
	"log"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2"
)

var (
	mIPRangePool       = MongoDriver{}
	mpublicCloudTenant = MongoDriver{}
	mIPWhiteList       = MongoDriver{}
	ipRangePool        *mgo.Collection
	publicCloudTenant  *mgo.Collection
	ipWhiteList        *mgo.Collection
)

type IPRangePool struct {
	IPRange    string `bson:"ip_range"`
	Flag       string `bson:"flag"`
	IsPrivate  int    `bson:"is_private"`
	SeaNetType string `bson:"sea_net_type"`
}

//func init() {
//	mIPRangePool = MongoDriver{TableName: "iprange_pool"}
//	mpublicCloudTenant = MongoDriver{TableName: "public_cloud_tenant"}
//	mIPWhiteList = MongoDriver{TableName: "port_whitelist"}
//	err := mIPRangePool.Init()
//	err = mpublicCloudTenant.Init()
//	err = mIPWhiteList.Init()
//	ipRangePool, err = mIPRangePool.NewTable()
//	publicCloudTenant, err = mpublicCloudTenant.NewTable()
//	ipWhiteList, err = mIPWhiteList.NewTable()
//	if err != nil {
//		log.Println("INIT MONGODB ERRPR:" + err.Error())
//	}
//}

func findPublicCloudTenantIPs() []string {
	publicCloudTenantType := []string{}
	publicCloudTenant.Find(bson.M{}).Distinct("iprange", &publicCloudTenantType)
	ipList := []string{}
	for _, v := range publicCloudTenantType {
		for _, w := range IpRangeToIPs(v) {
			ipList = append(ipList, w)
		}
	}
	return ipList
}

func findIPWithoutTenant(ipList1 []string, ipList2 []string) []string {
	m := make(map[string]int)
	ipList := []string{}

	for _, v := range ipList1 {
		m[v] = 1
	}
	for _, v := range ipList2 {
		m[v] = 2
	}
	for key, value := range m {
		if value == 1 {
			ipList = append(ipList, key)
		}
	}
	return ipList
}

func findIPInWhiteList() []string {
	ipRangeList := []string{}
	ipList := []string{}
	ipWhiteList.Find(bson.M{}).Distinct("iprange", &ipRangeList)
	for _, v := range ipRangeList {
		for _, w := range IpRangeToIPs(v) {
			ipList = append(ipList, w)
		}
	}
	return ipList
}

func findAllIPs(ipList []string, whiteiplist []string) (allips []string, allips_slow []string) {
	m := make(map[string]int)

	for _, v := range ipList {
		m[v] = 1
	}
	for _, v := range whiteiplist {
		if m[v] == 1 {
			m[v] = 3
		} else {
			m[v] = 2
		}
	}
	for key, value := range m {
		if value == 1 {
			allips = append(allips, key)
		} else if value == 3 {
			allips_slow = append(allips_slow, key)
		}
	}

	iplist := findPublicCloudTenantIPs()
	iplist2 := []string{}
	//判断allips不在云租户里
	flag := 0
	for _, v := range allips_slow {
		for _, w := range iplist {
			if v == w {
				flag = 1
			}
		}
		if flag == 0 {
			iplist2 = append(iplist2, v)
		}
	}

	flag = 0

	return allips, iplist2
}

func FindInitIPS() (allips []string, allips_white []string) {
	// 初始化数据库连接
	iPRangePool := new([]IPRangePool)
	ipRangePool.Find(bson.M{"is_private": 2, "sea_net_type": bson.M{"$nin": []string{"VPC租户", "企业租户"}}, "flag": bson.M{"$ne": "已下线"}}).All(iPRangePool)
	ipList := []string{}
	for _, v := range *iPRangePool {
		for _, w := range IpRangeToIPs(v.IPRange) {
			ipList = append(ipList, w)
		}
	}
	publicCloudTenantIPs := findPublicCloudTenantIPs()
	ipWithoutTenant := findIPWithoutTenant(ipList, publicCloudTenantIPs)
	ipWhite := findIPInWhiteList()
	allips, allips_white = findAllIPs(ipWithoutTenant, ipWhite)

	log.Println(len(allips))
	log.Println(len(allips_white))

	return allips, allips_white
}
