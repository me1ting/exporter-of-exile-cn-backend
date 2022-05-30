package main

import (
	"log"
	"os"
)

var Logger *log.Logger

func InitGlobalLogger(file string) error {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	Logger = log.New(f, "", 0)
	return nil
}
