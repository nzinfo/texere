package transport

import (
	"context"
	"time"

	"github.com/coreseekdev/texere/pkg/ot"
)

// LegacyMessageType represents the type of legacy transport message.
type LegacyMessageType int

const (
	LegacyMsgOperation LegacyMessageType = iota
	LegacyMsgSync
	LegacyMsgSyncAck
	LegacyMsgAck
	LegacyMsgError
	LegacyMsgHello
	LegacyMsgWelcome
)

// Message represents a message sent over the transport layer.
type Message struct {
	Type      LegacyMessageType
	DocID     string
	ClientID  string
	Timestamp int64
	Operation *ot.Operation
	Content   string // For sync messages
	Version   int64  // Document version
	SeqNum    int64  // Sequence number
	Error     string
	Metadata  map[string]interface{}
}

// NewOperationMessage creates a new operation message.
func NewOperationMessage(docID, clientID string, op *ot.Operation) *Message {
	return &Message{
		Type:      LegacyMsgOperation,
		DocID:     docID,
		ClientID:  clientID,
		Timestamp: time.Now().Unix(),
		Operation: op,
	}
}

// NewSyncMessage creates a new sync message.
func NewSyncMessage(docID, clientID string, version int64) *Message {
	return &Message{
		Type:      LegacyMsgSync,
		DocID:     docID,
		ClientID:  clientID,
		Timestamp: time.Now().Unix(),
		Version:   version,
	}
}

// NewSyncAckMessage creates a new sync acknowledgment message.
func NewSyncAckMessage(docID string, content string, version int64) *Message {
	return &Message{
		Type:      LegacyMsgSyncAck,
		DocID:     docID,
		Timestamp: time.Now().Unix(),
		Content:   content,
		Version:   version,
	}
}

// NewAckMessage creates a new acknowledgment message.
func NewAckMessage(docID, clientID string) *Message {
	return &Message{
		Type:      LegacyMsgAck,
		DocID:     docID,
		ClientID:  clientID,
		Timestamp: time.Now().Unix(),
	}
}

// NewErrorMessage creates a new error message.
func NewErrorMessage(docID string, err error) *Message {
	return &Message{
		Type:      LegacyMsgError,
		DocID:     docID,
		Timestamp: time.Now().Unix(),
		Error:     err.Error(),
	}
}

// ========== Transport Interface ==========

// Transport represents a bidirectional transport for collaborative editing.
type Transport interface {
	// ID returns the unique identifier for this transport.
	ID() string

	// Send sends a message over the transport.
	Send(ctx context.Context, msg *Message) error

	// Receive returns a channel for receiving messages.
	Receive() <-chan *Message

	// Close closes the transport.
	Close() error

	// Connect establishes the connection.
	Connect(ctx context.Context) error

	// IsConnected returns true if the transport is connected.
	IsConnected() bool
}

// BaseTransport provides common functionality for transport implementations.
type BaseTransport struct {
	id        string
	clientID  string
	docID     string
	sendCh    chan *Message
	recvCh    chan *Message
	closeCh   chan struct{}
	connected bool
}

// NewBaseTransport creates a new base transport.
func NewBaseTransport(id, clientID, docID string) *BaseTransport {
	return &BaseTransport{
		id:        id,
		clientID:  clientID,
		docID:     docID,
		sendCh:    make(chan *Message, 100),
		recvCh:    make(chan *Message, 100),
		closeCh:   make(chan struct{}),
		connected: false,
	}
}

// ID returns the transport ID.
func (t *BaseTransport) ID() string {
	return t.id
}

// Send sends a message over the transport.
func (t *BaseTransport) Send(ctx context.Context, msg *Message) error {
	select {
	case <-t.closeCh:
		return ErrTransportClosed
	case <-ctx.Done():
		return ctx.Err()
	case t.sendCh <- msg:
		return nil
	}
}

// Receive returns a channel for receiving messages.
func (t *BaseTransport) Receive() <-chan *Message {
	return t.recvCh
}

// Close closes the transport.
func (t *BaseTransport) Close() error {
	select {
	case <-t.closeCh:
		return nil
	default:
		close(t.closeCh)
		t.connected = false
		return nil
	}
}

// IsConnected returns true if the transport is connected.
func (t *BaseTransport) IsConnected() bool {
	return t.connected
}

// ========== Error Definitions ==========

var (
	// ErrTransportClosed is returned when the transport is closed.
	ErrTransportClosed = &TransportError{Code: "closed", Message: "transport closed"}

	// ErrSendFailed is returned when sending a message fails.
	ErrSendFailed = &TransportError{Code: "send_failed", Message: "failed to send message"}

	// ErrReceiveFailed is returned when receiving a message fails.
	ErrReceiveFailed = &TransportError{Code: "receive_failed", Message: "failed to receive message"}
)

// TransportError represents a transport-related error.
type TransportError struct {
	Code    string
	Message string
}

func (e *TransportError) Error() string {
	return e.Message
}
