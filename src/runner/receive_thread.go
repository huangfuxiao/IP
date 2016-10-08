package runner

import (
	"../handler"
	"../linklayer"
	"../pkg"
	"fmt"
)

func Receive_thread(udp linklayer.UDPLink, node *pkg.Node) {
	for {
		ipp := udp.Receive()
		handler.HandleIpPackage(ipp, node)

	}
}
