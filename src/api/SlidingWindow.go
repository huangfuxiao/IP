package api

import (
	"../tcp"
	"fmt"
)

type SendWindow struct {
	AdvertisedWindow int
	SendBuffer       []byte
	LastByteWritten  int
	LastByteAcked    int
	LastByteSent     int
	PkgInFlight      []tcp.TCPPackage
	BytesInFlight    int
	Size             int
}

type RecvWindow struct {
	RecvBuffer       []byte
	LastSeq          int
	LastByteRead     int
	NextByteExpected int
	LastByteRcvd     int
	Size             int
}

func BuildSendWindow() SendWindow {
	Sb := make([]byte, 65535)
	PIF := make([]tcp.TCPPackage, 0)
	return SendWindow{65535, Sb, 0, 0, 0, PIF, 0, 65535}
}

func BuildRecvWindow() RecvWindow {
	Rb := make([]byte, 65535)
	return RecvWindow{Rb, -1, 0, 1, 0, 65535}

}

func (sw *SendWindow) CheckSendAvaliable() bool {
	if sw.LastByteSent-sw.LastByteAcked <= sw.AdvertisedWindow {
		return true
	}
	return false
}

func (sw *SendWindow) CheckWriteAvaliable() bool {
	if sw.LastByteWritten-sw.LastByteAcked <= sw.Size {
		return true
	}
	return false
}

func (sw *SendWindow) EffectiveWindow() int {
	return sw.AdvertisedWindow - sw.BytesInFlight
}

func (sw *SendWindow) Write(buf []byte) int {
	if !sw.CheckWriteAvaliable() || !sw.CheckSendAvaliable() {
		return -1
	}

	count := 0
	i := 0
	for i < len(buf) {
		// if sw.LastByteWritten == sw.AdvertisedWindow {
		// 	break
		// }
		sw.SendBuffer[sw.LastByteWritten] = buf[i]
		sw.LastByteWritten++
		count++
		i++
	}
	fmt.Println("Count after loop ", count)
	fmt.Println("sending buffer ", sw.SendBuffer[sw.LastByteAcked:sw.LastByteSent+20])

	//If cannot write, count would be 0
	return count

}

func (rw *RecvWindow) AdvertisedWindow() int {
	return rw.Size - ((rw.NextByteExpected - 1) - rw.LastByteRead)
}

// Write the received data into the receiving window buffer
func (rw *RecvWindow) Receive(data []byte, se int) int {
	// TODO IMPLEMENTATION
	if len(data) > rw.AdvertisedWindow() {
		return 0
	}
	idx := se - rw.LastSeq
	//fmt.Println("last seq ", rw.LastSeq)
	//fmt.Println(se)
	i := 0
	//fmt.Println("write into receive buffer:", rw.LastByteRead)
	for i < len(data) {
		//fmt.Println("index and data ", rw.LastByteRead+idx+i, data[i])
		rw.RecvBuffer[rw.LastByteRead+idx+i] = data[i]

		rw.LastByteRcvd++
		rw.NextByteExpected++
		i++
		if rw.LastByteRcvd == rw.Size {
			rw.LastByteRcvd = 0
		}
		if rw.NextByteExpected == rw.Size {
			rw.NextByteExpected = 0
		}
	}
	// fmt.Println("recebuff remaining:", rw.RecvBuffer[rw.LastByteRead:rw.LastByteRead+20])
	// fmt.Println("recebuff remaining:", string(rw.RecvBuffer[rw.LastByteRead:rw.LastByteRead+20]))
	return 1

}

func (rw *RecvWindow) Read(nbyte int) ([]byte, int) {
	buf := make([]byte, 0)
	count := 0
	// fmt.Println("lastbyteread before ", rw.LastByteRead)
	// fmt.Println("recebuff remaining before:", rw.RecvBuffer[rw.LastByteRead:rw.LastByteRead+20])
	for i := 0; i < nbyte; i++ {
		if rw.LastByteRead == rw.NextByteExpected-1 {
			break
		}
		fmt.Println(string(rw.RecvBuffer[rw.LastByteRead]))
		buf = append(buf, rw.RecvBuffer[rw.LastByteRead])
		rw.LastByteRead++
		count++
	}
	rw.LastSeq += count
	// fmt.Println("recebuff remaining after:", rw.RecvBuffer[rw.LastByteRead:rw.LastByteRead+20])
	// fmt.Println("count ", count)
	// fmt.Println("......read.......", string(buf))
	// fmt.Println("lastbyteread after ", rw.LastByteRead)

	//If canno read, count would be 0
	return buf, count

}
