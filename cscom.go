// client server communication
package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

var ErrNoServerInstanceRunning = errors.New("no running server instance found - start one with 'lsleases -s'")

func clientListener(clientChan chan []byte) {
	ln, err := openListener()
	exitOnError(err, "openListener")
	defer ln.Close()

	for {
		con, err := ln.Accept()
		exitOnError(err, "uds accept")

		buf := make([]byte, 1024)
		if n, err := con.Read(buf); err != nil {
			fmt.Println("uds read error: ", err)
		} else {
			clientChan <- buf[:n]
			con.Write(<-clientChan)
			con.Close()
		}
	}
}

func tellServer(cmd string) error {
	_, err := tellServerAndThen(cmd, func(con net.Conn) ([]byte, error) { return nil, nil })
	return err
}

func askServer(cmd string) ([]byte, error) {
	return tellServerAndThen(cmd, func(con net.Conn) ([]byte, error) {
		var buffer []byte
		for {
			b := make([]byte, 512)
			n, err := con.Read(b)
			if err != nil && err != io.EOF {
				return nil, err
			}

			if n == 0 {
				break
			}

			buffer = append(buffer, b[:n]...)
		}
		return buffer, nil
	})
}

func tellServerAndThen(cmd string, andThen func(net.Conn) ([]byte, error)) ([]byte, error) {
	con, err := bind()
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") || // no sock file
			strings.Contains(err.Error(), "connection refused") || // sock file without running listener
			strings.Contains(err.Error(), "Timed out waiting for pipe") { // windows - no running server
			return nil, ErrNoServerInstanceRunning
		}
		return nil, err
	}
	defer con.Close()

	_, err = con.Write([]byte(cmd))
	if err != nil {
		return nil, err
	}

	return andThen(con)
}
