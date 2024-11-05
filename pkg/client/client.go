package client

import (
	"io"
	"net"
	"reflect"

	"github.com/yezzey-gp/yproxy/pkg/message"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

type YproxyClient interface {
	ID() uint
	ReplyError(err error, msg string) error
	GetRW() io.ReadWriteCloser

	SetOPType(optype byte)
	OPType() byte

	SetExternalFilePath(path string)
	ExternalFilePath() string

	Close() error
}

type YClient struct {
	Conn net.Conn
	op   byte
	path string
}

// ExternalFilePath implements YproxyClient.
func (y *YClient) ExternalFilePath() string {
	return y.path
}

// SetExternalFilePath implements YproxyClient.
func (y *YClient) SetExternalFilePath(path string) {
	y.path = path
}

// OPType implements YproxyClient.
func (y *YClient) OPType() byte {
	return y.op
}

// SetOPType implements YproxyClient.
func (y *YClient) SetOPType(optype byte) {
	y.op = optype
}

// Close implements YproxyClient.
func (y *YClient) Close() error {
	return y.Conn.Close()
}

// GetPointer do the same thing like fmt.Sprintf("%p", &num) but fast
// GetPointer returns the memory address of the given value as an unsigned integer.
func GetPointer(value interface{}) uint {
	ptr := reflect.ValueOf(value).Pointer()
	uintPtr := uintptr(ptr)
	return uint(uintPtr)
}

// ID implements YproxyClient.
func (y *YClient) ID() uint {
	return GetPointer(y)
}

func NewYClient(c net.Conn) YproxyClient {
	return &YClient{
		Conn: c,
	}
}

func (y *YClient) ReplyError(err error, msg string) error {
	ylogger.Zero.Error().Err(err).Msg(msg)

	_, _ = y.Conn.Write(message.NewErrorMessage(err, msg).Encode())
	return nil
}

func (y *YClient) GetRW() io.ReadWriteCloser {
	return y.Conn
}
