package runner

import (
	"../api"
	"../linklayer"
	"../pkg"
	"../tcp"
	//"bufio"
	//"bytes"
	"fmt"
	//"io"
	"os"
	"strconv"
	"sync"
	"time"
)

func Sendfile_thread(udp linklayer.UDPLink, thisNode *pkg.Node, mutex *sync.RWMutex, thisSocketManager *api.SocketManager, cmds []string) {
	port, err := strconv.Atoi(cmds[3])
	if err != nil {
		fmt.Println("syntax error (usage: send [interface] [payload])\n")
		return
	}
	fmt.Printf("connect ip port %d", port)
	socketFd := thisSocketManager.V_socket(thisNode, udp)
	fmt.Println(socketFd)
	thisSocketManager.V_bind(socketFd, "", -1)
	thisSocketManager.V_connect(socketFd, cmds[2], port)

	//How to time out?????????????
	for {
		//time.Sleep(3000 * time.Millisecond)
		tcb, _ := thisSocketManager.FdToSocket[socketFd]

		if tcb.State.State == tcp.ESTAB {
			//fmt.Println("v_connect() error: No route to host or invalid port\n")
			break
		}
	}

	//Read in the file
	file, err := os.OpenFile(cmds[1], os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("File cannot be open!\n")
		return
	}
	fmt.Println("STARTING SENDFILE! ")

	//Loop through each line and write to SendBuffer
	for {
		toSend := make([]byte, 1024)
		n, _ := file.Read(toSend)
		if n == 0 {
			break
		}
		toSend = toSend[:n]

		ok := thisSocketManager.V_write(socketFd, toSend)
		if ok > -1 {
			//fmt.Println("V_write successfully wrote ", ok, " bytes")
		}
		//How long to sleep?
		//time.Sleep(10 * time.Millisecond)
	}
	fmt.Println("FINISHED SENDFILE! ")
	//TO DO;
	time.Sleep(1000 * time.Millisecond)
	thisSocketManager.V_close(socketFd)
	// time.Sleep(12000 * time.Millisecond)
	// thisSocketManager.V_shutdown(socketFd, 3)
	// for {
	// 	time.Sleep(1000 * time.Millisecond)
	// 	if tcb.State.State == tcp.FINWAIT2 || tcb.State.State == tcp.CLOSED {
	// 		break
	// 	}
	// }

	// //Close connection at my side
	// if tcb.State.State == tcp.FINWAIT2 {
	// 	newState, _ := tcp.StateMachine(tcb.State.State, tcp.FIN, "")
	// 	tcb.State.State = newState
	// }

	//go thisSocketManager.TimeWaitTimeOut(tcb, 1000)

	// for {
	// 	time.Sleep(1000 * time.Millisecond)
	// 	if tcb.State.State == tcp.CLOSED {
	// 		thisSocketManager.PrintSockets()
	// 		break
	// 	}

	// }
	// fmt.Println("FINISHED SENDFILE! ")
}
