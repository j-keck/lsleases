// +build windows

package cscom

import (
	// native named pipe in stdlib missing: http://code.google.com/p/go/issues/detail?id=3599
	"github.com/j-keck/npipe"
	"github.com/j-keck/lsleases/pkg/config"
	"net"
	"time"
)


func connect() (net.Conn, error) {
	return npipe.DialTimeout(config.SOCK_PATH, time.Second)
}
