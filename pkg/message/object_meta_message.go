package message

import (
	"bytes"
	"encoding/binary"

	"github.com/open-gpdb/yproxy/pkg/object"
)

type ObjectInfoMessage struct {
	Content []*object.ObjectInfo
}

var _ ProtoMessage = &ObjectInfoMessage{}

func NewObjectMetaMessage(content []*object.ObjectInfo) *ObjectInfoMessage {
	return &ObjectInfoMessage{
		Content: content,
	}
}

func (c *ObjectInfoMessage) Encode() []byte {
	bt := []byte{
		byte(MessageTypeObjectMeta),
		0,
		0,
		0,
	}

	for _, objMeta := range c.Content {
		bt = append(bt, []byte(objMeta.Path)...)
		bt = append(bt, 0)

		bn := make([]byte, 8)
		binary.BigEndian.PutUint64(bn, uint64(objMeta.Size))
		bt = append(bt, bn...)
	}

	ln := len(bt) + 8
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, uint64(ln))
	return append(bs, bt...)
}

func (c *ObjectInfoMessage) Decode(body []byte) {
	body = body[4:]
	c.Content = make([]*object.ObjectInfo, 0)
	for len(body) > 0 {
		name, index := c.GetString(body)
		size := int64(binary.BigEndian.Uint64(body[index : index+8]))

		c.Content = append(c.Content, &object.ObjectInfo{
			Path: name,
			Size: size,
		})
		body = body[index+8:]
	}
}

func (c *ObjectInfoMessage) GetString(b []byte) (string, int) {
	buff := bytes.NewBufferString("")

	i := 0
	for ; i < len(b); i++ {
		if b[i] == 0 {
			break
		}
		buff.WriteByte(b[i])
	}

	return buff.String(), i + 1
}
