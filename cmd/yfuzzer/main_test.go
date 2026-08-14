package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/crypt"
	"github.com/yezzey-gp/yproxy/pkg/message"
	mockcl "github.com/yezzey-gp/yproxy/pkg/mock/client"
	"github.com/yezzey-gp/yproxy/pkg/storage"
)

type captureRW struct {
	buf    bytes.Buffer
	failAt int
	calls  int
}

func (rw *captureRW) Read(p []byte) (int, error) { return 0, io.EOF }

func (rw *captureRW) Write(p []byte) (int, error) {
	rw.calls++
	if rw.failAt > 0 && rw.calls > rw.failAt {
		return 0, errors.New("simulated write failure")
	}
	return rw.buf.Write(p)
}

func (rw *captureRW) Close() error { return nil }

type packet struct {
	msgType message.MessageType
	body    []byte
}

func parsePackets(t *testing.T, data []byte) []packet {
	t.Helper()
	var packets []packet
	for len(data) > 0 {
		assert.GreaterOrEqual(t, len(data), 8, "truncated length prefix")
		packetLength := binary.BigEndian.Uint64(data[:8])
		assert.GreaterOrEqual(t, packetLength, uint64(8), "invalid packet length")
		assert.LessOrEqual(t, int(packetLength), len(data), "packet length exceeds buffer")
		body := data[8:packetLength]
		packets = append(packets, packet{
			msgType: message.MessageType(body[0]),
			body:    body,
		})
		data = data[packetLength:]
	}
	return packets
}

func fuzz(cl *mockcl.MockYproxyClient) {
	_ = (&FuzzerProtoMgr{}).ProcessListExtended(
		"test/prefix",
		nil,
		storage.StorageInteractor(nil),
		crypt.Crypter(nil),
		cl,
		&config.Vacuum{},
	)
}

func TestProcessListExtended_ReturnsNilAndSetsPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	cl := mockcl.NewMockYproxyClient(ctrl)
	rw := &captureRW{}

	var capturedPath string
	cl.EXPECT().SetExternalFilePath("test/prefix").Do(func(p string) { capturedPath = p }).Times(1)
	cl.EXPECT().GetRW().Return(rw).AnyTimes()

	fuzz(cl)
	assert.Equal(t, "test/prefix", capturedPath)
}

func TestProcessListExtended_EndsWithReadyForQuery(t *testing.T) {
	for range 50 {
		ctrl := gomock.NewController(t)
		cl := mockcl.NewMockYproxyClient(ctrl)
		rw := &captureRW{}

		cl.EXPECT().SetExternalFilePath(gomock.Any()).AnyTimes()
		cl.EXPECT().GetRW().Return(rw).AnyTimes()

		fuzz(cl)

		pkts := parsePackets(t, rw.buf.Bytes())
		assert.NotEmpty(t, pkts, "should send at least one packet")
		assert.Equal(t, message.MessageTypeReadyForQuery, pkts[len(pkts)-1].msgType,
			"last packet must be ReadyForQuery")
	}
}
