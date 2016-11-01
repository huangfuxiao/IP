package api

import (
	//".././ipv4"
	".././linklayer"
	".././tcp"
)

type SocketManager struct {
	Fdnum        int
	Portnum      int
	FdToSocket   map[int]*TCB
	AddrToSocket map[SockAddr]*TCB
	//interfaces   []ipv4.Interface
}

func BuildSocketManager() SocketManager {
	map1 := make(map[int]*TCB)
	map2 := make(map[SockAddr]*TCB)
	// Construct Interfaces array

	return SocketManager{0, 1024, map1, map2}

}

func (manager *SocketManager) PrintSockets() {
}

func (manager *SocketManager) V_socket() int {
	fd := manager.Fdnum
	tcb := BuildTCB(fd)

	manager.Fdnum += 1
	manager.FdToSocket[fd] = &tcb
	return fd
}

func (manager *SocketManager) V_bind(socket int, addr string, port int) int {
	sock, ok := manager.FdToSocket[socket]
	if !ok {
		return -1
	}
	if addr == "" {
		//Find the first free interface ip address and set the addr to it.
		//If port is null, set port to Portnum and Portnum += 1
	}
	saddr := SockAddr{addr, port, "0.0.0.0", 0}
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
