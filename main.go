package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

const (
	logFileName    = "log.txt"
	configFileName = "config.json"
)

func main() {
	var err error

	ex, err := os.Executable()
	if err != nil {
		fmt.Errorf("can't read exec path\n")
		return
	}

	exPath := filepath.Dir(ex)

	err = InitGlobalLogger(path.Join(exPath, logFileName))
	if err != nil {
		fmt.Errorf("init logger failed\n")
	}

	var config *Config

	configPath := path.Join(exPath, configFileName)
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		config = NewConfig()
		config.Save(configPath)
	} else {
		if config, err = LoadConfig(configPath); err != nil {
			Logger.Fatalln("load config failed")
			return
		}
	}

	listen := fmt.Sprintf("localhost:%v", config.ListenPort)
	startupFailed := make(chan struct{})

	go func() {
		select {
		case <-time.After(500 * time.Millisecond):
			Logger.Printf("Listening on %v\n", listen)
		case <-startupFailed:
			return
		}
	}()

	err = http.ListenAndServe(listen, NewServer(config))

	if err != nil {
		startupFailed <- struct{}{}
		Logger.Fatalf("启动失败，请检查端口是否被占用: %v\n", err)
	}
}
