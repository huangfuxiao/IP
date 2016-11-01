package api

import (
	".././ipv4"
	".././linklayer"
	".././tcp"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
)

type TCB struct {
	Fd         int
	State      tcp.State
	Addr       SockAddr
	seq        int
	ack        int
	RecvBuffer []byte
	SendBuffer []byte
}

type SockAddr struct {
	LocalAddr  string
	LocalPort  int
	RemoteAddr string
	RemotePort int
}

func BuildTCB(fd int) TCB {
	buf := make([]byte, 0)
	s := tcp.State{State: 1}
	add := SockAddr{"0.0.0.0", 0, "0.0.0.0", 0}
	seqn := int(rand.Uint32())
	ackn := 0
	return TCB{fd, s, add, seqn, ackn, buf, buf}
}

func SendCtrlMsg(laddr, raddr string, lport, rport int, ctrl uint8, u linklayer.UDPLink) {
	tcph := tcp.BuildTCPHeader(lport, rport, seq, 0, 2, 0xaaaa)
	data := tcph.Marshal()
	tcph.Checksum = tcp.Csum(data, to4byte(laddr), to4byte(raddr))
	data = tcph.Marshal()
	ipp := ipv4.BuildIpPacket(data, 6, laddr, raddr)
	//Search the interface and send to the actual address and port
	//------------TO DO--------------
	//u.Send(ipp, "localhost", rport)

}

func to4byte(addr string) [4]byte {
	parts := strings.Split(addr, ".")
	b0, err := strconv.Atoi(parts[0])
	if err != nil {
		log.Fatalf("to4byte: %s (latency works with IPv4 addresses only, but not IPv6!)\n", err)
	}
	b1, _ := strconv.Atoi(parts[1])
	b2, _ := strconv.Atoi(parts[2])
	b3, _ := strconv.Atoi(parts[3])
	return [4]byte{byte(b0), byte(b1), byte(b2), byte(b3)}
}

/*
func receiveSynAck(laddr, raddr string) {
	for {
		buf := make([]byte, 1024)
		numRead, raddr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Fatalf("ReadFrom: %s\n", err)
		}
		if raddr.String() != remoteAddress {
			// this is not the packet we are looking for
			continue
		}
		tcp := NewTCPHeader(buf[:numRead])
		// Closed port gets RST, open port gets SYN ACK
		if tcp.HasFlag(RST) || (tcp.HasFlag(SYN) && tcp.HasFlag(ACK)) {
			break
		}
	}
}
*/
