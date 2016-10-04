package main

import (
	"./pkg"
	"fmt"
	"log"
	"os"
)

type node struct {
	port           int
	interfaceArray []pkg.Interface
	routeTable     map[string]int
}

func main() {
	//readin file
	fileName := os.Args[1]

	//test go programming
	var interArray []pkg.Interface
	example := pkg.Interface{Status: 1, Addr: "10.10.168.73"}
	interArray = append(interArray, example)
	m := make(map[string]int)
	m["route"] = 66
	node1 := node{port: 50001, interfaceArray: interArray, routeTable: m}
	fmt.Println(node1.port)
	fmt.Println(node1.interfaceArray)
	fmt.Println(node1.routeTable)
	fmt.Println(fileName)

	//run main handler
	file, err := os.Open("fileName") // For read access.
	if err != nil {
		log.Fatal(err)
	}

	for {
		b1 := make([]byte, 50)
		n1, err2 := file.Read(b1)
		//check(err)
		if err2 != nil {
			log.Fatal(err2)
		}
		fmt.Printf("Reading %d bytes: %s\n", n1, string(b1))
	}

}
