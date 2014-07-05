package main

import (
	"testing"
)

func IgnoreTestNoServerInstanceRunning(t *testing.T) {

	version, err := askServer("version")
	if err != ErrNoServerInstanceRunning {
		t.Errorf("expected: '%s', received: '%s'", ErrNoServerInstanceRunning, version)
	}

}
