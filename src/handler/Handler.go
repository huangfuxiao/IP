package handler

import (
	"../ipv4"
	"../linklayer"
	"../pkg"
	"fmt"
	"strings"
)

//Convert RIP to IP Packet
func ConvertRipToIpPackage(rip ipv4.RIP, src string, dest string) ipv4.IpPackage {
	b := ipv4.ConvertRipToBytes(rip)
	ipPkt := ipv4.BuildIpPacket(b, 200, src, dest)
	return ipPkt
}

//Send Trigger Updates using RIP to All of Node's Neighbors
func SendTriggerUpdates(destIpAddr string, cost int, node *pkg.Node, u linklayer.UDPLink) {
	var newRip ipv4.RIP
	newRip.Command = 2
	newRip.Entries = append(newRip.Entries, ipv4.RIPEntry{Cost: cost, Address: destIpAddr})
	newRip.NumEntries = len(newRip.Entries)
	for _, link := range node.InterfaceArray {
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
func RunDataHandler(ipPkt ipv4.IpPackage, node *pkg.Node) {
	data := ipv4.String(ipPkt)
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
			rip := ipv4.ConvertBytesToRIP(payLoad)

			//RIP Request
			if rip.Command == 1 {
				/* First, insert this neighbor to the node.RouteTable with cost = 1 */
				fmt.Println("IPPackage arrived after rip.Command==1\n")
				node.RouteTable[srcIpAddr] = pkg.Entry{Dest: srcIpAddr, Next: dstIpAddr, Cost: 1}
				SendTriggerUpdates(srcIpAddr, 1, node, u)

				//Then, put all of this node.RouteTable into RIP and send back
				var newRip ipv4.RIP
				newRip.Command = 2
				newRip.NumEntries = 0
				//put all of this node's RT's entries to RIP
				for _, v := range node.RouteTable {
					newRip.Entries = append(newRip.Entries, ipv4.RIPEntry{Cost: v.Cost, Address: v.Dest})
					newRip.NumEntries++
				}

				//send back RIP to src
				ipPkt := ConvertRipToIpPackage(newRip, link.Src, srcIpAddr)
				u.Send(ipPkt, link.RemoteAddr, link.RemotePort)
				fmt.Printf("RIP response sent back to this address: %s %d \n", link.RemoteAddr, link.RemotePort)
				return

			} else if rip.Command == 2 {
				/* First, insert this neighbor to the node.RouteTable with cost = 1 */
				v, ok := node.RouteTable[srcIpAddr]
				if ok {
					if v.Cost > 1 {
						node.RouteTable[srcIpAddr] = pkg.Entry{Dest: srcIpAddr, Next: dstIpAddr, Cost: 1}
						SendTriggerUpdates(srcIpAddr, 1, node, u)
					}
				}
				/* Then, loop through all of the rip's entry
				   Compare if the RIPEntry's Address already on this node.RouteTable
				   If so, compare the cost
				*/
				for _, entry := range rip.Entries {
					value, ok := node.RouteTable[entry.Address]
					if ok {
						/*Implement poison reverse first*/
						if (entry.Cost == 16) && (dstIpAddr == value.Next) {
							node.RouteTable[entry.Address] = pkg.Entry{Dest: entry.Address, Next: dstIpAddr, Cost: 16}
							SendTriggerUpdates(entry.Address, node.RouteTable[entry.Address].Cost, node, u)
						} else if (entry.Cost + 1) < value.Cost {
							node.RouteTable[entry.Address] = pkg.Entry{Dest: entry.Address, Next: dstIpAddr, Cost: entry.Cost + 1}
							SendTriggerUpdates(entry.Address, node.RouteTable[entry.Address].Cost, node, u)
						}
					} else {
						node.RouteTable[entry.Address] = pkg.Entry{Dest: entry.Address, Next: dstIpAddr, Cost: entry.Cost + 1}
						SendTriggerUpdates(entry.Address, node.RouteTable[entry.Address].Cost, node, u)
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

	//Check IP destination
	//Local interface check
	for _, link := range node.InterfaceArray {
		if strings.Compare(dstIpAddr, link.Src) == 0 {
			fmt.Println("Local Arrived! ")
			//Payload is not empty
			if len(payLoad) == 0 {
				fmt.Println("But payload is empty.\n")
				return
			} else {
				fmt.Println("Payload is not empty. Start handling!\n")
				switch ipPkt.IpHeader.Protocol {
				case 0:
					RunDataHandler(ipPkt, node)
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
