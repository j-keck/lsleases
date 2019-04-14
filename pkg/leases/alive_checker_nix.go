// +build !windows

package leases

import "github.com/j-keck/arping"
import "github.com/j-keck/plog"
import "net"
import "errors"

func NewAliveChecker(log plog.Logger) (*aliveChecker, error) {

	// find any local ip where
	//  * the interface is not down
	//  * is no loopback interface (because it has no mac)
	//  * is no v6 address
	ip, err := findLocalIP(log)
	if err != nil {
		return nil, err
	}

	// arping on a local ip results always in a timeout.
	// if we got a other error, we know that we have no permissions for raw sockets
	log.Tracef("arping to '%s' to check permission", ip.String())
	if _, _, err := arping.Ping(ip); err == arping.ErrTimeout {
		log.Tracef("arping check succeed")
		return &aliveChecker{log}, nil
	} else {
		return nil, errors.New("no permission for arping")
	}
}

func (self *aliveChecker) IsAlive(ip string) (bool, error) {
	if _, _, err := arping.Ping(net.ParseIP(ip)); err == arping.ErrTimeout {
		return false, nil
	} else if err == nil {
		return true, nil
	} else {
		return false, err
	}
}

func findLocalIP(log plog.Logger) (net.IP, error) {
	log.Trace("search interface for arping test")
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	// loop over all interfaces
	for _, iface := range ifaces {

		if iface.Flags&net.FlagLoopback != 0 {
			log.Tracef("ignore interface '%s' - is loopback", iface.Name)
			continue
		}

		if iface.Flags&net.FlagUp == 0 {
			log.Tracef("ignore interface '%s' - is down", iface.Name)
			continue
		}

		// loop over all addresses from the current interface
		if addrs, err := iface.Addrs(); err == nil {
			log.Tracef("testing addresses from interface: '%s'", iface.Name)
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok {
					if ipNet.IP.To4() == nil {
						// only work on tcp4 addresses (this is tcp6)
						log.Tracef("ignore address '%s' - is v6", ipNet.IP.String())
						continue
					}
					log.Tracef("found usable address: '%s' on interface: '%s'", ipNet.IP.String(), iface.Name)
					return ipNet.IP, nil
				}
			}
		} else {
			log.Tracef("unable to get addresses from interface: '%s' - %s", iface.Name, err.Error())
		}
	}
	return nil, errors.New("no usable interface for arping found")
}
