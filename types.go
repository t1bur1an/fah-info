package main

import (
  "sync"
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
