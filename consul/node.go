package consul

import (
	"errors"
	"log"
	"sync"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"golang.org/x/net/context"
)

type Node struct {
	lock sync.RWMutex

	Key       string
	Index     uint64
	PreKvpair *consulapi.KVPair
	Kvpair    *consulapi.KVPair

	consuler *consuler

	closed bool
	exit   chan bool
}

var ErrNil = errors.New("nil")

func (c *consuler) NewNode(kvpair *consulapi.KVPair) (*Node, error) {
	if kvpair == nil {
		return nil, ErrNil
	}

	node := &Node{
		Key:       kvpair.Key,
		Index:     kvpair.ModifyIndex,
		PreKvpair: nil,
		Kvpair:    kvpair,
		consuler:  c,
		exit:      make(chan bool),
	}

	return node, nil
}

func (n *Node) Start() error {
	go n.watch_key()
}

func (n *Node) Stop() error {
	n.lock.Lock()
	if !n.closed {
		close(n.exit)
		n.closed = true
	}
	n.lock.Unlock()
}

func (n *Node) watch_key() error {
	client, err := consulapi.NewClient(consulapi.DefaultConfig())
	if err != nil {
		return err
	}

	engine := client.KV()

	query_opt := &consulapi.QueryOptions{
		Datacenter:        "",
		AlloStale:         false,
		RequireConsistent: false,
	}

	for {
		select {
		case <-n.exit:
			return ErrExited
		default:
		}

		query_opt.WaitIndex = n.Kvpair.ModifyIndex + 1
		pair, meta, err := engine.Get(n.Key, query_opt)
		if err != nil {
			log.Printf("[consul] Get(%s) error: %v", n.Key, err)
		} else if pair == nil { // deleted
			return nil
		} else {
			n.consuler.putResponse(&Response{
				Action:    ActionCreate,
				PreKvpair: n.PreKvpair,
				Kvpair:    n.Kvpair,
			})
		}
	}
}

func (n *Node) parseAction(newKvp *consulapi.KVPair) Action {
	if n.Kvpair.CreateIndex == n.Kvpair.ModifyIndex {
		return ActionUpdate
	}
}
