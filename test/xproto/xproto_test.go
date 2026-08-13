package xproto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yezzey-gp/yproxy/pkg/message"
)

func copyData(b []byte) *message.CopyDataMessage {
	return &message.CopyDataMessage{Sz: uint64(len(b)), Data: b}
}

func TestPutSimple(t *testing.T) {
	s := newServer(t)

	fname := "cat-f"

	payload := []byte("the quick brown fox jumps over the lazy dog")

	protoTestRunner(t, s, []MessageGroup{{
		Name: "put",
		Request: []wireMessage{
			message.NewPutMessage(fname, false),
			copyData(payload),
			message.NewCopyDoneMessage(),
		},
		Response: []wireMessage{message.NewReadyForQueryMessage()},
	}})

	conn := s.dial(t)
	defer func() { _ = conn.Close() }()

	_, err := conn.Write(message.NewCatMessage(fname, false, 0).Encode())
	require.NoError(t, err)
	assert.Equal(t, payload, readRaw(t, conn, len(payload)))
}

func TestCatWithStartOffset(t *testing.T) {
	s := newServer(t)

	payload := []byte("any random bytes")
	const offset = 6

	protoTestRunner(t, s, []MessageGroup{{
		Name: "put",
		Request: []wireMessage{
			message.NewPutMessage("offset.bin", false),
			copyData(payload),
			message.NewCopyDoneMessage(),
		},
		Response: []wireMessage{message.NewReadyForQueryMessage()},
	}})

	conn := s.dial(t)
	defer func() { _ = conn.Close() }()

	_, err := conn.Write(message.NewCatMessage("offset.bin", false, offset).Encode())
	require.NoError(t, err)
	assert.Equal(t, payload[offset:], readRaw(t, conn, len(payload)-offset))
}

func TestCatV2WithStartOffset(t *testing.T) {
	s := newServer(t)

	payload := []byte("any random bytes")
	const offset = 6

	protoTestRunner(t, s, []MessageGroup{{
		Name: "put",
		Request: []wireMessage{
			message.NewPutMessage("offset.bin", false),
			copyData(payload),
			message.NewCopyDoneMessage(),
		},
		Response: []wireMessage{message.NewReadyForQueryMessage()},
	}})

	conn := s.dial(t)
	defer func() { _ = conn.Close() }()

	_, err := conn.Write(message.NewCatMessageV2("offset.bin", false, false, offset, nil).Encode())
	require.NoError(t, err)
	assert.Equal(t, payload[offset:], readRaw(t, conn, len(payload)-offset))
}

func TestCatNonExistent(t *testing.T) {
	s := newServer(t)

	conn := s.dial(t)
	defer func() { _ = conn.Close() }()

	_, err := conn.Write(message.NewCatMessageV2("offset.bin-no-such", false, false, 67, nil).Encode())
	require.NoError(t, err)

	b := make([]byte, 10000)

	_, err = conn.Read(b)

	require.Error(t, err)
}

type badMessage struct{}

func (b *badMessage) Encode() []byte {
	body := []byte{0xFF, 0, 0, 0}
	out := make([]byte, 8)
	out[7] = byte(len(body) + 8)
	return append(out, body...)
}

func TestUnsupportedMessageType(t *testing.T) {
	s := newServer(t)

	protoTestRunner(t, s, []MessageGroup{{
		Name:    "bad_type",
		Request: []wireMessage{&badMessage{}},
		Response: []wireMessage{&message.ErrorMessage{
			Error:   "wrong request type: UNKNOWN",
			Message: "message is unsupported in ProcConn",
		}},
	}})
}
