package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"
)

func client() {

	if *printVersionFlag {
		clientVersion := VERSION
		serverVersionB, _ := askServer("version")
		serverVersion := string(serverVersionB)
		fmt.Printf("client: %s, server: %s\n", clientVersion, serverVersion)
	} else if *clearLeasesFlag {
		tellServer("clearLeases")
	} else if *shutdownServerFlag {
		tellServer("shutdown")
	} else {
		var leases DHCPLeases
		b, err := askServer("leases")
		exitOnError(err, "query leases")
		if err := json.Unmarshal(b, &leases); err != nil {
			fmt.Println("unmarshall error - ", err)
		} else {

			// sort leases
			if !*listNewestLeasesFirstFlag {
				sort.Sort(SortedByCreated(leases))
			} else {
				sort.Sort(sort.Reverse(SortedByCreated(leases)))
			}

			listLeases(leases)
		}
	}
}

func listLeases(leases DHCPLeases) {
	format := "%-15s  %-17s  %s"
	formatVerboseActive := "%-11s  %-9s  %-15s  %-17s  %s"
	formatVerbosePassive := "%-11s  %-11s  %-15s  %-17s  %s"

	serverModeB, err := askServer("mode")
	exitOnError(err, "query mode")
	serverMode := string(serverModeB)

	// format a DHCPLease for output
	leaseFormatter := func(l *DHCPLease) string {
		dateFormatter := func(t time.Time) string {
			if *scriptedModeFlag {
				return strconv.FormatInt(t.Unix(), 10)
			}
			return t.Format("02.01 15:04")
		}

		createdStr := dateFormatter(l.Created)
		if *verboseFlag && serverMode == "active" {
			return fmt.Sprintf(formatVerboseActive,
				createdStr, strconv.Itoa(l.MissedPings), l.IP, l.Mac, l.Name)
		}

		if *verboseFlag && serverMode == "passive" {
			expireStr := dateFormatter(l.Expire)
			return fmt.Sprintf(formatVerbosePassive,
				createdStr, expireStr, l.IP, l.Mac, l.Name)
		}

		return fmt.Sprintf(format, l.IP, l.Mac, l.Name)
	}

	// print header if is in 'scriptedMode'
	if !*scriptedModeFlag {
		if *verboseFlag && serverMode == "active" {
			fmt.Printf(formatVerboseActive, "Created", "Ping miss", "Ip", "Mac", "Name\n")
		} else if *verboseFlag && serverMode == "passive" {
			fmt.Printf(formatVerbosePassive, "Created", "Expire", "Ip", "Mac", "Name\n")
		} else {
			fmt.Printf(format, "Ip", "Mac", "Name\n")
		}
	}

	leases.Foreach(func(l *DHCPLease) {
		fmt.Println(leaseFormatter(l))
	})
}
