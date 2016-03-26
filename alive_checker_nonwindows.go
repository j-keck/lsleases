// +build !windows

package main

import (
	"fmt"
	"net"

	"github.com/j-keck/arping"
)

func isAlive(host string) (bool, error) {
	if _, _, err := arping.Ping(net.ParseIP(host)); err == arping.ErrTimeout {
		return false, nil
	} else if err == nil {
		return true, nil
	} else {
		return false, err
	}
}

func hasPermissionForAliveCheck() (bool, error) {
	verboseLog.Println("verify if i have permissions for arping")

	// find any local ip where
	//  * the interface is not down
	//  * is no loopback interface (because it has no mac)
	//  * is no v6 address
	findAnyLocalIP := func() (net.IP, bool) {

		ifaces, err := net.Interfaces()
		panicOnError(err)

		// loop over all interfaces
		for _, iface := range ifaces {

			if iface.Flags&net.FlagLoopback != 0 {
				verboseLog.Printf("ignore interface '%s' - is loopback", iface.Name)
				continue
			}

			if iface.Flags&net.FlagUp == 0 {
				verboseLog.Printf("ignore interface '%s' - is down", iface.Name)
				continue
			}

			// loop over all addresses from the current interface
			if addrs, err := iface.Addrs(); err == nil {
				verboseLog.Printf("testing addresses from interface: '%s'", iface.Name)
				for _, addr := range addrs {
					if ipNet, ok := addr.(*net.IPNet); ok {
						if ipNet.IP.To4() == nil {
							// only work on tcp4 addresses (this is tcp6)
							verboseLog.Printf("ignore address '%s' - is v6", ipNet.IP.String())
							continue
						}

						verboseLog.Printf("found usable address: '%s' on interface: '%s'", ipNet.IP.String(), iface.Name)
						return ipNet.IP, true
					}
				}
			} else {
				verboseLog.Printf("unable to get addresses from interface: '%s' - %s", iface.Name, err.Error())
			}
		}
		return nil, false
	}

	if localIP, ok := findAnyLocalIP(); ok {
		// arping any local ip results always in a timeout
		// but if we don't get a other error, we know that we have permissions for raw sockets
		verboseLog.Printf("arping '%s' to check permission\n", localIP.String())
		if _, _, err := arping.Ping(localIP); err == arping.ErrTimeout {
			return true, nil
		}
	}

	return false, fmt.Errorf("no usable interface for arping found")
}
