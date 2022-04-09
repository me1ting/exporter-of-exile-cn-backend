package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	// 1025~49151
	// https://docs.microsoft.com/zh-cn/troubleshoot/windows-server/networking/default-dynamic-port-range-tcpip-chang
	defaultListenPort = "8655"

	configFileName = "config.json"
)

type Config struct {
	ListenPort string `json:"listen_port"`
}

func NewConfig() *Config {
	return &Config{
		ListenPort: defaultListenPort,
	}
}

func (c *Config) ToJson() ([]byte, error) {
	b, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return nil, err
	}
	return b, nil
}

func ConfigFromJson(b []byte) (*Config, error) {
	c := NewConfig()
	err := json.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func LoadConfig() (*Config, error) {
	var c *Config
	ex, err := os.Executable()
	if err != nil {
		return nil, err
	}
	exPath := filepath.Dir(ex)
	configPath := fmt.Sprintf("%v\\%v", exPath, configFileName)

	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		c = NewConfig()
		err := SaveConifg(c)
		if err != nil {
			log.Printf("error: %v", err)
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	return ConfigFromJson(data)
}

func SaveConifg(c *Config) error {
	ex, err := os.Executable()
	if err != nil {
		return err
	}
	exPath := filepath.Dir(ex)
	configPath := fmt.Sprintf("%v\\%v", exPath, configFileName)

	data, err := c.ToJson()
	if err != nil {
		return err
	}
	err = os.WriteFile(configPath, data, 0777)
	return err
}
