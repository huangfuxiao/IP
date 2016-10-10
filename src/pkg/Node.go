package pkg

import (
	"../ipv4"
	"../linklayer"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Node struct {
	LocalAddr      string
	Port           int
	InterfaceArray []*Interface
	RouteTable     map[string]Entry
}

func (n *Node) PrintInterfaces() {
	fmt.Println("id\tdst\t\tsrc\t\tenabled")
	i := 0
	for _, link := range n.InterfaceArray {
		fmt.Printf("%d\t%s\t%s\t%d\n", i, link.Dest, link.Src, link.Status)
		i++
	}
}

func (n *Node) PrintRoutes() {
	//fmt.Println("Here is the Rounting Table: \n")
	fmt.Println("dst\t\tsrc\t\tcost")
	for k, v := range n.RouteTable {
		if len(k) < 0 {
			fmt.Println("The rounting table currently has no routes!\n")
		}
		fmt.Printf("%s\t%s\t%d\n", v.Dest, v.Next, v.Cost)
	}
}

func (n *Node) InterfacesDown(id int) {
	if id >= len(n.InterfaceArray) {
		fmt.Println("invalid interface id\n")
		return
	}
	n.InterfaceArray[id].Status = 0
	src := n.InterfaceArray[id].Src
	for k, v := range n.RouteTable {
		if strings.Compare(src, v.Next) == 0 {
			n.RouteTable[k] = Entry{Dest: v.Dest, Next: v.Next, Cost: 16}
		}
	}
}

func (n *Node) InterfacesUp(id int) {
	if id >= len(n.InterfaceArray) {
		fmt.Println("invalid interface id\n")
		return
	}
	n.InterfaceArray[id].Status = 1
	src := n.InterfaceArray[id].Src
	n.RouteTable[src] = Entry{Dest: src, Next: src, Cost: 0}
}

func (n *Node) PrepareAndSendPacket(cmds []string, u linklayer.UDPLink) {
	//Check length of cmds
	if len(cmds) < 4 {
		fmt.Println("invalid args\n")
		return
	} else {
		//Check valid IP
		dest := cmds[1]
		ip := net.ParseIP(dest)
		if ip == nil {
			fmt.Println("invalid ipv4 address\n")
		}
		//Check valid protocol ID
		id, err := strconv.Atoi(cmds[2])
		if err != nil {
			fmt.Println("invalid args\n")
			return
		}

		//Loop through all interfaces to send
		for _, link := range n.InterfaceArray {
			if strings.Compare(dest, link.Dest) == 0 {
				if link.Status == 0 {
					return
				}
				payLoad := cmds[3]
				ipPkt := ipv4.BuildIpPacket([]byte(payLoad), id, link.Src, link.Dest)
				u.Send(ipPkt, link.RemoteAddr, link.RemotePort)
				fmt.Println(ipPkt)

				return
			}
		}

	}

}

func (n *Node) GetRemotePhysAddr(virIP string) (string, int) {
	for _, link := range n.InterfaceArray {
		if strings.Compare(virIP, link.Dest) == 0 {
			return link.RemoteAddr, link.RemotePort
		}
	}
	err := "error"
	return err, -1
}

func (n *Node) GetLearnFrom(virIP string) string {
	for _, link := range n.InterfaceArray {
		if strings.Compare(virIP, link.Src) == 0 {
			return link.Dest
		}
	}
	err := "error"
	return err
}
