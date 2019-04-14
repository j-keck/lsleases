package sniffer

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type rawDatagram struct {
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

func rawDatagramFromBytes(bs []byte) (rawDatagram, error) {

	fields := rawDatagram{}
	buffer := bytes.NewBuffer(bs)

	ebr := &errBufferReader{buffer: buffer}
	ebr.read(&fields.op, "op")
	ebr.read(&fields.htype, "htype")
	ebr.read(&fields.hlen, "hlen")
	ebr.read(&fields.hops, "hpos")
	ebr.read(&fields.xid, "xid")
	ebr.read(&fields.secs, "secs")
	ebr.read(&fields.flags, "flags")
	ebr.read(&fields.ciaddr, "ciaddr")
	ebr.read(&fields.yiaddr, "yiaddr")
	ebr.read(&fields.siaddr, "siaddr")
	ebr.read(&fields.giaddr, "giaddr")
	ebr.read(&fields.chaddr, "chaddr")
	ebr.read(&fields.sname, "sname")
	ebr.read(&fields.file, "file")

	if ebr.err != nil {
		return rawDatagram{}, ebr.err
	}

	buffer.Next(4) // skip Magic cookie
	options := make(map[uint8][]byte)
	for {
		var code, length uint8

		ebr.read(&code, "options code")
		if code == 255 { // END Flag
			break
		}

		ebr.read(&length, "options length")
		if ebr.err != nil {
			return rawDatagram{}, ebr.err
		}

		content := buffer.Next(int(length))
		options[code] = content
	}
	fields.options = options

	return fields, nil
}

type errBufferReader struct {
	buffer *bytes.Buffer
	err    error
}

func (ebr *errBufferReader) read(data interface{}, elem string) {
	if ebr.err != nil {
		return
	}

	err := binary.Read(ebr.buffer, binary.BigEndian, data)
	if err != nil {
		ebr.err = fmt.Errorf("read %s - %s", elem, err.Error())
	}
}
