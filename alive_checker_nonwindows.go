// +build !windows

package main

import (
	"github.com/j-keck/arping"
	"net"
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
	var localIP net.IP

	// find any local ip
	addrs, err := net.InterfaceAddrs()
	panicOnError(err)

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok {
			if ipNet.IP.IsLoopback() {
				// loopback not usable, because it has no mac
				continue
			}

			localIP = ipNet.IP
			break
		}
	}

	// arping any local ip results always in timeout
	verboseLog.Printf("arping '%s' to check permission\n", localIP.String())
	if _, _, err = arping.Ping(localIP); err == arping.ErrTimeout {
		return true, nil
	}

	return false, err
}
