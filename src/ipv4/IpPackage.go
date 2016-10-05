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

/*
func (ipp *IpPackage) String() string {
	if h == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ver=%d hdrlen=%d tos=%#x totallen=%d id=%#x flags=%#x fragoff=%#x ttl=%d proto=%d cksum=%#x src=%v dst=%v", h.Version, h.Len, h.TOS, h.TotalLen, h.ID, h.Flags, h.FragOff, h.TTL, h.Protocol, h.Checksum, h.Src, h.Dst)
}*/
