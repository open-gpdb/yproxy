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
	MessageTypeCopyDone
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
	MessageTypeCopyV2
	MessageTypeCopyComplete
	MessageTypeListV2

	MessageCollectObsolete = MessageType(64)
	MessageDeleteObsolete  = MessageType(65)

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
	case MessageTypeCopyDone:
		return "COMMAND COMPLETE"
	case MessageTypeReadyForQuery:
		return "READY FOR QUERY"
	case MessageTypeCopyData:
		return "COPY DATA"
	case MessageTypeDelete:
		return "DELETE"
	case MessageTypeList:
		return "LIST"
	case MessageTypeListV2:
		return "LISTV2"
	case MessageTypeObjectMeta:
		return "OBJECT META"
	case MessageTypeCopy:
		return "COPY"
	case MessageTypeCopyV2:
		return "COPYV2"
	case MessageTypeGool:
		return "GOOL"
	case MessageTypeError:
		return "ERROR"
	case MessageTypeUntrashify:
		return "UNTRASHIFY"
	case MessageCollectObsolete:
		return "COLLECT OBSOLETE"
	case MessageDeleteObsolete:
		return "DELETE OBSOLETE"
	case MessageTypeCopyComplete:
		return "COPY COMPLETE"
	}
	return "UNKNOWN"
}
