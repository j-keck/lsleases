package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

var (
	discover     = 1
	offer        = 2
	request      = 3
	decline      = 4
	ack          = 5
	nak          = 6
	release      = 7
	inform       = 8
	messageTypes = map[int]string{
		discover: "DHCP Discover",
		offer:    "DHCP Offer",
		request:  "DHCP Request",
		decline:  "DHCP Decline",
		ack:      "DHCP Ack",
		nak:      "DHCP NAK",
		release:  "DHCP Release",
		inform:   "DHCP Inform",
	}
)

type dhcpDatagram struct {
	op      uint8 // Message op code / message type (1 = BOOTREQUEST, 2 = BOOTREPLY)
	htype   uint8 // Hardware address type
	hlen    uint8 // Hardware address length
	hops    uint8
	xid     uint32
	secs    uint16
	flags   uint16
	ciaddr  [4]byte
	yiaddr  [4]byte
	siaddr  [4]byte
	giaddr  [4]byte
	chaddr  [16]byte
	sname   [64]byte
	file    [128]byte
	options map[uint8][]byte
}

func (datagram dhcpDatagram) Mac() string {
	return net.HardwareAddr(datagram.chaddr[:datagram.hlen]).String()
}

func (datagram dhcpDatagram) HostName() (string, error) {
	if h, ok := datagram.options[12]; ok {
		return string(h), nil
	}
	return "", fmt.Errorf("field 'host name' not found - datagram: %+v", datagram)
}

func (datagram dhcpDatagram) MessageType() (string, error) {
	if id, ok := datagram.options[53]; ok {
		return messageTypes[int(id[0])], nil
	}
	return "", fmt.Errorf("field 'message type' not found - datagram: %+v", datagram)
}

func (datagram dhcpDatagram) RequestedIPAddress() (string, error) {
	if reqIPBytes, ok := datagram.options[50]; ok {
		return net.IP(reqIPBytes).String(), nil
	}
	return "", fmt.Errorf("field 'RequestedIPAddress not found - datagram: %+v", datagram)
}

func parseDhcpDatagram(rawDatagram []byte) (datagram dhcpDatagram, err error) {
	// "panic catcher"
	//    converts a panic to an error and save it into 'err'
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(r.(string))
		}
	}()

	buffer := bytes.NewBuffer(rawDatagram)

	panicOnError(binary.Read(buffer, binary.BigEndian, &datagram.op), "read htype")
	panicOnError(binary.Read(buffer, binary.BigEndian, &datagram.htype), "read htype")
	panicOnError(binary.Read(buffer, binary.BigEndian, &datagram.hlen), "read hlen")
	panicOnError(binary.Read(buffer, binary.BigEndian, &datagram.hops), "read hpos")
	panicOnError(binary.Read(buffer, binary.BigEndian, &datagram.xid), "read xid")
	panicOnError(binary.Read(buffer, binary.BigEndian, &datagram.secs), "read secs")
	panicOnError(binary.Read(buffer, binary.BigEndian, &datagram.flags), "read flags")
	panicOnError(binary.Read(buffer, binary.BigEndian, &datagram.ciaddr), "read ciaddr")
	panicOnError(binary.Read(buffer, binary.BigEndian, &datagram.yiaddr), "read yiaddr")
	panicOnError(binary.Read(buffer, binary.BigEndian, &datagram.siaddr), "read siaddr")
	panicOnError(binary.Read(buffer, binary.BigEndian, &datagram.giaddr), "read giaddr")
	panicOnError(binary.Read(buffer, binary.BigEndian, &datagram.chaddr), "read chaddr")
	panicOnError(binary.Read(buffer, binary.BigEndian, &datagram.sname), "read sname")
	panicOnError(binary.Read(buffer, binary.BigEndian, &datagram.file), "read file")

	buffer.Next(4) // skip Magic cookie
	options := make(map[uint8][]byte)
	for {
		var code, length uint8

		panicOnError(binary.Read(buffer, binary.BigEndian, &code), "read options code")
		if code == 255 { // END Flag
			break
		}

		panicOnError(binary.Read(buffer, binary.BigEndian, &length), "read options length")

		content := buffer.Next(int(length))
		options[code] = content
	}
	datagram.options = options

	return datagram, err
}

func dhcpListener(dhcpLeaseChan chan<- *DHCPLease) {
	addr, err := net.ResolveUDPAddr("udp", ":67")
	exitOnError(err, "resolve udp addr")

	con, err := net.ListenUDP("udp", addr)
	exitOnError(err, "listen udp")

	for {
		rawBuffer := make([]byte, 512)
		n, err := con.Read(rawBuffer)
		exitOnError(err, "error reading from connection")

		verboseLog.Println("new dhcp datagram received")
		if datagram, err := parseDhcpDatagram(rawBuffer[:n]); err != nil {
			log.Printf("parse dhcp datagram error: '%s'\n", err.Error())
		} else {
			messageType, _ := datagram.MessageType()

			if messageType == messageTypes[request] {
				verboseLog.Printf("process datagram with type: '%s' - src mac: '%s'", messageType, datagram.Mac())

				reqIPAddr, err := datagram.RequestedIPAddress()
				if err != nil {
					fmt.Println("RequestIPAddress not found! - error: ", err)
					continue
				}
				hostName, err := datagram.HostName()
				if err != nil {
					fmt.Println("HostName not found! use <UNKNOW> - error: ", err)
					hostName = "<UNKNOW>"
				}

				var expire time.Time
				if *expireBasedFlag {
					expire = time.Now().Add(leaseExpiredDuration)
				}
				lease := &DHCPLease{
					Created: time.Now(),
					Expire:  expire,
					IP:      reqIPAddr,
					Mac:     datagram.Mac(),
					Name:    hostName}
				verboseLog.Printf("trigger new DHCPLease event: ip='%s', mac='%s', name='%s'", lease.IP, lease.Mac, lease.Name)
				dhcpLeaseChan <- lease
			} else {
				verboseLog.Printf("ignore datagram with type: '%s' - src mac: '%s'", messageType, datagram.Mac())
			}
		}

	}
}
