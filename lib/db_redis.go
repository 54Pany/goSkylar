package lib

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

var (
	cfg             = NewConfigUtil("")
	redisHost, _    = cfg.GetString("redis_default", "host")
	redisPort, _    = cfg.GetString("redis_default", "port")
	redisPass, _    = cfg.GetString("redis_default", "pass")
	redisDbStr, _   = cfg.GetString("redis_default", "db")
	redisDb, _      = strconv.Atoi(redisDbStr)
	redisChannel, _ = cfg.GetString("redis_default", "channel")

	RedisDriver = redis.NewClient(&redis.Options{
		Addr:        redisHost + ":" + redisPort,
		Password:    redisPass,
		DB:          redisDb,
		DialTimeout: time.Second * 2,
		//IdleTimeout: time.Second * 1000000,
	})

	DsnAddr = fmt.Sprintf("redis://root:%s@%s:%s/%s", redisPass, redisHost, redisPort, redisDbStr)
)

func PushPortInfoToRedis(infoStr string, taskTime string, selfIp string) error {
	var redisDriver = RedisDriver
	infoStr = infoStr + "§§§§" + taskTime + "§§§§" + selfIp
	err := redisDriver.Publish(redisChannel, infoStr).Err()
	if err != nil {
		log.Println("publish failed. err=", err)
		//continue
	}
	log.Println(infoStr)
	return err
}
