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
	//q := make(chan in, 0)
	//
	//if len(os.Args) != 2 {
	//	fmt.Println("must input topic")
	//	return
	//}

	go func() {
		t1 := time.NewTicker(time.Second * 10)
		for {
			<-t1.C
			fmt.Println(q.String())
		}
	}()

	go func() {
		obj := new(sink.MongoSink)
		obj.Run(q)
	}()

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
