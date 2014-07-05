package main

import (
	"testing"
)

func TestFilter(t *testing.T) {
	ls := DHCPLeases{}
	for i := 0; i < 10; i++ {
		ls.Leases = append(ls.Leases, &DHCPLease{MissedPings: i})
	}

	filtered := ls.Filter(func(l *DHCPLease) bool {
		return l.MissedPings < 3
	})

	if len(filtered.Leases) != 3 {
		t.Errorf("expected 3 leases - %d leases found", len(filtered.Leases))
	}
}
