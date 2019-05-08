// +build !windows

package cscom

import (
	"github.com/j-keck/lsleases/pkg/config"
	"github.com/j-keck/plog"
	"net"
	"os"
	"fmt"
	"path"
)

func startListener(log plog.Logger) (net.Listener, error) {
	// remove stale sock file if it exists
	if _, err := os.Stat(config.SOCK_PATH); err == nil {
		if _, err := AskServer(log, GetVersion); err != nil {
			log.Debugf("remove stale socket file %s", config.SOCK_PATH)
			if err = os.Remove(config.SOCK_PATH); err != nil {
				return nil, fmt.Errorf("delete stale socket file failed - %s", err.Error())
			}
		}
	}

	// create sock file directory if it's missing
	dir := path.Dir(config.SOCK_PATH)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Debugf("sock directory %s doesn't exists - create it", dir)
		os.MkdirAll(dir, os.ModePerm)
	}

    // create uds socket
	ls, err := net.Listen("unix", config.SOCK_PATH)
	if err != nil {
		return nil, err
	}

	// fix permissions - allow anyone communicate with the server
	err = os.Chmod(config.SOCK_PATH, 0666)
	if err != nil {
		return nil, err
	}

	return ls, nil
}

func stopListener(log plog.Logger) error {
	return os.Remove(config.SOCK_PATH)
}
