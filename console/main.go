package main

import (
	"github.com/c-bata/go-prompt"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"net/url"
	"strings"
	"goSkylar/lib/redispool"
	"time"
	"goSkylar/server/conf"
)

var (
	RedisPool *redis.Pool
)

func init() {
	u, err := url.Parse(conf.REDIS_URI)
	if err != nil {
		panic(err)
	}

	redisAddr := u.Host
	redisPass, ok := u.User.Password()
	if !ok {
		redisPass = ""
	}
	redisDB := strings.Trim(u.Path, "/")
	RedisPool = redispool.NewRedisPool(redispool.Options{
		RedisAddr:        redisAddr,         //redis链接地址
		RedisPass:        redisPass,         //redis认证密码
		RedisDB:          redisDB,           //redis数据库
		RedisMaxActive:   1,                 // 最大的激活连接数，表示同时最多有N个连接
		RedisMaxIdle:     1,                 //最大的空闲连接数，表示即使没有redis连接时依然可以保持N个空闲的连接，而不被清除，随时处于待命状态
		RedisIdleTimeout: 180 * time.Second, // 最大的空闲连接等待时间，超过此时间后，空闲连接将被关闭
	})
}

func main() {
	for {
		in := prompt.Input("goSkylar> ", Completer)
		switch in {
		case "agent":
			AgentConsole()
		case "task":
			TaskConsole()
		case "exit":
			return
		default:
			fmt.Println("无效命令", in)
		}
	}
}
