package main

import (
	"encoding/json"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Basic struct {
	Version string `json:"version"`
	User    string `json:"user"`
	Team    int    `json:"team"`
	Passkey string `json:"passkey"`
	Cause   string `json:"cause"`
	Power   string `json:"power"`
	Paused  bool   `json:"paused"`
	Idle    bool   `json:"idle"`
}

type Slot struct {
	ID             string   `json:"id"`
	Status         string   `json:"status"`
	Description    string   `json:"description"`
	Options        struct{} `json:"options"`
	Reason         string   `json:"reason"`
	Idle           bool     `json:"idle"`
	UnitID         int      `json:"unit_id"`
	Project        int      `json:"project"`
	Run            int      `json:"run"`
	Clone          int      `json:"clone"`
	Gen            int      `json:"gen"`
	PercentDone    string   `json:"percentdone"`
	ETA            string   `json:"eta"`
	PPD            string   `json:"ppd"`
	CreditEstimate string   `json:"creditestimate"`
	WaitingOn      string   `json:"waitingon"`
	NextAttempt    string   `json:"nextattempt"`
	TimeRemaining  string   `json:"timeremaining"`
}

type SlotInfo struct {
	BasicInfo Basic  `json:"basicInfo"`
	SlotsInfo []Slot `json:"slotsInfo"`
	Token     string `json:"token"`
	IpAddress string `json:"ipAddress"`
}

type Config struct {
	Targets []string `yaml:"targets"`
}

type Status struct {
	mu    sync.RWMutex
	value SlotInfo // MyStruct is the struct type returned by getStatus()
}

func (s *Status) Set(value SlotInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.value = value
}

func (s *Status) Get() SlotInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.value
}

type StatusCache struct {
	mu     sync.RWMutex
	status map[string]*Status // map of target ID to Status
}

func (sc *StatusCache) Set(id string, value SlotInfo) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if _, ok := sc.status[id]; !ok {
		sc.status[id] = &Status{}
	}

	sc.status[id].Set(value)
}

func (sc *StatusCache) Get(id string) SlotInfo {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if status, ok := sc.status[id]; ok {
		return status.Get()
	}

	return SlotInfo{} // return zero value if there is no status for this ID
}

func sendRequest(ipAddress string, method string, url string) (error, string) {
	// Create a new HTTP request
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return err, ""
	}

	// Set the headers
	req.Header.Set("sec-ch-ua", `"(Not(A:Brand";v="8", "Chromium";v="100"`)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.75 Safari/537.36`)
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("Origin", "http://"+ipAddress+":7396")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Referer", "http://"+ipAddress+":7396/?nocache=0.4642702738328919")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "close")

	// Send the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending HTTP request:", err)
		return err, ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return err, ""
	}

	return nil, string(body)
}

func startSession(ipAddress string) (error, string) {
	err, data := sendRequest(ipAddress, http.MethodPut, "http://"+ipAddress+":7396/api/session")
	return err, data
}

func fahConfigured(ipAddress string, token string) {
	err, _ := sendRequest(ipAddress, http.MethodGet, "http://"+ipAddress+":7396/api/configured?sid="+token+"&_=1685187908670")
	if err != nil {
		return
	}
}

func fahSetApi(ipAddress string, token string) {
	err, _ := sendRequest(ipAddress, http.MethodPut, "http://"+ipAddress+":7396/api/updates/set?sid="+token+"&update_id=1&update_rate=1&update_path=%2Fapi%2Fslots&_=1685187908669")
	if err != nil {
		return
	}
	err, _ = sendRequest(ipAddress, http.MethodPut, "http://"+ipAddress+":7396/api/updates/set?sid="+token+"&update_id=0&update_rate=1&update_path=%2Fapi%2Fbasic&_=1685187908670")
	if err != nil {
		return
	}
}

func getStatus(ipAddress string, token string) (error, []Slot) {
	err, body := sendRequest(ipAddress, http.MethodGet, "http://"+ipAddress+":7396/api/slots?sid="+token+"&_=1685187908670")
	if err != nil {
		return err, nil
	}

	var responses []Slot
	err = json.Unmarshal([]byte(body), &responses)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return err, nil
	}

	return nil, responses
}

func getBasic(ipAddress string, token string) (error, Basic) {
	var responses Basic
	err, body := sendRequest(ipAddress, http.MethodGet, "http://"+ipAddress+":7396/api/basic?sid="+token+"&_=1685187908670")
	if err != nil {
		return err, responses
	}

	err = json.Unmarshal([]byte(body), &responses)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return err, responses
	}

	return nil, responses
}

func printProgressBar(percent float64, data Slot, basicInfo Basic, ipAddress string) {
	totalBars := 50
	doneBars := int(percent / 2)
	fmt.Printf("[")
	for i := 0; i < totalBars; i++ {
		if i < doneBars {
			fmt.Print("=")
		} else {
			fmt.Print(" ")
		}
	}
	fmt.Print("] ")
	fmt.Printf(
		"%.2f%% rig: %15s user: %s team: %d project: %d estimate: %8s ppd: %9s eta: %s",
		percent,
		ipAddress,
		basicInfo.User,
		basicInfo.Team,
		data.Project,
		data.CreditEstimate,
		data.PPD,
		data.ETA)
}

func taskPercentage(data Slot) (error, float64) {
	percentStr := strings.TrimRight(data.PercentDone, "%")
	percent, err := strconv.ParseFloat(percentStr, 64)
	if err != nil {
		fmt.Println("Error parsing percent:", err)
		return err, 0
	}
	return nil, percent
}

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
				time.Sleep(time.Second * 5) // wait 5 seconds before updating the status again
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
