package yio

import (
	"io"

	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/client"
	"github.com/yezzey-gp/yproxy/pkg/metrics"
	"github.com/yezzey-gp/yproxy/pkg/proc/yio/limiter"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
	"golang.org/x/time/rate"
)

/* TBD: support restart */
type YproxyWriter struct {
	underlying io.WriteCloser

	lim *rate.Limiter

	selfCl client.YproxyClient

	offsetReached int64
}

// Close implements io.WriteCloser.
func (y *YproxyWriter) Close() error {
	return y.underlying.Close()
}

func (y *YproxyWriter) Write(p []byte) (n int, err error) {

	n, err = y.underlying.Write(p)
	metrics.WiteReqProcessed.Inc()
	y.offsetReached += int64(n)
	y.selfCl.SetByteOffset(y.offsetReached)

	if err != nil {
		metrics.WriteReqErrors.Inc()
		ylogger.Zero.Error().Uint("client id", y.selfCl.ID()).Int("bytes write", n).Err(err).Msg("failed to write into underlying connection")
	}

	return n, err
}

func NewYproxyWriter(under io.WriteCloser, selfCl client.YproxyClient) io.WriteCloser {

	w := &YproxyWriter{
		underlying:    under,
		selfCl:        selfCl,
		offsetReached: 0,
	}

	/* with limiter ? */

	if config.InstanceConfig().StorageCnf.EnableRateLimiter {
		w.lim = limiter.GetLimiter()
		w.underlying = limiter.NewWriter(under, w.lim)
	} else {
		w.underlying = under
	}

	return w
}

var _ io.WriteCloser = &YproxyWriter{}
