package message

import (
	"encoding/binary"
)

type UntrashifyMessage struct { //seg port
	Name    string
	Segnum  uint64
	Confirm bool
}

var _ ProtoMessage = &UntrashifyMessage{}

func NewUntrashifyMessage(name string, seg uint64, confirm bool) *UntrashifyMessage {
	return &UntrashifyMessage{
		Name:    name,
		Segnum:  seg,
		Confirm: confirm,
	}
}

func (c *UntrashifyMessage) Encode() []byte {
	bt := []byte{
		byte(MessageTypeUntrashify),
		0,
		0,
		0,
	}

	if c.Confirm {
		bt[1] = 1
	}

	bt = append(bt, []byte(c.Name)...)
	bt = append(bt, 0)

	p := make([]byte, 8)
	binary.BigEndian.PutUint64(p, uint64(c.Segnum))
	bt = append(bt, p...)

	ln := len(bt) + 8
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, uint64(ln))
	return append(bs, bt...)
}

func (c *UntrashifyMessage) Decode(body []byte) {
	if body[1] == 1 {
		c.Confirm = true
	}
	c.Name, _ = GetCstring(body[4:])
	c.Segnum = binary.BigEndian.Uint64(body[len(body)-8:])
}
