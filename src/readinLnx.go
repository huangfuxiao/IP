package main

import (
	"./pkg"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

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
func perror(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	//Read in a lnx file
	fileName := os.Args[1]
	lines, err := readLines(fileName)

	//Initialize a node struct and its rounting table
	var thisNode pkg.Node
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

		thisNode.InterfaceArray = append(thisNode.InterfaceArray, thisLink)

		//Put this link into the rounting table
		var thisEntry pkg.Entry
		thisEntry.Cost = 0
		thisEntry.Dest = thisLink.Src
		thisEntry.Next = thisLink.Src
		thisEntry.Time_to_live = 12
		thisRT[thisLink.Src] = thisEntry
	}

	fmt.Println(thisRT)
	thisNode.RouteTable = thisRT
}
