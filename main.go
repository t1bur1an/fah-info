package main

import (
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"os"
	"time"
)

func main() {
	statusCache := &StatusCache{
		status: make(map[string]*Status),
	}

	data, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	var config Config

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}

	for _, target := range config.Targets {
		go func(target string) {
			var status SlotInfo
			status.IpAddress = target
			err, sessionToken := startSession(status.IpAddress)
			if err != nil {
				panic(err)
			}

			status.Token = sessionToken
			fahConfigured(status.IpAddress, status.Token)
			fahSetApi(status.IpAddress, status.Token)
			err, basicInfo := getBasic(status.IpAddress, status.Token)
			if err != nil {
				panic(err)
			}

			status.BasicInfo = basicInfo
			for {
				err, statusData := getStatus(status.IpAddress, status.Token)
				if err != nil {
				} else {
					status.SlotsInfo = statusData
					statusCache.Set(target, status)
				}
				time.Sleep(time.Second * 5)
			}
		}(target)
	}

	for {
		fmt.Print("\033[H\033[2J")
		for _, target := range config.Targets {
			statusData := statusCache.Get(target)
			for _, data := range statusData.SlotsInfo {
				if data.Status != "DISABLED" {
					err, percentage := taskPercentage(data)
					if err != nil {
						panic(err)
					}
					printProgressBar(percentage, data, statusData.BasicInfo, statusData.IpAddress)
					fmt.Println()
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}
