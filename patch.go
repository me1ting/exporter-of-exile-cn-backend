package main

import (
	"bytes"
	"errors"
	"os"
	"path"
	"regexp"
	"strings"
)

const (
	importTabName                     = "ImportTab.lua"
	relativePathOfImportTabToPOECharm = "PathOfBuildingCommunity\\Classes\\ImportTab.lua"
	relativePathOfImportTabToPOB      = "Classes\\ImportTab.lua"
	hostNameOfTencent                 = "https://poe.game.qq.com/"
	hostNameOfLocalPattern            = `http://localhost:\d{1,5}/`
)

func doPatch(file string, hostName string) error {
	dat, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	replaced := false
	if bytes.Contains(dat, []byte(hostNameOfTencent)) {
		dat = bytes.Replace(dat, []byte(hostNameOfTencent), []byte(hostName), 1)
		replaced = true
	} else {
		matched, err := regexp.Match(hostNameOfLocalPattern, dat)
		if err != nil {
			return err
		}
		if matched {
			re := regexp.MustCompile(hostNameOfLocalPattern)
			dat = re.ReplaceAll(dat, []byte(hostName))
			replaced = true
		}
	}

	if !replaced {
		return errors.New("file contains no host name")
	}

	return os.WriteFile(file, dat, 0777)
}

func Patch(filePath string, hostName string) error {
	importTabPath := path.Join(filePath, relativePathOfImportTabToPOECharm)
	if strings.HasSuffix(filePath, importTabName) {
		if _, err := os.Stat(importTabPath); err != nil {
			return errors.New("ImportTab.lua not found")
		}
	} else {
		if _, err := os.Stat(importTabPath); err != nil {
			importTabPath = path.Join(filePath, relativePathOfImportTabToPOB)
			if _, err := os.Stat(importTabPath); err != nil {
				return errors.New("ImportTab.lua not found")
			}
		}
	}

	return doPatch(importTabPath, hostName)
}
