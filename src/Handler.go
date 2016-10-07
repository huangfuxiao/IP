package main

import (
	"./ipv4"
	"./linklayer"
	"./pkg"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

var u linklayer.UDPLink

//Build a test node; will be removed after testing
func buildTestNode() (testNode pkg.Node) {
	var testLink1, testLink2 pkg.Interface
	testRT := make(map[string]pkg.Entry)
	var testEntry1, testEntry2 pkg.Entry
	testNode.LocalAddr = "localhost"
	testNode.Port = 5003
	testLink1.Status = 1
	testLink1.Src = "192.168.0.12"
	testLink1.Dest = "192.168.0.11"
	testLink1.RemotePort = 5000
	testLink1.RemoteAddr = "localhost"
	testNode.InterfaceArray = append(testNode.InterfaceArray, &testLink1)
	testLink2.Status = 1
	testLink2.Src = "192.168.0.14"
	testLink2.Dest = "192.168.0.13"
	testLink2.RemotePort = 5004
	testLink2.RemoteAddr = "localhost"
	testNode.InterfaceArray = append(testNode.InterfaceArray, &testLink2)
	testEntry1.Cost = 0
	testEntry1.Dest = testLink1.Src
	testEntry1.Next = testLink1.Src
	testRT[testLink1.Src] = testEntry1
	testEntry2.Cost = 0
	testEntry2.Dest = testLink2.Src
	testEntry2.Next = testLink2.Src
	testRT[testLink2.Src] = testEntry2
	testNode.RouteTable = testRT
	return testNode
}

//Conver payLoad to RIP: to be implemented.
//Currently just create a test RIP
func convertBytesToRIP(data []byte) ipv4.RIP {
	fmt.Println(data)
	var newRip ipv4.RIP
	/*
		opcode := binary.BigEndian.Uint16(data) // this will get first 2 bytes to be interpreted as uint16 number
		raw_data := data[2:len(data)]           // this will copy rest of the raw data in to raw_data byte stream
	*/
	newRip.Command = 2
	newRip.Entries = append(newRip.Entries, ipv4.RIPEntry{Cost: 0, Address: "192.168.0.1"})
	newRip.Entries = append(newRip.Entries, ipv4.RIPEntry{Cost: 0, Address: "192.168.0.13"})
	newRip.NumEntries = len(newRip.Entries)
	return newRip
}

//Convert RIP to IP Packet
func convertRipToIpPackage(rip ipv4.RIP, src string, dest string) ipv4.IpPackage {
	b, err := json.Marshal(rip)
	if err != nil {
		fmt.Println("error with json.Marshal:", err)
	}
	os.Stdout.Write(b)
	ipPkt := ipv4.BuildIpPacket(b, 200, src, dest)
	return ipPkt
}

//IP protocol=0
func runForwardHandler(ipPkt ipv4.IpPackage, node pkg.Node) {
	fmt.Println("Start runForwardHandler: \n")

	dstIpAddr := ipPkt.IpHeader.Dst.String()
	srcIpAddr := ipPkt.IpHeader.Src.String()
	fmt.Println(srcIpAddr)
	ttl := ipPkt.IpHeader.TTL
	payLoad := ipPkt.Payload

	//Check TTL
	if ttl <= 0 {
		fmt.Println("Time to live runs out. Packet has to be dropped\n")
		return
	}

	//Local interface check
	for _, link := range node.InterfaceArray {
		if strings.Compare(dstIpAddr, link.Src) == 0 {
			//Payload is not empty
			if len(payLoad) > 0 {
				fmt.Println("Local Arrived! Here is the payload: ")
				os.Stdout.Write(payLoad)
				fmt.Println("\n")
			} else {
				fmt.Println("Payload is empty")
			}
			return
		}
	}

	//forward to upper node
	for k, v := range node.RouteTable {
		if len(k) < 0 {
			fmt.Println("The rounting table currently has no routes!\n")
			return
		}
		//Destination matches one on this Node's Rounting Table
		if strings.Compare(dstIpAddr, v.Dest) == 0 {

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

					ipPkt.IpHeader.TTL--
					if ipPkt.IpHeader.TTL <= 0 {
						fmt.Println("Time to live runs out. Packet has to be dropped\n")
						return
					}

					//Forward ip packet to next node
					u.Send(ipPkt, link.RemoteAddr, link.RemotePort)
					fmt.Printf("Node can get through this interface: %s to %s with cost%d\n", v.Next, v.Dest, v.Cost)
					return
				}
			}
		}
	}
	fmt.Printf("Cannot find a interface in this node's routing table. Packet has to be dropped\n")
	return
}

//IP protocol=200
func runRIPHandler(ipPkt ipv4.IpPackage, node pkg.Node) {
	fmt.Println("Start runRIPHandler: \n")
	dstIpAddr := ipPkt.IpHeader.Dst.String()
	srcIpAddr := ipPkt.IpHeader.Src.String()
	fmt.Println(srcIpAddr)
	payLoad := ipPkt.Payload

	//Check if the srcIpAddr is connected with me
	for _, link := range node.InterfaceArray {
		fmt.Println(link.Dest)
		if strings.Compare(srcIpAddr, link.Dest) == 0 {
			//Arrive the interface
			rip := convertBytesToRIP(payLoad)

			//RIP Request
			if rip.Command == 1 {

				//Put all of this node.RouteTable into RIP and send back

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
				/* Loop through all my the rip's entry
				   Compare if the RIPEntry's Address already on this node.RouteTable
				   If so, compare the cost
				*/
				for _, entry := range rip.Entries {
					value, ok := node.RouteTable[entry.Address]
					if ok {
						fmt.Println("Check cost! RIPEntry's address already exist: ", value)
						if value.Next != dstIpAddr && (entry.Cost+1) < value.Cost {
							node.RouteTable[entry.Address] = pkg.Entry{Dest: entry.Address, Next: dstIpAddr, Cost: entry.Cost + 1}
						}
						/* Split horizon with poison reverse
						   To be implemented*/
					} else {
						node.RouteTable[entry.Address] = pkg.Entry{Dest: entry.Address, Next: dstIpAddr, Cost: entry.Cost + 1}
					}
				}
				/* Send updated route table as RIP to all the neibors
				   To be implemented
				*/

				/* Last, insert this neighbor to to node.RouteTable with cost = 1 */
				node.RouteTable[srcIpAddr] = pkg.Entry{Dest: srcIpAddr, Next: dstIpAddr, Cost: 1}
				return
			} else {
				fmt.Println("Unrecognized RIP protocol!\n")
				return
			}

		}
	}

	fmt.Printf("Cannot find a interface in this node's routing table. Packet has to be dropped\n")
	return
}

func main() {
	//Create a test node, IP header and pkt
	testNode := buildTestNode()
	var testHeader ipv4.Header
	var testPkt ipv4.IpPackage
	testHeader.Protocol = 0
	fmt.Printf("testHeader's protocol: %d\n", testHeader.Protocol)
	payload := []byte("hello")
	testPkt = ipv4.BuildIpPacket(payload, testHeader.Protocol, "192.168.0.13", "192.168.0.14")

	u = linklayer.InitUDP(testNode.LocalAddr, testNode.Port)

	fmt.Println("Test Handler Main Begins: \n")

	switch testPkt.IpHeader.Protocol {
	case 0:
		runForwardHandler(testPkt, testNode)
	case 200:
		runRIPHandler(testPkt, testNode)
	}

	/*
		for{
			for _, link := range node.InterfaceArray {
				if link.status >0 {
					var pkt = node.bufferRead(link)
					int protocol = getIPProtocol(pkt.head)
					switch protocol {
					case 0:
						runIpHandler(IPPackage, node)
					case 200:
						runRIPHandler(IPPackage, node)
				}

		     }
		}
	*/

}
