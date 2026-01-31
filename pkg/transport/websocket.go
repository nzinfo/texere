package transport

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for testing
	},
}

// WebSocketTransport implements Transport using WebSocket.
type WebSocketTransport struct {
	*BaseTransport
	conn     *websocket.Conn
	client   *http.Client
	mu       sync.Mutex
	closed   bool
	endpoint string
}

// NewWebSocketTransport creates a new WebSocket transport.
func NewWebSocketTransport(id, clientID, docID string) *WebSocketTransport {
	base := NewBaseTransport(id, clientID, docID)
	return &WebSocketTransport{
		BaseTransport: base,
		client:        &http.Client{},
	}
}

// Connect establishes WebSocket connection (client mode).
func (t *WebSocketTransport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	if t.endpoint == "" {
		return fmt.Errorf("no endpoint set")
	}

	// Connect to WebSocket server
	dialer := websocket.Dialer{}
	conn, _, err := dialer.DialContext(ctx, t.endpoint, nil)
	if err != nil {
		return err
	}

	t.conn = conn
	t.connected = true

	// Start receiving messages
	go t.receiveLoop(ctx)

	return nil
}

// SetEndpoint sets the WebSocket endpoint URL.
func (t *WebSocketTransport) SetEndpoint(endpoint string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.endpoint = endpoint
}

// Send sends a message via WebSocket (client mode).
func (t *WebSocketTransport) Send(ctx context.Context, msg *Message) error {
	err := t.BaseTransport.Send(ctx, msg)
	if err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed || t.conn == nil {
		return ErrTransportClosed
	}

	// Send message as JSON
	return t.conn.WriteJSON(msg)
}

// Close closes the WebSocket transport.
func (t *WebSocketTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	t.connected = false

	if t.conn != nil {
		return t.conn.Close()
	}

	return t.BaseTransport.Close()
}

// receiveLoop receives WebSocket messages (client mode).
func (t *WebSocketTransport) receiveLoop(ctx context.Context) {
	defer func() {
		t.mu.Lock()
		if t.conn != nil {
			t.conn.Close()
			t.conn = nil
		}
		t.connected = false
		t.mu.Unlock()
	}()

	for {
		t.mu.Lock()
		if t.closed {
			t.mu.Unlock()
			return
		}
		conn := t.conn
		if conn == nil {
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

		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			return
		}

		select {
		case t.recvCh <- &msg:
		case <-t.closeCh:
			return
		}
	}
}

// ========== WebSocket Server ==========

// WebSocketServer handles WebSocket connections.
type WebSocketServer struct {
	addr       string
	mu         sync.RWMutex
	clients    map[string]*WebSocketConn
	closeCh    chan struct{}
	server     *http.Server
	handler    func(*Message)
	rawHandler func(clientID string, message []byte)
}

// WebSocketConn represents a WebSocket client connection.
type WebSocketConn struct {
	id   string
	conn *websocket.Conn
	send chan *Message
	hub  *WebSocketServer
}

// NewWebSocketServer creates a new WebSocket server.
func NewWebSocketServer(addr string) *WebSocketServer {
	return &WebSocketServer{
		addr:    addr,
		clients: make(map[string]*WebSocketConn),
		closeCh: make(chan struct{}),
	}
}

// SetMessageHandler sets the message handler for incoming messages.
func (s *WebSocketServer) SetMessageHandler(handler func(*Message)) {
	s.handler = handler
}

// SetRawMessageHandler sets the raw message handler for incoming messages.
func (s *WebSocketServer) SetRawMessageHandler(handler func(clientID string, message []byte)) {
	s.rawHandler = handler
}

// RegisterHandler registers the WebSocket handler with the given mux.
func (s *WebSocketServer) RegisterHandler(mux *http.ServeMux) {
	mux.HandleFunc("/ws", s.handleWebSocket)
}

// Start starts the WebSocket server.
func (s *WebSocketServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	go func() {
		s.server.ListenAndServe()
	}()

	return nil
}

// handleWebSocket handles WebSocket connections.
func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("[WebSocket] Incoming connection from %s", r.RemoteAddr)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] Upgrade failed: %v", err)
		return
	}

	log.Printf("[WebSocket] Connection established for %s", r.URL.String())

	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		clientID = fmt.Sprintf("client-%d", time.Now().UnixNano())
	}

	wsConn := &WebSocketConn{
		id:   clientID,
		conn: conn,
		send: make(chan *Message, 256),
		hub:  s,
	}

	s.mu.Lock()
	s.clients[clientID] = wsConn
	s.mu.Unlock()

	// Start reading from connection
	go wsConn.readPump()
	go wsConn.writePump()
}

// readPump pumps messages from the WebSocket connection to the hub.
func (c *WebSocketConn) readPump() {
	defer func() {
		log.Printf("[WebSocket] %s: readPump closing", c.id)
		c.conn.Close()
		c.hub.mu.Lock()
		delete(c.hub.clients, c.id)
		c.hub.mu.Unlock()
		close(c.send)
	}()

	for {
		// Read raw message
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("[WebSocket] %s: Read error: %v", c.id, err)
			break
		}

		log.Printf("[WebSocket] %s: Received raw message: %s", c.id, string(messageBytes))

		// Call message handler if set
		if c.hub.rawHandler != nil {
			c.hub.rawHandler(c.id, messageBytes)
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection.
func (c *WebSocketConn) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		log.Printf("[WebSocket] %s: writePump closing", c.id)
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Check if message contains raw JSON
			if rawJSON, ok := msg.Metadata["raw_json"].(string); ok {
				log.Printf("[WebSocket] %s: Sending raw JSON", c.id)
				c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := c.conn.WriteMessage(websocket.TextMessage, []byte(rawJSON)); err != nil {
					log.Printf("[WebSocket] %s: Write error: %v", c.id, err)
					return
				}
			} else {
				log.Printf("[WebSocket] %s: Sending message type=%s", c.id, msg.Type)
				c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				err := c.conn.WriteJSON(msg)
				if err != nil {
					log.Printf("[WebSocket] %s: Write error: %v", c.id, err)
					return
				}
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[WebSocket] %s: Ping error: %v", c.id, err)
				return
			}
		case <-c.hub.closeCh:
			return
		}
	}
}

// Broadcast sends a message to all connected clients.
func (s *WebSocketServer) Broadcast(msg *Message) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		select {
		case client.send <- msg:
		case <-s.closeCh:
			return
		}
	}
}

// Send sends a message to a specific client.
func (s *WebSocketServer) Send(clientID string, msg *Message) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[clientID]
	if !ok {
		return fmt.Errorf("client not found: %s", clientID)
	}

	select {
	case client.send <- msg:
		return nil
	case <-s.closeCh:
		return ErrTransportClosed
	}
}

// SendJSON sends raw JSON data to a specific client.
func (s *WebSocketServer) SendJSON(clientID string, data []byte) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[clientID]
	if !ok {
		return fmt.Errorf("client not found: %s", clientID)
	}

	// Create a message wrapper for JSON data
	msg := &Message{
		Type: LegacyMsgOperation, // Placeholder
		Metadata: map[string]interface{}{
			"raw_json": string(data),
		},
	}

	select {
	case client.send <- msg:
		return nil
	case <-s.closeCh:
		return ErrTransportClosed
	}
}

// Close closes the WebSocket server.
func (s *WebSocketServer) Close() error {
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
		close(client.send)
		client.conn.Close()
	}

	s.clients = make(map[string]*WebSocketConn)
	return nil
}
