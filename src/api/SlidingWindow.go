package api

type SendWindow struct {
	SendBuffer      []byte
	LastByteWritten int
	LastByteAcked   int
	LastByteSent    int
}

type RecvWindow struct {
	SendBuffer       []byte
	LastByteRead     int
	NextByteExpected int
	LastByteRcvd     int
}
