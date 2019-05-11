package main

import "github.com/j-keck/lsleases/pkg/sniffer"
import "github.com/j-keck/lsleases/pkg/config"
import "github.com/j-keck/plog"

func main() {
  // create a logger instance
  log := plog.NewDefaultConsoleLogger()

  // create the sniffer with the default configuration
  cfg := config.NewDefaultConfig()
  sniffer := sniffer.NewSniffer(cfg, log)

  // subscribe to DHCP leases events and log the events
  go func() {
    leasesC := sniffer.Subscribe(10)
    for {
      lease := <-leasesC
      log.Infof("new lease: %s", lease.String())
    }
  }()

  if err := sniffer.Start(); err == nil {
    log.Info("sniffing ... - hit <CTRL-C> to abort -")
    select {}
  } else {
    panic(err)
  }
}
