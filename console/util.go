package main

import "strings"

// 解析命令
func ParseCommand(in string) []string {
	// TODO 待处理特殊字符
	return strings.Split(in, " ")
}
