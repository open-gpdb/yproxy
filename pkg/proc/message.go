package proc

type ProtoMessage interface {
	Decode([]byte)
	Encode() []byte
}

type MessageType byte

type RequestEncryption byte

const (
	MessageTypeCat             = MessageType(42)
	MessageTypePut             = MessageType(43)
	MessageTypeCommandComplete = MessageType(44)
	MessageTypeReadyForQuery   = MessageType(45)
	MessageTypeCopyData        = MessageType(46)

	DecryptMessage   = RequestEncryption(1)
	NoDecryptMessage = RequestEncryption(0)

	EncryptMessage   = RequestEncryption(1)
	NoEncryptMessage = RequestEncryption(0)
)

func (m MessageType) String() string {
	switch m {
	case MessageTypeCat:
		return "CAT"
	case MessageTypePut:
		return "PUT"
	case MessageTypeCommandComplete:
		return "COMMAND COMPLETE"
	case MessageTypeReadyForQuery:
		return "READY FOR QUERY"
	case MessageTypeCopyData:
		return "COPY DATA"
	}
	return "UNKNOWN"
}
