package message

import (
	"encoding/binary"
)

type CatMessage struct {
	Decrypt     bool
	Name        string
	StartOffset uint64
}

var _ ProtoMessage = &CatMessage{}

func NewCatMessage(name string, decrypt bool, StartOffset uint64) *CatMessage {
	return &CatMessage{
		Name:        name,
		Decrypt:     decrypt,
		StartOffset: StartOffset,
	}
}

func (c *CatMessage) Encode() []byte {
	bt := []byte{
		byte(MessageTypeCat),
		0,
		0,
		0,
	}

	if c.Decrypt {
		bt[1] = byte(DecryptMessage)
	} else {
		bt[1] = byte(NoDecryptMessage)
	}

	if c.StartOffset != 0 {
		bt[2] = byte(ExtendedMesssage)
	}

	bt = append(bt, []byte(c.Name)...)
	bt = append(bt, 0)
	if c.StartOffset != 0 {
		bt = binary.BigEndian.AppendUint64(bt, c.StartOffset)
	}
	ln := len(bt) + 8

	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, uint64(ln))
	return append(bs, bt...)
}

func (c *CatMessage) Decode(body []byte) {
	c.Name, _ = GetCstring(body[4:])
	if body[1] == byte(DecryptMessage) {
		c.Decrypt = true
	}
	if body[2] == byte(ExtendedMesssage) {
		c.StartOffset = binary.BigEndian.Uint64(body[4+len(c.Name)+1:])
	}
}
