package leases

import (
	"testing"
)

func TestAddOrUpdate(t *testing.T) {
	var cache LeasesCache

	byMac := func(mac string) func(Lease) bool {
		return func(cur Lease) bool {
			return cur.Mac == mac
		}
	}

	lease := Lease{Host: "test", IP: "1.1.1.1", Mac: "01:02:03:04:05:06"}

	// check lease is not in the cache
	if _, ok := cache.FindBy(byMac(lease.Mac)); ok {
		t.Error("lease already in the cache")
	}

	cache.AddOrUpdate(lease)
	if _, ok := cache.FindBy(byMac(lease.Mac)); !ok {
		t.Error("lease was not added to the cache")
	}

	lease.Host = "new name"
	cache.AddOrUpdate(lease)
	if updated, ok := cache.FindBy(byMac(lease.Mac)); ok {
		if updated.Host != lease.Host {
			t.Errorf("lease not updated - actual: %s, expected: %s", updated.String(), lease.String())
		}
	} else {
		t.Error("lease not found")
	}
}

func TestCache(t *testing.T) {
	var cache LeasesCache

	if len(cache.List()) != 0 {
		t.Error("new cache not empty")
	}

	lease := Lease{Host: "test", IP: "1.1.1.1", Mac: "01:02:03:04:05:06"}
	cache.AddOrUpdate(lease)

	if len(cache.List()) != 1 {
		t.Error("cache empty - expected one lease")
	}
}

func TestLeasesFilter(t *testing.T) {
	//	cache := LeasesCache{}
	// for i := 0; i < 10; i++ {
	//	cache.Leases = append(cache.Leases, &sniffer.Lease{MissedPings: i})
	// }

	// filtered := cache.Filter(func(l *Lease) bool {
	//	return l.MissedPings < 3
	// })

	// if len(filtered.Leases) != 3 {
	//	t.Errorf("expected 3 leases - %d leases found", len(filtered.Leases))
	// }
}
