package main

import (
  "io"
  "fmt"
	"strconv"
	"strings"
  "net/http"
	"encoding/json"
)

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
	req.Header.Set("Origin", "http://"+ipAddress+":7396")
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
		"%.2f%% rig: %15s user: %s team: %d project: %d estimate: %8s ppd: %9s eta: %20s description: %s",
		percent,
		ipAddress,
		basicInfo.User,
		basicInfo.Team,
		data.Project,
		data.CreditEstimate,
		data.PPD,
		data.ETA,
		data.Description)
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
