package pkg

import (
	"fmt"
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
		fmt.Println("Unvalid id!\n")
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
		fmt.Println("Unvalid id!\n")
		return
	}
	n.InterfaceArray[id].Status = 1
	src := n.InterfaceArray[id].Src
	n.RouteTable[src] = Entry{Dest: src, Next: src, Cost: 0}
}

func (n *Node) PrepareAndSendPacket() {
	fmt.Println("Do nothing for prepareAndSendPacket for now")
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
