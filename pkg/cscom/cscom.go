package cscom

import (
	"encoding/json"
	"github.com/j-keck/lsleases/pkg/leases"
	"io"
	"net"
)

type ClientRequest string

const (
	GetVersion       = ClientRequest("get-version")
	GetLeases        = ClientRequest("get-leases")
	GetCleanupMethod = ClientRequest("get-cleanup-method")
	ClearLeases      = ClientRequest("clear-leases")
	Shutdown         = ClientRequest("shutdown")
)

type ServerResponse interface {
	Serialize() []byte
}

type Version string

func (self Version) Serialize() []byte {
	return []byte(self)
}

type Leases []leases.Lease

func (self Leases) Serialize() []byte {
	b, err := json.Marshal(self)
	if err != nil {
		panic(err)
	}
	return b
}

func read(con net.Conn) []byte {
	var buffer []byte
	for {
		tmp := make([]byte, 1024)
		n, err := con.Read(tmp)

		if err != nil && err != io.EOF {
			panic(err)
		}

		buffer = append(buffer, tmp[:n]...)

		if n < 1024 {
			break
		}
	}
	return buffer
}

func readString(con net.Conn) string {
	return string(read(con))
}

func readLeases(con net.Conn) []leases.Lease {
	raw := read(con)

	var leases []leases.Lease
	err := json.Unmarshal(raw, &leases)
	if err != nil {
		panic(err)
	}

	return leases
}
