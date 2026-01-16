package clientpool

import (
	"fmt"
	"sync"
	"time"

	"github.com/caio/go-tdigest"
	"github.com/yezzey-gp/yproxy/pkg/client"
	"github.com/yezzey-gp/yproxy/pkg/metrics"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Pool interface {
	ClientPoolForeach(cb func(client client.YproxyClient) error) error

	Put(client client.YproxyClient) error
	Pop(id uint) (bool, error)

	Quantile(ct int, q []float64) []QuantInfo

	Shutdown() error
}

type PoolImpl struct {
	mu   sync.Mutex
	pool map[uint]client.YproxyClient
	/* size category -> optype -> speed quantiles */
	opSpeed map[int]map[string]*tdigest.TDigest
}

var _ Pool = &PoolImpl{}

func SizeToCat(sz int64) int {
	if sz < 1024*1024 {
		return 0
	}
	if sz < 16*1024*1024 {
		return 1
	}
	return 2
}

func (c *PoolImpl) Put(client client.YproxyClient) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.pool[client.ID()] = client

	return nil
}

type QuantInfo struct {
	Op string
	Q  []float64
}

func (c *PoolImpl) Quantile(ct int, q []float64) []QuantInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	var ret []QuantInfo
	for k, v := range c.opSpeed[ct] {
		var Qs []float64
		for _, qq := range q {
			Qs = append(Qs, v.Quantile(qq))
		}
		ret = append(ret, QuantInfo{
			Op: k,
			Q:  Qs,
		})
	}
	return ret
}

func (c *PoolImpl) Pop(id uint) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cl, ok := c.pool[id]
	if ok {

		total := cl.ByteOffset()
		if total > 0 {
			ct := SizeToCat(total)
			timeTotal := time.Since(cl.OPStart()).Nanoseconds()

			optyp := cl.OPType().String()
			if c.opSpeed[ct][optyp] == nil {
				c.opSpeed[ct][optyp], _ = tdigest.New()
			}

			_ = c.opSpeed[ct][optyp].Add(float64(total) / float64(timeTotal))
			metrics.StoreLatencyAndSizeInfo(cl.OPType().String(), float64(total), float64(timeTotal))
		}

		delete(c.pool, id)
		/* be conservative */
		_ = cl.Close()
		return true, nil
	}

	return ok, fmt.Errorf("failed to find client %d connection", id)
}

func (c *PoolImpl) Shutdown() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, cl := range c.pool {
		go func(cl client.YproxyClient) {
			if err := cl.Close(); err != nil {
				ylogger.Zero.Error().Err(err).Msg("")
			}
		}(cl)
	}

	return nil
}
func (c *PoolImpl) ClientPoolForeach(cb func(client client.YproxyClient) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, cl := range c.pool {
		if err := cb(cl); err != nil {
			ylogger.Zero.Error().Err(err).Msg("")
		}
	}

	return nil
}

func NewClientPool() Pool {
	pool := &PoolImpl{
		pool: map[uint]client.YproxyClient{},
		mu:   sync.Mutex{},
		opSpeed: map[int]map[string]*tdigest.TDigest{
			0: {},
			1: {},
			2: {},
		},
	}
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "client_connections",
		Help: "The number of client connections to yproxy",
	},
		func() float64 {
			pool.mu.Lock()
			defer pool.mu.Unlock()
			return float64(len(pool.pool))
		})
	return pool
}
