package clientpool

import (
	"fmt"
	"sync"
	"time"

	"github.com/caio/go-tdigest"
	"github.com/yezzey-gp/yproxy/pkg/client"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

type Pool interface {
	ClientPoolForeach(cb func(client client.YproxyClient) error) error

	Put(client client.YproxyClient) error
	Pop(id uint) (bool, error)

	Quantile(q []float64) []QuantInfo

	Shutdown() error
}

type PoolImpl struct {
	mu   sync.Mutex
	pool map[uint]client.YproxyClient
	/* optype -> speed quantiles */
	opSpeed map[string]*tdigest.TDigest
}

var _ Pool = &PoolImpl{}

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

func (c *PoolImpl) Quantile(q []float64) []QuantInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	var ret []QuantInfo
	for k, v := range c.opSpeed {
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
		if total != 0 {
			timeTotal := time.Now().Sub(cl.OPStart()).Nanoseconds()

			optyp := cl.OPType().String()
			if c.opSpeed[optyp] == nil {
				c.opSpeed[optyp], _ = tdigest.New()
			}

			c.opSpeed[optyp].Add(float64(total) / float64(timeTotal))
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
	return &PoolImpl{
		pool:    map[uint]client.YproxyClient{},
		mu:      sync.Mutex{},
		opSpeed: map[string]*tdigest.TDigest{},
	}
}
