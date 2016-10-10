package ipv4

import (
	"encoding/binary"
	"fmt"
	"net"
)

type IpPackage struct {
	IpHeader Header
	Payload  []byte
}

func BuildIpPacket(payload []byte, protocol int, src string, dest string) IpPackage {
	header := Header{
		Version:  4,
		Len:      20,
		TOS:      0,
		TotalLen: 20 + len(payload),
		TTL:      16,
		Protocol: protocol,
		Src:      net.ParseIP(src),
		Dst:      net.ParseIP(dest),
		Options:  []byte{},
	}
	header.Checksum = Csum(header)
	return IpPackage{header, payload}
}

// Read the IpPackage and return it as a string message
func String(ipp IpPackage) string {
	return fmt.Sprintf("src_ip: %v\ndst_ip: %v\nbody_len: %d\nheader:\n    tos:   %#x\n    id:    %#x\n    prot:   %d\npayload: %s", ipp.IpHeader.Src, ipp.IpHeader.Dst, ipp.IpHeader.TotalLen-20, ipp.IpHeader.TOS, ipp.IpHeader.ID, ipp.IpHeader.Protocol, ipp.Payload)
}

// Change the IpPackage into buffer and ready for UDP transmission
func IpPkgToBuffer(ipp IpPackage) []byte {
	buffer, e := ipp.IpHeader.Marshal()
	if buffer == nil {
		fmt.Println(e)
	}
	buffer = append(buffer, ipp.Payload...)
	return buffer
}

// Read the buffer from UDP transmission and change it into a IpPackage
func BufferToIpPkg(buffer []byte) IpPackage {
	h, e := ParseHeader(buffer[0:20])
	if h == nil {
		fmt.Println(e)
	}
	p := buffer[20:]
	return IpPackage{*h, p}
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
