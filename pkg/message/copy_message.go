package message

import (
	"encoding/binary"

	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

type CopyMessage struct {
	Decrypt    bool
	Encrypt    bool
	Name       string
	OldCfgPath string
	Port       uint64
	Confirm    bool
}

var _ ProtoMessage = &CopyMessage{}

func NewCopyMessage(name, oldCfgPath string, encrypt, decrypt, confirm bool, port uint64) *CopyMessage {
	return &CopyMessage{
		Name:       name,
		Encrypt:    encrypt,
		Decrypt:    decrypt,
		OldCfgPath: oldCfgPath,
		Port:       port,
		Confirm:    confirm,
	}
}

func (message *CopyMessage) Encode() []byte {
	encodedMessage := []byte{
		byte(MessageTypeCopy),
		byte(NoDecryptMessage),
		byte(NoEncryptMessage),
		0,
	}

	if message.Decrypt {
		encodedMessage[1] = byte(DecryptMessage)
	}

	if message.Encrypt {
		encodedMessage[2] = byte(EncryptMessage)
	}

	byteName := []byte(message.Name)
	byteLen := make([]byte, 8)
	binary.BigEndian.PutUint64(byteLen, uint64(len(byteName)))
	encodedMessage = append(encodedMessage, byteLen...)
	encodedMessage = append(encodedMessage, byteName...)

	byteOldCfg := []byte(message.OldCfgPath)
	binary.BigEndian.PutUint64(byteLen, uint64(len(byteOldCfg)))
	encodedMessage = append(encodedMessage, byteLen...)
	encodedMessage = append(encodedMessage, byteOldCfg...)

	port := make([]byte, 8)
	binary.BigEndian.PutUint64(port, uint64(message.Port))
	encodedMessage = append(encodedMessage, port...)

	if message.Confirm {
		encodedMessage = append(encodedMessage, 1)
	} else {
		encodedMessage = append(encodedMessage, 0)
	}

	binary.BigEndian.PutUint64(byteLen, uint64(len(encodedMessage)+8))
	ylogger.Zero.Debug().Str("type", MessageType(encodedMessage[0]).String()).Msg("send")
	ylogger.Zero.Debug().Str("object-path", MessageType(encodedMessage[0]).String()).Msg("decrypt object")
	return append(byteLen, encodedMessage...)
}

func (encodedMessage *CopyMessage) Decode(data []byte) {
	if data[1] == byte(DecryptMessage) {
		encodedMessage.Decrypt = true
	}
	if data[2] == byte(EncryptMessage) {
		encodedMessage.Encrypt = true
	}

	nameLen := binary.BigEndian.Uint64(data[4:12])
	encodedMessage.Name = string(data[12 : 12+nameLen])
	oldConfLen := binary.BigEndian.Uint64(data[12+nameLen : 12+nameLen+8])
	encodedMessage.OldCfgPath = string(data[12+nameLen+8 : 12+nameLen+8+oldConfLen])
	encodedMessage.Port = binary.BigEndian.Uint64(data[12+nameLen+8+oldConfLen : 12+nameLen+8+oldConfLen+8])

	ind := 12 + nameLen + 8 + oldConfLen + 8
	if data[ind] == 1 {
		encodedMessage.Confirm = true
	}
}
