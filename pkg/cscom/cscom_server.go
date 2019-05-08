package cscom

import (
	"github.com/j-keck/lsleases/pkg/config"
	"github.com/j-keck/plog"
	"net"
	"strings"
)

type comServer struct {
	log  plog.Logger
	lsnr net.Listener
}

func NewComServer(log plog.Logger) (*comServer, error) {
	log.Tracef("start listener on %s", config.SOCK_PATH)
	lsnr, err := startListener(log)
	if err != nil {
		return new(comServer), err
	}

	return &comServer{log, lsnr}, nil
}

func (self *comServer) Listen(cb func(ClientRequest, string) ServerResponse) error {
	self.log.Trace("waiting for client connection")
	con, err := self.lsnr.Accept()
	if err != nil {
		return err
	}

	self.log.Trace("client connected - waiting for message")
	raw := strings.TrimSpace(readString(con))
	self.log.Tracef("client message received: '%s'", raw)

	// try to split the given message in a request and payload part.
	var req ClientRequest
	var payload string
	fields := strings.SplitN(raw, ":", 2)
	if len(fields) == 1 {
		req = ClientRequest(raw)
	} else {
		req = ClientRequest(fields[0])
		payload = fields[1]
	}

	// call the callback with the received request and payload
	if resp := cb(req, payload); resp != nil {
		con.Write(resp.Serialize())
	}
	con.Close()

	return nil
}

func (self *comServer) Stop() {
	stopListener(self.log)
}
