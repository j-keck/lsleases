package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

const VERSION = "1.3.dev"

var (
	//
	// Flags
	//   flag description in flag.Usage
	//

	// common
	printHelpFlag    = flag.Bool("h", false, "")
	verboseFlag      = flag.Bool("v", false, "")
	printVersionFlag = flag.Bool("V", false, "")

	// server
	serverModeFlag  = flag.Bool("s", false, "")
	expireBasedFlag = flag.Bool("p", false, "")
	// flag.Duration not useful because there is not unit for days
	leaseExpiredDurationFlag  = flag.String("e", "7d", "")
	cleanupLeaseTimerFlag     = flag.String("t", "15m", "")
	missedPingsThresholdFlag  = flag.Int("m", 3, "")
	keepLeasesOverRestartFlag = flag.Bool("k", false, "")

	// client
	scriptedModeFlag          = flag.Bool("H", false, "")
	clearLeasesFlag           = flag.Bool("c", false, "")
	listNewestLeasesFirstFlag = flag.Bool("n", false, "")
	shutdownServerFlag        = flag.Bool("x", false, "")
)

var (
	verboseLog           *log.Logger
	leaseExpiredDuration time.Duration
	cleanupLeaseTimer    time.Duration
)

var appDataPath = osDependAppDataPath()

func main() {
	flag.Usage = func() {
		fmt.Println("Common:")
		fmt.Println("  -h: print help")
		fmt.Println("  -v: verbose output")
		fmt.Println("  -V: print version")
		fmt.Println("Client mode:")
		fmt.Println("  -c: clear leases")
		fmt.Println("  -H: scripted mode: no headers, dates as unix time")
		fmt.Println("  -n: list newest leases first")
		fmt.Println("  -x: shutdown server")
		fmt.Println("Server mode:")
		fmt.Println("  -s: server mode")
		fmt.Println("  -p: passive mode - no active availability host check - clear leases expire based")
		fmt.Println("  -e: in passive mode: lease expire duration (valid units: 'd', 'h', 'm', 's') - default:",
			*leaseExpiredDurationFlag)
		fmt.Println("  -t: interval for checking of leases validity (valid units: 'd', 'h', 'm', 's') - default:", *cleanupLeaseTimerFlag)
		fmt.Println("  -m: in active mode: missed arpings threshold - default:", *missedPingsThresholdFlag)
		fmt.Println("  -k: keep leases over restart")
	}
	flag.Parse()

	if *printHelpFlag {
		flag.Usage()
		os.Exit(0)
	}

	var err error
	leaseExpiredDuration, err = parseDuration(*leaseExpiredDurationFlag)
	exitOnError(err)

	cleanupLeaseTimer, err = parseDuration(*cleanupLeaseTimerFlag)
	exitOnError(err)

	//
	// action
	//
	if *serverModeFlag {
		server()
	} else {
		client()
	}
}

func osDependAppDataPath() string {
	//
	// set os depend application data path
	//
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE") + "/lsleases"
	} else {
		return "/var/lib/lsleases"
	}
}
