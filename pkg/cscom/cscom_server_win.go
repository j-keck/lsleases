// +build windows

package cscom

import (
	"github.com/j-keck/lsleases/pkg/config"
	"github.com/j-keck/plog"
	"github.com/j-keck/npipe"
	"net"
)

func startListener() (net.Listener, error) {
	return npipe.Listen(config.SOCK_PATH)
}

func stopListener() error {
	return nil
}
