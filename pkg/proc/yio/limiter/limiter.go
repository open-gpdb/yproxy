package limiter

import (
	"context"
	"fmt"
	"io"

	"github.com/yezzey-gp/yproxy/pkg/ylogger"
	"golang.org/x/time/rate"
)

type Reader struct {
	reader  io.ReadCloser
	limiter *rate.Limiter
	ctx     context.Context
}

// Close implements io.ReadCloser.
func (r *Reader) Close() error {
	return r.reader.Close()
}

func NewReader(reader io.ReadCloser, limiter *rate.Limiter) *Reader {
	return &Reader{
		reader:  reader,
		limiter: limiter,
	}
}

func (r *Reader) Read(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, fmt.Errorf("empty buffer passed")
	}

	end := len(buf)
	if r.limiter.Burst() < end {
		end = r.limiter.Burst()
	}
	n, err := r.reader.Read(buf[:end])

	if err != nil {
		N := n
		if n < 0 {
			N = 0
		}
		limiterErr := r.limiter.WaitN(r.ctx, N)
		if limiterErr != nil {
			ylogger.Zero.Error().Err(limiterErr).Msg("Error happened while limiting")
		}
		return n, err
	}

	err = r.limiter.WaitN(r.ctx, n)
	return n, err
}
