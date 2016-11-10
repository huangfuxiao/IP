package api

import (
	".././linklayer"
	".././pkg"
	".././tcp"
	"fmt"
	"time"
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

func (manager *SocketManager) PrintSockets() {
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
	if addr != "" {
		sock.Addr = saddr
		manager.AddrToSocket[saddr] = sock
		return 0
	} else {
		fmt.Println("v_bind() error: Cannot assign requested address")
		delete(manager.FdToSocket, socket)
		manager.Fdnum -= 1
	}

	return -1
}

func (manager *SocketManager) V_listen(socket int) int {
	tcb, ok := manager.FdToSocket[socket]
	if ok {
		curState := tcb.State.State
		nextState, _ := tcp.StateMachine(curState, 0, "passive")
		tcb.State.State = nextState
		return 0
	}
	return -1
}

func (manager *SocketManager) V_connect(socket int, addr string, port int) int {

	// Send syn and change state to SynSent
	tcb := manager.FdToSocket[socket]
	//fmt.Println(saddr)

	//Set remote address to be input addr and port
	tcb.Addr.RemoteAddr = addr
	tcb.Addr.RemotePort = port
	newSaddr := SockAddr{tcb.Addr.LocalAddr, tcb.Addr.LocalPort, addr, port}
	manager.AddrToSocket[newSaddr] = tcb

	//Set state to SYN SENT
	curState := tcb.State.State
	nextState, ctrl := tcp.StateMachine(curState, 0, "active")
	//fmt.Println(ctrl)
	tcb.State.State = nextState
	tcb.SendCtrlMsg(ctrl, true, true)

	//SendCtrlMsg(saddr.LocalAddr, saddr.RemoteAddr, saddr.LocalPort, saddr.RemotePort, tcp.FIN, u)
	//-----------
	//Change State
	//-----------
	//Check ACK
	//-----------

	return 0
}

func (manager *SocketManager) V_accept(socket int, addr string, port int, node *pkg.Node, u linklayer.UDPLink) int {
	//idx := manager.V_socket(node, u)
	//saddr :=
	return 0
}

func (manager *SocketManager) V_read(socket int, nbyte int, check string) (int, []byte) {
	buf := make([]byte, 0)
	tcb, ok := manager.FdToSocket[socket]
	if !ok {
		fmt.Println("Socket doesn't exist!\n")
		return -1, buf
	}
	if tcb.BlockRead {
		fmt.Println("v_read() error: Operation not permitted")
		return -1, buf
	}
	readLen := 0

	//When receive buffer is empty
	if tcb.RecvW.LastByteRead == tcb.RecvW.NextByteExpected-1 {
		for {
			retbuf, ret := tcb.RecvW.Read(nbyte)
			readLen += ret
			buf = append(buf, retbuf...)
			if readLen > 0 {
				break
			}
		}
	} else {
		//When receive buffer is not empty
		retbuf, ret := tcb.RecvW.Read(nbyte)
		readLen += ret
		buf = append(buf, retbuf...)
	}

	//Check if block flag is yes;
	if check == "y" {
		fmt.Println("reach y.......")
		for readLen < nbyte {
			addBuf, count := tcb.RecvW.Read(1)
			buf = append(buf, addBuf...)
			readLen = readLen + count
		}
	}

	return readLen, buf
}

func (manager *SocketManager) V_write(socket int, data []byte) int {
	tcb, ok := manager.FdToSocket[socket]
	if !ok {
		fmt.Println("Socket doesn't exist!\n")
		return -1
	}
	if tcb.BlockWrite {
		fmt.Println("v_write() error: Operation not permitted")
		return -1
	}
	if len(data) > tcb.SendW.AdvertisedWindow {
		return -1
	}

	if tcb.State.State != 5 {
		return -1
	}

	writeLen := tcb.SendW.Write(data)
	if writeLen != -1 {
		tcb.SendData(data)
	}
	return writeLen
}

func (manager *SocketManager) V_shutdown(socket int, ntype int) int {
	// 1 = write; 2 = read; 3 = both
	tcb, ok := manager.FdToSocket[socket]
	if !ok {
		fmt.Println("Socket doesn't exist!\n")
		return -1
	}
	//Socket is closing or closed already
	if tcb.State.State == tcp.CLOSED {
		fmt.Println("Socket is closed already!\n")
		return 0
	}
	//Socket is listening
	if tcb.State.State == tcp.LISTEN {
		fmt.Println("Socket is listening and cannot be shutdown! \n")
		return -1
	}

	switch ntype {
	case 1:
		//Block write
		//Check tcb state first
		if !(tcb.State.State == tcp.ESTAB) && !(tcb.State.State == tcp.CLOSEWAIT) && !(tcb.State.State == tcp.SYNRCVD) {
			fmt.Println("Socket state is not ESTAB, CLOSEWAIT, nor SYNRCVD, and cannot be shutdown!\n")
			return -1
		}

		if tcb.BlockWrite {
			return 0
		}

		tcb.BlockWrite = true
		//Send FIN
		newState, cf := tcp.StateMachine(tcb.State.State, 0, "CLOSE")
		tcb.State.State = newState

		fmt.Printf("Shutdown Write: this is the cf: %d \n", cf)
		tcb.SendCtrlMsg(cf, false, true)
		fmt.Println("Shutdown Write seqnum and ack num : ", tcb.Seq, tcb.Ack)
	case 2:
		//Block read;
		tcb.BlockRead = true
	case 3:
		//Block both write and read
		//Check tcb state first
		if !(tcb.State.State == tcp.ESTAB) && !(tcb.State.State == tcp.CLOSEWAIT) && !(tcb.State.State == tcp.SYNRCVD) {
			fmt.Println("Socket state is not ESTAB, CLOSEWAIT, nor SYNRCVD, and cannot be shutdown!\n")
			return -1
		}

		tcb.BlockRead = true
		if tcb.BlockWrite {
			return 0
		}
		tcb.BlockWrite = true

		//Send FIN
		newState, cf := tcp.StateMachine(tcb.State.State, 0, "CLOSE")
		tcb.State.State = newState

		fmt.Printf("Shutdown both: this is the cf: %d \n", cf)
		tcb.SendCtrlMsg(cf, false, true)
		fmt.Println("Shutdown Write seqnum and ack num : ", tcb.Seq, tcb.Ack)

	}
	return 0
}

func (manager *SocketManager) V_close(socket int) int {
	tcb, ok := manager.FdToSocket[socket]
	if !ok {
		fmt.Println("Socket doesn't exist!\n")
		return -1
	}
	//Socket closed already
	if tcb.State.State == tcp.CLOSED {
		fmt.Println("Socket closed already!\n")
		return 0
	} else if tcb.State.State == tcp.LISTEN {
		fmt.Printf("V_accept() error on socket %d: Software caused connection abort", socket)
		//Delete TCB
		delete(manager.FdToSocket, socket)
		delete(manager.AddrToSocket, tcb.Addr)
		return 0
	} else if tcb.State.State == tcp.SYNSENT {
		//Delete TCB
		delete(manager.FdToSocket, socket)
		delete(manager.AddrToSocket, tcb.Addr)
		return 0
	} else {
		manager.V_shutdown(socket, 3)
		if tcb.State.State == tcp.FINWAIT2 {
			newState, _ := tcp.StateMachine(tcb.State.State, tcp.FIN, "")
			tcb.State.State = newState
			go manager.TimeWaitTimeOut(tcb, 1000)
		} else if tcb.State.State == tcp.LASTACK {
			newState, _ := tcp.StateMachine(tcb.State.State, tcp.ACK, "")
			tcb.State.State = newState
		}

		return 0
	}
}

//helper Function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (manager *SocketManager) TimeWaitTimeOut(tcb *TCB, num int) {
	for {
		if tcb.State.State == tcp.TIMEWAIT {
			fmt.Println("You are inside of the TimeWaitTimeOut loop!\n")
			time.Sleep(time.Duration(num) * time.Millisecond)
			newState, _ := tcp.StateMachine(tcb.State.State, 0, "")
			tcb.State.State = newState
			break
		}

	}
}

func (manager *SocketManager) GetEstabSocket(localAddr string, localPort int) int {
	for {
		for socket, tcb := range manager.FdToSocket {
			if tcb.State.State == tcp.ESTAB {
				if tcb.Addr.LocalAddr == localAddr && tcb.Addr.LocalPort == localPort {
					return socket
				}
			}
		}
	}
}
