// +build !windows

package sniffer

import (
	"net"
	"syscall"
)

func listenConfig() net.ListenConfig {
	return net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) (err error) {
			c.Control(func(fd uintptr) {
				err = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
				if err != nil {
					return
				}

				err = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, 0xf /*syscall.SO_REUSEPORT*/, 1)
				if err != nil {
					return
				}
			})
			return
		},
	}
}
