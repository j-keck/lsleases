package main

import (
	"encoding/json"
	"fmt"
	"github.com/j-keck/arping"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"time"
)

var leases DHCPLeases

func server() {
	log.Println("startup -  version: ", VERSION)

	if *verboseFlag {
		verboseLog = log.New(os.Stdout, "", log.LstdFlags)
	} else {
		verboseLog = log.New(ioutil.Discard, "", 0)
	}

	//
	// init listeners
	//
	dhcpLeaseChan := make(chan *DHCPLease)
	go dhcpListener(dhcpLeaseChan)

	clientChan := make(chan []byte)
	go clientListener(clientChan)

	//
	// init cleanup old leases routine
	//
	var clearExpiredLeasesTickerChan <-chan time.Time
	var clearOfflineHostsTickerChan <-chan time.Time

	if !*expireBasedFlag {
		log.Printf("enable active check - arping every: %s\n", *cleanupLeaseTimerFlag)
		if hasPermission, err := hasRawSocketPermission(); hasPermission {
			clearOfflineHostsTickerChan = time.NewTicker(cleanupLeaseTimer).C
			// clear one time on startup
			clearOfflineHosts()
		} else {
			log.Printf("enable active check failed: '%s' -  fallback to passive mode\n", err)
			*expireBasedFlag = true
		}
	}

	if *expireBasedFlag {
		timer := cleanupLeaseTimer
		if leaseExpiredDuration < timer {
			timer = leaseExpiredDuration / 2
		}
		log.Printf("enable expired based - check timer: %s, expire duation: %s\n", timer, leaseExpiredDuration)
		clearExpiredLeasesTickerChan = time.NewTicker(timer).C
		// clear on time on startup
		clearExpiredLeases()
	}

	//
	// if persistent leases are enable, load saved leases
	//
	if *keepLeasesOverRestartFlag {
		var err error
		leases, err = loadLeases()
		logOnError(err, "unable to load leases - start with emtpy leases")
	}

	//
	// init CTRL-C - catcher
	//
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		// block
		<-sigchan
		shutdown()
	}()

	//
	// main loop
	//
	for {
		select {
		case cmd := <-clientChan:
			switch string(cmd) {
			case "shutdown":
				shutdown()

			case "clearLeases":
				log.Println("clear leases")
				leases.Clear()
				clientChan <- []byte("done")

			case "leases":
				j, err := json.Marshal(leases)
				panicOnError(err)
				clientChan <- j

			case "version":
				clientChan <- []byte(VERSION)

			case "mode":
				if *expireBasedFlag {
					clientChan <- []byte("passive")
				} else {
					clientChan <- []byte("active")
				}
			}

		case l := <-dhcpLeaseChan:
			log.Printf("new DHCP Lease: '%s'\n", l.String())
			leases.UpdateOrAdd(l)

		case <-clearExpiredLeasesTickerChan:
			clearExpiredLeases()

		case <-clearOfflineHostsTickerChan:
			clearOfflineHosts()
		}
	}
}

func shutdown() {
	if *keepLeasesOverRestartFlag {
		log.Println("save leases")
		logOnError(saveLeases(), "unable to save leases")
	}

	log.Println("shutdown")
	closeListener()
	os.Exit(0)
}

func clearExpiredLeases() {
	verboseLog.Println("check expired leases")
	leases.Foreach(func(l *DHCPLease) {
		if time.Now().After(l.Expire) {
			log.Printf("expired: '%s'\n", l.String())
			leases.Delete(l)
		}
	})
}
func clearOfflineHosts() {
	verboseLog.Println("check offline hosts per ping")
	pingHosts()

	leases.Foreach(func(l *DHCPLease) {
		if l.MissedPings > *missedPingsThresholdFlag {
			log.Printf("remove lease: '%s'\n", l.String())
			leases.Delete(l)
		}
	})
}

func pingHosts() {
	leases.Foreach(func(l *DHCPLease) {
		if _, _, err := arping.Ping(net.ParseIP(l.IP)); err == arping.ErrTimeout {
			l.MissedPings++
			verboseLog.Printf("%s is offline\n", l.String())
		} else if err != nil {
			log.Printf("unable to execute ping: '%s'\n", err.Error())
		} else {
			verboseLog.Printf("%s is online", l.String())
			l.MissedPings = 0
		}
	})
}
func hasRawSocketPermission() (bool, error) {
	var localIP net.IP

	// find any local ip
	addrs, err := net.InterfaceAddrs()
	panicOnError(err)

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok {
			if ipNet.IP.IsLoopback() {
				continue
			}

			localIP = ipNet.IP
			break
		}
	}

	// arping any local ip results always in timeout
	verboseLog.Printf("arping '%s' to check permission\n", localIP.String())
	if _, _, err = arping.Ping(localIP); err == arping.ErrTimeout {
		return true, nil
	}

	return false, err
}

func saveLeases() error {
	j, err := json.Marshal(leases)
	if err != nil {
		return err
	}

	path := leasesPersistenceFilePath()

	verboseLog.Printf("save leases under %s\n", path)
	return ioutil.WriteFile(path, []byte(j), 0644)
}

func loadLeases() (DHCPLeases, error) {
	var leases DHCPLeases

	path := leasesPersistenceFilePath()

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

func leasesPersistenceFilePath() string {

	var basePath string
	if runtime.GOOS == "windows" {
		basePath = os.Getenv("APPDATA")
	} else {
		basePath = "/var/lib/lsleases"
	}
	return fmt.Sprintf("%s/lsleases.json", basePath)
}
