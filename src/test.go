package main

import (
	"./ipv4"
	"fmt"
	"net"
)

func main() {
	src := "192.168.1.1"
	dest := "192.168.0.1"
	payload := []byte("hello")

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

	ret := header.String()
	fmt.Println(ret)

	a, b := header.Marshal()
	fmt.Println(a)
	fmt.Println(b)

	c, d := ipv4.ParseHeader(a)
	fmt.Println(c)
	fmt.Println(d)

}
