package message_test

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yezzey-gp/yproxy/pkg/message"
	"github.com/yezzey-gp/yproxy/pkg/settings"
)

func TestCatMsg(t *testing.T) {
	assert := assert.New(t)

	type tcase struct {
		name    string
		decrypt bool
		off     uint64
		err     error
	}

	for _, tt := range []tcase{
		{
			"nam1",
			true,
			0,
			nil,
		},
		{
			"nam1",
			true,
			10,
			nil,
		},
	} {

		msg := message.NewCatMessage(tt.name, tt.decrypt, tt.off)
		body := msg.Encode()

		msg2 := message.CatMessage{}

		msg2.Decode(body[8:])

		assert.Equal(msg.Name, msg2.Name)
		assert.Equal(msg.Decrypt, msg2.Decrypt)
	}
}

func TestPutMsg(t *testing.T) {
	assert := assert.New(t)

	type tcase struct {
		name    string
		encrypt bool
		err     error
	}

	for _, tt := range []tcase{
		{
			"nam1",
			true,
			nil,
		},
	} {

		msg := message.NewPutMessage(tt.name, tt.encrypt)
		body := msg.Encode()

		msg2 := message.PutMessage{}

		msg2.Decode(body[8:])

		assert.Equal(msg.Name, msg2.Name)
		assert.Equal(msg.Encrypt, msg2.Encrypt)
	}
}

func TestPutV2Msg(t *testing.T) {
	assert := assert.New(t)

	type tcase struct {
		name     string
		encrypt  bool
		err      error
		settings []settings.StorageSettings
	}

	for _, tt := range []tcase{
		{
			"nam1",
			true,
			nil,
			[]settings.StorageSettings{
				{
					Name:  "a",
					Value: "b",
				},
				{
					Name:  "cdsdsd",
					Value: "ds",
				},
			},
		},
	} {

		msg := message.NewPutMessageV2(tt.name, tt.encrypt, tt.settings)
		body := msg.Encode()

		msg2 := message.PutMessageV2{}

		msg2.Decode(body[8:])

		assert.Equal(msg.Name, msg2.Name)
		assert.Equal(msg.Encrypt, msg2.Encrypt)
		assert.Equal(msg.Settings, msg2.Settings)
	}
}

func TestPutMsgV3(t *testing.T) {
	assert := assert.New(t)

	type tcase struct {
		name     string
		encrypt  bool
		err      error
		settings []settings.StorageSettings
	}

	for _, tt := range []tcase{
		{
			"nam1",
			true,
			nil,
			[]settings.StorageSettings{
				{
					Name:  "a",
					Value: "b",
				},
				{
					Name:  "cdsdsd",
					Value: "ds",
				},
			},
		},
	} {

		msg := message.NewPutMessageV3(tt.name, tt.encrypt, tt.settings)
		body := msg.Encode()

		msg2 := message.PutMessageV3{}

		msg2.Decode(body[8:])

		assert.Equal(msg.Name, msg2.Name)
		assert.Equal(msg.Encrypt, msg2.Encrypt)
		assert.Equal(msg.Settings, msg2.Settings)
	}
}

func TestPutCompleteMsg(t *testing.T) {
	assert := assert.New(t)

	type tcase struct {
		keyVersion uint16
	}

	for _, tt := range []tcase{
		{
			1,
		},
	} {

		msg := message.NewPutCompleteMessage(tt.keyVersion)
		body := msg.Encode()

		msg2 := message.PutCompleteMessage{}

		msg2.Decode(body[8:])

		assert.Equal(msg.KeyVersion, msg2.KeyVersion)
	}
}

func TestCatMsgV2(t *testing.T) {
	assert := assert.New(t)

	type tcase struct {
		name    string
		decrypt bool
		off     uint64

		settings []settings.StorageSettings
		err      error
	}

	for _, tt := range []tcase{
		{
			"nam1",
			true,
			0,
			[]settings.StorageSettings{
				{
					Name:  "a",
					Value: "b",
				},
				{
					Name:  "cdsdsd",
					Value: "ds",
				},
			},
			nil,
		},
		{
			"nam1",
			true,
			10,
			[]settings.StorageSettings{
				{
					Name:  "a",
					Value: "b",
				},
				{
					Name:  "cdsdsd",
					Value: "ds",
				},
			},
			nil,
		},
	} {

		msg := message.NewCatMessageV2(tt.name, tt.decrypt, tt.off, tt.settings)
		body := msg.Encode()

		msg2 := message.CatMessageV2{}

		msg2.Decode(body[8:])

		assert.Equal(msg.Name, msg2.Name)
		assert.Equal(msg.Decrypt, msg2.Decrypt)
		assert.Equal(msg.StartOffset, msg2.StartOffset)
		assert.Equal(msg.Settings, msg2.Settings)
	}
}

func TestPatchMsg(t *testing.T) {
	assert := assert.New(t)

	type tcase struct {
		name    string
		encrypt bool
		off     uint64
		err     error
	}

	for _, tt := range []tcase{
		{
			"nam1",
			true,
			1235,
			nil,
		},
	} {

		msg := message.NewPatchMessage(tt.name, tt.off, tt.encrypt)
		body := msg.Encode()

		msg2 := message.PatchMessage{}

		msg2.Decode(body[8:])

		assert.Equal(msg.Name, msg2.Name)
		assert.Equal(msg.Encrypt, msg2.Encrypt)
		assert.Equal(msg.Offset, msg2.Offset)
	}
}

func TestCopyDataMsg(t *testing.T) {
	assert := assert.New(t)

	type tcase struct {
		body []byte
		err  error
	}

	for _, tt := range []tcase{
		{
			[]byte(
				"hiuefheiufheuif",
			),
			nil,
		},
	} {

		msg := message.NewCopyDataMessage()
		msg.Data = tt.body
		msg.Sz = uint64(len(tt.body))
		body := msg.Encode()

		msg2 := message.CopyDataMessage{}

		msg2.Decode(body[8:])

		sz := binary.BigEndian.Uint64(body[:8])

		assert.Equal(int(sz), len(body))

		assert.Equal(msg.Data, msg2.Data)
		assert.Equal(msg.Sz, msg2.Sz)
	}
}

func TestListMsg(t *testing.T) {
	assert := assert.New(t)

	type tcase struct {
		prefix string
	}

	for _, tt := range []tcase{
		{
			"nam1",
		},
	} {

		msg := message.NewListMessage(tt.prefix)
		body := msg.Encode()

		msg2 := message.ListMessage{}

		msg2.Decode(body[8:])

		assert.Equal(msg.Prefix, msg2.Prefix)
	}
}

func TestCopyMsg(t *testing.T) {
	assert := assert.New(t)

	msg := message.NewCopyMessage("myname/mynextname", "myoldcfg/path", true, true, true, 5432)
	body := msg.Encode()

	assert.Equal(body[8], byte(message.MessageTypeCopy))

	msg2 := message.CopyMessage{}
	msg2.Decode(body[8:])

	assert.Equal("myname/mynextname", msg2.Name)
	assert.Equal("myoldcfg/path", msg2.OldCfgPath)
	assert.True(msg2.Decrypt)
	assert.True(msg2.Encrypt)
	assert.Equal(uint64(5432), msg2.Port)
	assert.True(msg2.Confirm)
}

func TestDeleteMsg(t *testing.T) {
	assert := assert.New(t)

	msg := message.NewDeleteMessage("myname/mynextname", 5432, 42, true, true)
	body := msg.Encode()

	assert.Equal(body[8], byte(message.MessageTypeDelete))

	msg2 := message.DeleteMessage{}
	msg2.Decode(body[8:])

	assert.Equal("myname/mynextname", msg2.Name)
	assert.Equal(uint64(5432), msg2.Port)
	assert.Equal(uint64(42), msg2.Segnum)
	assert.True(msg2.Confirm)
	assert.True(msg2.Garbage)
}

func TestErrorMsg(t *testing.T) {
	assert := assert.New(t)

	errString := "could not list objects: NoCredentialProviders: no valid providers in chain."
	messageString := "Deprecated. For verbose messaging see aws.Config.CredentialsChainVerboseErrors"
	err := fmt.Errorf("%s", errString)

	msg := message.NewErrorMessage(err, messageString)
	body := msg.Encode()

	assert.Equal(body[8], byte(message.MessageTypeError))

	msg2 := message.ErrorMessage{}
	msg2.Decode(body[8:])

	assert.Equal(errString, msg2.Error)
	assert.Equal(messageString, msg2.Message)
}
