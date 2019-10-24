package main

import "github.com/j-keck/lsleases/pkg/cscom"
import "github.com/j-keck/lsleases/pkg/leases"
import "github.com/j-keck/lsleases/pkg/sniffer"
import "github.com/j-keck/lsleases/pkg/config"
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

	Standalone

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

	log := plog.GlobalLogger().Add(plog.NewDefaultConsoleLogger())
	log.SetLevel(cfg.logLevel)

	switch cfg.action {
	case PrintVersion:
		var serverVersion string
		if version, err := cscom.AskServer(cscom.GetVersion); err == nil {
			serverVersion = version.(cscom.Version).String()
		} else {
			serverVersion = err.Error()
		}
		fmt.Printf("lsleases  (client): %s\nlsleasesd (server): %s\n", version, serverVersion)

	case PrintHelp:
		flag.Usage()

	case Standalone:
		sniffer := sniffer.NewSniffer(config.NewDefaultConfig())
		go func() {
			leasesC := sniffer.Subscribe(10)

			// TODO: combine the output from here with the code in WatchLeases
			format := "%-9s  %-15s  %-17s  %s\n"
			fmt.Printf(format, "Captured", "Ip", "Mac", "Host")
			for {
				lease := <-leasesC
				ts := lease.Created.Format("15:04:05")
				fmt.Printf(format, ts, lease.IP, lease.Mac, lease.Host)
			}
		}()

		if err := sniffer.Start(); err == nil {
			select {}
		}


	case ListLeases:
		if leases, err := cscom.AskServer(cscom.GetLeases); err == nil {
			if cfg.jsonOutput {
				listLeasesAsJson(leases.(cscom.Leases))
			} else {
				listLeases(cfg, leases.(cscom.Leases))
			}
		} else {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}

	case WatchLeases:

		if leases, err := cscom.AskServer(cscom.GetLeases); err == nil {

			format := "%-9s  %-15s  %-17s  %s\n"
			fmt.Printf(format, "Captured", "Ip", "Mac", "Host")
			var ts int64
			for {
				if leases, err = cscom.AskServerWithPayload(
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
					fmt.Fprintf(os.Stderr, "%s\n", err.Error())
					break
				}

			}
		} else {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}

	case ClearLeases:
		cscom.TellServer(cscom.ClearLeases)

	case Shutdown:
		cscom.TellServer(cscom.Shutdown)
	}
}

func listLeases(cfg CliConfig, leases []leases.Lease) {
	format := "%-15s  %-17s  %s\n"
	fmt.Printf(format, "Ip", "Mac", "Host")
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
	standalone := flag.Bool("s", false, "standalone mode - no daemon necessary")
	watchLeases := flag.Bool("w", false, "watch leases")
	clearLeases := flag.Bool("c", false, "clear leases history")
	shutdown := flag.Bool("x", false, "shutdown server")

	// options
	flag.BoolVar(&cfg.jsonOutput, "j", false, "json output")

	flag.Parse()

	// action
	if *printHelp {
		cfg.action = PrintHelp
	} else if *printVersion {
		cfg.action = PrintVersion
	} else if *standalone {
		cfg.action = Standalone
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
