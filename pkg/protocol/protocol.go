package protocol

// MessageType represents the type of protocol message exchanged over the tunnel.
type MessageType byte

const (
	MsgAuth MessageType = iota
	MsgAuthResponse
	MsgData
	MsgClose
	MsgHeartbeat
)

// Message is the base frame type for tunnel communication.
type Message struct {
	Type    MessageType
	Payload []byte
}
