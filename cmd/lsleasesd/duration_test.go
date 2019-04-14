package main

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	test := func(s, expected string) {
		if actual, err := ParseDuration(s); err != nil {
			t.Error(err)
		} else {
			expected, _ := time.ParseDuration(expected)
			if actual != expected {
				t.Errorf("expected: %s, actual: %s", expected, actual)
			}
		}
	}

	test("13m", "13m")
	test("1d", "24h")
	test("2d3h10m", "51h10m")
	test("600d", "14400h")
}
