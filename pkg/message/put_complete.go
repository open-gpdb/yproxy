package message

import (
	"encoding/binary"
)

type PutCompleteMessage struct {
	KeyVersion uint16
}

var _ ProtoMessage = &PutCompleteMessage{}

func NewPutCompleteMessage(keyVersion uint16) *PutCompleteMessage {
	return &PutCompleteMessage{
		KeyVersion: keyVersion,
	}
}

func (c *PutCompleteMessage) Encode() []byte {
	bt := []byte{
		byte(MessageTypePutComplete),
		0,
		0,
		0,
	}

	kv := make([]byte, 2)
	binary.BigEndian.PutUint16(kv, uint16(c.KeyVersion))
	bt = append(bt, kv...)

	ln := len(bt) + 8

	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, uint64(ln))
	return append(bs, bt...)
}

func (c *PutCompleteMessage) Decode(body []byte) {
	c.KeyVersion = binary.BigEndian.Uint16(body[4 : 4+2])
}
