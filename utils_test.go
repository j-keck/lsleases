package main

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	testValid := func(parse, expected string) {
		if d, err := parseDuration(parse); err != nil {
			t.Error(err)
		} else {
			expected, _ := time.ParseDuration(expected)
			if d != expected {
				t.Errorf("expected: %s, received: %s", expected, d)
			}
		}
	}

	testValid("13m", "13m")
	testValid("3h10m", "3h10m")
	testValid("2d", "48h")
	testValid("2d3h10m", "51h10m")
}
