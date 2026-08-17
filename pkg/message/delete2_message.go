package message

import (
	"encoding/binary"
)

// Requests deletion of all objects under a given prefix.
type DisposalPrefixMessage struct { // Seg port
	Prefix  string // Object key prefix to delete under
	Confirm bool   // Execute deletion; false means dry-run
	Garbage bool   // Restrict deletion to garbage (trash) objects past retention
}

var _ ProtoMessage = &DisposalPrefixMessage{}

func BuildDisposalPrefixMessage(prefix string, confirm bool, garbage bool) *DisposalPrefixMessage {
	return &DisposalPrefixMessage{
		Prefix:  prefix,
		Confirm: confirm,
		Garbage: garbage,
	}
}

func (c *DisposalPrefixMessage) Encode() []byte {
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

func (c *DisposalPrefixMessage) Decode(body []byte) {
	if body[1] == 1 {
		c.Confirm = true
	}
	if body[2] == 1 {
		c.Garbage = true
	}
	c.Prefix, _ = GetCstring(body[4:])
}
