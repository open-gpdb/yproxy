package message

import (
	"encoding/binary"
)

type ListMessage struct {
	Prefix string
}

var _ ProtoMessage = &ListMessage{}

func NewListMessage(name string) *ListMessage {
	return &ListMessage{
		Prefix: name,
	}
}

func (c *ListMessage) Encode() []byte {
	bt := []byte{
		byte(MessageTypeList),
		0,
		0,
		0,
	}

	bt = append(bt, []byte(c.Prefix)...)
	bt = append(bt, 0)
	ln := len(bt) + 8

	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, uint64(ln))
	return append(bs, bt...)
}

func (c *ListMessage) Decode(body []byte) {
	c.Prefix, _ = GetCstring(body[4:])
}
