package main

import (
	"./api"
	"./ipv4"
	"./linklayer"
	"./tcp"
	"fmt"
	"net"
)

func main() {

	h := tcp.BuildTCPHeader(5005, 50001, 0, 0, 0, 0)
	p := tcp.BuildTCPPacket([]byte("hello9"), h)
	buf := tcp.TCPPkgToBuffer(p)

	src := "192.168.1.1"
	dest := "192.168.0.1"
	payload := buf

	header := ipv4.Header{
		Version:  4,
		Len:      20,
		TOS:      0,
		TotalLen: 20 + len(payload),
		TTL:      16,
		Protocol: 0,
		Src:      net.ParseIP(src),
		Dst:      net.ParseIP(dest),
		Options:  []byte{},
	}
	/*
		ret := header.String()
		fmt.Println(ret)

		a, b := header.Marshal()
		fmt.Println(a)
		fmt.Println(b)

		c, d := ipv4.ParseHeader(a)
		fmt.Println(c)
		fmt.Println(d)
	*/
	x := ipv4.IpPackage{header, payload}
	//fmt.Println(x)
	//y := ipv4.String(x)
	//fmt.Println(y)
	//z := tcp.String(p)
	//fmt.Println(z)
	/*
		z := ipv4.IpPkgToBuffer(x)
		fmt.Println(z)

		t := ipv4.BufferToIpPkg(z)
		fmt.Println(ipv4.String(t))
	*/
	addr := "localhost"
	port := 5002
	u := linklayer.InitUDP(addr, port)
	fmt.Println(u)
	u.Send(x, "localhost", 5003)
	fmt.Println("reach")
	api.SendSyn(src, dest, 5002, 5003, u)

}
