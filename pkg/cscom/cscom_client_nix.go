// +build !windows

package cscom

import (
	"github.com/j-keck/lsleases/pkg/config"
	"net"
	"strings"
	"errors"
)


func connect() (net.Conn, error) {
	if con, err := net.Dial("unix", config.SOCK_PATH); err == nil {
		return con, nil
	} else {
		sockFileNotFound := "no such file or directory"
		serverNotRunning := "connection refused"

		if strings.Contains(err.Error(), sockFileNotFound) ||
			strings.Contains(err.Error(), serverNotRunning) {
			return nil, errors.New("no running server found - start one with 'lsleasesd'")
		}
		return nil, err
	}
}
