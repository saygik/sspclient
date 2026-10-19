package sspclient

// MessageType — сообщение SSP.
type MessageType struct {
	IeType     uint
	MessAtr    uint
	SendSSPAdr string
	DestSSPAdr string
	Kvit       string
	Body       []rune
}
