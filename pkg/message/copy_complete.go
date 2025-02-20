package message

import (
	"encoding/binary"
)

type CopyCompleteMessage struct {
	KeyVersion byte
}

var _ ProtoMessage = &CopyCompleteMessage{}

func NewCopyCompleteMessage(keyVersion byte) *CopyCompleteMessage {
	return &CopyCompleteMessage{
		KeyVersion: keyVersion,
	}
}

func (c *CopyCompleteMessage) Encode() []byte {
	bt := []byte{
		byte(MessageTypeCopyComplete),
		c.KeyVersion,
		0,
		0,
	}
	ln := len(bt) + 8

	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, uint64(ln))
	return append(bs, bt...)
}

func (c *CopyCompleteMessage) Decode(body []byte) {
	c.KeyVersion = body[1]
}
