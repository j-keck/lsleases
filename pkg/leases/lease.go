package leases

import (
	"fmt"
	"time"
)

type Lease struct {
	Created     time.Time
	ExpiryDate  time.Time
	MissedPings int
	IP          string
	Mac         string
	Host        string
}

func (self *Lease) String() string {
	return fmt.Sprintf("host: %s, ip: %s, mac: %s", self.Host, self.IP, self.Mac)
}


type Leases []Lease

func (self *Leases) Filter(pred func(Lease) bool) Leases {
	var leases []Lease
	for _, lease := range *self {
		if pred(lease) {
			leases = append(leases, lease)
		}
	}
	return leases
}

func (self *Leases) Map(f func(Lease) Lease) Leases {
	var leases []Lease
	for _, lease := range *self {
		leases = append(leases, f(lease))
	}
	return leases
}

type SortByCreated Leases

func (self SortByCreated) Len() int {
	return len(self)
}
func (self SortByCreated) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}
func (self SortByCreated) Less(i, j int) bool {
	return self[i].Created.Before(self[j].Created)
}
