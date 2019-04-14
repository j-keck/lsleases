package main

import "github.com/j-keck/lsleases/pkg/cscom"
import "github.com/j-keck/lsleases/pkg/leases"
import "github.com/j-keck/plog"
import "flag"
import "fmt"
import "os"
import "encoding/json"

type Action int

const (
	PrintVersion Action = iota
	PrintHelp

	ListLeases
	ClearLeases

	Shutdown
)

type CliConfig struct {
	logLevel     plog.LogLevel
	action       Action
	jsonOutput   bool
}

func main() {
	cfg := parseFlags()

	log := plog.NewConsoleLogger()
	defer log.Flush()
	log.SetLevel(cfg.logLevel)

	switch cfg.action {
	case PrintVersion:
		version, _ := cscom.AskServer(log, cscom.GetVersion)
		println(version.(cscom.Version))
	case PrintHelp:
		flag.Usage()
	case ListLeases:
		leases, _ := cscom.AskServer(log, cscom.GetLeases)

		if cfg.jsonOutput {
			listLeasesAsJson(leases.(cscom.Leases))
		} else {
			listLeases(cfg, leases.(cscom.Leases))
		}

	case ClearLeases:
		cscom.TellServer(log, cscom.ClearLeases)
	case Shutdown:
		cscom.TellServer(log, cscom.Shutdown)
	}
}

func listLeases(cfg CliConfig, leases []leases.Lease) {
	format := "%-15s  %-17s  %s\n"
	for _, lease := range leases {
		fmt.Printf(format, lease.IP, lease.Mac, lease.Host)
	}
}

func listLeasesAsJson(leases []leases.Lease) {
	b, _ := json.Marshal(leases)
	os.Stdout.Write(b)
}
func parseFlags() CliConfig {
	cfg := CliConfig{logLevel: plog.Info}

	// log level
	plog.FlagDebugVar(&cfg.logLevel, "v", "debug output")
	plog.FlagTraceVar(&cfg.logLevel, "vv", "trace output")

	// action
	printHelp := flag.Bool("h", false, "print help and exit")
	printVersion := flag.Bool("V", false, "print version and exit")
	clearLeases := flag.Bool("c", false, "clear leases")
	shutdown := flag.Bool("x", false, "shutdown server")

	// options
	flag.BoolVar(&cfg.jsonOutput, "j", false, "json output")

	flag.Parse()

	// action
	if *printHelp {
		cfg.action = PrintHelp
	} else if *printVersion {
		cfg.action = PrintVersion
	} else if *clearLeases {
		cfg.action = ClearLeases
	} else if *shutdown {
		cfg.action = Shutdown
	} else {
		cfg.action = ListLeases
	}

	return cfg
}
