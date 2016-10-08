package main

import (
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
	buffer := bytes.NewBuffer(make([]byte, 1024))
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
	fmt.Printf("local's phyAddr: %s\n", localInfor[0])
	fmt.Printf("local's port number: %d\n", p)
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
		fmt.Printf("remote's phyAddr: %s\n", remoteAddr[0])
		fmt.Printf("remote's port number: %d\n", q)
		thisLink.RemoteAddr = remoteAddr[0]
		thisLink.RemotePort = q

		thisNode.InterfaceArray = append(thisNode.InterfaceArray, &thisLink)

		//Put this link into the rounting table
		var thisEntry pkg.Entry
		thisEntry.Cost = 0
		thisEntry.Dest = thisLink.Src
		thisEntry.Next = thisLink.Src
		thisRT[thisLink.Src] = thisEntry
	}

	fmt.Println(thisRT)
	thisNode.RouteTable = thisRT
	return thisNode
}

func printHelp() {
	fmt.Println("******************************")
	fmt.Println("help\t\t\t\tHelp Printing")
	fmt.Println("interfaces\t\t\tInterface Information")
	fmt.Println("routes\t\t\tRouting table")
	fmt.Println("down <id>\t\t\t\tBring one interface down")
	fmt.Println("up <id>\t\t\t\tBring one interface up")
	fmt.Println("send <dst_ip> <prot> <payload>\t\t\t\tSend the message to a virtual IP")
	fmt.Println("quit\t\t\t\tQUIT")
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

	//test go programming
	/*example := pkg.Interface{Status: 1, Addr: "10.10.168.73"}
	fmt.Println(example)

	entryEx := pkg.Entry{"dest", "next", 1, 1}
	fmt.Println(entryEx)

	m := make(map[string]pkg.Entry)
	m["k1"] = entryEx

	arrEx := []pkg.Interface{example}
	nodeEx := pkg.Node{1, arrEx, m}
	fmt.Println(nodeEx)*/

	fileName := os.Args[1]
	fmt.Printf("Args' length: %d \n", len(os.Args))
	if len(os.Args) < 2 {
		println("ERROR: please input a link file")
		os.Exit(1)
	}
	fmt.Println(fileName)

	thisNode := readinLnx(fileName)
	fmt.Printf("thisNode made successfully and has local physical addr: %s\n", thisNode.LocalAddr)

	udp := linklayer.InitUDP(thisNode.LocalAddr, thisNode.Port)

	initRIP(thisNode, udp)

	var wg sync.WaitGroup
	wg.Add(3)

	go runner.Receive_thread(udp, &thisNode)
	go runner.Send_thread(&thisNode, udp)
	//main handler
	//This is silly but works
	for {
		fmt.Println(">")
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Enter text: ")
		text, _ := reader.ReadString('\n')
		cmds := strings.Fields(text)

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
				id, _ := strconv.Atoi(cmds[1])
				thisNode.InterfacesDown(id)
			}
		case "up":
			if len(cmds) == 1 {
				fmt.Println("invalid interface id")
			} else {
				id, _ := strconv.Atoi(cmds[1])
				thisNode.InterfacesUp(id)
			}
		case "send":
			thisNode.PrepareAndSendPacket()
		case "quit":
			os.Exit(1)
		default:
			fmt.Println("Invalid Command!\n")
			printHelp()
		}
	}
	wg.Wait()

}
