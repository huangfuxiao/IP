package tcp

/*
Copied and modified from
https://github.com/grahamking/latency/blob/07ba2b7cc2210227cdb9afbec07e65fcfcd2c968/tcp.go
*/

import (
	"bytes"
	"encoding/binary"
)

const (
	NOTHING = 0
	FIN     = 1  // 00 0001
	SYN     = 2  // 00 0010
	RST     = 4  // 00 0100
	PSH     = 8  // 00 1000
	ACK     = 16 // 01 0000
	URG     = 32 // 10 0000
)

type TCPHeader struct {
	Source      uint16
	Destination uint16
	SeqNum      uint32
	AckNum      uint32
	DataOffset  uint8 // 4 bits
	Reserved    uint8 // 3 bits
	ECN         uint8 // 3 bits
	Ctrl        uint8 // 6 bits
	Window      uint16
	Checksum    uint16 // Kernel will set this if it's 0
	Urgent      uint16
	Options     []TCPOption
}

type TCPOption struct {
	Kind   uint8
	Length uint8
	Data   []byte
}

func BuildTCPHeader(src int, dst int, seq int, ack int, ctrl int, ws int) TCPHeader {
	header := TCPHeader{
		Source:      uint16(src),
		Destination: uint16(dst),
		SeqNum:      uint32(seq),
		AckNum:      uint32(ack),
		DataOffset:  uint8(5),
		ECN:         uint8(0),
		Ctrl:        uint8(ctrl),
		Reserved:    uint8(0),
		Window:      uint16(ws),
		Checksum:    uint16(0),
		Urgent:      uint16(0),
		Options:     []TCPOption{},
	}
	return header
}

// Parse packet into TCPHeader structure
func ParseTCPHeader(data []byte) *TCPHeader {
	var tcp TCPHeader
	r := bytes.NewReader(data)
	binary.Read(r, binary.BigEndian, &tcp.Source)
	binary.Read(r, binary.BigEndian, &tcp.Destination)
	binary.Read(r, binary.BigEndian, &tcp.SeqNum)
	binary.Read(r, binary.BigEndian, &tcp.AckNum)

	var mix uint16
	binary.Read(r, binary.BigEndian, &mix)
	tcp.DataOffset = byte(mix >> 12)  // top 4 bits
	tcp.Reserved = byte(mix >> 9 & 7) // 3 bits
	tcp.ECN = byte(mix >> 6 & 7)      // 3 bits
	tcp.Ctrl = byte(mix & 0x3f)       // bottom 6 bits

	binary.Read(r, binary.BigEndian, &tcp.Window)
	binary.Read(r, binary.BigEndian, &tcp.Checksum)
	binary.Read(r, binary.BigEndian, &tcp.Urgent)

	return &tcp
}

func (tcp *TCPHeader) HasFlag(flagBit byte) bool {
	return tcp.Ctrl&flagBit != 0
}

func (tcp *TCPHeader) Marshal() []byte {

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, tcp.Source)
	binary.Write(buf, binary.BigEndian, tcp.Destination)
	binary.Write(buf, binary.BigEndian, tcp.SeqNum)
	binary.Write(buf, binary.BigEndian, tcp.AckNum)

	var mix uint16
	mix = uint16(tcp.DataOffset)<<12 | // top 4 bits
		uint16(tcp.Reserved)<<9 | // 3 bits
		uint16(tcp.ECN)<<6 | // 3 bits
		uint16(tcp.Ctrl) // bottom 6 bits
	binary.Write(buf, binary.BigEndian, mix)

	binary.Write(buf, binary.BigEndian, tcp.Window)
	binary.Write(buf, binary.BigEndian, tcp.Checksum)
	binary.Write(buf, binary.BigEndian, tcp.Urgent)

	for _, option := range tcp.Options {
		binary.Write(buf, binary.BigEndian, option.Kind)
		if option.Length > 1 {
			binary.Write(buf, binary.BigEndian, option.Length)
			binary.Write(buf, binary.BigEndian, option.Data)
		}
	}

	out := buf.Bytes()

	// Pad to min tcp header size, which is 20 bytes (5 32-bit words)

	if len(out) > 20 {
		pad := 24 - len(out)
		for i := 0; i < pad; i++ {
			out = append(out, 0)
		}
	}
	/*
		pad := 20 - len(out)
		for i := 0; i < pad; i++ {
			out = append(out, 0)
		}
	*/

	return out
}

// TCP Checksum
func Csum(data []byte, srcip, dstip [4]byte) uint16 {

	pseudoHeader := []byte{
		srcip[0], srcip[1], srcip[2], srcip[3],
		dstip[0], dstip[1], dstip[2], dstip[3],
		0,    // zero
		6,    // protocol number (6 == TCP)
		0, 0, // TCP length (16 bits), not inc pseudo header
	}

	binary.BigEndian.PutUint16(pseudoHeader[10:12], uint16(len(data)))

	sumThis := make([]byte, 0, len(pseudoHeader)+len(data))
	sumThis = append(sumThis, pseudoHeader...)
	sumThis = append(sumThis, data...)
	//fmt.Printf("% x\n", sumThis)

	lenSumThis := len(sumThis)
	var nextWord uint16
	var sum uint32
	for i := 0; i+1 < lenSumThis; i += 2 {
		nextWord = uint16(sumThis[i])<<8 | uint16(sumThis[i+1])
		sum += uint32(nextWord)
	}
	if lenSumThis%2 != 0 {
		//fmt.Println("Odd byte")
		sum += uint32(sumThis[len(sumThis)-1])
	}

	// Add back any carry, and any carry from adding the carry
	sum = (sum >> 16) + (sum & 0xffff)
	sum = sum + (sum >> 16)

	// Bitwise complement
	return uint16(^sum)
}
