package main

import (
	"./pkg"
	"fmt"
)

func main() {

	//test go programming
	example := pkg.Interface{Status: 1, Addr: "10.10.168.73"}
	fmt.Println(example)

	entryEx := pkg.Entry{"dest", "next", 1, 1}
	fmt.Println(entryEx)

	m := make(map[string]pkg.Entry)
	m["k1"] = entryEx

	arrEx := []pkg.Interface{example}
	nodeEx := pkg.Node{1, arrEx, m}
	fmt.Println(nodeEx)

}
