package message

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
	MessageTypeDelete          = MessageType(47)
	MessageTypeList            = MessageType(48)
	MessageTypeObjectMeta      = MessageType(49)
	MessageTypePatch           = MessageType(50)
	MessageTypeCopy            = MessageType(51)
	MessageTypeGool            = MessageType(52)
	MessageTypePutV2           = MessageType(53)
	MessageTypeCatV2           = MessageType(54)
	MessageTypeError           = MessageType(55)

	DecryptMessage   = RequestEncryption(1)
	NoDecryptMessage = RequestEncryption(0)

	EncryptMessage   = RequestEncryption(1)
	NoEncryptMessage = RequestEncryption(0)

	ExtendedMesssage = byte(1)
)

func (m MessageType) String() string {
	switch m {
	case MessageTypeCat:
		return "CAT"
	case MessageTypeCatV2:
		return "CATV2"
	case MessageTypePut:
		return "PUT"
	case MessageTypePutV2:
		return "PUTV2"
	case MessageTypeCommandComplete:
		return "COMMAND COMPLETE"
	case MessageTypeReadyForQuery:
		return "READY FOR QUERY"
	case MessageTypeCopyData:
		return "COPY DATA"
	case MessageTypeDelete:
		return "DELETE"
	case MessageTypeList:
		return "LIST"
	case MessageTypeObjectMeta:
		return "OBJECT META"
	case MessageTypeCopy:
		return "COPY"
	case MessageTypeGool:
		return "GOOL"
	case MessageTypeError:
		return "ERROR"
	}
	return "UNKNOWN"
}
