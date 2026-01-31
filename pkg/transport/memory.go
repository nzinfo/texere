package transport

import (
	"context"
	"fmt"
	"sync"
)

// MemoryTransport is an in-memory transport for testing.
type MemoryTransport struct {
	*BaseTransport
	mu     sync.RWMutex
	peers  map[string]*MemoryTransport
	closed bool
}

// NewMemoryTransport creates a new in-memory transport.
func NewMemoryTransport(id, clientID, docID string) *MemoryTransport {
	base := NewBaseTransport(id, clientID, docID)
	return &MemoryTransport{
		BaseTransport: base,
		peers:         make(map[string]*MemoryTransport),
	}
}

// Connect establishes the connection (no-op for memory transport).
func (t *MemoryTransport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	t.connected = true
	return nil
}

// ConnectTo connects this transport to another transport.
func (t *MemoryTransport) ConnectTo(other *MemoryTransport) {
	t.mu.Lock()
	other.mu.Lock()
	defer t.mu.Unlock()
	defer other.mu.Unlock()

	t.peers[other.ID()] = other
	other.peers[t.ID()] = t
}

// Send sends a message to connected peers.
func (t *MemoryTransport) Send(ctx context.Context, msg *Message) error {
	if err := t.BaseTransport.Send(ctx, msg); err != nil {
		return err
	}

	// Broadcast to all peers
	t.mu.RLock()
	peers := make(map[string]*MemoryTransport, len(t.peers))
	for id, peer := range t.peers {
		peers[id] = peer
	}
	t.mu.RUnlock()

	for _, peer := range peers {
		peer.mu.RLock()
		recvCh := peer.recvCh
		peer.mu.RUnlock()

		select {
		case recvCh <- msg:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// Close closes the transport.
func (t *MemoryTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	t.BaseTransport.Close()

	// Disconnect from all peers
	for _, peer := range t.peers {
		peer.mu.Lock()
		delete(peer.peers, t.ID())
		peer.mu.Unlock()
	}

	t.peers = make(map[string]*MemoryTransport)
	return nil
}

// Start starts the message processing loop.
func (t *MemoryTransport) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-t.closeCh:
				return
			case <-ctx.Done():
				return
			case msg := <-t.sendCh:
				t.mu.RLock()
				peers := make(map[string]*MemoryTransport, len(t.peers))
				for id, peer := range t.peers {
					peers[id] = peer
				}
				t.mu.RUnlock()

				for _, peer := range peers {
					select {
					case peer.recvCh <- msg:
					case <-t.closeCh:
						return
					}
				}
			}
		}
	}()
}
