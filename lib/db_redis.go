package lib

import (
	"strconv"
	"time"
	"github.com/go-redis/redis"
	"fmt"
	"log"
)

var cfg = NewConfigUtil("")
var redisHost, _ = cfg.GetString("redis_default", "host")
var redisPort, _ = cfg.GetString("redis_default", "port")
var redisPass, _ = cfg.GetString("redis_default", "pass")
var redisDbStr, _ = cfg.GetString("redis_default", "db")
var redisDb, _ = strconv.Atoi(redisDbStr)
var redisChannel, _ = cfg.GetString("redis_default", "channel")

var RedisDriver = redis.NewClient(&redis.Options{
	Addr:        redisHost + ":" + redisPort,
	Password:    redisPass,
	DB:          redisDb,
	DialTimeout: time.Second * 2,
	//IdleTimeout: time.Second * 1000000,
})

var DsnAddr = fmt.Sprintf("redis://root:%s@%s:%s/%s", redisPass, redisHost, redisPort, redisDbStr)

func PushPortInfoToRedis(infoStr string, taskTime string, selfIp string) error {
	var redisDriver = RedisDriver
	infoStr = infoStr + "§§§§" + taskTime + "§§§§" + selfIp
	err := redisDriver.Publish(redisChannel, infoStr).Err()
	if err != nil {
		fmt.Println("publish failed. err=", err)
		//continue
	}
	log.Println(infoStr)
	return err
}
