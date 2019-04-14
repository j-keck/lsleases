// +build windows

package sniffer

import (
	"golang.org/x/sys/windows"
	"net"
	"syscall"
)

func listenConfig() net.ListenConfig {
	return net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) (err error) {
			c.Control(func(fd uintptr) {
				err = windows.SetsockoptInt(windows.Handle(fd), windows.SOL_SOCKET, windows.SO_REUSEADDR, 1)
			})
			return
		},
	}
}
