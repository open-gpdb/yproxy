package proc_test

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/message"
	"github.com/yezzey-gp/yproxy/pkg/proc"
)

type procConnTestClient struct {
	rw           *procConnTestRW
	op           message.MessageType
	opStart      time.Time
	byteOffset   int64
	externalPath string
	closed       bool
}

func newProcConnTestClient(input []byte) *procConnTestClient {
	return &procConnTestClient{rw: &procConnTestRW{reader: bytes.NewReader(input)}}
}

func (c *procConnTestClient) ID() uint {
	return 1
}

func (c *procConnTestClient) ReplyError(err error, msg string) error {
	_, _ = c.rw.Write(message.NewErrorMessage(err, msg).Encode())
	return nil
}

func (c *procConnTestClient) GetRW() io.ReadWriteCloser {
	return c.rw
}

func (c *procConnTestClient) SetOPType(optype message.MessageType) {
	c.op = optype
	c.opStart = time.Now()
}

func (c *procConnTestClient) OPType() message.MessageType {
	return c.op
}

func (c *procConnTestClient) OPStart() time.Time {
	return c.opStart
}

func (c *procConnTestClient) SetByteOffset(offset int64) {
	c.byteOffset = offset
}

func (c *procConnTestClient) ByteOffset() int64 {
	return c.byteOffset
}

func (c *procConnTestClient) SetExternalFilePath(path string) {
	c.externalPath = path
}

func (c *procConnTestClient) ExternalFilePath() string {
	return c.externalPath
}

func (c *procConnTestClient) Close() error {
	c.closed = true
	return c.rw.Close()
}

type procConnTestRW struct {
	reader *bytes.Reader
	writer bytes.Buffer
	closed bool
}

func (rw *procConnTestRW) Read(p []byte) (int, error) {
	return rw.reader.Read(p)
}

func (rw *procConnTestRW) Write(p []byte) (int, error) {
	return rw.writer.Write(p)
}

func (rw *procConnTestRW) Close() error {
	rw.closed = true
	return nil
}

func (rw *procConnTestRW) Written() []byte {
	return rw.writer.Bytes()
}

func TestProcConnUnknownMessageTypeRepliesErrorWithoutPanic(t *testing.T) {
	const unknownType = message.MessageType(200)

	ycl := newProcConnTestClient(testPacket(unknownType))

	require.NotPanics(t, func() {
		err := proc.ProcConn(nil, nil, nil, ycl, &config.Vacuum{})
		require.NoError(t, err)
	})
	require.True(t, ycl.closed)
	require.Equal(t, unknownType, ycl.OPType())

	body := decodeWrittenPacket(t, ycl.rw.Written())
	require.Equal(t, message.MessageTypeError, message.MessageType(body[0]))

	errorMessage := message.ErrorMessage{}
	require.NotPanics(t, func() { errorMessage.Decode(body) })
	require.Equal(t, "wrong request type", errorMessage.Message)
	require.Contains(t, errorMessage.Error, "wrong request type")
}

func TestProcConnCopyDoneDoesNothing(t *testing.T) {
	ycl := newProcConnTestClient(testPacket(message.MessageTypeCopyDone))

	require.NotPanics(t, func() {
		err := proc.ProcConn(nil, nil, nil, ycl, &config.Vacuum{})
		require.NoError(t, err)
	})
	require.True(t, ycl.closed)
	require.Equal(t, message.MessageTypeCopyDone, ycl.OPType())
	require.Empty(t, ycl.rw.Written())
}

func testPacket(tp message.MessageType) []byte {
	body := []byte{byte(tp), 0, 0, 0}
	packet := make([]byte, 8, 8+len(body))
	binary.BigEndian.PutUint64(packet, uint64(8+len(body)))
	return append(packet, body...)
}

func decodeWrittenPacket(t *testing.T, packet []byte) []byte {
	t.Helper()
	require.GreaterOrEqual(t, len(packet), 9)

	packetLen := binary.BigEndian.Uint64(packet[:8])
	require.Equal(t, uint64(len(packet)), packetLen)

	return packet[8:]
}
