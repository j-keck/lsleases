package main

import (
	"flag"
	"fmt"
	"github.com/j-keck/lsleases/pkg/config"
	"github.com/j-keck/lsleases/pkg/daemon"
	"github.com/j-keck/plog"
)

type CliConfig struct {
	logLevel      plog.LogLevel
	logTimestamps bool
	printVersion  bool
	printHelp     bool
}

func main() {
	cliCfg, daemonCfg := parseFlags()

	if cliCfg.printHelp {
		flag.Usage()
		return
	}

	if cliCfg.printVersion {
		fmt.Printf("lsleasesd (server): %s\n", daemon.Version())
		return
	}

	log := newLogger(cliCfg)
	if err := daemon.Start(daemonCfg, log); err != nil {
		log.Errorf("unable to start daemon - %s", err.Error())
	}
}

func newLogger(cliCfg CliConfig) plog.Logger {

	var formatters []plog.Formatter

	if cliCfg.logTimestamps {
		formatters = append(formatters, plog.TimestampUnixDate)
	}

	formatters = append(formatters, plog.Level)

	if cliCfg.logLevel == plog.Trace {
		formatters = append(formatters, plog.Location)
	}

	formatters = append(formatters, plog.Message)


	// initalize the logger with the given formatters
	log := plog.NewConsoleLogger(" | ", formatters...)
	log.SetLevel(cliCfg.logLevel)

	return log
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

	// log timestamps
	flag.BoolVar((*bool)(&cliCfg.logTimestamps), "log-timestamps", false, "log messages with timestamps in unix format")

	//
	// daemon config
	//
	daemonCfg := config.NewDefaultConfig()

	validDurationUnits := "(valid units: 'd', 'h', 'm', 's')"

	CleanupMethodVar(&daemonCfg.CleanupMethod, "p", "passive mode")
	DurationVar(&daemonCfg.CleanupLeasesInterval, "t", "cleanup interval "+validDurationUnits)
	flag.IntVar(&daemonCfg.MissedPingsThreshold, "m", daemonCfg.MissedPingsThreshold, "missed arping threshold")
	DurationVar(&daemonCfg.LeasesExpiryDuration, "e", "leases expiry duration "+validDurationUnits)
	flag.BoolVar(&daemonCfg.PersistentLeases, "k", daemonCfg.PersistentLeases, "keep leases over restart")

	flag.Parse()

	return cliCfg, daemonCfg
}
