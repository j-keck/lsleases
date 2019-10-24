// +build windows

package cscom

import (
	// native named pipe in stdlib missing: http://code.google.com/p/go/issues/detail?id=3599
	"errors"
	"github.com/j-keck/lsleases/pkg/config"
	"github.com/j-keck/npipe"
	"net"
	"strings"
	"time"
)

func connect() (net.Conn, error) {
	if con, err := npipe.DialTimeout(config.SOCK_PATH, time.Second); err == nil {
		return con, nil
	} else {
		if strings.Contains(err.Error(), "Timed out waiting for pipe") {
			msg := strings.Join([]string{
				"no running server found",
				"  - start the server per 'lsleasesd.exe'",
				"  - or use the standalone mode per 'lsleases.exe -s'",
			}, "\n")
			return nil, errors.New(msg)
		}
		return nil, err
	}
}
