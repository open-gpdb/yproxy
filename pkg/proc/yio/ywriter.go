package yio

import (
	"io"

	"github.com/yezzey-gp/yproxy/pkg/client"
)

/* TBD: support restart */
type YproxyWriter struct {
	underlying io.WriteCloser

	selfCl client.YproxyClient

	offsetReached int64
}

// Close implements io.WriteCloser.
func (y *YproxyWriter) Close() error {
	return y.underlying.Close()
}

func (y *YproxyWriter) Write(p []byte) (n int, err error) {
	n, err = y.underlying.Write(p)
	y.offsetReached += int64(n)
	y.selfCl.SetByteOffset(y.offsetReached)

	return n, err
}

func NewYproxyWriter(under io.WriteCloser, selfCl client.YproxyClient) io.WriteCloser {
	return &YproxyWriter{
		underlying:    under,
		selfCl:        selfCl,
		offsetReached: 0,
	}
}

var _ io.WriteCloser = &YproxyWriter{}
