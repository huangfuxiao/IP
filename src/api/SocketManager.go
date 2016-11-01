package api

import (
	".././linklayer"
	".././pkg"
	".././tcp"
	"fmt"
)

type SocketManager struct {
	Fdnum        int
	Portnum      int
	FdToSocket   map[int]*TCB
	AddrToSocket map[SockAddr]*TCB
	Interfaces   map[string]bool
}

func BuildSocketManager(interfaceArray []*pkg.Interface) SocketManager {
	map1 := make(map[int]*TCB)
	map2 := make(map[SockAddr]*TCB)
	set := make(map[string]bool)

	// Construct Interfaces array
	for _, link := range interfaceArray {
		set[link.Src] = true
	}

	return SocketManager{0, 1024, map1, map2, set}
}

//Build a fake manager for testing; Will be removed after testing
/*
func buildTestMananger(interfaceArray []*pkg.Interface) *SocketManager {
	testManager := BuildSocketManager(interfaceArray)
	fd := testManager.Fdnum
	port := 1
	for addr, _ := range testManager.Interfaces {
		saddr := SockAddr{addr, port, "0.0.0.0", 0}
		buf := make([]byte, 0)
		s := tcp.State{State: 1}
		tcb := TCB{fd, s, saddr, 0, 0, buf, buf}
		testManager.FdToSocket[fd] = &tcb
		testManager.AddrToSocket[saddr] = &tcb
		fd++
	}
	return &testManager
}
*/

func (manager *SocketManager) PrintSockets(interfaceArray []*pkg.Interface) {
	//Build a test manager for testing; Will be removed after testing
	//manager = buildTestMananger(interfaceArray)

	//Print Sockets
	fmt.Println("socket\tlocal-addr\tport\t\tdst-addr\tport\tstatus")
	fmt.Println("--------------------------------------------------------------")
	for fd, tcb := range manager.FdToSocket {
		state := tcp.StateString(tcb.State.State)
		fmt.Printf("%d\t%s\t%d\t\t%s\t\t%d\t%s\n", fd, tcb.Addr.LocalAddr, tcb.Addr.LocalPort, tcb.Addr.RemoteAddr, tcb.Addr.RemotePort, state)
	}
}

func (manager *SocketManager) V_socket(node *pkg.Node, u linklayer.UDPLink) int {
	fd := manager.Fdnum
	tcb := BuildTCB(fd, node, u)
	fmt.Println(tcb)
	manager.Fdnum += 1
	manager.FdToSocket[fd] = &tcb
	return fd
}

func (manager *SocketManager) V_bind(socket int, addr string, port int) int {
	sock, ok := manager.FdToSocket[socket]
	if !ok {
		return -1
	}

	//If port is null, set port to Portnum and Portnum += 1
	if port == -1 {
		port = manager.Portnum
		manager.Portnum++
	}
	saddr := SockAddr{addr, port, "0.0.0.0", 0}
	if addr == "" {
		//Find the first free interface ip address and set the addr to it.
		for ipAddr, _ := range manager.Interfaces {
			saddr.LocalAddr = ipAddr
			_, ok := manager.AddrToSocket[saddr]
			if !ok {
				addr = ipAddr
				break
			}
		}
	}
	sock.Addr = saddr
	manager.AddrToSocket[saddr] = sock
	return 0
}

func (manager *SocketManager) V_listen(socket int) int {
	tcb := manager.FdToSocket[socket]
	curState := tcb.State.State
	nextState, _ := tcp.StateMachine(curState, 0, "passive")
	tcb.State.State = nextState
	return 0
}

func (manager *SocketManager) V_connect(socket int, addr string, port int, u linklayer.UDPLink) int {

	// Send syn and change state to SynSent
	tcb := manager.FdToSocket[socket]
	saddr := tcb.Addr
	fmt.Println(saddr)

	//Set remote address to be input addr and port
	tcb.Addr.RemoteAddr = addr
	tcb.Addr.RemotePort = port
	newSaddr := SockAddr{tcb.Addr.LocalAddr, tcb.addr.LocalPort, addr, port}
	manager.AddrToSocket[newSaddr] = &tcb

	//Set state to SYN SENT
	curState := tcb.State.State
	nextState, ctrl := tcp.StateMachine(curState, 0, "active")
	tcb.State.State = nextState
	tcb.SendCtrlMsg(ctrl)

	//SendCtrlMsg(saddr.LocalAddr, saddr.RemoteAddr, saddr.LocalPort, saddr.RemotePort, tcp.FIN, u)
	//-----------
	//Change State
	//-----------
	//Check ACK
	//-----------

	return 0
}

func (manager *SocketManager) V_accept(socket int, addr string) int {
	return 0
}

func (manager *SocketManager) V_read(socket int, buf []byte, nbyte int) int {
	return 0
}

func (manager *SocketManager) V_write(socket int, buf []byte, nbyte int) int {
	return 0
}

func (manager *SocketManager) V_shutdown(socket int, ntype int) int {
	return 0
}

func (manager *SocketManager) v_close(socket int) int {
	return 0
}
