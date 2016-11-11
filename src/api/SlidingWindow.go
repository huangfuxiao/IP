package api

import (
	//"../tcp"
	"fmt"
)

type SendWindow struct {
	Back             bool
	WSback           bool
	WAback           bool
	AdvertisedWindow int
	SendBuffer       []byte
	LastByteWritten  int
	LastByteAcked    int
	LastByteSent     int
	// PkgInFlight      []tcp.TCPPackage
	BytesInFlight int
	Size          int
}

type RecvWindow struct {
	back             bool
	RecvBuffer       []byte
	LastSeq          int
	LastByteRead     int
	NextByteExpected int
	Size             int
}

func BuildSendWindow() SendWindow {
	Sb := make([]byte, 65535)
	// PIF := make([]tcp.TCPPackage, 0)
	return SendWindow{false, false, false, 65535, Sb, 0, 0, 0, 0, 65535}
}

func BuildRecvWindow() RecvWindow {
	Rb := make([]byte, 65535)
	return RecvWindow{false, Rb, -1, 0, 1, 65535}

}

// func (sw *SendWindow) CheckSendAvaliable() bool {
// 	if sw.LastByteSent-sw.LastByteAcked <= sw.AdvertisedWindow {
// 		return true
// 	}
// 	return false
// }

func (sw *SendWindow) BytesCanBeWritten() int {
	if sw.WAback {
		return sw.Size - (sw.LastByteWritten + sw.Size - sw.LastByteAcked)
	}
	return sw.Size - (sw.LastByteWritten - sw.LastByteAcked)
}

func (sw *SendWindow) EffectiveWindow() int {
	return sw.AdvertisedWindow - sw.BytesInFlight
}

func (sw *SendWindow) Write(buf []byte) int {
	num := len(buf)
	//fmt.Println("length of buf in Write ", num)
	//fmt.Println("length that can be written ", sw.BytesCanBeWritten())
	if sw.BytesCanBeWritten() < num {
		//fmt.Println("less bytes can be written ", sw.BytesCanBeWritten())
		num = sw.BytesCanBeWritten()
	}
	// fmt.Println("bytescanbewritten ", sw.BytesCanBeWritten())
	// fmt.Println("num ", num)

	i := 0
	for i < num {

		sw.SendBuffer[sw.LastByteWritten] = buf[i]
		sw.LastByteWritten++
		if sw.LastByteWritten >= sw.Size {
			sw.WAback = true
			sw.WSback = true
			sw.LastByteWritten = 0
		}
		i++
	}
	//fmt.Println("Count after loop ", count)
	//fmt.Println("sending buffer ", sw.SendBuffer[sw.LastByteAcked:sw.LastByteAcked+20])
	//If cannot write, count would be 0
	return num

}

func (rw *RecvWindow) AdvertisedWindow() int {
	if rw.back {
		return rw.Size - ((rw.NextByteExpected - 1 + rw.Size) - rw.LastByteRead)
	}
	return rw.Size - ((rw.NextByteExpected - 1) - rw.LastByteRead)
}

// Write the received data into the receiving window buffer
func (rw *RecvWindow) Receive(data []byte, se int, order bool) (int, int) {
	// TODO IMPLEMENTATION
	pad := 0
	if len(data) > rw.AdvertisedWindow() {
		fmt.Println("data length larger than advertise window")
		return 0, 0
	}
	idx := se - rw.LastSeq
	i := 0
	//fmt.Println("write into receive buffer:", rw.LastByteRead)
	for i < len(data) {
		//fmt.Println("index and data ", rw.LastByteRead+idx+i, data[i])
		index := rw.LastByteRead + idx + i
		if index >= rw.Size {
			index -= rw.Size
		}
		rw.RecvBuffer[index] = data[i]
		i++
	}
	//fmt.Println(rw.RecvBuffer[rw.LastByteRead : rw.LastByteRead+20])
	i = 0

	if order {
		for i < len(data) {
			rw.NextByteExpected++
			i++
			if rw.NextByteExpected == rw.Size {
				rw.back = true
				rw.NextByteExpected = 0
			}
		}
	}

	if order {
		for {
			if rw.NextByteExpected == 0 {
				if rw.RecvBuffer[rw.Size-1] == 0 {
					break
				}
			} else {
				if rw.RecvBuffer[rw.NextByteExpected-1] == 0 {
					break
				}
			}
			if rw.NextByteExpected == rw.LastByteRead+1 {
				break
			}
			rw.NextByteExpected++
			if rw.NextByteExpected == rw.Size {
				rw.back = true
				rw.NextByteExpected = 0
			}
			pad++

		}
	}

	// fmt.Println("recebuff remaining:", rw.RecvBuffer[rw.LastByteRead:rw.LastByteRead+20])
	// fmt.Println("recebuff remaining:", string(rw.RecvBuffer[rw.LastByteRead:rw.LastByteRead+20]))
	return 1, pad

}

func (rw *RecvWindow) Read(nbyte int) ([]byte, int) {
	buf := make([]byte, 0)
	count := 0
	// fmt.Println("lastbyteread before ", rw.LastByteRead)
	// fmt.Println("recebuff remaining before:", rw.RecvBuffer[rw.LastByteRead:rw.LastByteRead+20])
	for i := 0; i < nbyte; i++ {
		if rw.back == false {
			if rw.LastByteRead == rw.NextByteExpected-1 {
				break
			}
		}

		buf = append(buf, rw.RecvBuffer[rw.LastByteRead])
		rw.LastByteRead++
		rw.RecvBuffer[rw.LastByteRead-1] = 0
		if rw.LastByteRead == rw.Size {
			rw.back = false
			rw.LastByteRead = 0
		}
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
