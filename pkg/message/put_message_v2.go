package message

import (
	"encoding/binary"

	"github.com/yezzey-gp/yproxy/pkg/settings"
)

const StorageClassSetting = "StorageClass"
const TableSpaceSetting = "TableSpace"
const MultipartChunksize = "MultipartChunksize"
const DoMultipart = "DoMultipart"

type PutMessageV2 struct {
	Encrypt bool
	Name    string

	Settings []settings.StorageSettings
}

var _ ProtoMessage = &PutMessageV2{}

func NewPutMessageV2(name string, encrypt bool, settings []settings.StorageSettings) *PutMessageV2 {
	return &PutMessageV2{
		Name:     name,
		Encrypt:  encrypt,
		Settings: settings,
	}
}

func (c *PutMessageV2) Encode() []byte {
	bt := []byte{
		byte(MessageTypePutV2),
		0,
		0,
		0,
	}

	if c.Encrypt {
		bt[1] = byte(EncryptMessage)
	} else {
		bt[1] = byte(NoEncryptMessage)
	}

	bt = append(bt, []byte(c.Name)...)
	bt = append(bt, 0)

	slen := make([]byte, 8)
	binary.BigEndian.PutUint64(slen, uint64(len(c.Settings)))
	bt = append(bt, slen...)

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

func (c *PutMessageV2) Decode(body []byte) {
	if body[1] == byte(EncryptMessage) {
		c.Encrypt = true
	}
	var off uint64
	c.Name, off = GetCstring(body[4:])

	settLen := binary.BigEndian.Uint64(body[4+off : 4+off+8])

	totalOff := 4 + off + 8

	c.Settings = make([]settings.StorageSettings, settLen)

	for i := 0; i < int(settLen); i++ {

		var currOff uint64

		c.Settings[i].Name, currOff = GetCstring(body[totalOff:])
		totalOff += currOff

		c.Settings[i].Value, currOff = GetCstring(body[totalOff:])
		totalOff += currOff
	}
}
