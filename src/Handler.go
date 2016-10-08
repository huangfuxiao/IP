package main

import (
	"./ipv4"
	"./linklayer"
	"./pkg"
	"fmt"
	"strings"
)

var u linklayer.UDPLink

//Build a test node; will be removed after testing
func buildTestNode() (testNode pkg.Node) {
	testRT := make(map[string]pkg.Entry)
	testNode.LocalAddr = "localhost"
	testNode.Port = 5003
	testLink1 := pkg.Interface{Status: 1, Src: "192.168.0.12", Dest: "192.168.0.11", RemotePort: 5000, RemoteAddr: "localhost"}
	testLink2 := pkg.Interface{Status: 1, Src: "192.168.0.14", Dest: "192.168.0.13", RemotePort: 5004, RemoteAddr: "localhost"}
	testNode.InterfaceArray = append(testNode.InterfaceArray, &testLink1)
	testNode.InterfaceArray = append(testNode.InterfaceArray, &testLink2)
	testRT[testLink1.Src] = pkg.Entry{Cost: 0, Dest: testLink1.Src, Next: testLink1.Src}
	testRT[testLink2.Src] = pkg.Entry{Cost: 0, Dest: testLink2.Src, Next: testLink2.Src}
	testRT["192.168.0.15"] = pkg.Entry{Cost: 1, Dest: "192.168.0.15", Next: "192.168.0.12"}
	testNode.RouteTable = testRT
	fmt.Println(testRT)
	return testNode
}

//Convert RIP to IP Packet
func convertRipToIpPackage(rip ipv4.RIP, src string, dest string) ipv4.IpPackage {
	b := ipv4.ConvertRipToBytes(rip)
	ipPkt := ipv4.BuildIpPacket(b, 200, src, dest)
	return ipPkt
}

//Send Trigger Updates using RIP to All of Node's Neighbors
func sendTriggerUpdates(destIpAddr string, cost int, node *pkg.Node) {
	var newRip ipv4.RIP
	newRip.Command = 2
	newRip.Entries = append(newRip.Entries, ipv4.RIPEntry{Cost: cost, Address: destIpAddr})
	newRip.NumEntries = len(newRip.Entries)
	for _, link := range node.InterfaceArray {
		ipPkt := convertRipToIpPackage(newRip, link.Src, link.Dest)
		u.Send(ipPkt, link.RemoteAddr, link.RemotePort)
		fmt.Printf("Trigger update RIP sent to this address: %s to %d with cost: %d\n", link.RemoteAddr, link.RemotePort, cost)
	}
}

//IP is not locally arrived
func forwardIpPackage(ipPkt ipv4.IpPackage, node pkg.Node) {
	fmt.Println("Start Forwarding IP Package: \n")

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

					if ipPkt.IpHeader.TTL < 0 {
						fmt.Println("Time to live runs out. Packet has to be dropped\n")
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
func runDataHandler(ipPkt ipv4.IpPackage, node *pkg.Node) string {
	payLoad := string(ipPkt.Payload)
	fmt.Println(payLoad)
	return payLoad
}

//IP protocol=200
func runRIPHandler(ipPkt ipv4.IpPackage, node *pkg.Node) {
	fmt.Println("Start runRIPHandler: \n")
	dstIpAddr := ipPkt.IpHeader.Dst.String()
	srcIpAddr := ipPkt.IpHeader.Src.String()
	fmt.Println(srcIpAddr)
	payLoad := ipPkt.Payload

	for _, link := range node.InterfaceArray {
		if strings.Compare(srcIpAddr, link.Dest) == 0 {
			//Arrive the interface
			rip := ipv4.ConvertBytesToRIP(payLoad)

			//RIP Request
			if rip.Command == 1 {
				/* First, insert this neighbor to the node.RouteTable with cost = 1 */
				v, ok := node.RouteTable[srcIpAddr]
				if ok {
					fmt.Println("Check cost! Address already exist: ", v)
					if v.Cost > 1 {
						node.RouteTable[srcIpAddr] = pkg.Entry{Dest: srcIpAddr, Next: dstIpAddr, Cost: 1}
						sendTriggerUpdates(srcIpAddr, 1, node)
					}
				}

				//Then, put all of this node.RouteTable into RIP and send back
				fmt.Println("Start building a response RIP and send back\n")
				var newRip ipv4.RIP
				newRip.Command = 2
				newRip.NumEntries = 0
				//put all of this node's RT's entries to RIP
				for _, v := range node.RouteTable {
					newRip.Entries = append(newRip.Entries, ipv4.RIPEntry{Cost: v.Cost, Address: v.Dest})
					newRip.NumEntries++
				}
				//send back RIP to src
				ipPkt := convertRipToIpPackage(newRip, link.Src, srcIpAddr)
				u.Send(ipPkt, link.RemoteAddr, link.RemotePort)
				return

			} else if rip.Command == 2 {

				/* First, insert this neighbor to the node.RouteTable with cost = 1 */
				v, ok := node.RouteTable[srcIpAddr]
				if ok {
					fmt.Println("Check cost! Address already exist: ", v)
					if v.Cost > 1 {
						node.RouteTable[srcIpAddr] = pkg.Entry{Dest: srcIpAddr, Next: dstIpAddr, Cost: 1}
						sendTriggerUpdates(srcIpAddr, 1, node)
					}
				}
				/* Then, loop through all of the rip's entry
				   Compare if the RIPEntry's Address already on this node.RouteTable
				   If so, compare the cost
				*/
				for _, entry := range rip.Entries {
					value, ok := node.RouteTable[entry.Address]
					if ok {
						fmt.Println("Check cost! RIPEntry's address already exist: ", value)
						if (entry.Cost + 1) < value.Cost {
							node.RouteTable[entry.Address] = pkg.Entry{Dest: entry.Address, Next: dstIpAddr, Cost: entry.Cost + 1}
							sendTriggerUpdates(entry.Address, node.RouteTable[entry.Address].Cost, node)
						}
						/* Split horizon with poison reverse
						   To be implemented*/
					} else {
						node.RouteTable[entry.Address] = pkg.Entry{Dest: entry.Address, Next: dstIpAddr, Cost: entry.Cost + 1}
						sendTriggerUpdates(entry.Address, node.RouteTable[entry.Address].Cost, node)
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

func main() {
	//Create a test node, IP header and pkt, and UDP connection
	testNode := buildTestNode()
	var testHeader ipv4.Header
	var testPkt ipv4.IpPackage
	testHeader.Protocol = 0
	payload := []byte("hello")
	testPkt = ipv4.BuildIpPacket(payload, testHeader.Protocol, "192.168.0.13", "192.168.0.15")
	u = linklayer.InitUDP(testNode.LocalAddr, testNode.Port)

	fmt.Println("Test Handler Main Begins: \n")

	//Open the IP package
	dstIpAddr := testPkt.IpHeader.Dst.String()
	srcIpAddr := testPkt.IpHeader.Src.String()
	fmt.Printf("Forward Handler starts! This is the ip package's src address: %s\n", srcIpAddr)
	payLoad := testPkt.Payload

	//Check TTL
	if testPkt.IpHeader.TTL < 0 {
		fmt.Println("Time to live runs out. Packet has to be dropped\n")
		return
	}
	//Check IP destination
	//Local interface check
	for _, link := range testNode.InterfaceArray {
		if strings.Compare(dstIpAddr, link.Src) == 0 {
			fmt.Println("Local Arrived!")
			//Payload is not empty
			if len(payLoad) == 0 {
				fmt.Println("But payload is empty.\n")
				return
			} else {
				fmt.Println("Payload is not empty. Start handling!\n")
				switch testPkt.IpHeader.Protocol {
				case 0:
					runDataHandler(testPkt, &testNode)
					return
				case 200:
					runRIPHandler(testPkt, &testNode)
					return
				}
			}
		}
	}

	//Forward IP package to upper layer
	forwardIpPackage(testPkt, testNode)
	return
}
