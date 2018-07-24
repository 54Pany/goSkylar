package lib

import (
	"strings"
	"strconv"
	"time"
	"gopkg.in/mgo.v2/bson"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
)

func TimeToStr(intTime int64) string {
	timeLayout := "2006-01-02 15:04:05"                     //转化所需模板
	dataTimeStr := time.Unix(intTime, 0).Format(timeLayout) //设置时间戳 使用模板格式化为日期字符串
	return dataTimeStr
}

func TimeToData(intTime int64) string {
	timeLayout := "2006-01-02"                     //转化所需模板
	dataTimeStr := time.Unix(intTime, 0).Format(timeLayout) //设置时间戳 使用模板格式化为日期字符串
	return dataTimeStr
}

func CurrentTime() string {
	return TimeToStr(time.Now().Unix())
}

func DateToStr(intTime int64) string {
	timeLayout := "2006-01-02"
	dataTimeStr := time.Unix(intTime, 0).Format(timeLayout)
	return dataTimeStr
}

// 获取当天的日期
func CurrentDate() string {
	return DateToStr(time.Now().Unix())
}

//ip转换，如果是ip则转换成ip段
func Iptransfer(ip string) string {
	if strings.Contains(ip, "/") {
		return ip
	}
	return ip + "/32"
}

func InterfaceToStr(inter interface{}) (s string) {
	tempStr := ""
	switch inter.(type) {
	case nil:
		tempStr = ""
		break
	case string:
		tempStr = inter.(string)
		break
	case float64:
		tempStr = strconv.FormatFloat(inter.(float64), 'f', -1, 64)
		break
	case int64:
		tempStr = strconv.FormatInt(inter.(int64), 10)
		break
	case int:
		tempStr = strconv.Itoa(inter.(int))
		break
	case bool:
		tempStr = strconv.FormatBool(inter.(bool))
	case bson.ObjectId:
		tempStr = inter.(bson.ObjectId).Hex()
	case []interface{}:
		tempStr, _ = JsonToString(inter)
	case []int:
		tempStr, _ = JsonToString(inter)
	case []int64:
		tempStr, _ = JsonToString(inter)
	case []float32:
		tempStr, _ = JsonToString(inter)
	case []float64:
		tempStr, _ = JsonToString(inter)
	case map[string]interface{}:
		tempStr, _ = JsonToString(inter)
	case map[string]string:
		tempStr, _ = JsonToString(inter)
	case time.Time:
		tempStr = inter.(time.Time).String()
	default:
		tempStr = "Error! Not Found Type!"
	}
	return tempStr
}

func JsonToString(inter interface{}) (string, error) {
	by, err := json.Marshal(inter)
	if err != nil {
		return "", err
	} else {
		return string(by), nil
	}
}

//md5加密
func Md5Str(str string) string {
	strMd5 := md5.New()
	strMd5.Write([]byte(str))
	return hex.EncodeToString(strMd5.Sum(nil))
}
