// +build !windows

package main

import (
	"net"
	"os"
)

const sockFileName = "/var/lib/lsleases/lsleases.sock"

func openListener() (net.Listener, error) {
	// remove old stale sock file if no other instance is running
	if _, err := os.Stat(sockFileName); err == nil {
		if err := tellServer("version"); err != nil {
			os.Remove(sockFileName)
		}
	}

	// open listener
	ln, err := net.Listen("unix", sockFileName)
	if err != nil {
		return nil, err
	}

	// change permissions
	err = os.Chmod(sockFileName, 0666)
	if err != nil {
		return nil, err
	}

	return ln, err
}

func closeListener() {
	os.Remove(sockFileName)
}

func bind() (net.Conn, error) {
	return net.Dial("unix", sockFileName)
}
