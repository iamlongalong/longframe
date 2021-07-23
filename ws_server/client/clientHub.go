package client

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

var Hub *ClientHub

var ErrClientExist = errors.New("client already exist")

func init() {
	Hub = &ClientHub{
		mu:      &sync.RWMutex{},
		clients: make(map[string]*Client, 100),
	}
}

type ClientHub struct {
	mu      sync.Locker
	clients map[string]*Client
}

func (h *ClientHub) Register(c *Client) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[c.ID]; ok {
		return errors.WithMessage(ErrClientExist, fmt.Sprintf("clientID : %s", c.ID))
	}

	h.clients[c.ID] = c
	c.hub = h

	return nil
}

func (h *ClientHub) UnRegister(clientID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[clientID] = nil
}

func (h *ClientHub) Reset(c *Client) {
	h.UnRegister(c.ID)

	h.Register(c)
}
