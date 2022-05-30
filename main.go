package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/me1ting/exporter-of-exile-cn-backend/log"
	"github.com/me1ting/exporter-of-exile-cn-backend/ui"
)

const (
	logFileName    = "log.txt"
	configFileName = "config.json"
)

func main() {
	var err error

	ex, err := os.Executable()
	if err != nil {
		ui.ShowError(err, nil)
		return
	}
	exPath := filepath.Dir(ex)

	err = log.InitGlobalLogger(path.Join(exPath, logFileName))
	if err != nil {
		ui.ShowError(err, nil)
		return
	}

	var config *Config

	configPath := path.Join(exPath, configFileName)
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		config = NewConfig()
		config.Save(configPath)
	} else {
		if config, err = LoadConfig(configPath); err != nil {
			ui.ShowError(err, nil)
			return
		}
	}

	app, err := ui.NewAppMainWindow()
	if err != nil {
		ui.ShowError(err, nil)
		return
	}

	listen := fmt.Sprintf("localhost:%v", config.ListenPort)
	startupFailed := make(chan struct{})

	go func() {
		select {
		case <-time.After(500 * time.Millisecond):
			log.Printf("Listening on %v", listen)
		case <-startupFailed:
			return
		}
	}()

	app.Run()

	err = http.ListenAndServe(listen, NewServer(config))

	if err != nil {
		startupFailed <- struct{}{}
		fmt.Printf("启动失败，请检查端口是否被占用: %v\n", err)
		pause()
	}
}

func pause() {
	fmt.Print("\n")
	fmt.Print("关闭窗口或按回车继续...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	os.Exit(-1)
}
