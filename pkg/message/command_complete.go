package message

import "encoding/binary"

type CopyDoneMessage struct {
}

var _ ProtoMessage = &CopyDoneMessage{}

func NewCopyDoneMessage() *CopyDoneMessage {
	return &CopyDoneMessage{}
}

func (cc *CopyDoneMessage) Encode() []byte {
	bt := []byte{
		byte(MessageTypeCopyDone),
		0,
		0,
		0,
	}

	ln := len(bt) + 8

	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, uint64(ln))
	return append(bs, bt...)
}

func (c *CopyDoneMessage) Decode(body []byte) {
}
