package ipv4

import (
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
