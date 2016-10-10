package runner

import (
	"../pkg"
	"fmt"
	"strings"
	"time"
)

func Timeout_thread(node *pkg.Node) {
	for {
		for _, v := range node.RouteTable {
			time_now := time.Now().Unix()
			if strings.Compare(v.Next, v.Dest) != 0 {
				if v.Ttl < time_now {
					fmt.Println(v.Ttl)
					fmt.Println(time_now)
					//fmt.Println(node.RouteTable[v.Dest].Cost)
					node.RouteTable[v.Dest] = pkg.Entry{Dest: v.Dest, Next: v.Next, Cost: 16, Ttl: time.Now().Unix()}
					fmt.Println(node.RouteTable[v.Dest])
					//fmt.Println(node.RouteTable[v.Dest].Cost)
				}
			}

		}
		time.Sleep(12 * time.Second)

	}
}
