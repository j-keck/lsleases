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
