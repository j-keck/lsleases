// +build !windows

package main

import (
	"net"
	"os"
)

var sockPath = appDataPath + "/lsleases.sock"

func openListener() (net.Listener, error) {
	// remove old stale sock file if no other instance is running
	if _, err := os.Stat(sockPath); err == nil {
		if err := tellServer("version"); err != nil {
			os.Remove(sockPath)
		}
	}

	// open listener
	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		return nil, err
	}

	// change permissions
	err = os.Chmod(sockPath, 0666)
	if err != nil {
		return nil, err
	}

	return ln, err
}

func closeListener() {
	os.Remove(sockPath)
}

func bind() (net.Conn, error) {
	return net.Dial("unix", sockPath)
}
