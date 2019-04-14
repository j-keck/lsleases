package leases

import (
	"encoding/json"
	"fmt"
	"github.com/j-keck/lsleases/pkg/config"
	"github.com/j-keck/plog"
	"io/ioutil"
	"os"
	"path"
	"time"
)

type Cache []Lease

func NewCache(cfg config.Config, log plog.Logger) *Cache {
	cache := new(Cache)
	go func() {
		cleaner := NewCleaner(cfg, log)
		ticker := time.NewTicker(cfg.CleanupLeasesInterval)
		defer ticker.Stop()
		for {
			*cache = cleaner.FilterObsoleteLeases(*cache)
			<-ticker.C
		}
	}()
	return cache
}

func (self *Cache) AddOrUpdate(lease Lease) {
	byMac := func(cur Lease) bool {
		return cur.Mac == lease.Mac
	}

	if orig, ok := self.findByForUpdate(byMac); ok {
		*orig = lease
	} else {
		*self = append(*self, lease)
	}
}

func (self *Cache) ContainsMac(mac string) bool {
	_, found := self.FindBy(func(l Lease) bool {
		return l.Mac == mac
	})
	return found
}

func (self *Cache) LoadLeases(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("no persistence file found under %s\n", filePath)
	}

	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, &self)
}

func (self *Cache) SaveLeases(filePath string) error {
	j, err := json.Marshal(*self)
	if err != nil {
		return err
	}

	fileDir := path.Dir(filePath)
	if _, err := os.Stat(fileDir); os.IsNotExist(err) {
		os.MkdirAll(fileDir, os.ModePerm)
	}

	return ioutil.WriteFile(filePath, []byte(j), 0644)
}

func (self *Cache) Append(leases []Lease) {
	*self = append(*self, leases...)
}

func (self *Cache) FindBy(pred func(Lease) bool) (Lease, bool) {
	for _, lease := range *self {
		if pred(lease) {
			return lease, true
		}
	}
	return *new(Lease), false
}

func (self *Cache) findByForUpdate(pred func(Lease) bool) (*Lease, bool) {
	for i := 0; i < len(*self); i++ {
		if pred((*self)[i]) {
			return &(*self)[i], true
		}
	}

	// TODO: why does the following code not work?
	// i _think_ it copies the values in the range operator
	// so the later update are on a copied value
	// for _, lease := range *self {
	//	if pred(lease) {
	//		return &lease, true
	//	}
	// }

	return new(Lease), false
}

func (self *Cache) List() []Lease {
	return *self
}

func (self *Cache) Clear() {
	*self = []Lease{}
}

func (self *Cache) Filter(pred func(Lease) bool) Cache {
	var leases []Lease
	for _, lease := range *self {
		if pred(lease) {
			leases = append(leases, lease)
		}
	}
	return leases
}

func (self *Cache) Map(f func(Lease) Lease) Cache {
	var leases []Lease
	for _, lease := range *self {
		leases = append(leases, f(lease))
	}
	return leases
}

type SortByCreated Cache

func (self SortByCreated) Len() int {
	return len(self)
}
func (self SortByCreated) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}
func (self SortByCreated) Less(i, j int) bool {
	return self[i].Created.Before(self[j].Created)
}
