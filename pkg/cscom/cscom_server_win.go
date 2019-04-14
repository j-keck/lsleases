// +build windows

package cscom

import (
	"github.com/j-keck/lsleases/pkg/config"
	"github.com/j-keck/plog"
	"github.com/j-keck/npipe"
	"net"
)

func startListener(log plog.Logger) (net.Listener, error) {
	return npipe.Listen(config.SOCK_PATH)
}
