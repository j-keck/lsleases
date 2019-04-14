package cscom

import (
	"github.com/j-keck/lsleases/pkg/config"
	"github.com/j-keck/plog"
	"net"
)

func TellServer(log plog.Logger, clientReq ClientRequest) error {
	log.Tracef("connect to unix-domain-socket at: %s", config.SOCK_PATH)
	con, err := connect()
	if err != nil {
		return err
	}
	defer con.Close()

	log.Tracef("send client request: %s", clientReq)
	_, err = con.Write([]byte(clientReq))
	return err
}

func AskServer(log plog.Logger, clientReq ClientRequest) (ServerResponse, error) {
	log.Tracef("connect to unix-domain-socket at: %s", config.SOCK_PATH)
	con, err := net.Dial("unix", config.SOCK_PATH)
	if err != nil {
		return nil, err
	}
	defer con.Close()

	log.Tracef("send client request: %s", clientReq)
	_, err = con.Write([]byte(clientReq))
	if err != nil {
		return nil, err
	}

	// depending on the request, we may the response
	switch clientReq {
	case GetVersion:
		return Version(readString(con)), nil
	case GetLeases:
		return Leases(readLeases(con)), nil
	default:
		log.Warnf("unhandled client request: '%s'", clientReq)
		return nil, nil
	}
}
