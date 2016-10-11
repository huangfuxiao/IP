package runner

import (
	"../pkg"
	//"fmt"
	"strings"
	"sync"
	"time"
)

func Timeout_thread(node *pkg.Node, mutex *sync.RWMutex) {
	i := 0
	for {
		mutex.RLock()
		i = 1
		for _, v := range node.RouteTable {
			time_now := time.Now().Unix()
			if strings.Compare(v.Next, v.Dest) != 0 {
				if v.Ttl < time_now {
					if i == 1 {
						mutex.RUnlock()
						i = 0
					}
					//fmt.Println(v.Ttl)
					//fmt.Println(time_now)
					//fmt.Println(node.RouteTable[v.Dest].Cost)
					mutex.Lock()
					node.RouteTable[v.Dest] = pkg.Entry{Dest: v.Dest, Next: v.Next, Cost: 16, Ttl: time.Now().Unix()}
					mutex.Unlock()
					//fmt.Println(node.RouteTable[v.Dest])
					//fmt.Println(node.RouteTable[v.Dest].Cost)
				}
			}

		}
		if i == 1 {
			mutex.RUnlock()
			i = 0
		}

		time.Sleep(5 * time.Second)

	}
}
