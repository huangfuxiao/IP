package api

import (
	//"../tcp"
	"fmt"
	"sync"
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
	Mutex         *sync.RWMutex
}

type RecvWindow struct {
	back             bool
	Fill             []bool
	RecvBuffer       []byte
	LastSeq          int
	LastByteRead     int
	NextByteExpected int
	Size             int
	mutex            *sync.RWMutex
}

func BuildSendWindow() SendWindow {
	Sb := make([]byte, 65535)
	// PIF := make([]tcp.TCPPackage, 0)
	return SendWindow{false, false, false, 65535, Sb, 0, 0, 0, 0, 65535, &sync.RWMutex{}}
}

func BuildRecvWindow() RecvWindow {
	Rb := make([]byte, 65535)
	Fil := make([]bool, 65535)
	return RecvWindow{false, Fil, Rb, -1, 0, 1, 65535, &sync.RWMutex{}}
}

// func (sw *SendWindow) CheckSendAvaliable() bool {
// 	if sw.LastByteSent-sw.LastByteAcked <= sw.AdvertisedWindow {
// 		return true
// 	}
// 	return false
// }

func (sw *SendWindow) BytesCanBeWritten() int {
	sw.Mutex.RLock()
	defer sw.Mutex.RUnlock()
	if sw.WAback {
		return sw.Size - (sw.LastByteWritten + sw.Size - sw.LastByteAcked)
	}
	return sw.Size - (sw.LastByteWritten - sw.LastByteAcked)
}

func (sw *SendWindow) EffectiveWindow() int {
	sw.Mutex.RLock()
	defer sw.Mutex.RUnlock()
	return sw.AdvertisedWindow - sw.BytesInFlight
}

func (sw *SendWindow) Write(buf []byte) int {

	num := len(buf)
	//fmt.Println("length of buf in Write ", num)
	//fmt.Println("length that can be written ", sw.BytesCanBeWritten())
	writebytes := sw.BytesCanBeWritten()
	if writebytes < num {
		//fmt.Println("less bytes can be written ", sw.BytesCanBeWritten())
		//fmt.Println("num ", num)
		//fmt.Println("buf to write ", buf)
		num = writebytes
		//fmt.Println("bcbw ", writebytes)

		// if writebytes < 0 {
		// 	fmt.Println("bcbw ", writebytes)
		// 	fmt.Println("lbw ", sw.LastByteWritten)
		// 	fmt.Println("lba ", sw.LastByteAcked)
		// }
	}
	// fmt.Println("bytescanbewritten ", sw.BytesCanBeWritten())
	// fmt.Println("num ", num)
	sw.Mutex.Lock()
	defer sw.Mutex.Unlock()
	i := 0
	for i < num {
		if sw.LastByteWritten < 0 || i < 0 || sw.LastByteWritten >= 65535 || i > len(buf) {

			fmt.Println("sw write lbw ", sw.LastByteWritten)
			fmt.Println("sw write i ", i)
		}
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
	//fmt.Println("recebuff remaining:", rw.RecvBuffer[rw.LastByteRead:rw.LastByteRead+20])
	pad := 0
	if se >= rw.LastSeq {
		rw.mutex.Lock()
		defer rw.mutex.Unlock()

		if len(data) > rw.AdvertisedWindow() {
			fmt.Println("data length larger than advertise window")
			return 0, 0
		}
		idx := se - rw.LastSeq
		//fmt.Println("idx before ", idx)
		idx %= rw.Size
		i := 0
		//fmt.Println("write into receive buffer:", rw.LastByteRead)
		for i < len(data) {
			//fmt.Println("index and data ", rw.LastByteRead+idx+i, data[i])
			n := rw.NextByteExpected - 1
			if rw.NextByteExpected == 0 {
				n = rw.Size - 1
			}

			index := n + idx + i
			if index >= rw.Size {
				index -= rw.Size
			}
			//fmt.Println("write rb index  and i ", index, i)
			if index < 0 {
				fmt.Println("error index ", index)
			}
			rw.RecvBuffer[index] = data[i]
			rw.Fill[index] = true
			i++
		}
		//fmt.Println(rw.RecvBuffer[rw.LastByteRead : rw.LastByteRead+20])
		i = 0

		if order {
			for i < len(data) {
				rw.NextByteExpected++
				rw.LastSeq++
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
					if rw.Fill[rw.Size-1] == false {
						break
					}
				} else {
					if rw.Fill[rw.NextByteExpected-1] == false {
						break
					}
				}
				if rw.NextByteExpected == rw.LastByteRead+1 {
					break
				}
				rw.NextByteExpected++
				rw.LastSeq++
				if rw.NextByteExpected == rw.Size {
					rw.back = true
					rw.NextByteExpected = 0
				}
				pad++

			}
		}

	}
	//fmt.Println("recebuff remaining:", rw.RecvBuffer[rw.LastByteRead:rw.LastByteRead+20])
	// fmt.Println("recebuff remaining:", string(rw.RecvBuffer[rw.LastByteRead:rw.LastByteRead+20]))
	return 1, pad

}

func (rw *RecvWindow) ReadableSize() int {
	return rw.Size - rw.AdvertisedWindow()
}

func (rw *RecvWindow) Read(nbyte int) ([]byte, int) {
	rw.mutex.Lock()
	defer rw.mutex.Unlock()
	buf := make([]byte, 0)
	count := 0

	tempsize := rw.ReadableSize()
	if nbyte > tempsize {
		nbyte = tempsize
	}

	// fmt.Println("lastbyteread before ", rw.LastByteRead)
	// fmt.Println("recebuff remaining before:", rw.RecvBuffer[rw.LastByteRead:rw.LastByteRead+20])
	for i := 0; i < nbyte; i++ {
		// if rw.back == false {
		// 	if rw.LastByteRead == rw.NextByteExpected-1 {
		// 		break
		// 	}
		// }

		buf = append(buf, rw.RecvBuffer[rw.LastByteRead])
		rw.Fill[rw.LastByteRead] = false
		rw.LastByteRead++
		if rw.LastByteRead == rw.Size {
			rw.back = false
			rw.LastByteRead = 0
		}
		count++
	}

	// fmt.Println("recebuff remaining after:", rw.RecvBuffer[rw.LastByteRead:rw.LastByteRead+20])
	// fmt.Println("......read.......", string(buf))
	// fmt.Println("lastbyteread after ", rw.LastByteRead)

	//If canno read, count would be 0
	return buf, count

}
