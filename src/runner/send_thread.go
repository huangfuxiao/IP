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
		//Put all of this node.RouteTable into RIP
		var newRip ipv4.RIP
		newRip.Command = 2
		newRip.NumEntries = 0
		//put all of this node's RT's entries to RIP
		for _, v := range node.RouteTable {
			newRip.Entries = append(newRip.Entries, ipv4.RIPEntry{Cost: v.Cost, Address: v.Dest})
			newRip.NumEntries++
		}

		//Loop through interfaces and send to all neighbors
		for _, link := range node.InterfaceArray {
			ipPkt := handler.ConvertRipToIpPackage(newRip, link.Src, link.Dest)
			u.Send(ipPkt, link.RemoteAddr, link.RemotePort)
		}

		time.Sleep(5000 * time.Millisecond)

	}
}
