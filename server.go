// dhcp leases sniffer
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
	// create application data directory if not existent
	//
	exitOnError(createAppData(), "unable to create application data directory", appDataPath)

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

	cleanupLeaseTimer, err := parseDuration(*cleanupLeaseTimerFlag)
	exitOnError(err)

	if !*expireBasedFlag {
		log.Printf("enable active check - ping every: %s\n", *cleanupLeaseTimerFlag)
		if hasPermission, err := hasPermissionForAliveCheck(); hasPermission {
			clearOfflineHostsTickerChan = time.NewTicker(cleanupLeaseTimer).C
			// clear on startup
			clearOfflineHosts()
		} else {
			log.Printf("enable active check failed: '%s' -  fallback to passive mode\n", err)
			*expireBasedFlag = true
		}
	}

	var leaseExpiredDuration time.Duration
	if *expireBasedFlag {
		leaseExpiredDuration, err = parseDuration(*leaseExpiredDurationFlag)
		exitOnError(err)

		timer := cleanupLeaseTimer
		if leaseExpiredDuration < timer {
			timer = leaseExpiredDuration / 2
		}
		log.Printf("enable expired based (passive mode) - check timer: %s, expire duation: %s\n", timer, leaseExpiredDuration)
		clearExpiredLeasesTickerChan = time.NewTicker(timer).C
		// clear on startup
		clearExpiredLeases()
	}

	//
	// if persistent leases are enable, load saved leases
	//
	if *keepLeasesOverRestartFlag {
		var err error
		leases, err = LoadLeases()
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
		case cmdBytes := <-clientChan:
			cmd := string(cmdBytes)
			switch {
			case cmd == "shutdown":
				shutdown()

			case cmd == "clearLeases":
				log.Println("clear leases")
				leases.Clear()
				clientChan <- []byte("done")

			case cmd == "leases":
				j, err := json.Marshal(leases)
				panicOnError(err)
				clientChan <- j

			case strings.HasPrefix(cmd, "leases-since:"):
				fields := strings.Split(cmd, ":")
				if ts, err := strconv.ParseInt(fields[1], 10, 64); err != nil {
					log.Println("invalid ts in cmd 'leases-since", err)
					clientChan <- []byte("invalid ts in cmd 'leases-since")
				} else {
					ls := leases.Filter(func(l *DHCPLease) bool {
						return l.Created.UnixNano() >= ts
					})
					j, err := json.Marshal(ls)
					panicOnError(err)
					clientChan <- j
				}

			case cmd == "version":
				clientChan <- []byte(VERSION)

			case cmd == "mode":
				if *expireBasedFlag {
					clientChan <- []byte("passive")
				} else {
					clientChan <- []byte("active")
				}

			default:
				log.Println("invalid cmd: ", cmd)
				clientChan <- []byte("invalid cmdk")
			}

		case l := <-dhcpLeaseChan:
			// update expire entry
			if *expireBasedFlag {
				l.Expire = time.Now().Add(leaseExpiredDuration)
			}

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
		logOnError(leases.SaveLeases(), "unable to save leases")
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

func createAppData() error {
	if _, err := os.Stat(appDataPath); os.IsNotExist(err) {
		err = os.MkdirAll(appDataPath, 0644)
		if err != nil {
			return err
		}

		// set permission
		return os.Chmod(appDataPath, 0755)
	}
	return nil
}

func pingHosts() {
	leases.Foreach(func(l *DHCPLease) {
		if hostIsAlive, err := isAlive(l.IP); err != nil {
			log.Printf("unable to execute ping: '%s'\n", err.Error())
			// noop
		} else if hostIsAlive {
			verboseLog.Printf("%s is online", l.String())
			l.MissedPings = 0
		} else {
			l.MissedPings++
			verboseLog.Printf("%s is offline\n", l.String())
		}
	})
}
