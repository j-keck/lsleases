package main

import "github.com/j-keck/lsleases/pkg/cscom"
import "github.com/j-keck/lsleases/pkg/leases"
import "github.com/j-keck/plog"
import "flag"
import "fmt"
import "os"
import "encoding/json"
import "time"

type Action int

const (
	PrintVersion Action = iota
	PrintHelp

	ListLeases
	WatchLeases
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
	log.SetLevel(cfg.logLevel)

	switch cfg.action {
	case PrintVersion:
		var serverVersion string
		if version, err := cscom.AskServer(log, cscom.GetVersion); err == nil {
			serverVersion = version.(cscom.Version).String()
		} else {
			serverVersion = err.Error()
		}
		fmt.Printf("lsleases  (client): %s\nlsleasesd (server): %s\n", version, serverVersion)

	case PrintHelp:
		flag.Usage()

	case ListLeases:
		if leases, err := cscom.AskServer(log, cscom.GetLeases); err == nil {
			if cfg.jsonOutput {
				listLeasesAsJson(leases.(cscom.Leases))
			} else {
				listLeases(cfg, leases.(cscom.Leases))
			}
		} else {
			os.Stderr.WriteString(err.Error())
		}

	case WatchLeases:
		if leases, err := cscom.AskServer(log, cscom.GetLeases); err == nil {
			// TODO: json output?

			format := "%-9s  %-15s  %-17s  %s\n"
			fmt.Printf(format, "Captured", "IP", "Mac", "Host")
			var ts int64
			for {
				if leases, err = cscom.AskServerWithPayload(
					log,
					cscom.GetLeasesSince,
					fmt.Sprintf("%d", ts),
				); err == nil {

					ts = time.Now().UnixNano()
					for _, lease := range leases.(cscom.Leases) {
						ts := lease.Created.Format("15:04:05")
						fmt.Printf(format, ts, lease.IP, lease.Mac, lease.Host)
					}
					time.Sleep(1 * time.Second)

				} else {
					os.Stderr.WriteString(err.Error())
					break
				}

			}
		} else {
			os.Stderr.WriteString(err.Error())
		}

	case ClearLeases:
		cscom.TellServer(log, cscom.ClearLeases)

	case Shutdown:
		cscom.TellServer(log, cscom.Shutdown)
	}
}

func listLeases(cfg CliConfig, leases []leases.Lease) {
	format := "%-15s  %-17s  %s\n"
	fmt.Printf(format, "IP", "Mac", "Host")
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
	watchLeases := flag.Bool("w", false, "watch leases")
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
	} else if *watchLeases {
		cfg.action = WatchLeases
	} else if *clearLeases {
		cfg.action = ClearLeases
	} else if *shutdown {
		cfg.action = Shutdown
	} else {
		cfg.action = ListLeases
	}

	return cfg
}
