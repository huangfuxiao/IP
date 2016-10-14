package ipv4

import (
	"encoding/binary"
	//"fmt"
	"net"
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

/*
//Convert []byte to RIP packet
func ConvertBytesToRIP(data []byte) RIP {
	var newRip RIP
	err := json.Unmarshal(data, &newRip)
	if err != nil {
		fmt.Println("error:", err)
	}
	return newRip
}
*/

//Convert RIP to []byte
/*
func ConvertRipToBytes(rip RIP) []byte {

	b, err := json.Marshal(rip)
	if err != nil {
		fmt.Println("error with json.Marshal:", err)
	}
	return b
}
*/

func ConvertBytesToRIP(data []byte) RIP {
	command := int(binary.BigEndian.Uint16(data[0:2]))
	num := int(binary.BigEndian.Uint16(data[2:4]))
	entries := make([]RIPEntry, 0)
	for i := 0; i < num; i++ {
		cost := int(binary.BigEndian.Uint32(data[4+8*i : 8+8*i]))
		address := int2ip(net.IPv4(data[8+8*i], data[9+8*i], data[10+8*i], data[11+8*i]))
		entries = append(entries, RIPEntry{cost, address})
	}
	return RIP{command, num, entries}
}

func ConvertRipToBytes(rip RIP) []byte {
	riplen := 4 + 8*rip.NumEntries
	b := make([]byte, riplen)
	binary.BigEndian.PutUint16(b[0:2], uint16(rip.Command))
	binary.BigEndian.PutUint16(b[2:4], uint16(rip.NumEntries))
	for i := 0; i < rip.NumEntries; i++ {
		binary.BigEndian.PutUint32(b[4+8*i:8+8*i], uint32(rip.Entries[i].Cost))
		ipbyte4 := ip2int(rip.Entries[i].Address)
		b[8+8*i] = ipbyte4[0]
		b[9+8*i] = ipbyte4[1]
		b[10+8*i] = ipbyte4[2]
		b[11+8*i] = ipbyte4[3]
	}
	return b

}

func ip2int(ipa string) []byte {
	return net.ParseIP(ipa).To4()
}

func int2ip(nip net.IP) string {
	return nip.String()
}
