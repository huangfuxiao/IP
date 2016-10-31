package main

import (
	/*"./ipv4"
	"./linklayer"*/
	"fmt"
	//"net"
	"./api"
)

func main() {
	v := api.BuildSocketManager()
	n := v.V_socket()
	fmt.Println(n)
	n1 := v.V_socket()
	fmt.Println(n1)
	fmt.Println(" old state here ", v.FdToSocket[0].State.State)
	n2 := v.V_listen(0)
	fmt.Println(n2, " new state here ", v.FdToSocket[0].State.State)
	/*

		addr := "localhost"
		port := 5003
		u := linklayer.InitUDP(addr, port)
		fmt.Println(u)
		for {
			ret := u.Receive()
			fmt.Println("Receive successfullly")
			fmt.Println(ret)
		}
	*/
}
