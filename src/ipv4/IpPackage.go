package ipv4

import (
	"net"
)

type IpPackage struct {
	IpHeader Header
	Payload  []byte
}

func BuildIpPacket(payload []byte, protocol int, src string, dest string) IpPackage {
	header := Header{
		Version: 4,
		Len: len(payload),
		TOS: 0,
		TotalLen: 20 + len(payload),
		TTL: 16,
		Protocol: protocol,
		Src: net.ParseIP(src),
		Dst: net.ParseIP(dest),
		Options: []byte{}
	}
	return IpPackage(header, payload)
}
