// +build windows

package config

import (
	"os"
)

var (
	SOCK_PATH              = `\\.\pipe\lsleases`
	PERSISTENT_LEASES_PATH = os.Getenv("USERPROFILE") + "/lsleases/leases.json"
)
