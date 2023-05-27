package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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

type Response struct {
	Path  string
	Value []Slot
}

type mainResponse interface {
}

func sendRequest(ipAddress string, url string) (error, string) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPut, url, nil)
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
	err, data := sendRequest(ipAddress, "http://"+ipAddress+":7396/api/session")
	return err, data
}

func fahConfigured(ipAddress string, token string) {
	err, _ := sendRequest(ipAddress, "http://"+ipAddress+":7396/api/configured?sid="+token+"&_=1685187908670")
	if err != nil {
		return
	}
}

func fahSetApi(ipAddress string, token string) {
	err, _ := sendRequest(ipAddress, "http://"+ipAddress+":7396/api/updates/set?sid="+token+"&update_id=1&update_rate=1&update_path=%2Fapi%2Fslots&_=1685187908669")
	if err != nil {
		return
	}
	err, _ = sendRequest(ipAddress, "http://"+ipAddress+":7396/api/updates/set?sid="+token+"&update_id=0&update_rate=1&update_path=%2Fapi%2Fbasic&_=1685187908670")
	if err != nil {
		return
	}
}

func getStatus(ipAddress string, token string) (error, []Slot) {
	err, body := sendRequest(ipAddress, "http://"+ipAddress+":7396/api/slots?sid="+token+"&_=1685187908670")
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

func printProgressBar(percent float64) {
	totalBars := 50
	doneBars := int(percent / 2)
	fmt.Printf("\r[")
	for i := 0; i < totalBars; i++ {
		if i < doneBars {
			fmt.Print("=")
		} else {
			fmt.Print(" ")
		}
	}
	fmt.Print("] ")
	fmt.Printf("%.2f%%", percent)
}

func taskPercentage(data []Slot) (error, float64) {
	for _, slot := range data {
		percentStr := strings.TrimRight(slot.PercentDone, "%")
		percent, err := strconv.ParseFloat(percentStr, 64)
		if err != nil {
			fmt.Println("Error parsing percent:", err)
			return err, 0
		}
		return nil, percent
	}

	return nil, 0
}

func main() {
	err, sessionToken := startSession("192.168.0.20")
	if err != nil {
		panic(err)
	}
	fahConfigured("192.168.0.20", sessionToken)
	fahSetApi("192.168.0.20", sessionToken)
	for {
		err, statusData := getStatus("192.168.0.20", sessionToken)
		if err != nil {
			panic(err)
		}
		err, percentage := taskPercentage(statusData)
		if err != nil {
			panic(err)
		}
		printProgressBar(percentage)

		time.Sleep(1)
	}
}
