package api

import (
	"../tcp"
)

type SendWindow struct {
	AdvertisedWindow int

	SendBuffer      []byte
	LastByteWritten int
	LastByteAcked   int
	LastByteSent    int
	PkgInFlight     []tcp.TCPPackage
	BytesInFlight   int
	Size            int
}

type RecvWindow struct {
	RecvBuffer       []byte
	LastByteRead     int
	NextByteExpected int
	LastByteRcvd     int
	Size             int
}

func (rw *RecvWindow) AdvertisedWindow() int {
	return rw.Size - ((rw.NextByteExpected - 1) - rw.LastByteRead)
}

func (sw *SendWindow) CheckSendAvaliable(adw int) bool {
	if sw.LastByteSent-sw.LastByteAcked <= adw {
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

func (sw *SendWindow) EffectiveWindow(adw int) int {
	return adw - sw.BytesInFlight
}

func (rw *RecvWindow) Read(nbyte int) ([]byte, int) {
	buf := make([]byte, 0, nbyte)
	count := 0
	for i := 0; i < nbyte; i++ {
		if rw.LastByteRead == rw.NextByteExpected-1 {
			break
		}
		buf = append(buf, []byte{rw.RecvBuffer[rw.LastByteRead]}...)
		rw.LastByteRead++
		count++
	}

	//If canno read, count would be 0
	return buf, count

}

func (rw *SendWindow) Write(buf []byte) int {
	count := 0
	for i := 0; i < len(buf); i++ {
		if !rw.CheckWriteAvaliable() {
			break
		}
		if rw.LastByteWritten == rw.AdvertisedWindow {
			break
		}
		rw.SendBuffer = append(rw.SendBuffer, []byte{buf[i]}...)
		rw.LastByteWritten++
		count++
	}

	//If cannot write, count would be 0
	return count

}
