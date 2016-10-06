package linklayer

import (
	"../ipv4"
	"fmt"
	"net"
	"strconv"
)

// UDP link structure
// Store the UDP socket
type UDPLink struct {
	socket *net.UDPConn
}

// Check the error message
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

// Initiate the UDP socket for a given local address and local port
func InitUDP(addr string, port int) UDPLink {
	portStr := strconv.Itoa(port)
	service := addr + ":" + portStr
	udpAddr, err := net.ResolveUDPAddr("udp", service)
	CheckError(err)
	socket, err := net.ListenUDP("udp", udpAddr)
	CheckError(err)
	return UDPLink{socket}
}

// Convert an IpPackage to a buffer and send it through UDP to reAddr:rePort
// @Parameter IpPackage, remote address, remote port
func (l *UDPLink) Send(ipp ipv4.IpPackage, reAddr string, rePort int) {
	buf := ipv4.IpPkgToBuffer(ipp)
	portStr := strconv.Itoa(rePort)
	service := reAddr + ":" + portStr
	reUDPAddr, err := net.ResolveUDPAddr("udp", service)
	CheckError(err)

	l.socket.WriteToUDP(buf, reUDPAddr)

}

// Receive buffer from the UDP link and convert it to an IpPackage
// @Return a IpPackage
func (l *UDPLink) Receive() ipv4.IpPackage {
	buf := make([]byte, 1400)
	for {
		n, addr, err := l.socket.ReadFromUDP(buf)

		if err != nil {
			fmt.Println("Error: ", err)
		}

		if n != 0 {
			fmt.Println("Driver Received Packet from ", addr)
			ipp := ipv4.BufferToIpPkg(buf[0:n])

			return ipp
		}
	}
}
