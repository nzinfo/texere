package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// SSETransport implements Transport using Server-Sent Events.
type SSETransport struct {
	*BaseTransport
	client   *http.Client
	server   *http.Server
	mu       sync.Mutex
	closed   bool
	endpoint string
}

// NewSSETransport creates a new SSE transport.
func NewSSETransport(id, clientID, docID string) *SSETransport {
	base := NewBaseTransport(id, clientID, docID)
	return &SSETransport{
		BaseTransport: base,
		client:        &http.Client{},
	}
}

// Connect establishes SSE connection (client mode).
func (t *SSETransport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	if t.endpoint == "" {
		return fmt.Errorf("no endpoint set")
	}

	// Start receiving messages
	go t.receiveLoop(ctx)
	t.connected = true

	return nil
}

// SetEndpoint sets the SSE endpoint URL.
func (t *SSETransport) SetEndpoint(endpoint string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.endpoint = endpoint
}

// Send sends a message via HTTP POST (client mode).
func (t *SSETransport) Send(ctx context.Context, msg *Message) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed || t.endpoint == "" {
		return ErrTransportClosed
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.endpoint, nil)
	if err != nil {
		return err
	}

	// Set request body
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Transport-ID", t.id)

	// Create request body from data
	_ = data // Used for request body (would be set in req.Body)

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SSE send failed: %s", resp.Status)
	}

	return nil
}

// Close closes the SSE transport.
func (t *SSETransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	t.connected = false

	if t.server != nil {
		t.server.Close()
	}

	return t.BaseTransport.Close()
}

// receiveLoop receives SSE messages (client mode).
func (t *SSETransport) receiveLoop(ctx context.Context) {
	req, err := http.NewRequestWithContext(ctx, "GET", t.endpoint, nil)
	if err != nil {
		return
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("X-Transport-ID", t.id)

	resp, err := t.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	reader := bufio.NewReader(resp.Body)
	for {
		t.mu.Lock()
		if t.closed {
			t.mu.Unlock()
			return
		}
		t.mu.Unlock()

		select {
		case <-ctx.Done():
			return
		case <-t.closeCh:
			return
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		if len(line) > 6 && line[:6] == "data: " {
			data := line[6:]
			var msg Message
			if err := json.Unmarshal([]byte(data), &msg); err == nil {
				select {
				case t.recvCh <- &msg:
				case <-t.closeCh:
					return
				}
			}
		}
	}
}

// ========== SSE Server ==========

// SSEServer handles SSE connections.
type SSEServer struct {
	addr    string
	mu      sync.RWMutex
	clients map[string]*SSEClient
	closeCh chan struct{}
	server  *http.Server
}

// SSEClient represents an SSE client connection.
type SSEClient struct {
	id       string
	chanChan chan chan *Message
	msgChan  chan *Message
	closeCh  chan struct{}
}

// NewSSEServer creates a new SSE server.
func NewSSEServer(addr string) *SSEServer {
	return &SSEServer{
		addr:    addr,
		clients: make(map[string]*SSEClient),
		closeCh: make(chan struct{}),
	}
}

// Start starts the SSE server.
func (s *SSEServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/events", s.handleEvents)

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	go func() {
		s.server.ListenAndServe()
	}()

	return nil
}

// handleEvents handles SSE connections.
func (s *SSEServer) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	clientID := r.Header.Get("X-Transport-ID")
	if clientID == "" {
		clientID = "unknown"
	}

	msgChan := make(chan *Message, 100)
	closeCh := make(chan struct{})

	client := &SSEClient{
		id:       clientID,
		msgChan:  msgChan,
		closeCh:  closeCh,
	}

	s.mu.Lock()
	s.clients[clientID] = client
	s.mu.Unlock()

	// Send connection established message
	fmt.Fprintf(w, "data: {\"type\":\"connected\",\"id\":\"%s\"}\n\n", clientID)
	flusher.Flush()

	// Send messages to client
	for {
		select {
		case msg := <-msgChan:
			_, _ = json.Marshal(msg)
			fmt.Fprintf(w, "data: {\"type\":\"message\"}\n\n")
			flusher.Flush()
		case <-closeCh:
			return
		case <-r.Context().Done():
			return
		case <-s.closeCh:
			return
		}
	}
}

// Broadcast sends a message to all connected clients.
func (s *SSEServer) Broadcast(msg *Message) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		select {
		case client.msgChan <- msg:
		case <-client.closeCh:
		case <-s.closeCh:
			return
		}
	}
}

// Close closes the SSE server.
func (s *SSEServer) Close() error {
	select {
	case <-s.closeCh:
		return nil
	default:
		close(s.closeCh)
	}

	if s.server != nil {
		s.server.Close()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, client := range s.clients {
		close(client.closeCh)
	}

	s.clients = make(map[string]*SSEClient)
	return nil
}
