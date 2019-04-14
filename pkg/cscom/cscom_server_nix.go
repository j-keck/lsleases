// +build !windows

package cscom

import (
	"github.com/j-keck/lsleases/pkg/config"
	"github.com/j-keck/plog"
	"net"
	"os"
	"fmt"
)

func startListener(log plog.Logger) (net.Listener, error) {
	if _, err := os.Stat(config.SOCK_PATH); err == nil {
		if _, err := AskServer(log, GetVersion); err != nil {
			log.Debugf("remove old socket file %s", config.SOCK_PATH)
			if err = os.Remove(config.SOCK_PATH); err != nil {
				return nil, fmt.Errorf("delete old socket file failed - %s", err.Error())
			}
		}
	}

	ls, err := net.Listen("unix", config.SOCK_PATH)
	if err != nil {
		return nil, err
	}

	err = os.Chmod(config.SOCK_PATH, 0666)
	if err != nil {
		return nil, err
	}

	return ls, nil
}
