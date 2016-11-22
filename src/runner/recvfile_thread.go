package runner

import (
	"../api"
	"../linklayer"
	"../pkg"
	//"bufio"
	"../tcp"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

func Recvfile_thread(udp linklayer.UDPLink, thisNode *pkg.Node, mutex *sync.RWMutex, thisSocketManager *api.SocketManager, cmds []string) {
	//V_listen to the port
	port, err := strconv.Atoi(cmds[2])
	if err != nil {
		fmt.Println("syntax error (usage: recvfile [filename] [port])\n")
		return
	}
	fmt.Printf("Listen to port %d\n", port)
	socketFd := thisSocketManager.V_socket(thisNode, udp)
	thisSocketManager.V_bind(socketFd, "", port)
	thisSocketManager.V_listen(socketFd)
	time.Sleep(6000 * time.Millisecond)

	//Find the new socket after connection is established
	tcb, _ := thisSocketManager.FdToSocket[socketFd]
	newSocket := thisSocketManager.GetEstabSocket(tcb.Addr.LocalAddr, tcb.Addr.LocalPort)
	newTcb, _ := thisSocketManager.FdToSocket[newSocket]
	//Delete old socket
	saddr := tcb.Addr
	delete(thisSocketManager.FdToSocket, socketFd)
	delete(thisSocketManager.AddrToSocket, saddr)

	//Open the file to write
	f, err := os.OpenFile(cmds[1], os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	check(err)
	fmt.Println("Start Receiving File")
	//Keep readin and write out, until the buffer is empty
	for {
		ok, bufReadIn := thisSocketManager.V_read(newSocket, 4096, "n")
		//time.Sleep(10 * time.Millisecond)
		if ok != -1 {
			//fmt.Printf("This is the string readin: %v\n", string(bufReadIn))
			writeLines(f, bufReadIn)
		}

		if newTcb.State.State == tcp.CLOSEWAIT {
			//fmt.Println("break")
			break
		}
	}

	//If readbuffer is not empty, continue read
	//This is silly but works
	for {
		ok, bufReadIn := thisSocketManager.V_read(newSocket, 4096, "n")
		if ok != -1 {
			// fmt.Println("Read bytes ", ok)
			// fmt.Printf("This is the string readin: %v\n", string(bufReadIn))
			writeLines(f, bufReadIn)
		}
		if newTcb.RecvW.LastByteRead == newTcb.RecvW.NextByteExpected-1 {
			//fmt.Println("break out 2")
			break
		}
		//fmt.Println("loop 2")
	}

	fmt.Println("End Receiving File")
	thisSocketManager.V_close(socketFd)
	f.Close()
	// thisSocketManager.V_shutdown(newSocket, 3)

	//Close connection at my side

	// fmt.Println("This is the new state after FIN: ", newState)
	// for {
	// 	time.Sleep(1000 * time.Millisecond)
	// 	if newTcb.State.State == tcp.TIMEWAIT {
	// 		break
	// 	}
	// }
	// go thisSocketManager.TimeWaitTimeOut(newTcb, 1000)

}

func writeLines(f *os.File, toWrite []byte) {

	// You can `Write` byte slices as you'd expect.
	_, err2 := f.Write(toWrite)
	check(err2)

	return
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
