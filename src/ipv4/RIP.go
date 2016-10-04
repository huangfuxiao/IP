package ipv4

type RIP struct {
	command    int // command: 1 - request, 2 - response
	numEntries int
	entries    []RIPEntry
}

type RIPEntry struct {
	cost    int
	address string
}
