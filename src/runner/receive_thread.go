package runner

import (
	"../api"
	"../handler"
	"../linklayer"
	"../pkg"
	"sync"
)

func Receive_thread(udp linklayer.UDPLink, node *pkg.Node, mutex *sync.RWMutex, manager *api.SocketManager) {
	for {
		ipp := udp.Receive()
		handler.HandleIpPackage(ipp, node, udp, mutex, manager)

	}
}
