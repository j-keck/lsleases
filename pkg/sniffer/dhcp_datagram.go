package sniffer

import (
	"errors"
	"fmt"
	"net"
)

type DHCPDatagram struct {
	Mac         string
	MessageType MessageType
	raw         rawDatagram
}

func DHCPDatagramFromBytes(bytes []byte) (*DHCPDatagram, error) {
	var datagram DHCPDatagram

	// parse the raw datagram
	if raw, err := rawDatagramFromBytes(bytes); err != nil {
		return nil, err
	} else {
		datagram.raw = raw
	}

	// update mac address
	datagram.Mac = net.HardwareAddr(datagram.raw.chaddr[:datagram.raw.hlen]).String()

	// update message type
	if id, ok := datagram.raw.options[53]; !ok {
		return nil, fmt.Errorf("field 'MessageType' not found - raw: %+v", datagram.raw)
	} else {
		if messageType, err := messageTypeFromCode(int(id[0])); err != nil {
			return nil, err
		} else {
			datagram.MessageType = messageType
		}
	}

	return &datagram, nil
}

func (self *DHCPDatagram) String() string {
	// ignore errors, use empty fields
	host, _ := self.Host()
	ip, _ := self.IP()

	return fmt.Sprintf("DHCPDatagram{ MessageType: %s, Host: %s, IP: %s, Mac: %s}",
		self.MessageType.String(), host, ip, self.Mac)
}

func (self *DHCPDatagram) Host() (string, error) {
	if h, ok := self.raw.options[12]; ok {
		return string(h), nil
	}
	return "", errors.New("dhcp-option  '12: Host Name' not found")
}

func (self *DHCPDatagram) IP() (string, error) {
	if ip, ok := self.raw.options[50]; ok {
		return net.IP(ip).String(), nil
	}
	return "", errors.New("dhcp-option '50: Requested IP Address' not found")
}

type MessageType int

const (
	DHCPDiscover = MessageType(1)
	DHCPOffer    = MessageType(2)
	DHCPRequest  = MessageType(3)
	DHCPDecline  = MessageType(4)
	DHCPACK      = MessageType(5)
	DHCPNAK      = MessageType(6)
	DHCPRelease  = MessageType(7)
	DHCPInform   = MessageType(8)
)

func (mt *MessageType) IsRequest() bool {
	return *mt == DHCPRequest
}

func (mt MessageType) String() string {
	var str string
	switch mt {
	case 1:
		str = "DHCP Discover"
	case 2:
		str = "DHCP Offer"
	case 3:
		str = "DHCP Request"
	case 4:
		str = "DHCP Decline"
	case 5:
		str = "DHCP ACK"
	case 6:
		str = "DHCP NAK"
	case 7:
		str = "DHCP Release"
	case 8:
		str = "DHCP Inform"
	}
	return str
}
func messageTypeFromCode(code int) (MessageType, error) {
	switch code {
	case 1:
		return DHCPDiscover, nil
	case 2:
		return DHCPOffer, nil
	case 3:
		return DHCPRequest, nil
	case 4:
		return DHCPDecline, nil
	case 5:
		return DHCPACK, nil
	case 6:
		return DHCPNAK, nil
	case 7:
		return DHCPRelease, nil
	case 8:
		return DHCPInform, nil
	default:
		return 0, fmt.Errorf("invalid code: %d", code)
	}
}
