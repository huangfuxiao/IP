package main

import (
	//"./ipv4"
	"./linklayer"
	"fmt"
	//"net"
)

func main() {

	addr := "localhost"
	port := 5003
	u := linklayer.InitUDP(addr, port)
	fmt.Println(u)
	for {
		ret := u.Receive()
		fmt.Println("Receive successfullly")
		fmt.Println(ret)
	}

}
