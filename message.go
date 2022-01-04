package gowebsocket

const (
	TextMessage   = 1
	BinaryMessage = 2
	CloseMessage  = 8
	PingMessage   = 9
	PongMessage   = 10
)

type MessageBody struct {
	Type int
	Data []byte
}
