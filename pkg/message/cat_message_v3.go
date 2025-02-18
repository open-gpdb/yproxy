package message

import (
	"encoding/binary"

	"github.com/yezzey-gp/yproxy/pkg/settings"
)

type CatMessageV3 struct {
	Decrypt     bool
	KEK         bool
	Name        string
	StartOffset uint64

	Settings []settings.StorageSettings
}

var _ ProtoMessage = &CatMessage{}

func NewCatMessageV3(name string, decrypt bool, kek bool, StartOffset uint64, Settings []settings.StorageSettings) *CatMessageV3 {
	return &CatMessageV3{
		Name:        name,
		Decrypt:     decrypt,
		KEK:         kek,
		StartOffset: StartOffset,
		Settings:    Settings,
	}
}

func (c *CatMessageV3) Encode() []byte {
	bt := []byte{
		byte(MessageTypeCatV3),
		0,
		0,
		0,
	}

	if c.Decrypt {
		bt[1] = byte(DecryptMessage)
	} else {
		bt[1] = byte(NoDecryptMessage)
	}

	if c.KEK {
		bt[2] = byte(UseKEK)
	} else {
		bt[2] = byte(NoUseKEK)
	}

	bt = append(bt, []byte(c.Name)...)
	bt = append(bt, 0)
	bt = binary.BigEndian.AppendUint64(bt, c.StartOffset)

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

func (c *CatMessageV3) Decode(body []byte) {
	var off uint64
	c.Name, off = GetCstring(body[4:])
	if body[1] == byte(DecryptMessage) {
		c.Decrypt = true
	}
	if body[2] == byte(UseKEK) {
		c.KEK = true
	}
	c.StartOffset = binary.BigEndian.Uint64(body[4+len(c.Name)+1:])

	settLen := binary.BigEndian.Uint64(body[4+8+off : 4+off+8+8])

	totalOff := 4 + off + 8 + 8

	c.Settings = make([]settings.StorageSettings, settLen)

	for i := 0; i < int(settLen); i++ {

		var currOff uint64

		c.Settings[i].Name, currOff = GetCstring(body[totalOff:])
		totalOff += currOff

		c.Settings[i].Value, currOff = GetCstring(body[totalOff:])
		totalOff += currOff
	}
}
