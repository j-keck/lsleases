package leases

import (
	"testing"
)

func TestAddOrUpdate(t *testing.T) {
	cache := new(Leases)

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
	cache := new(Leases)

	if len(cache.List()) != 0 {
		t.Error("new cache not empty")
	}

	lease := Lease{Host: "test", IP: "1.1.1.1", Mac: "01:02:03:04:05:06"}
	cache.AddOrUpdate(lease)

	if len(cache.List()) != 1 {
		t.Error("cache empty - expected one lease")
	}
}
