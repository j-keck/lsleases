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
			msg := strings.Join([]string{
				"no running server found",
				"  - start the server per 'sudo lsleasesd'",
				"  - or use the standalone mode per 'sudo lsleases -s'",
			}, "\n")
			return nil, errors.New(msg)
		}
		return nil, err
	}
}
