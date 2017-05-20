package consul

import (
	"errors"
	"sync"
	"time"

	consulapi "github.com/hashicorp/consul/api"
)

var ErrExited = errors.New("exited")
var ErrTimeout = errors.New("timeout")
var ErrExisted = errors.New("existed")

type Response struct {
	Action Action `json:"action"`

	PreKvpair *consulapi.KVPair
	Kvpair    *consulapi.KVPair

	watchIndex uint64
}

type consuler struct {
	lock sync.RWMutex

	nodes map[string]*Node //key->Node

	event_pool chan *Response

	chexit chan bool
}

func (c *consuler) Next(timeout time.Duration) (*Response, error) {
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	timer := time.NewTimer(timeout)

	for {
		select {
		case <-c.chexit:
			return nil, ErrExited
		case r := c.event_pool:
			return r, nil
		case <-timer.C:
			return nil, ErrTimeout
		}
	}
}

func (c *consuler) addNode(kv consulapi.KVPair) error {
	c.lock.Lock()
	node, ok := c.nodes[kv.Key]
	if ok {
		c.lock.Unlock()
		return ErrExisted
	}

	node, err := NewNode(&kv)
	if err != nil {
		c.lock.Unlock()
		return err
	}

	c.nodes[kv.Key] = node
	node.start()
}

func (c *consuler) removeNode(key string) bool {
	c.lock.Lock()
	node, ok := c.nodes[key]
	if ok {
		node.Stop()
		delete(c.nodes, key)
	}
	c.lock.Unlock()
	return ok
}

func (c *consuler) putResponse(resp *Response) {
	c.event_pool <- resp
}
