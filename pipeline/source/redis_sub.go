package source

import (
	"goSkylar/pipeline/monitor"
	"log"
	"goSkylar/pipeline/channel"
	"gopkg.in/redis.v5"
)

var (
	redis_host = "116.196.96.123:23177"
	redis_pass = "a1x06awvaBpD"
	channels   = "portinfo"
	redis_db   = 5
)

type RedisSub struct {
}

func (this *RedisSub) Run(q1 *queue.EsQueue) {
	log.Println("run redis sub source")
	flag := 0
	pubsub, err := this.ConnectRedis()
	if err != nil {
		log.Println(err)
	}
	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			log.Println(err)

			for {
				pubsub, err = this.ConnectRedis()
				if err == nil {
					log.Println("----Redis重连成功----")
					break
				}

				flag++
				if flag > 5 {
					log.Println("----重连五次失败-------")
					monitor.Sendmail(err)
					break
				}
			}
		} else {
			log.Println(msg.Payload)
			q1.Put(msg.Payload)
		}
	}

}
func (this *RedisSub) ConnectRedis() (*redis.PubSub, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     redis_host,
		Password: redis_pass, // no password set
		DB:       redis_db,   // use default DB
	})
	pubsub, e := client.Subscribe(channels)
	if e != nil {
		log.Println(e)
	}
	return pubsub, e
}
