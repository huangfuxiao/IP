package api

type SendWindow struct {
	SendBuffer       []byte
	AdvertisedWindow int
	LastByteWritten  int
	LastByteAcked    int
	LastByteSent     int
}

type RecvWindow struct {
	RecvBuffer       []byte
	LastByteRead     int
	NextByteExpected int
	LastByteRcvd     int
}

func (rw *RecvWindow) Read(nbyte int) ([]byte, int) {
	buf := make([]byte, 0, nbyte)
	count := 0
	for i := 0; i < nbyte; i++ {
		if rw.LastByteRead == rw.NextByteExpected-1 {
			break
		}
		buf = append(buf, rw.RecvBuffer[rw.LastByteRead]...)
		rw.LastByteRead++
		count++
	}

	//If canno read, count would be 0
	return buf, count

}

func (rw *SendWindow) Write(buf []byte) int {
	count := 0
	for i := 0; i < len(buf); i++ {
		if rw.LastByteWritten == rw.AdvertisedWindow {
			break
		}
		rw.SendBuffer = append(rw.SendBuffer, buf[i]...)
		rw.LastByteWritten++
		count++
	}

	//If cannot write, count would be 0
	return count

}
