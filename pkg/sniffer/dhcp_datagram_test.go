package sniffer

import (
	"encoding/hex"
	"testing"
)

func TestInvalidDatagram(t *testing.T) {
	if _, err := DHCPDatagramFromBytes([]byte("00000")); err == nil {
		t.Error("error expected")
	}
}

func TestParseValidDatagram(t *testing.T) {
	rawTelegram, _ := hex.DecodeString("010106002ccab3380000000000000000000000000000000000000000080027f2975a" +
		"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
		"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
		"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
		"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
		"638253633501033204c0a8016f0c0b64656269616e2d6c7864653712011c02030f06770c2c2f1a792a79f921fc2aff0000000" +
		"000000000000000000000000000")

	datagram, err := DHCPDatagramFromBytes(rawTelegram)
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	expectedMac := "08:00:27:f2:97:5a"
	if datagram.Mac != expectedMac {
		t.Errorf("mac '%s' expected - '%s' received", expectedMac, datagram.Mac)
	}

	expectedHostName := "debian-lxde"
	if hostName, err := datagram.Host(); err != nil {
		t.Errorf("unexpected error: '%s'", err)
	} else if hostName != expectedHostName {
		t.Errorf("host name '%s' expected - '%s' received", expectedHostName, hostName)
	}

	expectedMessageType := DHCPRequest
	if datagram.MessageType != expectedMessageType {
		t.Errorf("message type '%s' expected - '%s' received", expectedMessageType, datagram.MessageType)
	}

	expectedRequestedIPAddress := "192.168.1.111"
	if requestedIPAddress, err := datagram.IP(); err != nil {
		t.Errorf("unexpected error: '%s'", err)
	} else if requestedIPAddress != expectedRequestedIPAddress {
		t.Errorf("ip address '%s' expected - '%s' received", expectedRequestedIPAddress, requestedIPAddress)
	}

}
