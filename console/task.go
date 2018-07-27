package main

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"os"
)

// task控制模块
func TaskConsole() {
	for {
		in := prompt.Input("agent> ", TaskCompleter)
		args := ParseCommand(in)
		switch args[0] {
		case "count":
			CountTask()
		case "back":
			return
		case "exit":
			os.Exit(0)
		default:
			fmt.Println("无效指令", args[0])
		}
	}
}

// 统计任务数量
func CountTask() {
	conn := RedisPool.Get()

	queues := []string{"masscan", "nmap"}

	for _, queue := range queues {
		//n 剩余任务
		reply, err := conn.Do("LLEN", "goskylar:queue:"+queue)
		if err != nil {
			fmt.Println(err)
			return
		}
		if reply == nil {
			fmt.Println(err)
			return
		}
		fmt.Println(queue, "队列剩余任务数量：", reply)
	}

	conn.Close()
}
