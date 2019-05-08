package cscom

import (
	"github.com/j-keck/lsleases/pkg/config"
	"github.com/j-keck/plog"
	"fmt"
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
	return AskServerWithPayload(log, clientReq, "")
}

func AskServerWithPayload(log plog.Logger, clientReq ClientRequest, payload string) (ServerResponse, error) {
	log.Tracef("connect to unix-domain-socket at: %s", config.SOCK_PATH)
	con, err := connect()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	req := string(clientReq)
	if payload != "" {
		req = fmt.Sprintf("%s:%s", req, payload)
	}

	log.Tracef("send client request: %s", req)
	_, err = con.Write([]byte(req))
	if err != nil {
		return nil, err
	}

	// depending on the request, we may expected a response
	switch clientReq {
	case GetVersion:
		return Version(readString(con)), nil
	case GetLeases:
		return Leases(readLeases(con)), nil
	case GetLeasesSince:
		return Leases(readLeases(con)), nil
	default:
		log.Warnf("unhandled client request: '%s'", clientReq)
		return nil, nil
	}
}
