package limiter

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/metrics"
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

func (r *Reader) Wait(n int) error {
	if r.limiter == nil {
		return nil
	}
	start := time.Now()
	err := r.limiter.WaitN(r.ctx, n)
	waitTime := time.Since(start).Nanoseconds()
	metrics.StoreLatencyAndSizeInfo("LIMIT_READ", float64(n), float64(waitTime))
	return err
}

func (r *Reader) getBurstableLimit(n int) int {
	if r.limiter == nil {
		return n
	}
	return min(r.limiter.Burst(), n)
}

func (r *Reader) Read(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, fmt.Errorf("empty buffer passed")
	}

	end := r.getBurstableLimit(len(buf))
	// we do not know how many bytes we could read, so first - read data
	n, err := r.reader.Read(buf[:end])

	// and then wait in a limiter to correct measure processed data
	if err != nil {
		N := max(n, 0)
		limiterErr := r.Wait(N)
		if limiterErr != nil {
			ylogger.Zero.Error().Err(limiterErr).Msg("Error happened while limiting")
		}
		return n, err
	}

	err = r.Wait(n)
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

func (w *Writer) Wait(n int) error {
	if w.limiter == nil {
		return nil
	}
	start := time.Now()
	err := w.limiter.WaitN(w.ctx, n)
	waitTime := time.Since(start).Nanoseconds()
	metrics.StoreLatencyAndSizeInfo("LIMIT_WRITE", float64(n), float64(waitTime))
	return err
}

func (w *Writer) getBurstableLimit(n int) int {
	if w.limiter == nil {
		return n
	}
	return min(w.limiter.Burst(), n)
}

func (r *Writer) Write(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, fmt.Errorf("empty buffer passed")
	}

	end := r.getBurstableLimit(len(buf))
	// in a case of write we should wait before handling query
	limiterErr := r.Wait(end)
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
	if netLimit == 0 {
		return nil
	}
	ylogger.Zero.Debug().Uint64("bytes per sec", netLimit).Msg("allocate limiter")

	netLimiter = rate.NewLimiter(rate.Limit(netLimit),
		int(netLimit))

	return netLimiter
}
