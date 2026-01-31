package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// MultiDocWebSocketTransport handles multiple documents over a single WebSocket connection.
// This is more efficient than creating separate connections for each document.
//
// Example usage:
//   transport := NewMultiDocWebSocketTransport("client-1", "ws://localhost:8080/ws")
//   transport.Connect(ctx)
//
//   // Subscribe to documents
//   transport.Subscribe("/doc1.txt", doc1Handler)
//   transport.Subscribe("/doc2.txt", doc2Handler)
//
//   // Send operation for specific document
//   transport.SendOperation("/doc1.txt", operation)
type MultiDocWebSocketTransport struct {
	id       string
	clientID string
	endpoint string

	mu     sync.RWMutex
	conn   *websocket.Conn
	closed bool

	// Document subscriptions
	// docPath -> DocumentSubscription
	documents map[string]*DocumentSubscription

	// Message channels
	recvCh    chan *Message
	closeCh   chan struct{}
	connected bool

	// Protocol handler (optional, for automatic message handling)
	handler TransportMessageHandler
}

// DocumentSubscription represents a subscription to a document.
type DocumentSubscription struct {
	DocPath  string
	SessionID string
	ReadOnly bool

	// Message handlers
	onOperation   func(*RemoteOperationData)
	onRemoteOp    func(*RemoteOperationData)
	onSnapshot    func(*SnapshotData)
	onUserJoined  func(*UserJoinedData)
	onUserLeft    func(*UserLeftData)
	onSessionInfo func(*SessionInfoData)
	onError       func(*ErrorData)

	// Channels for document-specific messages
	opCh    chan *Message
	snapshotCh chan *Message
	eventCh  chan *Message
}

// TransportMessageHandler handles messages for a transport.
type TransportMessageHandler interface {
	// HandleMessage is called when a message is received
	HandleMessage(msg *Message) error
}

// NewMultiDocWebSocketTransport creates a new multi-document WebSocket transport.
func NewMultiDocWebSocketTransport(clientID, endpoint string) *MultiDocWebSocketTransport {
	return &MultiDocWebSocketTransport{
		id:        fmt.Sprintf("multidoc-%s", clientID),
		clientID:  clientID,
		endpoint:  endpoint,
		documents: make(map[string]*DocumentSubscription),
		recvCh:    make(chan *Message, 1000),
		closeCh:   make(chan struct{}),
	}
}

// Connect establishes a single WebSocket connection for all documents.
func (t *MultiDocWebSocketTransport) Connect(ctx context.Context) error {
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
		return fmt.Errorf("failed to connect: %w", err)
	}

	t.conn = conn
	t.connected = true

	// Start message processing
	go t.receiveLoop(ctx)
	go t.dispatchLoop()

	return nil
}

// Subscribe subscribes to a document and sets up message handlers.
// Returns the DocumentSubscription for further configuration.
func (t *MultiDocWebSocketTransport) Subscribe(docPath string) (*DocumentSubscription, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil, fmt.Errorf("transport is closed")
	}

	// Check if already subscribed
	if _, exists := t.documents[docPath]; exists {
		return t.documents[docPath], nil
	}

	// Create subscription
	sub := &DocumentSubscription{
		DocPath:  docPath,
		ReadOnly: false,
		opCh:     make(chan *Message, 100),
		snapshotCh: make(chan *Message, 10),
		eventCh:   make(chan *Message, 100),
	}

	t.documents[docPath] = sub

	// Send subscribe message
	if t.conn != nil && t.connected {
		subscribeData := &SubscribeData{
			FilePath: docPath,
			ReadOnly: false,
		}

		protocolMsg, err := NewProtocolMessage(MessageTypeSubscribe, "", subscribeData)
		if err != nil {
			return nil, fmt.Errorf("failed to create subscribe message: %w", err)
		}

		msg := &Message{
			Type:      LegacyMsgOperation,
			ClientID:  t.clientID,
			DocID:     docPath,
			Timestamp: protocolMsg.Timestamp,
			Metadata: map[string]interface{}{
				"protocol_message": protocolMsg,
			},
		}

		if err := t.conn.WriteJSON(msg); err != nil {
			return nil, fmt.Errorf("failed to send subscribe: %w", err)
		}

		log.Printf("[MultiDoc] Subscribed to %s", docPath)
	}

	return sub, nil
}

// Unsubscribe unsubscribes from a document.
func (t *MultiDocWebSocketTransport) Unsubscribe(docPath string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	sub, exists := t.documents[docPath]
	if !exists {
		return fmt.Errorf("not subscribed to %s", docPath)
	}

	// Send unsubscribe message
	if t.conn != nil && t.connected {
		unsubscribeData := &UnsubscribeData{
			SessionID: sub.SessionID,
		}

		protocolMsg, err := NewProtocolMessage(MessageTypeUnsubscribe, sub.SessionID, unsubscribeData)
		if err != nil {
			return fmt.Errorf("failed to create unsubscribe message: %w", err)
		}

		msg := &Message{
			Type:      LegacyMsgOperation,
			ClientID:  t.clientID,
			DocID:     docPath,
			Timestamp: protocolMsg.Timestamp,
			Metadata: map[string]interface{}{
				"protocol_message": protocolMsg,
			},
		}

		if err := t.conn.WriteJSON(msg); err != nil {
			return fmt.Errorf("failed to send unsubscribe: %w", err)
		}
	}

	// Close channels
	close(sub.opCh)
	close(sub.snapshotCh)
	close(sub.eventCh)

	// Remove from documents map
	delete(t.documents, docPath)

	log.Printf("[MultiDoc] Unsubscribed from %s", docPath)

	return nil
}

// SendOperation sends an OT operation for a specific document.
func (t *MultiDocWebSocketTransport) SendOperation(docPath string, operation []interface{}) error {
	return t.SendOperationWithContext(context.Background(), docPath, operation)
}

// SendOperationWithContext sends an OT operation with context.
func (t *MultiDocWebSocketTransport) SendOperationWithContext(ctx context.Context, docPath string, operation []interface{}) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.closed || t.conn == nil {
		return ErrTransportClosed
	}

	// Find subscription
	sub, exists := t.documents[docPath]
	if !exists {
		return fmt.Errorf("not subscribed to %s", docPath)
	}

	// Create operation message
	opData := &OperationData{
		SessionID: sub.SessionID,
		Operation: operation,
	}

	protocolMsg, err := NewProtocolMessage(MessageTypeOperation, sub.SessionID, opData)
	if err != nil {
		return fmt.Errorf("failed to create operation message: %w", err)
	}

	msg := &Message{
		Type:      LegacyMsgOperation,
		ClientID:  t.clientID,
		DocID:     docPath,
		Timestamp: protocolMsg.Timestamp,
		Metadata: map[string]interface{}{
			"protocol_message": protocolMsg,
		},
	}

	// Send with timeout
	t.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return t.conn.WriteJSON(msg)
}

// SendHeartbeat sends heartbeat for multiple sessions.
func (t *MultiDocWebSocketTransport) SendHeartbeat(sessionIDs []string) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.closed || t.conn == nil {
		return ErrTransportClosed
	}

	heartbeatData := &HeartbeatData{
		SessionIDs: sessionIDs,
	}

	protocolMsg, err := NewProtocolMessage(MessageTypeHeartbeat, "", heartbeatData)
	if err != nil {
		return fmt.Errorf("failed to create heartbeat message: %w", err)
	}

	msg := &Message{
		Type:      LegacyMsgHello, // Use Hello for heartbeat
		ClientID:  t.clientID,
		Timestamp: protocolMsg.Timestamp,
		Metadata: map[string]interface{}{
			"protocol_message": protocolMsg,
		},
	}

	t.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return t.conn.WriteJSON(msg)
}

// receiveLoop receives messages from WebSocket and dispatches to document channels.
func (t *MultiDocWebSocketTransport) receiveLoop(ctx context.Context) {
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
			log.Printf("[MultiDoc] Read error: %v", err)
			return
		}

		// Route message based on SessionID/DocID
		t.routeMessage(&msg)
	}
}

// routeRoute routes incoming messages to the appropriate document subscription.
func (t *MultiDocWebSocketTransport) routeMessage(msg *Message) {
	// Extract protocol message
	protocolMsgBytes, ok := msg.Metadata["protocol_message"]
	if !ok {
		return
	}

	var protocolMsg ProtocolMessage
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", protocolMsgBytes)), &protocolMsg); err != nil {
		log.Printf("[MultiDoc] Failed to parse protocol message: %v", err)
		return
	}

	// Find the subscription for this session
	t.mu.RLock()
	defer t.mu.RUnlock()

	var sub *DocumentSubscription
	for _, s := range t.documents {
		if s.SessionID == protocolMsg.SessionID {
			sub = s
			break
		}
	}

	if sub == nil {
		// No subscription found, might be for an unknown document
		log.Printf("[MultiDoc] No subscription found for session %s", protocolMsg.SessionID)
		return
	}

	// Route to appropriate channel based on message type
	switch protocolMsg.Type {
	case MessageTypeSnapshot:
		select {
		case sub.snapshotCh <- msg:
		case <-time.After(5 * time.Second):
			log.Printf("[MultiDoc] Snapshot channel full for %s", sub.DocPath)
		}

	case MessageTypeRemoteOperation:
		select {
		case sub.opCh <- msg:
		case <-time.After(5 * time.Second):
			log.Printf("[MultiDoc] Operation channel full for %s", sub.DocPath)
		}

	case MessageTypeUserJoined, MessageTypeUserLeft, MessageTypeSessionInfo, MessageTypeError:
		select {
		case sub.eventCh <- msg:
		case <-time.After(5 * time.Second):
			log.Printf("[MultiDoc] Event channel full for %s", sub.DocPath)
		}

	default:
		log.Printf("[MultiDoc] Unhandled message type: %s", protocolMsg.Type)
	}
}

// dispatchLoop dispatches messages from document channels to handlers.
func (t *MultiDocWebSocketTransport) dispatchLoop() {
	for {
		select {
		case <-t.closeCh:
			return
		case msg := <-t.recvCh:
			// Global message handler
			if t.handler != nil {
				t.handler.HandleMessage(msg)
			}
		}
	}
}

// GetSubscription returns the subscription for a document.
func (t *MultiDocWebSocketTransport) GetSubscription(docPath string) (*DocumentSubscription, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	sub, exists := t.documents[docPath]
	return sub, exists
}

// ListSubscriptions returns all subscribed documents.
func (t *MultiDocWebSocketTransport) ListSubscriptions() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	docs := make([]string, 0, len(t.documents))
	for docPath := range t.documents {
		docs = append(docs, docPath)
	}
	return docs
}

// IsConnected returns whether the WebSocket is connected.
func (t *MultiDocWebSocketTransport) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.connected
}

// Close closes the transport and all subscriptions.
func (t *MultiDocWebSocketTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	close(t.closeCh)

	// Close all document subscriptions
	for docPath, sub := range t.documents {
		log.Printf("[MultiDoc] Closing subscription to %s", docPath)
		close(sub.opCh)
		close(sub.snapshotCh)
		close(sub.eventCh)
	}

	t.documents = make(map[string]*DocumentSubscription)

	if t.conn != nil {
		return t.conn.Close()
	}

	return nil
}

// SetMessageHandler sets a global message handler.
func (t *MultiDocWebSocketTransport) SetMessageHandler(handler TransportMessageHandler) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.handler = handler
}

// ========== DocumentSubscription Methods ==========

// OnOperation sets a handler for operations from the current client.
func (s *DocumentSubscription) OnOperation(handler func(*RemoteOperationData)) {
	s.onOperation = handler
}

// OnRemoteOperation sets a handler for operations from other clients.
func (s *DocumentSubscription) OnRemoteOperation(handler func(*RemoteOperationData)) {
	s.onRemoteOp = handler
}

// OnSnapshot sets a handler for snapshot updates.
func (s *DocumentSubscription) OnSnapshot(handler func(*SnapshotData)) {
	s.onSnapshot = handler
}

// OnUserJoined sets a handler for user joined events.
func (s *DocumentSubscription) OnUserJoined(handler func(*UserJoinedData)) {
	s.onUserJoined = handler
}

// OnUserLeft sets a handler for user left events.
func (s *DocumentSubscription) OnUserLeft(handler func(*UserLeftData)) {
	s.onUserLeft = handler
}

// OnSessionInfo sets a handler for session info updates.
func (s *DocumentSubscription) OnSessionInfo(handler func(*SessionInfoData)) {
	s.onSessionInfo = handler
}

// OnError sets a handler for errors.
func (s *DocumentSubscription) OnError(handler func(*ErrorData)) {
	s.onError = handler
}

// StartMessageHandler starts a goroutine to handle messages for this subscription.
func (s *DocumentSubscription) StartMessageHandler(transport *MultiDocWebSocketTransport) {
	go func() {
		for {
			select {
			case <-transport.closeCh:
				return
			case msg := <-s.snapshotCh:
				s.handleSnapshot(msg)
			case msg := <-s.opCh:
				s.handleOperation(msg)
			case msg := <-s.eventCh:
				s.handleEvent(msg)
			}
		}
	}()
}

// handleSnapshot handles snapshot messages.
func (s *DocumentSubscription) handleSnapshot(msg *Message) {
	if s.onSnapshot == nil {
		return
	}

	protocolMsgBytes, ok := msg.Metadata["protocol_message"]
	if !ok {
		return
	}

	var protocolMsg ProtocolMessage
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", protocolMsgBytes)), &protocolMsg); err != nil {
		log.Printf("Failed to parse snapshot message: %v", err)
		return
	}

	var data SnapshotData
	if err := json.Unmarshal(protocolMsg.Data, &data); err != nil {
		log.Printf("Failed to parse snapshot data: %v", err)
		return
	}

	s.onSnapshot(&data)
}

// handleOperation handles operation messages.
func (s *DocumentSubscription) handleOperation(msg *Message) {
	protocolMsgBytes, ok := msg.Metadata["protocol_message"]
	if !ok {
		return
	}

	var protocolMsg ProtocolMessage
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", protocolMsgBytes)), &protocolMsg); err != nil {
		log.Printf("Failed to parse operation message: %v", err)
		return
	}

	var data RemoteOperationData
	if err := json.Unmarshal(protocolMsg.Data, &data); err != nil {
		log.Printf("Failed to parse operation data: %v", err)
		return
	}

	// Call appropriate handler
	if s.onRemoteOp != nil {
		s.onRemoteOp(&data)
	}
}

// handleEvent handles event messages (user joined/left, session info, errors).
func (s *DocumentSubscription) handleEvent(msg *Message) {
	protocolMsgBytes, ok := msg.Metadata["protocol_message"]
	if !ok {
		return
	}

	var protocolMsg ProtocolMessage
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", protocolMsgBytes)), &protocolMsg); err != nil {
		log.Printf("Failed to parse event message: %v", err)
		return
	}

	switch protocolMsg.Type {
	case MessageTypeUserJoined:
		if s.onUserJoined != nil {
			var data UserJoinedData
			if err := json.Unmarshal(protocolMsg.Data, &data); err == nil {
				s.onUserJoined(&data)
			}
		}

	case MessageTypeUserLeft:
		if s.onUserLeft != nil {
			var data UserLeftData
			if err := json.Unmarshal(protocolMsg.Data, &data); err == nil {
				s.onUserLeft(&data)
			}
		}

	case MessageTypeSessionInfo:
		if s.onSessionInfo != nil {
			var data SessionInfoData
			if err := json.Unmarshal(protocolMsg.Data, &data); err == nil {
				s.onSessionInfo(&data)
			}
		}

	case MessageTypeError:
		if s.onError != nil {
			var data ErrorData
			if err := json.Unmarshal(protocolMsg.Data, &data); err == nil {
				s.onError(&data)
			}
		}
	}
}
