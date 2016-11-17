package api

import (
	".././ipv4"
	".././linklayer"
	".././pkg"
	".././tcp"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type TCB struct {
	Fd          int
	State       tcp.State
	Addr        SockAddr
	Seq         int
	Ack         int
	RecvW       RecvWindow
	SendW       SendWindow
	node        *pkg.Node
	u           linklayer.UDPLink
	PIFCheck    map[int]*PkgInFlight
	Check       map[int]bool
	BlockWrite  bool
	BlockRead   bool
	ShouldClose bool
}

type PkgInFlight struct {
	Length int
	Count  int
	Ipp    ipv4.IpPackage
	Addr   string
	Port   int
}

type SockAddr struct {
	LocalAddr  string
	LocalPort  int
	RemoteAddr string
	RemotePort int
}

func BuildTCB(fd int, node *pkg.Node, u linklayer.UDPLink) TCB {
	//PIF := make([]tcp.TCPPackage)
	PIF := make(map[int]*PkgInFlight)
	Rw := BuildRecvWindow()
	Sw := BuildSendWindow()
	ch := make(map[int]bool)
	s := tcp.State{State: 1}
	add := SockAddr{"0.0.0.0", 0, "0.0.0.0", 0}
	seqn := int(rand.Uint32())
	ackn := 0
	return TCB{fd, s, add, seqn, ackn, Rw, Sw, node, u, PIF, ch, false, false, false}
}

func (tcb *TCB) SendCtrlMsg(ctrl int, c bool, notestb bool, ws int) {
	if c {
		tcb.Check[tcb.Seq] = false
	}

	taddr := tcb.Addr
	tcph := tcp.BuildTCPHeader(taddr.LocalPort, taddr.RemotePort, tcb.Seq, tcb.Ack, ctrl, ws)
	data := tcph.Marshal()
	tcph.Checksum = tcp.Csum(data, To4byte(taddr.LocalAddr), To4byte(taddr.RemoteAddr))
	data = tcph.Marshal()
	/*
		ipp := ipv4.BuildIpPacket(data, 6, taddr.LocalAddr, taddr.RemoteAddr)
		fmt.Println(ipp)
	*/
	//Search the interface and send to the actual address and port
	//------------TO DO--------------
	if notestb {
		tcb.Seq += 1
	}
	v, ok := tcb.node.RouteTable[taddr.RemoteAddr]
	if ok {
		for _, link := range tcb.node.InterfaceArray {
			if strings.Compare(v.Next, link.Src) == 0 {
				if link.Status == 0 {
					return
				}

				ipPkt := ipv4.BuildIpPacket(data, 6, taddr.LocalAddr, taddr.RemoteAddr)
				//fmt.Println(ipPkt.IpHeader.TTL)
				//fmt.Println(ipPkt.IpHeader.Protocol)
				tcb.u.Send(ipPkt, link.RemoteAddr, link.RemotePort)
				if c {
					go CheckACK(tcb.Seq-1, tcb, 0, ipPkt, link.RemoteAddr, link.RemotePort)
				}

				return
			}
		}

	}

}

func (tcb *TCB) SendDataThread() {
	for {
		// fmt.Println("Last Byte Sent ", tcb.SendW.LastByteSent)
		// fmt.Println("Last Byte Written ", tcb.SendW.LastByteWritten)
		time.Sleep(500 * time.Millisecond)
		if tcb.SendW.WSback {
			if tcb.SendW.LastByteSent+tcb.SendW.Size == tcb.SendW.LastByteWritten {
				fmt.Println("continue-----------")
				continue
			}
		} else {
			if tcb.SendW.LastByteSent == tcb.SendW.LastByteWritten {
				continue
			}
		}
		bytesToSent := tcb.SendW.EffectiveWindow()
		fmt.Println("effective window ", bytesToSent)
		payload := make([]byte, 0)

		i := 0
		for i < bytesToSent {
			if !tcb.SendW.WSback {
				if tcb.SendW.LastByteSent == tcb.SendW.LastByteWritten {
					break
				}
			}

			//fmt.Println("index of payload ", i)
			//fmt.Println("sendbuff index ", tcb.SendW.LastByteSent)
			payload = append(payload, tcb.SendW.SendBuffer[tcb.SendW.LastByteSent])
			tcb.SendW.LastByteSent++
			if tcb.SendW.LastByteSent >= tcb.SendW.Size {
				tcb.SendW.WSback = false
				tcb.SendW.Back = true
				tcb.SendW.LastByteSent -= tcb.SendW.Size
			}
			i++
		}
		//fmt.Println("payload ", payload)
		tcb.SendData(payload, tcb.RecvW.AdvertisedWindow())
		tcb.SendW.BytesInFlight += len(payload)
		//fmt.Println("recent bytes in flight: ", tcb.SendW.BytesInFlight)
		//fmt.Println("lastbyteSent ", tcb.SendW.LastByteSent)
	}
}

// Send actual data to the receiver
func (tcb *TCB) SendData(payload []byte, ws int) {
	//tcb.Check[tcb.Seq + len(payload)] = false
	taddr := tcb.Addr
	tcph := tcp.BuildTCPHeader(taddr.LocalPort, taddr.RemotePort, tcb.Seq, tcb.Ack, 0, ws)
	tcpp := tcp.BuildTCPPacket(payload, tcph)
	data := tcp.TCPPkgToBuffer(tcpp)
	tcph.Checksum = tcp.Csum(data, To4byte(taddr.LocalAddr), To4byte(taddr.RemoteAddr))
	tcpp2 := tcp.BuildTCPPacket(payload, tcph)
	tcpbuf := tcp.TCPPkgToBuffer(tcpp2)
	tcb.Seq += len(payload)

	v, ok := tcb.node.RouteTable[taddr.RemoteAddr]
	if ok {
		for _, link := range tcb.node.InterfaceArray {
			if strings.Compare(v.Next, link.Src) == 0 {
				if link.Status == 0 {
					return
				}
				ipPkt := ipv4.BuildIpPacket(tcpbuf, 6, taddr.LocalAddr, taddr.RemoteAddr)
				tcb.u.Send(ipPkt, link.RemoteAddr, link.RemotePort)
				tcb.PIFCheck[tcb.Seq] = &PkgInFlight{len(payload), 0, ipPkt, link.RemoteAddr, link.RemotePort}
				//if c {
				// fmt.Println("Check seq after, ", tcb.Seq-len(payload))
				// go CheckACK(tcb.Seq, tcb, 0, ipPkt, link.RemoteAddr, link.RemotePort)
				//}
				//fmt.Println("recent PIF ", tcb.PIFCheck)

				return
			}
		}

	}

}

func To4byte(addr string) [4]byte {
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

func CheckACK(idx int, tcb *TCB, count int, ipp ipv4.IpPackage, addr string, port int) {
	for {
		//fmt.Println("Count ", count)
		time.Sleep(3000 * time.Millisecond)
		if count == 3 {
			break
		}
		if tcb.Check[idx] == true {
			break
		} else {
			tcb.u.Send(ipp, addr, port)
		}
		//fmt.Println("Check thread and flag ctrl thread ---------------------")
		//fmt.Println("current seq in thread ", idx)
		count += 1

	}

}

func (tcb *TCB) DataACKThread() {
	for {
		time.Sleep(3000 * time.Millisecond)
		for k, v := range tcb.PIFCheck {
			fmt.Println("retransmission seq ", k)
			if v.Count < 3 {
				tcb.u.Send(v.Ipp, v.Addr, v.Port)
				v.Count += 1
				//fmt.Println("count, ", v.Count)

			} else {
				tcb.ShouldClose = true
			}

		}

	}
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
