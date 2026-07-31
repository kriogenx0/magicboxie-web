// Package events implements a tiny in-process pub/sub hub used to push
// live job-progress updates to connected SSE clients. The frontend still
// does a normal REST fetch for full state; this hub only layers live deltas
// on top, so a dropped connection is self-healing on reconnect/refetch.
package events

import "sync"

type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Hub struct {
	mu      sync.Mutex
	clients map[chan Event]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[chan Event]struct{})}
}

func (h *Hub) Subscribe() chan Event {
	ch := make(chan Event, 16)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(ch chan Event) {
	h.mu.Lock()
	if _, ok := h.clients[ch]; ok {
		delete(h.clients, ch)
		close(ch)
	}
	h.mu.Unlock()
}

// Broadcast sends evt to all subscribers. Slow/backed-up clients have the
// event dropped rather than blocking the sender -- clients always have a
// consistent REST fetch as the source of truth.
func (h *Hub) Broadcast(evt Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients {
		select {
		case ch <- evt:
		default:
		}
	}
}
