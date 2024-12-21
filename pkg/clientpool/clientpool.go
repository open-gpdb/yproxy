package clientpool

import (
	"fmt"
	"sync"

	"github.com/yezzey-gp/yproxy/pkg/client"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

type Pool interface {
	ClientPoolForeach(cb func(client client.YproxyClient) error) error

	Put(client client.YproxyClient) error
	Pop(id uint) (bool, error)

	Shutdown() error
}

type PoolImpl struct {
	mu   sync.Mutex
	pool map[uint]client.YproxyClient
}

var _ Pool = &PoolImpl{}

func (c *PoolImpl) Put(client client.YproxyClient) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.pool[client.ID()] = client

	return nil
}
func (c *PoolImpl) Pop(id uint) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cl, ok := c.pool[id]
	if ok {
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
		pool: map[uint]client.YproxyClient{},
		mu:   sync.Mutex{},
	}
}
