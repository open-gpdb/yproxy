package message

import (
	"encoding/binary"

	"github.com/yezzey-gp/yproxy/pkg/settings"
)

type ListMessageV2 struct {
	Prefix string

	Settings []settings.StorageSettings
}

var _ ProtoMessage = &ListMessage{}

func NewListMessageV2(name string, Settings []settings.StorageSettings) *ListMessageV2 {
	return &ListMessageV2{
		Prefix:   name,
		Settings: Settings,
	}
}

func (c *ListMessageV2) Encode() []byte {
	bt := []byte{
		byte(MessageTypeList),
		0,
		0,
		0,
	}

	bt = append(bt, []byte(c.Prefix)...)
	bt = append(bt, 0)

	bt = binary.BigEndian.AppendUint64(bt, uint64(len(c.Settings)))

	for _, s := range c.Settings {

		bt = append(bt, []byte(s.Name)...)
		bt = append(bt, 0)

		bt = append(bt, []byte(s.Value)...)
		bt = append(bt, 0)
	}

	ln := len(bt) + 8

	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, uint64(ln))
	return append(bs, bt...)
}

func (c *ListMessageV2) Decode(body []byte) {
	var off uint64
	c.Prefix, off = GetCstring(body[4:])

	settLen := binary.BigEndian.Uint64(body[4+off : 4+off+8])

	totalOff := 4 + off + 8

	c.Settings = make([]settings.StorageSettings, settLen)

	for i := range int(settLen) {

		var currOff uint64

		c.Settings[i].Name, currOff = GetCstring(body[totalOff:])
		totalOff += currOff

		c.Settings[i].Value, currOff = GetCstring(body[totalOff:])
		totalOff += currOff
	}
}
