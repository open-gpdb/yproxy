package proc

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/yezzey-gp/yproxy/pkg/client"
	"github.com/yezzey-gp/yproxy/pkg/message"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

type ProtoReader struct {
	c io.ReadWriteCloser
}

func NewProtoReader(ycl client.YproxyClient) *ProtoReader {
	return &ProtoReader{
		c: ycl.GetRW(),
	}
}

// 1mb of data + header
const maxMsgLen = 1<<20 | 1<<10

func (r *ProtoReader) ReadPacket() (message.MessageType, []byte, error) {
	msgLenBuf := make([]byte, 8)
	_, err := io.ReadFull(r.c, msgLenBuf)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read params: %w", err)
	}

	dataLen := binary.BigEndian.Uint64(msgLenBuf)

	if dataLen > maxMsgLen {
		return 0, nil, fmt.Errorf("message too big %d", dataLen)
	}

	if dataLen <= 8 {
		return 0, nil, fmt.Errorf("message empty")
	}

	dataLen -= 8

	ylogger.Zero.Debug().Uint64("size", dataLen).Msg("requested packet")

	data := make([]byte, dataLen)
	_, err = io.ReadFull(r.c, data)
	if err != nil {
		return 0, nil, err
	}

	return message.MessageType(data[0]), data, nil
}
