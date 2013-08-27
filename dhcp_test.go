package main

import (
	"encoding/hex"
	"testing"
)

func TestParseDHCPDatagram(t *testing.T) {
	//
	// invalid telegram
	//
	if _, err := parseDhcpDatagram([]byte("00000")); err == nil {
		t.Error("error expected")
	}

	//
	// valid telegram
	//
	rawTelegram, _ := hex.DecodeString("010106002ccab3380000000000000000000000000000000000000000080027f2975a" +
		"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
		"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
		"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
		"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
		"638253633501033204c0a8016f0c0b64656269616e2d6c7864653712011c02030f06770c2c2f1a792a79f921fc2aff0000000" +
		"000000000000000000000000000")
	datagram, err := parseDhcpDatagram(rawTelegram)
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	expectedMac := "08:00:27:f2:97:5a"
	if datagram.Mac() != expectedMac {
		t.Errorf("mac '%s' expected - '%s' received", expectedMac, datagram.Mac())
	}

	expectedHostName := "debian-lxde"
	if hostName, err := datagram.HostName(); err != nil {
		t.Errorf("unexpected error: '%s'", err)
	} else if hostName != expectedHostName {
		t.Errorf("host name '%s' expected - '%s' received", expectedHostName, hostName)
	}

	expectedMessageType := messageTypes[request]
	if messageType, err := datagram.MessageType(); err != nil {
		t.Errorf("unexpected error: '%s'", err)
	} else if messageType != expectedMessageType {
		t.Errorf("message type '%s' expected - '%s' received", expectedMessageType, messageType)
	}

	expectedRequestedIPAddress := "192.168.1.111"
	if requestedIPAddress, err := datagram.RequestedIPAddress(); err != nil {
		t.Errorf("unexpected error: '%s'", err)
	} else if requestedIPAddress != expectedRequestedIPAddress {
		t.Errorf("ip address '%s' expected - '%s' received", expectedRequestedIPAddress, requestedIPAddress)
	}

}
