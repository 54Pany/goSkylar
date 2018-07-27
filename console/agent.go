package main

import (
	"github.com/olekukonko/tablewriter"
	"os"
	"fmt"
	"strconv"
	"github.com/c-bata/go-prompt"
)

// agent控制模块
func AgentConsole() {
	for {
		in := prompt.Input("agent> ", AgentCompleter)
		args := ParseCommand(in)
		switch args[0] {
		case "show":
			ShowAgent()
		case "command:":
			if len(args) >= 3 {
				CommandAgent(args[1], args[2])
			} else {
				fmt.Println("缺少参数.\n e.g: command: 192.168.0.1 shutdown")
			}
		case "remove":
			if len(args) >= 2 {
				RemoveAgent(args[1])
			} else {
				fmt.Println("缺少参数.\n e.g: remove 192.168.0.1")
			}
		case "back":
			return
		case "exit":
			os.Exit(0)
		default:
			fmt.Println("无效指令", args[0])
		}
	}
}

// 列出所有agent主机
func ShowAgent() {
	conn := RedisPool.Get()
	reply, err := conn.Do("SMEMBERS", fmt.Sprintf("agent:ip"))
	if err != nil {
		fmt.Println("Server SMEMBERS Error:", err.Error())
	}
	if reply != nil {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Host", "Total"})
		agentList := reply.([]interface{})
		for _, v := range agentList {
			agentIp := string(v.([]byte))
			table.Append([]string{agentIp})
		}
		table.SetFooter([]string{"", strconv.Itoa(len(agentList))}) // Add Footer
		table.Render()                                              // Send output
	}
	conn.Close()
}

// 移除掉agent主机
func RemoveAgent(ip string) {
	conn := RedisPool.Get()
	_, err := conn.Do("SREM", fmt.Sprintf("agent:ip"), ip)
	if err != nil {
		fmt.Println("移除失败:", err.Error())
	} else {
		fmt.Println("移除", ip, "agent主机成功")
	}
	conn.Close()
}

// 给指定agent主机派发指定指令
func CommandAgent(ip string, command string) {
	var agents []string
	conn := RedisPool.Get()
	// 关闭所有主机
	if ip == "all" {
		reply, err := conn.Do("SMEMBERS", fmt.Sprintf("agent:ip"))
		if err != nil {
			fmt.Println("Server SMEMBERS Error:", err.Error())
		}
		if reply != nil {
			temp := reply.([]interface{})
			for _, v := range temp {
				agent := string(v.([]byte))
				agents = append(agents, agent)
			}
		}
	} else {
		agents = append(agents, ip)
	}

	for _, agent := range agents {
		_, err := conn.Do("HMSET", fmt.Sprintf("agent:command"), agent, command)
		if err != nil {
			fmt.Println("派发指令失败:", err.Error())
		} else {
			fmt.Println("往", agent, "派发", command, "指令成功")
		}
	}
	conn.Close()
}
