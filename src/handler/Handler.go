package handler

import (
	"../ipv4"
	"../linklayer"
	"../pkg"
	"fmt"
	"strings"
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
		fmt.Printf("Trigger update RIP sent to this address: %s %d \n", link.RemoteAddr, link.RemotePort)
	}
}

//IP is not locally arrived
func ForwardIpPackage(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink) {
	dstIpAddr := ipPkt.IpHeader.Dst.String()

	//Loop through node.RouteTable, and forward to upper node
	for k, v := range node.RouteTable {
		if len(k) < 0 {
			fmt.Println("The rounting table currently has no routes!\n")
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
						fmt.Println("Interface is down. Packet has to be dropped\n")
						return
					}

					//Check cost is not infinity
					if v.Cost >= 16 {
						fmt.Println("Inifinity loop. Packet has to be dropped\n")
						return
					}

					//Forward ip packet to next node
					ipPkt.IpHeader.TTL--
					ipPkt.IpHeader.Checksum = 0
					ipPkt.IpHeader.Checksum = ipv4.Csum(ipPkt.IpHeader)
					u.Send(ipPkt, link.RemoteAddr, link.RemotePort)
					fmt.Printf("Node can get through this interface: %s to %s with cost: %d\n", v.Next, v.Dest, v.Cost)
					return
				}
			}
		}
	}
	fmt.Printf("Cannot find a interface in this node's routing table. Packet has to be dropped\n")
	return
}

//IP protocol=0
func RunDataHandler(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink) {
	data := ipv4.String(ipPkt)
	fmt.Println("Driver Received Packet from ", u.Addr)
	fmt.Println(data)
}

//IP protocol=200
func RunRIPHandler(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink) {
	dstIpAddr := ipPkt.IpHeader.Dst.String()
	srcIpAddr := ipPkt.IpHeader.Src.String()
	payLoad := ipPkt.Payload

	for _, link := range node.InterfaceArray {
		if strings.Compare(srcIpAddr, link.Dest) == 0 {

			//Arrive the interface
			if link.Status == 0 {
				fmt.Println("Interface is down. Packet has to be dropped\n")
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

				//send back RIP to src
				ipPkt := ConvertRipToIpPackage(newRip, link.Src, srcIpAddr)
				u.Send(ipPkt, link.RemoteAddr, link.RemotePort)
				fmt.Printf("RIP response sent back to this address: %s %d \n", link.RemoteAddr, link.RemotePort)

				/* Then, insert this neighbor to the node.RouteTable with cost = 1 */
				fmt.Println("IPPackage arrived after rip.Command==1\n")
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

func HandleIpPackage(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink) {
	//Open the IP package
	dstIpAddr := ipPkt.IpHeader.Dst.String()
	payLoad := ipPkt.Payload

	//Check TTL
	if ipPkt.IpHeader.TTL < 0 {
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
				fmt.Println("Interface is down. Packet has to be dropped\n")
				return
			} else if len(payLoad) == 0 {
				fmt.Println("But payload is empty.\n")
				return
			} else {
				//fmt.Println("Payload is not empty. Start handling!\n")
				switch ipPkt.IpHeader.Protocol {
				case 0:
					RunDataHandler(ipPkt, node, u)
					return
				case 200:
					RunRIPHandler(ipPkt, node, u)
					return
				}
			}
		}
	}

	//Forward IP package to upper layer
	ForwardIpPackage(ipPkt, node, u)
	return
}

func CheckCsum(ipp ipv4.IpPackage) bool {
	sum := ipp.IpHeader.Checksum
	ipp.IpHeader.Checksum = 0
	temp := ipv4.Csum(ipp.IpHeader)
	return sum == temp
}
