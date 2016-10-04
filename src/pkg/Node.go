package pkg

type Node struct {
	Port           int
	InterfaceArray []Interface
	RouteTable     map[string]Entry
}

type Interface struct {
	Status int
	Addr   string
}

type Entry struct {
	Dest         string
	Next         string
	Cost         int
	Time_to_live int
}
