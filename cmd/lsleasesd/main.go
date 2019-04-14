package main

import (
	"flag"
	"fmt"
	"github.com/j-keck/lsleases/pkg/config"
	"github.com/j-keck/lsleases/pkg/daemon"
	"github.com/j-keck/plog"
)

type CliConfig struct {
	logLevel     plog.LogLevel
	printVersion bool
	printHelp    bool
}

func main() {
	cliCfg, daemonCfg := parseFlags()

	if cliCfg.printHelp {
		flag.Usage()
		return
	}

	if cliCfg.printVersion {
		fmt.Println(daemon.Version())
		return
	}

	log := plog.NewConsoleLogger()
	defer log.Flush()
	log.SetLevel(cliCfg.logLevel)

	if err := daemon.Start(daemonCfg, log); err != nil {
		log.Errorf("unable to start daemon - %s", err.Error())
	}
}

func parseFlags() (CliConfig, config.Config) {

	//
	// cli config
	//
	cliCfg := CliConfig{logLevel: plog.Info}

	flag.BoolVar((*bool)(&cliCfg.printHelp), "h", false, "print help and exit")
	flag.BoolVar((*bool)(&cliCfg.printVersion), "V", false, "print version and exit")

	// log level
	plog.FlagDebugVar(&cliCfg.logLevel, "v", "debug output")
	plog.FlagTraceVar(&cliCfg.logLevel, "vv", "trace output")

	//
	// daemon config
	//
	daemonCfg := config.DefaultConfig()

	validDurationUnits := "(valid units: 'd', 'h', 'm', 's')"

	CleanupMethodVar(&daemonCfg.CleanupMethod, "p", "passive mode")
	DurationVar(&daemonCfg.CleanupLeasesInterval, "t", "cleanup old leases interval "+validDurationUnits)
	flag.IntVar(&daemonCfg.MissedPingsThreshold, "m", daemonCfg.MissedPingsThreshold, "missed arping threshold")
	DurationVar(&daemonCfg.LeasesExpiryDuration, "e", "leases expiry duration "+validDurationUnits)
	flag.BoolVar(&daemonCfg.PersistentLeases, "k", daemonCfg.PersistentLeases, "keep leases over restart")

	flag.Parse()

	return cliCfg, daemonCfg
}
