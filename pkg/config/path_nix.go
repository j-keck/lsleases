// +build !windows

package config

const (
	SOCK_PATH              = "/var/run/lsleases.sock"
	PERSISTENT_LEASES_PATH = "/var/cache/lsleases/leases.json"
)
