package handler

import (
	"../api"
	"../ipv4"
	"../linklayer"
	"../pkg"
	"../tcp"
	"fmt"
	"strings"
	"sync"
	"time"
)

//Convert RIP to IP Packet
func ConvertRipToIpPackage(rip ipv4.RIP, src string, dest string) ipv4.IpPackage {
	b := ipv4.ConvertRipToBytes(rip)
	ipPkt := ipv4.BuildIpPacket(b, 200, src, dest)
	return ipPkt
}

//Send Trigger Updates using RIP to All of Node's Neighbors
func SendTriggerUpdates(destIpAddr string, route pkg.Entry, node *pkg.Node, u linklayer.UDPLink) {
	learnFrom := node.GetLearnFrom(route.Next)
	for _, link := range node.InterfaceArray {
		if link.Status == 0 {
			continue
		}

		var newRip ipv4.RIP
		newRip.Command = 2
		if learnFrom == link.Dest {
			newRip.Entries = append(newRip.Entries, ipv4.RIPEntry{Cost: 16, Address: destIpAddr})
		} else {
			newRip.Entries = append(newRip.Entries, ipv4.RIPEntry{Cost: route.Cost, Address: destIpAddr})
		}
		newRip.NumEntries = len(newRip.Entries)

		ipPkt := ConvertRipToIpPackage(newRip, link.Src, link.Dest)
		u.Send(ipPkt, link.RemoteAddr, link.RemotePort)
		//fmt.Printf("Trigger update RIP sent to this address: %s %d \n", link.RemoteAddr, link.RemotePort)
	}
}

//IP is not locally arrived
func ForwardIpPackage(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink, mutex *sync.RWMutex) {
	dstIpAddr := ipPkt.IpHeader.Dst.String()

	//Loop through node.RouteTable, and forward to upper node
	mutex.RLock()
	for k, v := range node.RouteTable {
		if len(k) < 0 {
			fmt.Println("The rounting table currently has no routes!\n")
			mutex.RUnlock()
			return
		}
		//Destination matches one on this Node's Rounting Table
		if strings.Compare(dstIpAddr, v.Dest) == 0 {

			//Find corresponding interface's remote physic address and port
			for _, link := range node.InterfaceArray {
				if strings.Compare(v.Next, link.Src) == 0 {
					//Arrives the target interface
					//Check status first
					if link.Status == 0 {
						mutex.RUnlock()
						//fmt.Println("Interface is down. Packet has to be dropped\n")
						return
					}

					//Check cost is not infinity
					if v.Cost >= 16 {
						mutex.RUnlock()
						//fmt.Println("Inifinity loop. Packet has to be dropped\n")
						return
					}

					//Forward ip packet to next node
					ipPkt.IpHeader.TTL--
					ipPkt.IpHeader.Checksum = 0
					ipPkt.IpHeader.Checksum = ipv4.Csum(ipPkt.IpHeader)
					u.Send(ipPkt, link.RemoteAddr, link.RemotePort)
					mutex.RUnlock()
					//fmt.Printf("Node can get through this interface: %s to %s with cost: %d\n", v.Next, v.Dest, v.Cost)
					return
				}
			}
		}
	}
	mutex.RUnlock()
	//fmt.Printf("Cannot find a interface in this node's routing table. Packet has to be dropped\n")
	return
}

//IP protocol=0
func RunDataHandler(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink) {
	data := ipv4.String(ipPkt)
	fmt.Println("Driver Received Packet")
	fmt.Println(data)
}

//IP protocol=200
func RunRIPHandler(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink, mutex *sync.RWMutex) {
	dstIpAddr := ipPkt.IpHeader.Dst.String()
	srcIpAddr := ipPkt.IpHeader.Src.String()
	payLoad := ipPkt.Payload

	for _, link := range node.InterfaceArray {
		if strings.Compare(srcIpAddr, link.Dest) == 0 {

			//Arrive the interface
			if link.Status == 0 {
				//fmt.Println("Interface is down. Packet has to be dropped\n")
				return
			}
			rip := ipv4.ConvertBytesToRIP(payLoad)

			//RIP Request
			if rip.Command == 1 {

				//First, put all of this node.RouteTable into RIP and send back
				var newRip ipv4.RIP
				newRip.Command = 2
				newRip.NumEntries = 0
				//put all of this node's RT's entries to RIP
				mutex.RLock()
				for _, v := range node.RouteTable {

					if newRip.NumEntries == 64 {
						ipPkt := ConvertRipToIpPackage(newRip, link.Src, link.Dest)
						u.Send(ipPkt, link.RemoteAddr, link.RemotePort)
						newRip.NumEntries = 0
						newRip.Entries = nil
					}

					newRip.Entries = append(newRip.Entries, ipv4.RIPEntry{Cost: v.Cost, Address: v.Dest})
					newRip.NumEntries++

				}
				mutex.RUnlock()
				//send back RIP to src
				ipPkt := ConvertRipToIpPackage(newRip, link.Src, srcIpAddr)
				u.Send(ipPkt, link.RemoteAddr, link.RemotePort)
				//fmt.Printf("RIP response sent back to this address: %s %d \n", link.RemoteAddr, link.RemotePort)

				/* Then, insert this neighbor to the node.RouteTable with cost = 1 */
				//fmt.Println("IPPackage arrived after rip.Command==1\n")
				node.RouteTable[srcIpAddr] = pkg.Entry{Dest: srcIpAddr, Next: dstIpAddr, Cost: 1, Ttl: time.Now().Unix() + 12}
				SendTriggerUpdates(srcIpAddr, node.RouteTable[srcIpAddr], node, u)
				return

			} else if rip.Command == 2 {
				/* First, insert this neighbor to the node.RouteTable with cost = 1 */
				v, ok := node.RouteTable[srcIpAddr]
				if ok {
					if v.Cost > 1 {
						node.RouteTable[srcIpAddr] = pkg.Entry{Dest: srcIpAddr, Next: dstIpAddr, Cost: 1, Ttl: time.Now().Unix() + 12}
						SendTriggerUpdates(srcIpAddr, node.RouteTable[srcIpAddr], node, u)
					}
				}
				/* Then, loop through all of the rip's entry
				   Compare if the RIPEntry's Address already on this node.RouteTable
				   If so, compare the cost
				*/
				for _, entry := range rip.Entries {
					value, ok := node.RouteTable[entry.Address]
					mutex.Lock()
					if ok {
						/*Check poison first*/

						learnFrom := node.GetLearnFrom(value.Next)
						if (entry.Cost == 16) && (learnFrom == srcIpAddr) && (entry.Address == value.Dest) && (value.Next != value.Dest) {
							time := node.RouteTable[entry.Address].Ttl
							node.RouteTable[entry.Address] = pkg.Entry{Dest: entry.Address, Next: dstIpAddr, Cost: 16, Ttl: time}
							//fmt.Println("111111111111111111111111")
							SendTriggerUpdates(entry.Address, node.RouteTable[entry.Address], node, u)
						} else if (entry.Cost + 1) < value.Cost {
							//fmt.Println("222222222222")
							//fmt.Println(node.RouteTable[entry.Address])
							node.RouteTable[entry.Address] = pkg.Entry{Dest: entry.Address, Next: dstIpAddr, Cost: entry.Cost + 1, Ttl: time.Now().Unix() + 12}
							SendTriggerUpdates(entry.Address, node.RouteTable[entry.Address], node, u)
						} else if (entry.Cost+1) == value.Cost && (learnFrom == srcIpAddr) && (entry.Address == value.Dest) && (value.Next != value.Dest) {
							dst := node.RouteTable[entry.Address].Dest
							next := node.RouteTable[entry.Address].Next
							cost := node.RouteTable[entry.Address].Cost
							//fmt.Println("3333333333333")
							//fmt.Println(node.RouteTable[entry.Address])
							node.RouteTable[entry.Address] = pkg.Entry{Dest: dst, Next: next, Cost: cost, Ttl: time.Now().Unix() + 12}
						}
					} else {
						//fmt.Println(node.RouteTable[entry.Address])
						//fmt.Println("4444444444444")

						node.RouteTable[entry.Address] = pkg.Entry{Dest: entry.Address, Next: dstIpAddr, Cost: entry.Cost + 1, Ttl: time.Now().Unix() + 12}
						SendTriggerUpdates(entry.Address, node.RouteTable[entry.Address], node, u)
					}
					mutex.Unlock()
				}

				return
			} else {
				fmt.Println("Unrecognized RIP protocol!\n")
				return
			}
		}
	}
	return
}

//IP protocol=6
func RunTCPHandler(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink, mutex *sync.RWMutex, manager *api.SocketManager) {
	//fmt.Println("RECEIVE !!!!!!!!")

	dstIpAddr := ipPkt.IpHeader.Dst.String()
	srcIpAddr := ipPkt.IpHeader.Src.String()

	tcpPkt := tcp.BufferToTCPPkg(ipPkt.Payload)
	tcpHeader := tcpPkt.TCPHeader

	tcpPayload := tcpPkt.Payload
	dstPort := int(tcpHeader.Destination)
	srcPort := int(tcpHeader.Source)
	//fmt.Println("tcp saddr ", srcIpAddr, srcPort, dstIpAddr, dstPort)
	//fmt.Println("tcp get ctrl", tcpHeader.Ctrl)

	//Check tcp Check sum
	tempbuff := ipPkt.Payload
	tempbuff[16] = 0
	tempbuff[17] = 0
	recvCheckSum := tcp.Csum(tempbuff, api.To4byte(srcIpAddr), api.To4byte(dstIpAddr))
	if tcpHeader.Checksum != recvCheckSum {
		fmt.Println(tcpHeader.Checksum)
		fmt.Println(recvCheckSum)

		fmt.Println("tcp Checksum error!")
		return
	}

	ws := int(tcpHeader.Window)

	// fmt.Println("receive seqnum and acknum: ", tcpHeader.SeqNum, tcpHeader.AckNum)
	// fmt.Println("receive payload ", tcpPayload)
	if len(tcpPayload) == 0 {
		if tcpHeader.HasFlag(tcp.SYN) && !tcpHeader.HasFlag(tcp.ACK) {
			//manager.V_accept()
			//fmt.Println("receive syn only")
			saddr := api.SockAddr{dstIpAddr, dstPort, "0.0.0.0", 0}
			tcb, ok := manager.AddrToSocket[saddr]
			if ok && tcb.State.State == 2 {
				saddr.RemoteAddr = srcIpAddr
				saddr.RemotePort = srcPort
				_, ok2 := manager.AddrToSocket[saddr]
				if !ok2 {
					newfd := manager.V_socket(node, u)
					newtcb := manager.FdToSocket[newfd]
					newtcb.Addr = saddr
					newState, cf := tcp.StateMachine(tcb.State.State, tcp.SYN, "")
					if newState == 0 {
						return
					}
					//fmt.Println("return flag", cf)
					newtcb.State.State = newState
					newtcb.Ack = int(tcpHeader.SeqNum) + 1
					newtcb.SendW.AdvertisedWindow = ws
					manager.AddrToSocket[saddr] = newtcb

					newtcb.SendCtrlMsg(cf, true, true, tcb.RecvW.AdvertisedWindow())
				}

			}
		} else if tcpHeader.HasFlag(tcp.SYN) && tcpHeader.HasFlag(tcp.ACK) {
			//fmt.Println("receive syn and ack")
			saddr := api.SockAddr{dstIpAddr, dstPort, srcIpAddr, srcPort}
			tcb, ok := manager.AddrToSocket[saddr]
			//fmt.Println("current seqnum and ack num : ", tcb.Seq, tcb.Ack)
			if ok {
				newState, cf := tcp.StateMachine(tcb.State.State, tcp.SYN+tcp.ACK, "")
				if newState == 0 {
					return
				}
				tcb.State.State = newState
				tcb.Ack = int(tcpHeader.SeqNum) + 1
				tcb.RecvW.LastSeq = int(tcpHeader.SeqNum) + 1
				tcb.SendW.AdvertisedWindow = ws
				tcb.Check[int(tcpHeader.AckNum-1)] = true
				//fmt.Println("receive syn and ack ", tcb.Seq-1)
				tcb.SendCtrlMsg(cf, false, false, tcb.RecvW.AdvertisedWindow())
				go tcb.SendDataThread()
				//go tcb.DataACKThread()

			}
		} else if tcpHeader.HasFlag(tcp.FIN) {
			//fmt.Println("tcpHeader seqnum and ack num : ", int(tcpHeader.SeqNum), int(tcpHeader.AckNum))
			//fmt.Println("tcpHeader: ", tcpHeader)

			saddr := api.SockAddr{dstIpAddr, dstPort, srcIpAddr, srcPort}
			tcb, ok := manager.AddrToSocket[saddr]
			if ok {
				newState, cf := tcp.StateMachine(tcb.State.State, tcp.FIN, "")
				if newState == 0 {
					return
				}
				tcb.State.State = newState
				//fmt.Println("This is the new state after FIN: ", newState)
				tcb.Ack = int(tcpHeader.SeqNum) + 1
				tcb.Check[int(tcpHeader.AckNum-1)] = true
				//fmt.Println("receive syn and ack ", tcb.Seq-1)
				tcb.SendCtrlMsg(cf, false, false, tcb.RecvW.AdvertisedWindow())

				if tcb.State.State == tcp.TIMEWAIT {
					go manager.TimeWaitTimeOut(tcb, 10000)
				}

				//Reset tcb seq num; works for now but not sure correctness
				//tcb.Seq = int(tcpHeader.AckNum)
			}
		} else if tcpHeader.HasFlag(tcp.ACK) {
			//fmt.Println("receive ack only")
			//fmt.Println("receive ack seq num ", tcpHeader.AckNum)
			saddr := api.SockAddr{dstIpAddr, dstPort, srcIpAddr, srcPort}
			tcb, ok := manager.AddrToSocket[saddr]
			//fmt.Println("current seqnum and ack num : ", tcb.Seq, int(tcpHeader.AckNum))
			if ok {
				if tcb.State.State == tcp.SYNRCVD {
					tcb.RecvW.LastSeq = int(tcpHeader.SeqNum)
					if tcb.Seq == int(tcpHeader.AckNum) {
						tcb.Ack = int(tcpHeader.SeqNum)
						tcb.Check[int(tcpHeader.AckNum-1)] = true
						newState, cf := tcp.StateMachine(tcb.State.State, tcp.ACK, "")
						if newState == 0 {
							return
						}
						tcb.State.State = newState
						tcb.SendW.AdvertisedWindow = ws
						if cf != 0 {
							tcb.SendCtrlMsg(cf, false, true, 65535)
						}
						go tcb.SendDataThread()
						//go tcb.DataACKThread()
					}

				} else if tcb.State.State == 5 {
					// fmt.Println("reach here with idx ", int(tcpHeader.AckNum))
					// fmt.Println("recent PIF ", tcb.PIFCheck)
					//fmt.Println("receive seq and recent ack ", tcpHeader.SeqNum, tcb.Ack)

					//fmt.Println("receive ack ", tcpHeader.AckNum)
					tcb.PIFCheck.Mutex.RLock()
					_, ok := tcb.PIFCheck.PIF[tcpHeader.AckNum]
					tcb.PIFCheck.Mutex.RUnlock()
					if ok {
						tcb.SendW.AdvertisedWindow = ws
						for k, v := range tcb.PIFCheck.PIF {
							//tcb.PIFCheck.Mutex.RLock()

							if k <= tcpHeader.AckNum {
								//fmt.Println("find PIFCheck ack ", k)
								tcb.SendW.Mutex.Lock()
								tcb.SendW.BytesInFlight -= v.Length
								tcb.SendW.LastByteAcked += v.Length
								if tcb.SendW.LastByteAcked >= tcb.SendW.Size {
									tcb.SendW.WAback = false
									tcb.SendW.Back = false
									tcb.SendW.LastByteAcked -= tcb.SendW.Size
								}
								tcb.SendW.Mutex.Unlock()
								//tcb.PIFCheck.Mutex.RUnlock()
								tcb.PIFCheck.Mutex.Lock()
								delete(tcb.PIFCheck.PIF, k)
								tcb.PIFCheck.Mutex.Unlock()
							}

						}
					}

					// fmt.Println("bytes can be written now", tcb.SendW.BytesCanBeWritten())
					// fmt.Println("pif ", tcb.PIFCheck)
					// fmt.Println("bytes in flight after ack", tcb.SendW.BytesInFlight)

				} else if tcb.State.State == tcp.FINWAIT1 || tcb.State.State == tcp.LASTACK {
					newState, _ := tcp.StateMachine(tcb.State.State, tcp.ACK, "")
					tcb.State.State = newState
				} else if tcb.State.State == tcp.CLOSING {
					newState, _ := tcp.StateMachine(tcb.State.State, tcp.ACK, "")
					tcb.State.State = newState
					go manager.TimeWaitTimeOut(tcb, 10000)

				}
			}
		}
	} else {
		//fmt.Println("receive data in handler ", len(tcpPayload))

		saddr := api.SockAddr{dstIpAddr, dstPort, srcIpAddr, srcPort}
		tcb, ok := manager.AddrToSocket[saddr]

		if ok && (tcb.State.State == 5 || tcb.State.State == 6 || tcb.State.State == 7 || tcb.State.State == 9) && len(tcpPayload) <= tcb.RecvW.AdvertisedWindow() {
			//fmt.Println("Handler receive order or not ", tcpHeader.SeqNum, tcb.Ack)

			if int(tcpHeader.SeqNum) == tcb.Ack {
				// Write into the receive buffer
				//fmt.Println("recv window seq ", tcb.RecvW.LastSeq)
				tcb.SendW.AdvertisedWindow = ws
				if !tcb.BlockRead {
					su, pad := tcb.RecvW.Receive(tcpPayload, int(tcpHeader.SeqNum), true)
					if su == 1 {
						//fmt.Println("success and send ack")
						tcb.Ack = int(tcpHeader.SeqNum) + len(tcpPayload) + pad

						//fmt.Println(tcpHeader.SeqNum)

						temp := tcb.RecvW.AdvertisedWindow()
						tcb.SendCtrlMsg(tcp.ACK, false, false, temp)
					}
				} else {
					tcb.Ack = int(tcpHeader.SeqNum) + len(tcpPayload)
					temp := tcb.RecvW.AdvertisedWindow()

					tcb.SendCtrlMsg(tcp.ACK, false, false, temp)
				}
			} else {
				tcb.SendW.AdvertisedWindow = ws
				if !tcb.BlockRead {

					su, _ := tcb.RecvW.Receive(tcpPayload, int(tcpHeader.SeqNum), false)

					//tcb.Seq += len(tcpPayload)
					if su == 1 {

						temp := tcb.RecvW.AdvertisedWindow()

						tcb.SendCtrlMsg(tcp.ACK, false, false, temp)
					}
				} else {
					temp := tcb.RecvW.AdvertisedWindow()

					tcb.SendCtrlMsg(tcp.ACK, false, false, temp)
				}
			}

		}

	}

	//To Be Continued

}

func HandleIpPackage(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink, mutex *sync.RWMutex, manager *api.SocketManager) {
	//Open the IP package
	dstIpAddr := ipPkt.IpHeader.Dst.String()
	payLoad := ipPkt.Payload

	//Check TTL
	if ipPkt.IpHeader.TTL == 0 {
		//fmt.Println(ipPkt.IpHeader.TTL)
		//fmt.Println(ipPkt.IpHeader.Protocol)
		fmt.Println("Time to live runs out. Packet has to be dropped\n")
		return
	}

	if !CheckCsum(ipPkt) {
		fmt.Println("Checksum error\n")
		return
	}

	//Check IP destination
	//Local interface check
	for _, link := range node.InterfaceArray {
		if strings.Compare(dstIpAddr, link.Src) == 0 {
			if link.Status == 0 {
				//fmt.Println("Interface is down. Packet has to be dropped\n")
				return
			} else if len(payLoad) == 0 {
				//fmt.Println("But payload is empty.\n")
				return
			} else {
				//fmt.Println("Payload is not empty. Start handling!\n")
				switch ipPkt.IpHeader.Protocol {
				case 0:
					RunDataHandler(ipPkt, node, u)
					return
				case 200:
					RunRIPHandler(ipPkt, node, u, mutex)
					return
				case 6:
					RunTCPHandler(ipPkt, node, u, mutex, manager)
					return
				default:
					fmt.Println("Unrecognized Protocol")
					return
				}
			}
		}
	}

	//Forward IP package to upper layer
	ForwardIpPackage(ipPkt, node, u, mutex)
	return
}

func CheckCsum(ipp ipv4.IpPackage) bool {
	sum := ipp.IpHeader.Checksum
	ipp.IpHeader.Checksum = 0
	temp := ipv4.Csum(ipp.IpHeader)
	return sum == temp
}
