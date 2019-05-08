package sniffer

import (
	"context"
	"errors"
	"fmt"
	"github.com/j-keck/lsleases/pkg/config"
	"github.com/j-keck/lsleases/pkg/leases"
	"github.com/j-keck/plog"
	"net"
	"time"
)

type sniffer struct {
	log         plog.Logger
	cfg         config.Config
	subscribers []chan leases.Lease
}

func NewSniffer(cfg config.Config, log plog.Logger) *sniffer {
	return &sniffer{cfg: cfg, log: log}
}

func (self *sniffer) Start() error {
	return self.listen(func(lease leases.Lease) {
		for _, sub := range self.subscribers {
			if len(sub) < cap(sub) {
				sub <- lease
			} else {
				self.log.Warnf("subscriber channel full - discard lease event: %s", lease)
			}
		}
	})
}

func (self *sniffer) Subscribe(bufferSize int) <-chan leases.Lease {
	c := make(chan leases.Lease, bufferSize)
	self.subscribers = append(self.subscribers, c)
	return c
}

type cachedSniffer struct {
	sniffer
	cache *leases.Leases
}

func NewCachedSniffer(cfg config.Config, log plog.Logger) *cachedSniffer {
	sniffer := NewSniffer(cfg, log)
	self := cachedSniffer{*sniffer, leases.NewCache(cfg, log)}

	go func() {
		leaseC := self.Subscribe(10)
		for {
			lease := <-leaseC
			if !self.cache.ContainsMac(lease.Mac) {
				log.Infof("new DHCP lease - %s", lease.String())
			}
			self.cache.AddOrUpdate(lease)
		}
	}()
	return &self
}

func (self *cachedSniffer) LoadLeases() error {
	self.log.Infof("load lease from %s", config.PERSISTENT_LEASES_PATH)
	return self.cache.LoadLeases(config.PERSISTENT_LEASES_PATH)
}

func (self *cachedSniffer) SaveLeases() error {
	self.log.Infof("save lease to %s", config.PERSISTENT_LEASES_PATH)
	return self.cache.SaveLeases(config.PERSISTENT_LEASES_PATH)
}

func (self *cachedSniffer) ListLeases() leases.Leases {
	return self.cache.List()
}

func (self *cachedSniffer) ClearLeases() {
	self.cache.Clear()
}

func (self *sniffer) listen(cb func(leases.Lease)) error {
	log := self.log
	log.Trace("setup listener on port :67")
	config := listenConfig()
	con, err := config.ListenPacket(context.Background(), "udp4", ":67")
	if err != nil {
		msg := fmt.Sprintf("listen on port :67 failed - %s", err.Error())
		log.Error(msg)
		// if we can't open a socket - we have nothing to do
		return errors.New(msg)
	}

	go func() {
		for {
			rawBuffer := make([]byte, 512)
			n, addr, err := con.ReadFrom(rawBuffer)
			if err != nil {
				log.Error("socket read error", err)
				// FIXME: continue, panic, restart?
				panic(err)
			}

			if datagram, err := DHCPDatagramFromBytes(rawBuffer[:n]); err == nil {
				log.Debugf("new dhcp datagram received: %s", datagram.String())
				log.Tracef("raw datagram: %+v", datagram.raw)

				if datagram.MessageType.IsRequest() {
					host, err := datagram.Host()
					if err != nil {
						log.Warnf("%s - sender mac: %s", err, datagram.Mac)
					}

					ip, err := datagram.IP()
					if err != nil {
						// don't check for 'type assertion' errors - let it crash, because it *must be* a UPDAddr
						ip = addr.(*net.UDPAddr).IP.String()
						log.Debugf("%s - sender mac: %s - use src ip: %s", err, datagram.Mac, ip)
					}

					lease := &leases.Lease{
						Created:    time.Now(),
						ExpiryDate: time.Now().Add(self.cfg.LeasesExpiryDuration),
						IP:         ip,
						Mac:        datagram.Mac,
						Host:       host}

					log.Tracef("trigger new DHCP lease event - %s", lease.String())
					cb(*lease)
				} else {
					log.Debugf("ignore datagram with type: '%s' - src mac: '%s'", datagram.MessageType, datagram.Mac)
				}
			} else {
				log.Warnf("unparsable new dhcp datagram received - error: '%s' - ignore datagram", err)
			}
		}
	}()

	return nil
}
