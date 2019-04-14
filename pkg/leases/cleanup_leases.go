package leases

import (
	"time"
)

import "github.com/j-keck/lsleases/pkg/config"
import "github.com/j-keck/plog"

type Cleaner interface {
	FilterObsoleteLeases(Cache) Cache
}

func NewCleaner(cfg config.Config, log plog.Logger) Cleaner {
	if cfg.CleanupMethod == config.PingBasedCleanup {
		log.Infof("enable active check - ping every: %s", cfg.CleanupLeasesInterval)
		if aliveChecker, err := NewAliveChecker(log); err != nil {
			log.Warnf("ping test failed - fallback to time based cleanup - %s", err.Error())
			return timeBasedCleanup{cfg, log}
		} else {
			return pingBasedCleanup{cfg, log, *aliveChecker}
		}
	}

	return timeBasedCleanup{cfg, log}
}

type timeBasedCleanup struct {
	cfg config.Config
	log plog.Logger
}

func (self timeBasedCleanup) FilterObsoleteLeases(leases Cache) Cache {
	return leases.Filter(func(lease Lease) bool {
		return time.Now().After(lease.ExpiryDate)
	})
}

type pingBasedCleanup struct {
	cfg          config.Config
	log          plog.Logger
	aliveChecker aliveChecker
}

func (self pingBasedCleanup) FilterObsoleteLeases(leases Cache) Cache {
	self.log.Debug("check online hosts")
	updated := leases.Map(func(lease Lease) Lease {
		if hostIsAlive, err := self.aliveChecker.IsAlive(lease.IP); err != nil {
			self.log.Warnf("unable to ping host: %s - %s", lease.String(), err.Error())
			lease.MissedPings++
		} else if hostIsAlive {
			self.log.Tracef("alive: %s", lease.String())
			lease.MissedPings = 0
		} else {
			self.log.Tracef("NOT alive: %s", lease.String())
			lease.MissedPings++
		}
		return lease
	})

	return updated.Filter(func(lease Lease) bool {
		keep := lease.MissedPings < self.cfg.MissedPingsThreshold
		if !keep {
			self.log.Debugf("remove old lease: %s", lease.String())
		}
		return keep
	})
}
