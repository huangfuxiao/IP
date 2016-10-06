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

func (n *Node) InterfacesDown() {
	for _, link := range n.InterfaceArray {
		link.Status = 0
	}
}

func (n *Node) InterfacesUp() {
	for _, link := range n.InterfaceArray {
		link.Status = 1
	}
}

func (n *Node) PrepareAndSendPacket() {
	fmt.Println("Do nothing for prepareAndSendPacket for now")
}

<<<<<<< HEAD
func (n *Node) SetMTU() {
	fmt.Println("Do nothing for setMTU for now")

}

func (n *Node) GetRemotePhysAddr(virIP string) (string, int) {
=======
func (n *Node) GetRemotePhysAddr(virIP string) (phyAddr string, port int) {
>>>>>>> 833d3cdeae0db3e222c9de039601a102dad3c84f
	for _, link := range n.InterfaceArray {
		if strings.Compare(virIP, link.Dest) == 0 {
			return link.RemoteAddr, link.RemotePort
		}
	}
	err := "error"
	return err, -1

}
