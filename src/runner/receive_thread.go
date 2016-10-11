package runner

import (
	"../handler"
	"../linklayer"
	"../pkg"
	"sync"
)

func Receive_thread(udp linklayer.UDPLink, node *pkg.Node, mutex *sync.RWMutex) {
	for {
		ipp := udp.Receive()
		handler.HandleIpPackage(ipp, node, udp, mutex)

	}
}
