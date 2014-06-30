package main

import (
	// native named pipe in stdlib missing: http://code.google.com/p/go/issues/detail?id=3599
	"github.com/j-keck/npipe"
	"net"
	"time"
)

const sockFileName = `\\.\pipe\lsleases`

func openListener() (net.Listener, error) {
	return npipe.Listen(sockFileName)
}
func closeListener() {
	// nothing to do
}

func bind() (net.Conn, error) {
	return npipe.DialTimeout(sockFileName, time.Second)
}
