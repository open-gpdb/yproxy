package message

import (
	"bytes"
	"encoding/binary"
)

type GoolMessage struct {
	Name string
}

func (c *GoolMessage) Encode() []byte {
	bt := []byte{
		byte(MessageTypeGool),
		0,
		0,
		0,
	}

	bt = append(bt, []byte("GOOL "+c.Name)...)
	bt = append(bt, 0)
	ln := len(bt) + 8

	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, uint64(ln))
	return append(bs, bt...)
}

func (c *GoolMessage) Decode(body []byte) {
	c.Name = c.GetGoolName(body[4:])
}

func (c *GoolMessage) GetGoolName(b []byte) string {
	buff := bytes.NewBufferString("")

	for i := range b {
		if b[i] == 0 {
			break
		}
		buff.WriteByte(b[i])
	}

	return buff.String()
}

var _ ProtoMessage = &GoolMessage{}

func NewGoolMessage(name string) *GoolMessage {
	return &GoolMessage{
		Name: name,
	}
}
