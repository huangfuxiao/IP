package tcp

type State struct {
	State  string
	Action string
}

type StateMachine struct {
	CurrentState string
	NextState    map[State]State
}
