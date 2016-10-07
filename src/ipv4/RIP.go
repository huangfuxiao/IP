package ipv4

type RIP struct {
	Command    int // command: 1 - request, 2 - response
	NumEntries int
	Entries    []RIPEntry
}

type RIPEntry struct {
	Cost    int
	Address string
}
