package limiter

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/yezzey-gp/yproxy/config"
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
		ctx:     context.Background(),
		reader:  reader,
		limiter: limiter,
	}
}

func (r *Reader) Read(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, fmt.Errorf("empty buffer passed")
	}

	end := min(r.limiter.Burst(), len(buf))
	n, err := r.reader.Read(buf[:end])

	if err != nil {
		N := max(n, 0)
		limiterErr := r.limiter.WaitN(r.ctx, N)
		if limiterErr != nil {
			ylogger.Zero.Error().Err(limiterErr).Msg("Error happened while limiting")
		}
		return n, err
	}

	err = r.limiter.WaitN(r.ctx, n)
	return n, err
}

type Writer struct {
	writer  io.WriteCloser
	limiter *rate.Limiter
	ctx     context.Context
}

// Close implements io.ReadCloser.
func (r *Writer) Close() error {
	return r.writer.Close()
}

func NewWriter(writer io.WriteCloser, limiter *rate.Limiter) *Writer {
	return &Writer{
		ctx:     context.Background(),
		writer:  writer,
		limiter: limiter,
	}
}

func (r *Writer) Write(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, fmt.Errorf("empty buffer passed")
	}

	end := min(r.limiter.Burst(), len(buf))
	// in a case of write we should wait before handling query
	limiterErr := r.limiter.WaitN(r.ctx, end)
	if limiterErr != nil {
		ylogger.Zero.Error().Err(limiterErr).Msg("Error happened while limiting")
	}
	n, err := r.writer.Write(buf[:end])

	return n, err
}

var (
	/* Single limiter for all external storage interaction */
	mu         sync.Mutex
	netLimiter *rate.Limiter = nil
)

func GetLimiter() *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	if netLimiter != nil {
		ylogger.Zero.Debug().Msg("reuse limiter")
		return netLimiter
	}

	netLimit := config.InstanceConfig().StorageCnf.StorageRateLimit
	ylogger.Zero.Debug().Uint64("bytes per sec", netLimit).Msg("allocate limiter")

	netLimiter = rate.NewLimiter(rate.Limit(netLimit),
		int(netLimit))

	return netLimiter
}
