// +build windows

package leases

import (
	"bytes"
	"encoding/binary"
	"github.com/j-keck/plog"
	"net"
	"time"
)

func NewAliveChecker() (*aliveChecker, error) {
	if con, err := net.Dial("ip4:icmp", "127.0.0.1"); err != nil {
		return nil, err
	} else {
		con.Close()
		return new(aliveChecker), nil
	}
}

const ICMP_ECHO_REQUEST = 8
const ICMP_ECHO_REPLY = 0

func (self *aliveChecker) IsAlive(ip string) (bool, error) {
	con, err := net.Dial("ip4:icmp", ip)
	if err != nil {
		return false, err
	}
	defer con.Close()

	if err := con.SetReadDeadline(time.Now().Add(time.Millisecond * 500)); err != nil {
		return false, err
	}

	//
	// build request
	//
	request := ICMPPingMessage{
		Type: ICMP_ECHO_REQUEST,
		Code: 0,
		Id:   12345, // "random"
		Data: bytes.Repeat([]byte("easy ping!"), 5)}

	//
	// send request
	//
	if _, err := con.Write(request.Marshal()); err != nil {
		return false, err
	}

	//
	// read response
	//
	buffer := make([]byte, 512)
	for {
		if _, err = con.Read(buffer); err != nil {
			// return nil for error (if ip offline, we get an timeout error)
			return false, nil
		} else {
			response, _ := ParseICMPPingMessage(buffer[20:])
			if response.IsResponseOf(&request) {
				return true, nil
			}
		}
	}
}

type ICMPPingMessage struct {
	Type     uint8
	Code     uint8
	Checksum uint16
	Id       uint16
	Seq      uint16
	Data     []byte
}

func (msg *ICMPPingMessage) Marshal() []byte {
	buffer := new(bytes.Buffer)

	binary.Write(buffer, binary.BigEndian, msg.Type)
	binary.Write(buffer, binary.BigEndian, msg.Code)
	binary.Write(buffer, binary.BigEndian, msg.Checksum)
	binary.Write(buffer, binary.BigEndian, msg.Id)
	binary.Write(buffer, binary.BigEndian, msg.Seq)
	buffer.Write(msg.Data)

	raw := buffer.Bytes()

	// checksum from: https://github.com/jnwhiteh/golang/blob/master/src/pkg/net/mockicmp_test.go
	csumcv := len(raw) - 1 // checksum coverage
	s := uint32(0)
	for i := 0; i < csumcv; i += 2 {
		s += uint32(raw[i+1])<<8 | uint32(raw[i])
	}
	if csumcv&1 == 0 {
		s += uint32(raw[csumcv])
	}
	s = s>>16 + s&0xffff
	s = s + s>>16
	// Place checksum back in header; using ^= avoids the
	// assumption the checksum bytes are zero.
	raw[2] ^= byte(^s)
	raw[3] ^= byte(^s >> 8)

	return raw
}

func (msg *ICMPPingMessage) IsResponseOf(request *ICMPPingMessage) bool {
	return msg.Type == ICMP_ECHO_REPLY && msg.Id == request.Id
}

func ParseICMPPingMessage(raw []byte) (ICMPPingMessage, error) {
	var msg ICMPPingMessage

	buffer := bytes.NewBuffer(raw)
	binary.Read(buffer, binary.BigEndian, &msg.Type)
	binary.Read(buffer, binary.BigEndian, &msg.Code)
	binary.Read(buffer, binary.BigEndian, &msg.Checksum)
	binary.Read(buffer, binary.BigEndian, &msg.Id)
	binary.Read(buffer, binary.BigEndian, &msg.Seq)
	msg.Data = buffer.Bytes()

	return msg, nil
}
