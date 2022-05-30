package main

import (
	"encoding/json"
	"errors"
	"os"
)

const (
	// 1025~49151
	// https://docs.microsoft.com/zh-cn/troubleshoot/windows-server/networking/default-dynamic-port-range-tcpip-chang
	defaultListenPort = 8655
)

type Config struct {
	ListenPort      int
	LastPatchedPath string
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

func LoadConfig(file string) (*Config, error) {
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return ConfigFromJson(data)
}

func (c *Config) Save(file string) error {
	data, err := c.ToJson()
	if err != nil {
		return err
	}
	return os.WriteFile(file, data, 0644)
}
