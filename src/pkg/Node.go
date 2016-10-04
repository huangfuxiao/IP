package pkg

import "../ipv4"

type Node struct {
	LocalAddr      string
	Port           int
	InterfaceArray []Interface
	RouteTable     map[string]Entry
}

type Interface struct {
	Status     int
	Src        string
	Dest       string
	RemotePort int
	RemoteAddr string
}

type Entry struct {
	Dest         string
	Next         string
	Cost         int
	Time_to_live int
}
