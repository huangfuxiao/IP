package api

import (
	"../tcp"
)

type SendWindow struct {
	SendBuffer      []byte
	LastByteWritten int
	LastByteAcked   int
	LastByteSent    int
	PkgInFlight     []tcp.TCPPackage
	BytesInFlight   int
	Size            int
}

type RecvWindow struct {
	SendBuffer       []byte
	LastByteRead     int
	NextByteExpected int
	LastByteRcvd     int
	Size             int
}

func (rw *RecvWindow) AdvertisedWindow() int {
	return size - ((rw.NextByteExpected - 1) - rw.LastByteRead)
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
