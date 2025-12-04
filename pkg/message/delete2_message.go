package message

import (
	"encoding/binary"
)

type Delete2Message struct { //seg port
	Prefix  string
	Confirm bool
	Garbage bool
}

var _ ProtoMessage = &Delete2Message{}

func NewDelete2Message(prefix string, confirm bool, garbage bool) *Delete2Message {
	return &Delete2Message{
		Prefix:  prefix,
		Confirm: confirm,
		Garbage: garbage,
	}
}

func (c *Delete2Message) Encode() []byte {
	bt := []byte{
		byte(MessageTypeDelete2),
		0,
		0,
		0,
	}

	if c.Confirm {
		bt[1] = 1
	}
	if c.Garbage {
		bt[2] = 1
	}

	bt = append(bt, []byte(c.Prefix)...)
	bt = append(bt, 0)

	ln := len(bt) + 8
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, uint64(ln))
	return append(bs, bt...)
}

func (c *Delete2Message) Decode(body []byte) {
	if body[1] == 1 {
		c.Confirm = true
	}
	if body[2] == 1 {
		c.Garbage = true
	}
	c.Prefix, _ = GetCstring(body[4:])
}
