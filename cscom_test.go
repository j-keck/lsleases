package main

import (
	"testing"
)

func TestNoServerInstanceRunning(t *testing.T) {

	_, err := askServer("version")
	if err != ErrNoServerInstanceRunning {
		t.Errorf("expected: '%s', received: '%s'", ErrNoServerInstanceRunning, err)
	}

}
