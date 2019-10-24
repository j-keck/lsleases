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
	if err := daemon.Start(daemonCfg); err != nil {
		log.Errorf("unable to start daemon - %s", err.Error())
	}
}

func newLogger(cliCfg CliConfig) plog.Logger {

	consoleLogger := plog.NewConsoleLogger(" | ")
	consoleLogger.SetLevel(cliCfg.logLevel)

	if cliCfg.logTimestamps {
		consoleLogger.AddLogFormatter(plog.TimestampUnixDate)
	}

	consoleLogger.AddLogFormatter(plog.Level)

	if cliCfg.logLevel == plog.Trace {
		consoleLogger.AddLogFormatter(plog.Location)
	}

	consoleLogger.AddLogFormatter(plog.Message)

	return plog.GlobalLogger().Add(consoleLogger)
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
