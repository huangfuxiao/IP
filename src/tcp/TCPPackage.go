package tcp

import (

//"encoding/binary"
//"fmt"
//"net"
)

type TCPPackage struct {
	TCPHeader TCPHeader
	Payload   []byte
}

func BuildTCPPacket(payload []byte, header TCPHeader) TCPPackage {
	return TCPPackage{header, payload}
}

func TCPPkgToBuffer(tcpp TCPPackage) []byte {
	buffer := tcpp.TCPHeader.Marshal()
	buffer = append(buffer, tcpp.Payload...)
	return buffer
}

func BufferToTCPPkg(buffer []byte) TCPPackage {
	h := ParseTCPHeader(buffer[0:20])
	index := (int)(4 * h.DataOffset)
	p := buffer[index:]
	return TCPPackage{*h, p}
}

/*
// Read the IpPackage and return it as a string message
func String(tcpp TCPPackage) string {
	return fmt.Sprintf("Source: %v\nDest: %v\npayload: %s", tcpp.TCPHeader.Source, tcpp.TCPHeader.Destination, tcpp.Payload)
}
func Csum(header Header) int {
	buf, _ := header.Marshal()
	l := len(buf)
	sum := uint32(0)
	for i := 0; i < l-1; i += 2 {
		sum += uint32(binary.BigEndian.Uint16(buf[i : i+2]))

	}
	if l%2 == 1 {
		sum += uint32(buf[l])
	}
	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)
	ret := uint16(0xffffffff ^ sum)
	return int(ret)
}
*/
