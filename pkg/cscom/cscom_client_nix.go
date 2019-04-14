// +build !windows

package cscom

import (
	"github.com/j-keck/lsleases/pkg/config"
	"net"
)


func connect() (net.Conn, error) {
	return net.Dial("unix", config.SOCK_PATH)
}
