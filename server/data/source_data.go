package data

import (
	"gopkg.in/mgo.v2"
	"log"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
	"goSkylar/server/lib"
	"os"
	"bufio"
	"io"
	"fmt"
	"goSkylar/lib/mongo"
	"goSkylar/server/conf"
)

var (
	mIPRangePool           = mongo.MongoDriver{}
	mpublicCloudTenant     = mongo.MongoDriver{}
	mIPWhiteList           = mongo.MongoDriver{}
	mExternalScanIprange   = mongo.MongoDriver{}
	mExternalScanWhiteList = mongo.MongoDriver{}
	mExternalScanBlackList = mongo.MongoDriver{}
	ipRangePool            *mgo.Collection
	publicCloudTenant      *mgo.Collection
	ipWhiteList            *mgo.Collection
	externalScanIprange    *mgo.Collection
	externalScanWhiteList  *mgo.Collection
	externalScanBlackList  *mgo.Collection
)

type IPRangePool struct {
	IPRange    string `bson:"ip_range"`
	Flag       string `bson:"flag"`
	IsPrivate  int    `bson:"is_private"`
	SeaNetType string `bson:"sea_net_type"`
}

func init() {
	// 数据库表实例化
	mIPRangePool = mongo.MongoDriver{TableName: "iprange_pool"}
	mpublicCloudTenant = mongo.MongoDriver{TableName: "monitor_public_cloud_tenant_v2"}
	mIPWhiteList = mongo.MongoDriver{TableName: "port_whitelist"}
	mExternalScanIprange = mongo.MongoDriver{TableName: "external_scan_iprange"}
	mExternalScanWhiteList = mongo.MongoDriver{TableName: "external_scan_white_list"}
	mExternalScanBlackList = mongo.MongoDriver{TableName: "external_scan_black_list"}
	err := mIPRangePool.Init()
	err = mpublicCloudTenant.Init()
	err = mIPWhiteList.Init()
	err = mExternalScanIprange.Init()
	err = mExternalScanWhiteList.Init()
	err = mExternalScanBlackList.Init()
	ipRangePool, err = mIPRangePool.NewTable()
	publicCloudTenant, err = mpublicCloudTenant.NewTable()
	ipWhiteList, err = mIPWhiteList.NewTable()
	externalScanIprange, err = mExternalScanIprange.NewTable()
	externalScanWhiteList, err = mExternalScanWhiteList.NewTable()
	externalScanBlackList, err = mExternalScanBlackList.NewTable()
	if err != nil {
		log.Println("Init Mongodb Error: " + err.Error())
	}
}

func FindInitIPS() (err error) {
	var ipBlack []string
	// 初始化数据库连接
	log.Println("......初始化数据库连接......", lib.CurrentTimeForPrint())
	iPRangePool := new([]IPRangePool)
	// 第一步：查询所有 外网-VPC租户-企业租户-已下线 的所有IP段
	log.Println("......第一步：查询所有 外网-VPC租户-企业租户-已下线 的所有IP段......", lib.CurrentTimeForPrint())
	err = ipRangePool.Find(bson.M{"is_private": 2, "sea_net_type": bson.M{"$nin": []string{"VPC租户", "企业租户", "物理云"}}, "flag": bson.M{"$ne": "已下线"}}).All(iPRangePool)
	if err != nil {
		log.Println(err)
		return err
	}
	ipList := []string{}
	ipRangeList := []string{}
	// 生成所有外网ip地址
	for _, v := range *iPRangePool {
		for _, w := range ipRangeToIPs(v.IPRange) {
			ipList = append(ipList, w)
		}
	}
	// 查询所有外网IP段
	for _, v := range *iPRangePool {
		ipRangeList = append(ipRangeList, v.IPRange)
	}
	// 第二步：查询公有云租户
	log.Println("......第二步：查询公有云租户......", lib.CurrentTimeForPrint())
	publicCloudTenantIPs := findPublicCloudTenantIPs()
	// 第三步：去除外网IP私用
	log.Println("......第三步：去除外网IP私用......", lib.CurrentTimeForPrint())
	ipWithoutPriUse := findIPWithoutPriUse(ipList)
	// 第四步：外网IP信息去除云租户
	log.Println("......第四步：外网IP信息去除云租户......", lib.CurrentTimeForPrint())
	ipWithoutTenant := findList1WithoutList2(ipWithoutPriUse, publicCloudTenantIPs)
	// 第五步：查询白名IP单
	log.Println("......第五步：查询白名IP单......", lib.CurrentTimeForPrint())
	ipWhite := findIPInWhiteList()
	// 第六步：白名单去除云租户 慢扫
	log.Println("......第六步：白名单去除云租户 慢扫......", lib.CurrentTimeForPrint())
	ipWhiteWithoutTenant := findList1WithoutList2(ipWhite, publicCloudTenantIPs)
	// 第七步：外网IP信息去除云租户后去除白名单 正常速度扫
	log.Println("......第七步：外网IP信息去除云租户后去除白名单 正常速度扫......", lib.CurrentTimeForPrint())
	ipWithoutTenantAndWihte := findList1WithoutList2(ipWithoutTenant, ipWhite)
	// 第八步：生成黑名单
	log.Println("......第八步：生成黑名单......", lib.CurrentTimeForPrint())
	ipBlack1 := findList1WithoutList2(ipList, ipWithoutPriUse)
	ipBlack2 := findList1WithoutList2(ipWithoutPriUse, ipWithoutTenant)
	ipBlack3 := findList1WithoutList2(ipWithoutTenant, ipWithoutTenantAndWihte)
	ipBlack = append(ipBlack1)
	ipBlack = append(ipBlack2)
	ipBlack = append(ipBlack3)
	log.Println("......最终生成结果统计......", lib.CurrentTimeForPrint())
	log.Println("......生成iprange个数：", len(ipRangeList), "......")
	log.Println("......生成iprange白名单个数：", len(ipWhiteWithoutTenant), "......")
	log.Println("......生成ip黑名单个数：", len(ipBlack), "......")

	// 分别插入数据库
	// iprange
	_, err = externalScanIprange.RemoveAll(bson.M{})
	if err != nil {
		log.Println(err)
		return
	}
	for _, v := range ipRangeList {
		err = externalScanIprange.Insert(bson.M{"ip_range": v})
		if err != nil {
			log.Println(err)
			return
		}
	}
	// ip_range_white
	_, err = externalScanWhiteList.RemoveAll(bson.M{})
	if err != nil {
		log.Println(err)
		return
	}
	for _, v := range ipWhiteWithoutTenant {
		err = externalScanWhiteList.Insert(bson.M{"ip_range_white": v})
		if err != nil {
			log.Println(err)
			return
		}
	}
	// ip_black
	_, err = externalScanBlackList.RemoveAll(bson.M{})
	if err != nil {
		log.Println(err)
		return
	}
	for _, v := range ipBlack {
		err = externalScanBlackList.Insert(bson.M{"ip_black": v})
		if err != nil {
			log.Println(err)
			return
		}
	}

	return err
}

// 将IP段数据打散成IP地址数据
func ipRangeToIPs(ipaddr string) []string {
	ipRangeList := strings.Split(ipaddr, "/")
	if len(ipRangeList) != 2 {
		return []string{}
	}
	ip := ipRangeList[0]
	mask, err := strconv.Atoi(ipRangeList[1])
	if err != nil {
		return []string{}
	}
	var result []string
	if mask > 32 || mask < 0 {
		log.Println("netmask error")
		return result
	}
	maskhead := 0xFFFFFFFF
	for i := 1; i <= 32-mask; i++ {
		maskhead = maskhead << 1
	}
	masktail := 0xFFFFFFFF
	for i := 1; i <= mask; i++ {
		masktail = masktail >> 1
	}
	ipint := lib.IpStringToInt(ip)
	IPintstart := ipint & maskhead
	IPintend := ipint | masktail
	for i := IPintstart; i <= IPintend; i++ {
		result = append(result, lib.IpIntToString(i))
	}
	return result
}

// 查询所有公有云租户租户IP信息
func findPublicCloudTenantIPs() []string {
	publicCloudTenantType := []string{}
	publicCloudTenant.Find(bson.M{}).Distinct("cidr", &publicCloudTenantType)
	ipList := []string{}
	for _, v := range publicCloudTenantType {
		for _, w := range ipRangeToIPs(v) {
			ipList = append(ipList, w)
		}
	}
	return ipList
}

// List1中去除List2
func findList1WithoutList2(ipList1 []string, ipList2 []string) []string {
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

// 查询IP白名单
func findIPInWhiteList() []string {
	ipRangeList := []string{}
	ipList := []string{}
	ipWhiteList.Find(bson.M{}).Distinct("iprange", &ipRangeList)
	for _, v := range ipRangeList {
		for _, w := range ipRangeToIPs(v) {
			ipList = append(ipList, w)
		}
	}
	return ipList
}

// 外网IP信息去除外网IP私用
func findIPWithoutPriUse(ipList []string) []string {
	privateIpUse := conf.PUBLIC_IP_PRIVATEUSE
	ipListStart := []string{}
	ipRangeStart := strings.Split(privateIpUse, ",")
	for k, v := range ipList {
		ipListStart = strings.Split(v, ".")
		if ipListStart[0] == ipRangeStart[0] {
			ipList = append(ipList[:k], ipList[k+1:]...)
		}
	}
	return ipList
}

// 处理新加入的白名单 主函数里单独执行即可
func UpdateWhite() {
	var whiteList []string
	f, err := os.Open("data/20180718.txt")
	if err != nil {
		log.Println()
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	for {
		line, _, err := rd.ReadLine()
		if err != nil || io.EOF == err {
			break
		}
		whiteList = append(whiteList, string(line))
	}
	for _, v := range whiteList {
		err := ipWhiteList.Insert(bson.M{"iprange": v})
		if err != nil {
			fmt.Println(err)
		}
	}
}
