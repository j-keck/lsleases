package daemon

import (
	"github.com/j-keck/lsleases/pkg/config"
	"github.com/j-keck/lsleases/pkg/cscom"
	"github.com/j-keck/lsleases/pkg/sniffer"
	"github.com/j-keck/plog"
	"os"
	"os/signal"
)

func Start(cfg config.Config, log plog.Logger) error {
	log.Infof("startup  - version: %s", version)

	// client-server communication
	com, err := cscom.NewComServer(log)
	if err != nil {
		return err
	}

	// initialize sniffer
	sniffer := sniffer.NewCachedSniffer(cfg, log)
	if cfg.PersistentLeases {
		log.Debug("load leases")
		if err := sniffer.LoadLeases(); err != nil {
			log.Infof("unable to load leases - start with empty leases cache - %s", err.Error())
		}
	}

	// start sniffer
	if err := sniffer.Start(); err != nil {
		return err
	}

	// shutdown handler
	shutdown := func() {
		if err := sniffer.SaveLeases(); err != nil {
			log.Warnf("unable to save leases - %s", err.Error())
		}
		log.Infof("shutdown - version: %s", version)
		os.Exit(0)
	}

	// catch CTRL-C and trigger shutdown
	go func() {
		interruptC := make(chan os.Signal)
		signal.Notify(interruptC, os.Interrupt)
		<-interruptC
		shutdown()
	}()


	// wait for 'lsleases' client requests
	for {
		if err := com.Listen(func(req cscom.ClientRequest) cscom.ServerResponse {
			switch req {
			case cscom.GetVersion:
				return cscom.Version(version)
			case cscom.GetLeases:
				return cscom.Leases(sniffer.ListLeases())
			case cscom.ClearLeases:
				sniffer.ClearLeases()
			case cscom.Shutdown:
				shutdown()
			}
			return nil
		}); err != nil {
			log.Warnf("cscom error - %s", err.Error())
		}
	}
}
