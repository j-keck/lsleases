package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type DHCPLease struct {
	Created     time.Time
	Expire      time.Time
	MissedPings int
	IP          string
	Mac         string
	Name        string
}

func (l *DHCPLease) String() string {
	return fmt.Sprintf("%-15s %-17s %s", l.IP, l.Mac, l.Name)
}

type DHCPLeases struct {
	Leases []*DHCPLease
}

type SortedByCreated DHCPLeases

func (ls SortedByCreated) Len() int {
	return len(ls.Leases)
}
func (ls SortedByCreated) Swap(i, j int) {
	ls.Leases[i], ls.Leases[j] = ls.Leases[j], ls.Leases[i]
}
func (ls SortedByCreated) Less(i, j int) bool {
	return ls.Leases[i].Created.Before(ls.Leases[j].Created)
}

func (ls *DHCPLeases) Clear() {
	ls.Leases = ls.Leases[:0]
}

func (ls *DHCPLeases) Delete(l *DHCPLease) {
	if i, ok := ls.IndexOfMac(l.Mac); ok {
		ls.Leases = append(ls.Leases[:i], ls.Leases[i+1:]...)
	} else {
		log.Printf("unable to delte lease: '%s' - not found\n", l.String())
	}
}

func (ls *DHCPLeases) Foreach(f func(*DHCPLease)) {
	for _, l := range ls.Leases {
		f(l)
	}
}

func (ls *DHCPLeases) UpdateOrAdd(l *DHCPLease) {
	if i, ok := ls.IndexOfMac(l.Mac); ok {
		ls.Leases[i] = l
	} else {
		ls.Leases = append(ls.Leases, l)
	}
}

func (ls *DHCPLeases) IndexOf(f func(*DHCPLease) bool) (int, bool) {
	for i, l := range ls.Leases {
		if f(l) {
			return i, true
		}
	}
	return -1, false
}
func (ls *DHCPLeases) IndexOfMac(mac string) (int, bool) {
	return ls.IndexOf(func(l *DHCPLease) bool {
		return l.Mac == mac
	})
}

func (ls *DHCPLeases) SaveLeases() error {
	j, err := json.Marshal(ls)
	if err != nil {
		return err
	}

	path := appDataPath + "/lsleases.json"

	verboseLog.Printf("save leases under %s\n", path)
	return ioutil.WriteFile(path, []byte(j), 0644)
}

func LoadLeases() (DHCPLeases, error) {
	var leases DHCPLeases

	path := appDataPath + "/lsleases.json"

	verboseLog.Printf("load saved leases from %s\n", path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return leases, fmt.Errorf("no persistence file found under %s\n", path)
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return leases, err
	}

	err = json.Unmarshal(b, &leases)
	return leases, err
}
