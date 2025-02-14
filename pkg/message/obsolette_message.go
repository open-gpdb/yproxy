package message

import (
	"bytes"
	"encoding/binary"
)

type CollectObsoletteMessage struct {
	Segnum  uint64
	Port    uint64
	DBName  string
	Message string
}

type DeleteObsoletteMessage struct {
	Segnum  uint64
	Port    uint64
	DBName  string
	Message string
}

var _ ProtoMessage = &ListMessage{}

// TODO
func NewCollectObsoletteMessage(dbname string, msg string) *CollectObsoletteMessage {
	return &CollectObsoletteMessage{
		DBName:  dbname,
		Message: msg,
	}
}

func (c *CollectObsoletteMessage) Encode() []byte {
	encodedMessage := []byte{
		byte(MessageCollectObsolette),
		0,
		0,
		0,
	}

	byteError := []byte(c.DBName)
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

func (c *CollectObsoletteMessage) Decode(data []byte) {
	c.Segnum = binary.BigEndian.Uint64(data[4:12])
	c.Port = binary.BigEndian.Uint64(data[12:20])
	n := bytes.IndexByte(data[20:], 0)
	c.DBName = string(data[20 : 20+n])
	c.Message = string(data[20+n+1 : len(data)-1])
}

func NewDeleteObsoletteMessage(dbname string, msg string) *DeleteObsoletteMessage {
	return &DeleteObsoletteMessage{
		DBName:  dbname,
		Message: msg,
	}
}

func (c *DeleteObsoletteMessage) Encode() []byte {
	encodedMessage := []byte{
		byte(MessageDeleteObsolette),
		0,
		0,
		0,
	}

	byteError := []byte(c.DBName)
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

func (c *DeleteObsoletteMessage) Decode(data []byte) {
	c.Segnum = binary.BigEndian.Uint64(data[4:12])
	c.Port = binary.BigEndian.Uint64(data[12:20])
	n := bytes.IndexByte(data[20:], 0)

	c.DBName = string(data[20 : 20+n])
	c.Message = string(data[20+n+1 : len(data)-1])
}
