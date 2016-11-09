package main

import (
	"./api"
	"./handler"
	"./ipv4"
	"./linklayer"
	"./pkg"
	"./runner"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Panic if catches an error
func perror(err error) {
	if err != nil {
		panic(err)
	}
}

// Read a whole file into the memory and store it as array of lines
func readLines(path string) (lines []string, err error) {
	var (
		file   *os.File
		part   []byte
		prefix bool
	)
	if file, err = os.Open(path); err != nil {
		return
	}
	reader := bufio.NewReader(file)
	buffer := bytes.NewBuffer(make([]byte, 0))
	for {
		if part, prefix, err = reader.ReadLine(); err != nil {
			break
		}
		buffer.Write(part)
		if !prefix {
			lines = append(lines, buffer.String())
			buffer.Reset()
		}
	}
	if err == io.EOF {
		err = nil
	}
	return
}

func readinLnx(fileName string) (thisNode pkg.Node) {
	//Read in a lnx file
	lines, err := readLines(fileName)

	//Initialize a node struct and its rounting table
	thisRT := make(map[string]pkg.Entry)

	//Put first line into this node's local address and port
	var localInfor []string
	localInfor = strings.Split(lines[0], ":")
	p, err := strconv.Atoi(strings.Trim(localInfor[1], " "))
	perror(err)
	//fmt.Printf("local's phyAddr: %s\n", localInfor[0])
	//fmt.Printf("local's port number: %d\n", p)
	thisNode.LocalAddr = localInfor[0]
	thisNode.Port = p

	//Loop through each link lines and add the link interface to this node's InterfaceArray
	for _, line := range lines[1:] {
		var thisLink pkg.Interface
		var linkInfor []string
		linkInfor = strings.Split(line, " ")
		thisLink.Status = 1
		thisLink.Src = linkInfor[1]
		thisLink.Dest = linkInfor[2]

		var remoteAddr []string
		remoteAddr = strings.Split(linkInfor[0], ":")
		q, err := strconv.Atoi(strings.Trim(remoteAddr[1], " "))
		perror(err)
		//fmt.Printf("remote's phyAddr: %s\n", remoteAddr[0])
		//fmt.Printf("remote's port number: %d\n", q)
		thisLink.RemoteAddr = remoteAddr[0]
		thisLink.RemotePort = q

		thisNode.InterfaceArray = append(thisNode.InterfaceArray, &thisLink)

		//Put this link into the rounting table
		var thisEntry pkg.Entry
		thisEntry.Cost = 0
		thisEntry.Dest = thisLink.Src
		thisEntry.Next = thisLink.Src
		thisEntry.Ttl = time.Now().Unix() + 12
		thisRT[thisLink.Src] = thisEntry
	}

	//fmt.Println(thisRT)
	thisNode.RouteTable = thisRT
	return thisNode
}

func printHelp() {
	fmt.Println("Commands: ")
	fmt.Println("accept [port]\t\t\t\t- Spawn a socket, bind it to the given port, and start ")
	fmt.Println("\t\t\t\t\t  accepting connections on that port")
	fmt.Println("connect [ip] [port] \t\t\t- Attempt to connect to the given ip address, ")
	fmt.Println("\t\t\t\t\t  in dot notition, on the given port.")
	fmt.Println("send [socket] [data]\t\t\t- Send a string on a socket.")
	fmt.Println("recv [socket] [numbytes] [y/n]\t\t- Try to read data from a given socket. ")
	fmt.Println("\t\t\t\t\t  If the last argument is y, then you should block ")
	fmt.Println("\t\t\t\t\t  until numbytes is received,")
	fmt.Println("\t\t\t\t\t  or the connection closes. If n, then don.t block;")
	fmt.Println("\t\t\t\t\t  return whatever recv returns. Default is n.")
	fmt.Println("sendfile [filename] [ip] [port]\t\t- Connect to the given ip and port, ")
	fmt.Println("\t\t\t\t\t  send the entirety of the specified file,")
	fmt.Println("\t\t\t\t\t  and close the connection.")
	fmt.Println("recvfile [filename] [port]\t\t- Listen for a connection on the given port. ")
	fmt.Println("\t\t\t\t\t  Once established, write everything you can read from ")
	fmt.Println("\t\t\t\t\t  the socket to the given file. ")
	fmt.Println("\t\t\t\t\t  Once the other side closes the connection, ")
	fmt.Println("\t\t\t\t\t  close the connection as well.")
	fmt.Println("shutdown [socket] [read/write/both]\t- v_shutdown on the given socket.")
	fmt.Println("close [socket] \t\t\t\t- v_close on the given socket.")
	fmt.Println("up <id>\t\t\t\t\t- Bring one interface up")
	fmt.Println("down <id>\t\t\t\t- Bring one interface down")
	fmt.Println("interfaces\t\t\t\t- List interfaces")
	fmt.Println("routes\t\t\t\t\t- List routing table rows")
	fmt.Println("socketss\t\t\t\t- List sockets (fd, ip, port, state)")
	fmt.Println("window [socket]\t\t\t\t- List window sizes for socket")
	fmt.Println("quit\t\t\t\t\t- No cleanup, exit(0)")
	fmt.Println("help\t\t\t\t\t- Show this help")
}

func initRIP(node pkg.Node, udp linklayer.UDPLink) {
	var newRip ipv4.RIP
	var ripEntries []ipv4.RIPEntry
	newRip.Command = 1
	newRip.NumEntries = 0
	newRip.Entries = ripEntries
	//Loop through interfaces and send to all neighbors
	for _, link := range node.InterfaceArray {
		ipPkt := handler.ConvertRipToIpPackage(newRip, link.Src, link.Dest)
		udp.Send(ipPkt, link.RemoteAddr, link.RemotePort)
	}
}

func main() {
	if len(os.Args) < 2 {
		println("ERROR: please input a link file")
		os.Exit(1)
	}

	//Read in lnx file and initialize this node
	fileName := os.Args[1]

	fmt.Println(fileName)
	thisNode := readinLnx(fileName)

	udp := linklayer.InitUDP(thisNode.LocalAddr, thisNode.Port)
	//Initialize this Socket Manager
	thisSocketManager := api.BuildSocketManager(thisNode.InterfaceArray)

	initRIP(thisNode, udp)

	var mutex = &sync.RWMutex{}

	var wg sync.WaitGroup
	wg.Add(4)

	go runner.Receive_thread(udp, &thisNode, mutex, &thisSocketManager)
	go runner.Send_thread(&thisNode, udp, mutex)
	go runner.Timeout_thread(&thisNode, mutex)

	//main handler
	for {
		fmt.Println(">")
		reader := bufio.NewReader(os.Stdin)
		//fmt.Println("Enter text: ")
		text, _ := reader.ReadString('\n')
		cmds := strings.Fields(text)

		if len(cmds) == 0 {
			continue
		} else {
			switch cmds[0] {
			case "help":
				printHelp()
			case "interfaces":
				thisNode.PrintInterfaces()
			case "routes":
				thisNode.PrintRoutes()
			case "down":
				if len(cmds) == 1 {
					fmt.Println("invalid interface id")
				} else {
					id, err := strconv.Atoi(cmds[1])
					if err != nil {
						fmt.Println("invalid interface id\n")
						continue
					}
					thisNode.InterfacesDown(id, mutex)
				}
			case "up":
				if len(cmds) == 1 {
					fmt.Println("invalid interface id")
				} else {
					id, err := strconv.Atoi(cmds[1])
					if err != nil {
						fmt.Println("invalid interface id\n")
						continue
					}
					thisNode.InterfacesUp(id, mutex)
				}
			case "send":
				if len(cmds) < 3 {
					fmt.Println("syntax error (usage: send [interface] [payload])\n")
				} else {
					//thisNode.PrepareAndSendPacket(cmds, udp, mutex)
					//TCP additions start here
					socketFd, err := strconv.Atoi(cmds[1])
					if err != nil {
						fmt.Println("syntax error (usage: send [interface] [payload])\n")
						continue
					}
					toSend := []byte(cmds[2])
					thisSocketManager.V_write(socketFd, toSend, len(toSend))
				}
			case "recv":
				if len(cmds) < 3 {
					fmt.Println("syntax error (usage: recv [interface] [bytes to read] [loop? (y/n), optional])\n")
				} else {
					//TCP additions start here
					socketFd, err1 := strconv.Atoi(cmds[1])
					nbyte, err2 := strconv.Atoi(cmds[2])
					if err1 != nil || err2 != nil {
						fmt.Println("syntax error (usage: recv [interface] [bytes to read] [loop? (y/n), optional])\n")
						continue
					}
					buf := make([]byte, 0, nbyte)
					_, bufReadIn := thisSocketManager.V_read(socketFd, buf, nbyte, "y")
					fmt.Printf("This is the string readin: %v\n", bufReadIn)
				}
			case "sockets":
				thisSocketManager.PrintSockets(thisNode.InterfaceArray)
			case "accept":
				if len(cmds) < 2 {
					fmt.Println("syntax error (usage: accept [port])\n")
				} else {
					port, err := strconv.Atoi(cmds[1])
					if err != nil {
						fmt.Println("syntax error (usage: accept [port])\n")
						continue
					}
					//fmt.Printf("accept port %d\n", port)
					socketFd := thisSocketManager.V_socket(&thisNode, udp)
					thisSocketManager.V_bind(socketFd, "", port)
					thisSocketManager.V_listen(socketFd)
				}
			case "connect":
				if len(cmds) < 3 {
					fmt.Println("syntax error (usage: connect [ip address] [port])\n")
				} else {
					port, err := strconv.Atoi(cmds[2])
					if err != nil {
						fmt.Println("syntax error (usage: accept [port])\n")
						continue
					}
					fmt.Printf("connect ip port %d", port)
					socketFd := thisSocketManager.V_socket(&thisNode, udp)
					fmt.Println(socketFd)
					thisSocketManager.V_bind(socketFd, "", -1)
					thisSocketManager.V_connect(socketFd, cmds[1], port)
				}
			case "shutdown":
				if len(cmds) < 3 {
					fmt.Println("syntax error (usage: shutdown [socket] [shutdown type])\n")
				} else {
					socketFD, err := strconv.Atoi(cmds[1])
					if err != nil {
						fmt.Println("syntax error (usage: shutdown [socket] [shutdown type])\n")
						continue
					}
					if cmds[2] == "write" {
						thisSocketManager.V_shutdown(socketFD, 1)
					} else if cmds[2] == "read" {
						thisSocketManager.V_shutdown(socketFD, 2)
					} else if cmds[2] == "both" {
						thisSocketManager.V_shutdown(socketFD, 3)
					} else {
						fmt.Println("syntax error (usage: shutdown [socket] [shutdown type])\n")
						continue
					}

				}
			case "close":
				if len(cmds) < 2 {
					fmt.Println("syntax error (usage: close [socket])\n")
				} else {
					socketFD, err := strconv.Atoi(cmds[1])
					if err != nil {
						fmt.Println("syntax error (usage: close [socket])\n")
						continue
					}
					thisSocketManager.V_close(socketFD)
				}

			case "quit":
				os.Exit(1)
			default:
				fmt.Println("Invalid Command!\n")
				printHelp()
			}
		}
	}
	wg.Wait()

}
