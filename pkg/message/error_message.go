package message

import (
	"encoding/binary"
)

type ErrorMessage struct {
	Error   string
	Message string
}

var _ ProtoMessage = &ListMessage{}

func NewErrorMessage(err error, msg string) *ErrorMessage {
	return &ErrorMessage{
		Error:   err.Error(),
		Message: msg,
	}
}

func (c *ErrorMessage) Encode() []byte {
	encodedMessage := []byte{
		byte(MessageTypeError),
		0,
		0,
		0,
	}

	byteError := []byte(c.Error)
	byteLen := make([]byte, 8)
	binary.BigEndian.PutUint64(byteLen, uint64(len(byteError)))
	encodedMessage = append(encodedMessage, byteLen...)
	encodedMessage = append(encodedMessage, byteError...)

	byteMessage := []byte(c.Message)
	binary.BigEndian.PutUint64(byteLen, uint64(len(byteMessage)))
	encodedMessage = append(encodedMessage, byteLen...)
	encodedMessage = append(encodedMessage, byteMessage...)

	binary.BigEndian.PutUint64(byteLen, uint64(len(encodedMessage)+8))
	return append(byteLen, encodedMessage...)
}

func (c *ErrorMessage) Decode(data []byte) {
	errorLen := binary.BigEndian.Uint64(data[4:12])
	c.Error = string(data[12 : 12+errorLen])
	messageLen := binary.BigEndian.Uint64(data[12+errorLen : 12+errorLen+8])
	c.Message = string(data[12+errorLen+8 : 12+errorLen+8+messageLen])
}
