package ipv4

import (
	"encoding/json"
	"fmt"
)

type RIP struct {
	Command    int // command: 1 - request, 2 - response
	NumEntries int
	Entries    []RIPEntry
}

type RIPEntry struct {
	Cost    int
	Address string
}

//Convert []byte to RIP packet
func ConvertBytesToRIP(data []byte) RIP {
	var newRip RIP
	err := json.Unmarshal(data, &newRip)
	if err != nil {
		fmt.Println("error:", err)
	}
	return newRip
}

//Convert RIP to []byte
func ConvertRipToBytes(rip RIP) []byte {
	b, err := json.Marshal(rip)
	if err != nil {
		fmt.Println("error with json.Marshal:", err)
	}
	return b
}
