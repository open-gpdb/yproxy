package message

import (
	"encoding/binary"
)

type DeleteMessage struct { //seg port
	Name      string
	Port      uint64
	Segnum    uint64
	Confirm   bool
	Garbage   bool
	CrazyDrop bool
}

var _ ProtoMessage = &DeleteMessage{}

func NewDeleteMessage(name string, port uint64, seg uint64, confirm bool, garbage bool) *DeleteMessage {
	return &DeleteMessage{
		Name:    name,
		Port:    port,
		Segnum:  seg,
		Confirm: confirm,
		Garbage: garbage,
	}
}

func (c *DeleteMessage) Encode() []byte {
	bt := []byte{
		byte(MessageTypeDelete),
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
	if c.CrazyDrop {
		bt[3] = 1
	}

	bt = append(bt, []byte(c.Name)...)
	bt = append(bt, 0)

	p := make([]byte, 8)
	binary.BigEndian.PutUint64(p, uint64(c.Port))
	bt = append(bt, p...)

	p = make([]byte, 8)
	binary.BigEndian.PutUint64(p, uint64(c.Segnum))
	bt = append(bt, p...)

	ln := len(bt) + 8
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, uint64(ln))
	return append(bs, bt...)
}

func (c *DeleteMessage) Decode(body []byte) {
	if body[1] == 1 {
		c.Confirm = true
	}
	if body[2] == 1 {
		c.Garbage = true
	}
	if body[3] == 1 {
		c.CrazyDrop = true
	}
	c.Name, _ = GetCstring(body[4:])
	c.Port = binary.BigEndian.Uint64(body[len(body)-16 : len(body)-8])
	c.Segnum = binary.BigEndian.Uint64(body[len(body)-8:])
}
