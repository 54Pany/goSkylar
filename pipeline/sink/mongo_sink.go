package sink

import (
	"goSkylar/pipeline/channel"
	"time"
	"log"
	"goSkylar/pipeline/data"
)

type MongoSink struct {
}

func (this *MongoSink) Run(q *queue.EsQueue) {

	//ProducerLoop:

	for {
		if q.Quantity() < 1 {
			time.Sleep(time.Second * 1)
			continue
		}
		msg, ok, _ := q.Get()
		if !ok {
			log.Println("Get.Fail")
		}
		if msg == nil {
			continue
		}

		err := data.DataTransfer(msg.(string))
		if err != nil {
			log.Println("info transfer error")
		}

	}
}
