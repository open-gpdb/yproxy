package message

type ProtoMessage interface {
	Decode([]byte)
	Encode() []byte
}

type MessageType byte

type RequestEncryption byte

type KEKUsage byte

const (
	MessageTypeCat = MessageType(iota + 42)
	MessageTypePut
	MessageTypeCommandComplete
	MessageTypeReadyForQuery
	MessageTypeCopyData
	MessageTypeDelete
	MessageTypeList
	MessageTypeObjectMeta
	MessageTypePatch
	MessageTypeCopy
	MessageTypeGool
	MessageTypePutV2
	MessageTypeCatV2
	MessageTypeError
	MessageTypePutV3
	MessageTypePutComplete
	MessageTypeUntrashify

	DecryptMessage   = RequestEncryption(1)
	NoDecryptMessage = RequestEncryption(0)

	EncryptMessage   = RequestEncryption(1)
	NoEncryptMessage = RequestEncryption(0)

	NoUseKEK = KEKUsage(0)
	UseKEK   = KEKUsage(1)

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
	case MessageTypePutV3:
		return "PUTV3"
	case MessageTypePutComplete:
		return "PUTCOMPLETE"
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
	case MessageTypeUntrashify:
		return "UNTRASHIFY"
	}
	return "UNKNOWN"
}
