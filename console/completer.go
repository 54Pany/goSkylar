package main

import "github.com/c-bata/go-prompt"

func Completer(in prompt.Document) []prompt.Suggest {
	suggest := []prompt.Suggest{
		{Text: "agent", Description: "Agent主机管理"},
		{Text: "task", Description: "任务管理"},
		{Text: "exit", Description: "退出"},
	}
	return prompt.FilterHasPrefix(suggest, in.GetWordBeforeCursor(), true)
}

func AgentCompleter(in prompt.Document) []prompt.Suggest {
	suggest := []prompt.Suggest{
		{Text: "show", Description: "列出所有agent主机"},
		{Text: "remove", Description: "移除掉指定agent主机,例如: remove 192.168.0.1"},
		{Text: "command:", Description: "往agent主机发送指令,例如: command: 192.168.0.1 shutdown"},
		{Text: "back", Description: "返回"},
		{Text: "exit", Description: "退出"},
	}
	return prompt.FilterHasPrefix(suggest, in.GetWordBeforeCursor(), true)
}

func TaskCompleter(in prompt.Document) []prompt.Suggest {
	suggest := []prompt.Suggest{
		{Text: "count", Description: "统计剩余任务数量"},
		{Text: "back", Description: "返回"},
		{Text: "exit", Description: "退出"},
	}
	return prompt.FilterHasPrefix(suggest, in.GetWordBeforeCursor(), true)
}
