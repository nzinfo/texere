package e2e

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coreseekdev/texere/pkg/ot"
	"github.com/coreseekdev/texere/pkg/transport"
)

// ========== Simulated Client ==========

// SimulatedClient represents a client in a collaborative editing session.
type SimulatedClient struct {
	id        string
	docID     string
	transport transport.Transport
	doc       ot.Document
	ctx       context.Context // Add context

	mu         sync.RWMutex
	connected  atomic.Bool
	content    string
	receivedOps []*transport.Message
	sentOps    []*transport.Message

	// Event channels
	opCh       chan *transport.Message
	closeCh    chan struct{}
}

// NewSimulatedClient creates a new simulated client.
func NewSimulatedClient(id, docID string) *SimulatedClient {
	return &SimulatedClient{
		id:          id,
		docID:       docID,
		ctx:         context.Background(),
		receivedOps: make([]*transport.Message, 0),
		sentOps:     make([]*transport.Message, 0),
		opCh:        make(chan *transport.Message, 100),
		closeCh:     make(chan struct{}),
		content:     "",
	}
}

// Connect connects the client to a server.
func (c *SimulatedClient) Connect(ctx context.Context, endpoint, transportType string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Store context for later use
	c.ctx = ctx

	switch transportType {
	case "websocket":
		c.transport = transport.NewWebSocketTransport(c.id, c.id, c.docID)
		wsTransport := c.transport.(*transport.WebSocketTransport)
		wsTransport.SetEndpoint(endpoint)
	case "sse":
		c.transport = transport.NewSSETransport(c.id, c.id, c.docID)
		sseTransport := c.transport.(*transport.SSETransport)
		sseTransport.SetEndpoint(endpoint)
	default:
		return fmt.Errorf("unsupported transport type: %s", transportType)
	}

	// Connect to server
	if err := c.transport.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.connected.Store(true)

	// Start message receiver
	go c.receiveLoop()

	return nil
}

// Close closes the client connection.
func (c *SimulatedClient) Close() error {
	c.connected.Store(false)
	close(c.closeCh)

	if c.transport != nil {
		return c.transport.Close()
	}
	return nil
}

// receiveLoop receives messages from the transport.
func (c *SimulatedClient) receiveLoop() {
	msgCh := c.transport.Receive()
	for {
		select {
		case <-c.closeCh:
			return
		case msg, ok := <-msgCh:
			if !ok {
				return
			}
			c.handleIncomingMessage(msg)
		}
	}
}

// handleIncomingMessage handles an incoming message.
func (c *SimulatedClient) handleIncomingMessage(msg *transport.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.receivedOps = append(c.receivedOps, msg)

	// Apply operation to local document
	if msg.Type == transport.LegacyMsgOperation {
		op := msg.Operation
		if op != nil {
			newContent, err := op.Apply(c.content)
			if err == nil {
				c.content = newContent
			}
		}
	}
}

// RunOperations runs a series of operations.
func (c *SimulatedClient) RunOperations(operations []*ClientOperation) error {
	for i, opSpec := range operations {
		// Add delay if specified
		if opSpec.Delay > 0 {
			time.Sleep(opSpec.Delay)
		}

		// Apply operation locally
		if err := c.applyOperation(opSpec); err != nil {
			return fmt.Errorf("operation %d failed: %w", i, err)
		}

		// Send operation to server
		if err := c.sendOperation(opSpec); err != nil {
			return fmt.Errorf("failed to send operation %d: %w", i, err)
		}
	}

	return nil
}

// applyOperation applies an operation locally.
func (c *SimulatedClient) applyOperation(opSpec *ClientOperation) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var op *ot.Operation

	switch opSpec.Type {
	case OpInsert:
		builder := ot.NewBuilder()
		builder.Retain(opSpec.Position)
		builder.Insert(opSpec.Content)
		op = builder.Build()

	case OpDelete:
		builder := ot.NewBuilder()
		builder.Retain(opSpec.Position)
		builder.Delete(opSpec.Length)
		op = builder.Build()

	case OpRetain:
		builder := ot.NewBuilder()
		builder.Retain(opSpec.Position)
		op = builder.Build()

	default:
		return fmt.Errorf("unknown operation type: %d", opSpec.Type)
	}

	// Apply operation
	newContent, err := op.Apply(c.content)
	if err != nil {
		return err
	}

	c.content = newContent
	return nil
}

// sendOperation sends an operation to the server.
func (c *SimulatedClient) sendOperation(opSpec *ClientOperation) error {
	if !c.connected.Load() {
		return fmt.Errorf("client not connected")
	}

	// Create operation
	var op *ot.Operation
	switch opSpec.Type {
	case OpInsert:
		builder := ot.NewBuilder()
		builder.Retain(opSpec.Position)
		builder.Insert(opSpec.Content)
		op = builder.Build()
	case OpDelete:
		builder := ot.NewBuilder()
		builder.Retain(opSpec.Position)
		builder.Delete(opSpec.Length)
		op = builder.Build()
	case OpRetain:
		builder := ot.NewBuilder()
		builder.Retain(opSpec.Position)
		op = builder.Build()
	}

	// Send message
	msg := &transport.Message{
		Type:      transport.LegacyMsgOperation,
		ClientID:  c.id,
		DocID:     c.docID,
		Timestamp: time.Now().Unix(),
		Operation: op,
	}

	if err := c.transport.Send(c.ctx, msg); err != nil {
		return err
	}

	c.mu.Lock()
	c.sentOps = append(c.sentOps, msg)
	c.mu.Unlock()

	return nil
}

// GetContent returns the current document content.
func (c *SimulatedClient) GetContent() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.content
}

// SetContent sets the document content.
func (c *SimulatedClient) SetContent(content string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.content = content
}

// GetID returns the client ID.
func (c *SimulatedClient) GetID() string {
	return c.id
}

// IsConnected returns true if the client is connected.
func (c *SimulatedClient) IsConnected() bool {
	return c.connected.Load()
}

// GetSentCount returns the number of sent messages.
func (c *SimulatedClient) GetSentCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.sentOps)
}

// GetReceivedCount returns the number of received messages.
func (c *SimulatedClient) GetReceivedCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.receivedOps)
}

// WaitForContent waits for the content to match the expected value.
func (c *SimulatedClient) WaitForContent(expected string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if c.GetContent() == expected {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}

	return false
}

// WaitForOperation waits for an operation to be received.
func (c *SimulatedClient) WaitForOperation(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	initialCount := c.GetReceivedCount()

	for time.Now().Before(deadline) {
		if c.GetReceivedCount() > initialCount {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}

	return false
}

// ========== Client Pool ==========

// ClientPool manages a pool of simulated clients.
type ClientPool struct {
	mu      sync.RWMutex
	clients map[string]*SimulatedClient
}

// NewClientPool creates a new client pool.
func NewClientPool() *ClientPool {
	return &ClientPool{
		clients: make(map[string]*SimulatedClient),
	}
}

// Add adds a client to the pool.
func (p *ClientPool) Add(client *SimulatedClient) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.clients[client.GetID()] = client
}

// Remove removes a client from the pool.
func (p *ClientPool) Remove(clientID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if client, ok := p.clients[clientID]; ok {
		client.Close()
		delete(p.clients, clientID)
	}
}

// Get retrieves a client by ID.
func (p *ClientPool) Get(clientID string) (*SimulatedClient, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	client, ok := p.clients[clientID]
	return client, ok
}

// GetAll returns all clients in the pool.
func (p *ClientPool) GetAll() []*SimulatedClient {
	p.mu.RLock()
	defer p.mu.RUnlock()

	clients := make([]*SimulatedClient, 0, len(p.clients))
	for _, client := range p.clients {
		clients = append(clients, client)
	}
	return clients
}

// Close closes all clients in the pool.
func (p *ClientPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, client := range p.clients {
		client.Close()
	}
	p.clients = make(map[string]*SimulatedClient)
}

// Broadcast sends an operation from all clients.
func (p *ClientPool) Broadcast(op *ClientOperation) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var wg sync.WaitGroup
	errCh := make(chan error, len(p.clients))

	for _, client := range p.clients {
		wg.Add(1)
		go func(c *SimulatedClient) {
			defer wg.Done()
			if err := c.sendOperation(op); err != nil {
				errCh <- err
			}
		}(client)
	}

	wg.Wait()
	close(errCh)

	// Return first error if any
	for err := range errCh {
		return err
	}

	return nil
}
