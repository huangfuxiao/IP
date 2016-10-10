package runner

import (
	"../handler"
	"../ipv4"
	"../linklayer"
	"../pkg"
	"time"
)

func Send_thread(node *pkg.Node, u linklayer.UDPLink) {
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
			for _, v := range node.RouteTable {
				/* Implement poison reverse
				Compare the learn from virIP to the RIP packege's destination
				If learnFrom == RIP.Dest, modify the cost to be INFINITY
				????????????????????????????????????????????????????????
				*/
				learnFrom := node.GetLearnFrom(v.Next)
				if string.Compare(v.Dest, v.Next) > 0 && string.Compare(link.Dest, learnFrom) == 0 && v.Cost != 16 {
					newRip.Entries = append(newRip.Entries, ipv4.RIPEntry{Cost: 16, Address: v.Dest})
				} else {
					newRip.Entries = append(newRip.Entries, ipv4.RIPEntry{Cost: v.Cost, Address: v.Dest})
				}

				newRip.NumEntries++
			}

			ipPkt := handler.ConvertRipToIpPackage(newRip, link.Src, link.Dest)
			u.Send(ipPkt, link.RemoteAddr, link.RemotePort)
		}

		time.Sleep(5000 * time.Millisecond)

	}
}
