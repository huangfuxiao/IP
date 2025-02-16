package runner

import (
	"../handler"
	"../ipv4"
	"../linklayer"
	"../pkg"
	"fmt"
	"strings"
	"sync"
	"time"
)

func Send_thread(node *pkg.Node, u linklayer.UDPLink, mutex *sync.RWMutex) {
	for {
		//Loop through interfaces and send to all neighbors
		for _, link := range node.InterfaceArray {
			if link.Status == 0 {
				continue
			}

			//Put all of this node.RouteTable into RIP
			var newRip ipv4.RIP
			newRip.Command = 2
			newRip.NumEntries = 0
			mutex.RLock()
			for _, v := range node.RouteTable {
				/* Implement poison reverse
				Compare the learn from virIP to the RIP packege's destination
				If learnFrom == RIP.Dest, modify the cost to be INFINITY
				????????????????????????????????????????????????????????
				*/
				if newRip.NumEntries == 64 {
					ipPkt := handler.ConvertRipToIpPackage(newRip, link.Src, link.Dest)
					u.Send(ipPkt, link.RemoteAddr, link.RemotePort)
					newRip.NumEntries = 0
					newRip.Entries = nil
				}

				learnFrom := node.GetLearnFrom(v.Next)
				if learnFrom == "error" {
					fmt.Println("ERROR in learn from")
					continue
				}
				if strings.Compare(v.Dest, v.Next) > 0 && strings.Compare(link.Dest, learnFrom) == 0 && v.Cost != 16 {
					newRip.Entries = append(newRip.Entries, ipv4.RIPEntry{Cost: 16, Address: v.Dest})
				} else {
					newRip.Entries = append(newRip.Entries, ipv4.RIPEntry{Cost: v.Cost, Address: v.Dest})
				}

				newRip.NumEntries++
			}
			mutex.RUnlock()

			ipPkt := handler.ConvertRipToIpPackage(newRip, link.Src, link.Dest)
			u.Send(ipPkt, link.RemoteAddr, link.RemotePort)
		}

		time.Sleep(5000 * time.Millisecond)

	}
}
