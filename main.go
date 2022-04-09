package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {

	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("加载配置文件失败： %v,\n", err)
		pause()
	}

	startErr := make(chan struct{})

	go func() {
		select {
		case <-time.After(500 * time.Millisecond):
			log.Printf("listening: %v", fmt.Sprintf(":%v", config.ListenPort))
		case <-startErr:
			return
		}
	}()

	err = http.ListenAndServe(fmt.Sprintf(":%v", config.ListenPort), NewServer(config))

	if err != nil {
		startErr <- struct{}{}
		fmt.Printf("启动程序失败，请检查端口是否被占用: %v\n", err)
		pause()
	}
}

func pause() {
	fmt.Print("\n")
	fmt.Print("关闭窗口或按回车继续...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
