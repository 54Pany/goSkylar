package main

import (
	"fmt"
	"os/exec"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			wg.Done()
			cmd := exec.Command("sleep", []string{"10",}...)
			err := cmd.Run()
			fmt.Println(err)
		}()
	}
	wg.Wait()
	time.Sleep(time.Second * 10000)
}
