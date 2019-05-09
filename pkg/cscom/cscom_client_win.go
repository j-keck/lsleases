// +build windows

package cscom

import (
	// native named pipe in stdlib missing: http://code.google.com/p/go/issues/detail?id=3599
	"github.com/j-keck/npipe"
	"github.com/j-keck/lsleases/pkg/config"
	"net"
	"time"
	"strings"
	"errors"
)


func connect() (net.Conn, error) {
	if con, err := npipe.DialTimeout(config.SOCK_PATH, time.Second); err == nil {
		return con, nil
	} else {
		if strings.Contains(err.Error(), "Timed out waiting for pipe") {
			return nil, errors.New("no running server found - start one with 'lsleasesd.exe'")
		}
		return nil, err
	}
}
