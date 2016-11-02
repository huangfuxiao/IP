package tcp

import (
//"fmt"
)

const (
	ERROR     = 0
	CLOSED    = 1
	LISTEN    = 2
	SYNSENT   = 3
	SYNRCVD   = 4
	ESTAB     = 5
	FINWAIT1  = 6
	FINWAIT2  = 7
	CLOSING   = 8
	TIMEWAIT  = 9
	CLOSEWAIT = 10
	LASTACK   = 11
)

type State struct {
	State int
}

func StateMachine(s int, flag int, api string) (int, int) {
	nextstate, nextflag := 0, 0
	if s == CLOSED {
		nextstate, nextflag = ClosedState(api)
	} else if s == LISTEN {
		nextstate, nextflag = ListenState(flag, api)
	} else if s == SYNSENT {
		nextstate, nextflag = SynSentState(flag, api)
	} else if s == SYNRCVD {
		nextstate, nextflag = SynRcvdState(flag, api)
	} else if s == ESTAB {
		nextstate, nextflag = EstabState(flag, api)
	} else if s == FINWAIT1 {
		nextstate, nextflag = FINWait1State(flag, api)
	} else if s == FINWAIT2 {
		nextstate, nextflag = FINWait2State(flag, api)
	} else if s == CLOSING {
		nextstate, nextflag = ClosingState(flag, api)
	} else if s == TIMEWAIT {
		nextstate, nextflag = TimeWaitState(flag, api)
	} else if s == CLOSEWAIT {
		nextstate, nextflag = CloseWaitState(flag, api)
	} else if s == LASTACK {
		nextstate, nextflag = LastAckState(flag, api)
	}
	return nextstate, nextflag
}

func ClosedState(api string) (int, int) {
	if api == "active" {
		return SYNSENT, SYN
	} else if api == "passive" {
		return LISTEN, NOTHING
	}
	return ERROR, NOTHING
}

func ListenState(flag int, api string) (int, int) {
	if flag == SYN {
		return SYNRCVD, SYN + ACK
	} else if flag == NOTHING {
		if api == "SEND" {
			return SYNSENT, SYN
		} else if api == "CLOSE" {
			return CLOSED, NOTHING
		}
	}
	return ERROR, NOTHING

}

func SynSentState(flag int, api string) (int, int) {
	if flag == SYN {
		return SYNRCVD, ACK
	} else if flag == SYN+ACK {
		return ESTAB, ACK
	} else if api == "CLOSE" {
		return CLOSED, NOTHING
	}
	return ERROR, NOTHING
}

func SynRcvdState(flag int, api string) (int, int) {
	if flag == ACK {
		return ESTAB, NOTHING
	} else if api == "CLOSE" {
		return FINWAIT1, FIN
	}
	return ERROR, NOTHING
}

func EstabState(flag int, api string) (int, int) {
	if flag == FIN {
		return CLOSEWAIT, ACK
	} else if api == "CLOSE" {
		return FINWAIT1, FIN
	}
	return ERROR, NOTHING
}

func FINWait1State(flag int, api string) (int, int) {
	if flag == FIN {
		return CLOSING, ACK
	} else if flag == ACK {
		return FINWAIT2, NOTHING
	}
	return ERROR, NOTHING
}

func FINWait2State(flag int, api string) (int, int) {
	if flag == FIN {
		return TIMEWAIT, ACK
	}
	return ERROR, NOTHING
}

func ClosingState(flag int, api string) (int, int) {
	if flag == ACK {
		return TIMEWAIT, NOTHING
	}
	return ERROR, NOTHING
}

func TimeWaitState(flag int, api string) (int, int) {
	/*
		NOT COMPLETE YET
		TIME OUT = 2MSL
		CHANGE TO CLOSED*/
	return CLOSED, NOTHING
}

func CloseWaitState(flag int, api string) (int, int) {
	if api == "CLOSE" {
		return LASTACK, FIN
	}
	return ERROR, NOTHING
}

func LastAckState(flag int, api string) (int, int) {
	if flag == ACK {
		return CLOSED, NOTHING
	}
	return ERROR, NOTHING
}

func StateString(s int) string {
	if s == 1 {
		return "CLOSED"
	} else if s == 2 {
		return "LISTEN"
	} else if s == 3 {
		return "SYN SENT"
	} else if s == 4 {
		return "SYN RCVD"
	} else if s == 5 {
		return "ESTAB"
	} else if s == 6 {
		return "FIN WAIT 1"
	} else if s == 7 {
		return "FIN WAIT 2"
	} else if s == 8 {
		return "CLOSING"
	} else if s == 9 {
		return "TIME WAIT"
	} else if s == 10 {
		return "CLOSE WAIT"
	} else if s == 11 {
		return "LAST ACK"
	}
	return "ERROR"
}
