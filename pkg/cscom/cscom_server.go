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

func (self *comServer) Listen(cb func(ClientRequest) ServerResponse) error {
	self.log.Trace("waiting for client connection")
	con, err := self.lsnr.Accept()
	if err != nil {
		return err
	}

	self.log.Trace("client connected - waiting for request")
	req := ClientRequest(strings.TrimSpace(readString(con)))
	self.log.Tracef("client request received: '%s'", req)

	if resp := cb(req); resp != nil {
		con.Write(resp.Serialize())
	}
	con.Close()

	return nil
}

