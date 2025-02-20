package message

import (
	"encoding/binary"
	"fmt"

	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

const (
	CopyEncrypt    = 0b1
	CopyDecrypt    = 0b10
	CopyUseKEK     = 0b100
	CopyServerSide = 0b1000
)

type CopyMessageV2 struct {
	Decrypt        bool
	Encrypt        bool
	Name           string
	OldCfgPath     string
	Port           uint64
	Confirm        bool
	KEKDecrypt     bool
	ServerSideCopy bool
}

var _ ProtoMessage = &CopyMessageV2{}

func NewCopyMessageV2(name, oldCfgPath string, encrypt, decrypt, confirm, kEKDecrypt, ssCopy bool, port uint64) *CopyMessageV2 {
	return &CopyMessageV2{
		Name:           name,
		Encrypt:        encrypt,
		Decrypt:        decrypt,
		OldCfgPath:     oldCfgPath,
		Port:           port,
		Confirm:        confirm,
		KEKDecrypt:     kEKDecrypt,
		ServerSideCopy: ssCopy,
	}
}

func (message *CopyMessageV2) Encode() []byte {
	flags := byte(0)
	if message.Decrypt {
		flags |= CopyDecrypt
	}
	if message.Encrypt {
		flags |= CopyEncrypt
	}
	if message.KEKDecrypt {
		flags |= CopyUseKEK
	}
	if message.ServerSideCopy {
		flags |= CopyServerSide
	}

	encodedMessage := []byte{
		byte(MessageTypeCopyV2),
		flags,
		0,
		0,
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
	fmt.Printf("send: %v\n", MessageType(encodedMessage[0]))
	ylogger.Zero.Debug().Str("object-path", MessageType(encodedMessage[0]).String()).Msg("decrypt object")
	return append(byteLen, encodedMessage...)
}

func (encodedMessage *CopyMessageV2) Decode(data []byte) {
	if data[1]&CopyDecrypt != 0 {
		encodedMessage.Decrypt = true
	}
	if data[1]&CopyEncrypt != 0 {
		encodedMessage.Encrypt = true
	}
	if data[1]&CopyUseKEK != 0 {
		encodedMessage.KEKDecrypt = true
	}
	if data[1]&CopyServerSide != 0 {
		encodedMessage.ServerSideCopy = true
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
