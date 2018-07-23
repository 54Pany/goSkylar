package main

import (
	"goSkylar/pipeline/channel"
	"goSkylar/pipeline/sink"
	"os"
	"os/signal"
	"goSkylar/pipeline/source"
	"time"
	"fmt"
)

func main() {

	q := queue.NewQueue(1024 * 1024 * 10)

	go func() {
		t1 := time.NewTicker(time.Second * 10)
		for {
			<-t1.C
			fmt.Println(q.String())
		}
	}()


	// nmap结果插入到mongo
	go func() {
		obj := new(sink.MongoSink)
		obj.Run(q)
	}()

	//
	go func() {
		obj := new(source.RedisSub)
		obj.Run(q)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for {
		select {
		case <-c:
			os.Exit(1)
		default:

		}
	}
}
