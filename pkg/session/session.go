package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/coreseekdev/texere/pkg/ot"
)

// DocumentType represents the type of document to use.
type DocumentType int

const (
	DocTypeString DocumentType = iota
	DocTypeRope
)

// SessionConfig represents session configuration.
type SessionConfig struct {
	DocID          string
	InitialContent string
	DocType        DocumentType
	EnableUndo     bool
	MaxHistory     int
	Auth           Authenticator
	Content        ContentStorage
}

// SessionEventType represents the type of session event.
type SessionEventType int

const (
	EventOperationApplied SessionEventType = iota
	EventUndo
	EventRedo
	EventContentChanged
	EventSessionClosed
)

// SessionEvent represents a session event.
type SessionEvent struct {
	Type      SessionEventType
	DocID     string
	Timestamp int64
	Data      interface{}
}

// Session represents a collaborative editing session.
type Session interface {
	ID() string
	Type() string
	GetDocument() ot.Document
	SetDocument(doc ot.Document) error
	GetContent() string
	SetContent(content string) error
	ApplyOperation(op *ot.Operation) error
	Undo() error
	Redo() error
	CanUndo() bool
	CanRedo() bool
	Close() error
}

// ========== SimpleSession Implementation ==========

// SimpleSession is a session using UndoableDocument.
type SimpleSession struct {
	id   string
	ctx  context.Context
	cancel context.CancelFunc

	mu     sync.RWMutex
	doc    ot.UndoableDocument
	config *SessionConfig

	// Event publishing
	subscribers map[chan *SessionEvent]bool
	subMu       sync.RWMutex
}

// NewSimpleSession creates a new simple session.
func NewSimpleSession(ctx context.Context, config SessionConfig) (*SimpleSession, error) {
	ctx, cancel := context.WithCancel(ctx)

	var doc ot.UndoableDocument
	var err error

	switch config.DocType {
	case DocTypeString:
		doc = ot.NewStringDocument(config.InitialContent)
	case DocTypeRope:
		// TODO: Use RopeDocument when available
		doc = ot.NewStringDocument(config.InitialContent)
	default:
		doc = ot.NewStringDocument(config.InitialContent)
	}

	if err != nil {
		cancel()
		return nil, err
	}

	return &SimpleSession{
		id:          config.DocID,
		ctx:         ctx,
		cancel:      cancel,
		doc:         doc,
		config:      &config,
		subscribers: make(map[chan *SessionEvent]bool),
	}, nil
}

// ID returns the session ID.
func (s *SimpleSession) ID() string {
	return s.id
}

// Type returns the session type.
func (s *SimpleSession) Type() string {
	return "simple"
}

// GetDocument returns the document.
func (s *SimpleSession) GetDocument() ot.Document {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.doc
}

// SetDocument sets the document.
func (s *SimpleSession) SetDocument(doc ot.Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	undoableDoc, ok := doc.(ot.UndoableDocument)
	if !ok {
		return fmt.Errorf("document must implement UndoableDocument")
	}

	s.doc = undoableDoc
	return nil
}

// GetContent returns the document content.
func (s *SimpleSession) GetContent() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.doc.String()
}

// SetContent sets the document content.
func (s *SimpleSession) SetContent(content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var doc ot.UndoableDocument
	switch s.config.DocType {
	case DocTypeString:
		doc = ot.NewStringDocument(content)
	case DocTypeRope:
		// TODO: Use RopeDocument when available
		doc = ot.NewStringDocument(content)
	default:
		doc = ot.NewStringDocument(content)
	}

	s.doc = doc
	return nil
}

// ApplyOperation applies an operation to the document.
func (s *SimpleSession) ApplyOperation(op *ot.Operation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Apply operation with history recording
	var err error
	s.doc, err = s.doc.ApplyOperationWithHistory(op)
	if err != nil {
		return fmt.Errorf("failed to apply operation: %w", err)
	}

	// Publish event
	s.publishEvent(&SessionEvent{
		Type:      EventOperationApplied,
		DocID:     s.id,
		Timestamp: time.Now().Unix(),
		Data:      op,
	})

	return nil
}

// Undo performs an undo operation using the document's undo capability.
func (s *SimpleSession) Undo() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.doc.CanUndo() {
		return fmt.Errorf("cannot undo")
	}

	err := s.doc.Undo()
	if err != nil {
		return err
	}

	// Publish event
	s.publishEvent(&SessionEvent{
		Type:      EventUndo,
		DocID:     s.id,
		Timestamp: time.Now().Unix(),
	})

	return nil
}

// Redo performs a redo operation using the document's redo capability.
func (s *SimpleSession) Redo() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.doc.CanRedo() {
		return fmt.Errorf("cannot redo")
	}

	err := s.doc.Redo()
	if err != nil {
		return err
	}

	// Publish event
	s.publishEvent(&SessionEvent{
		Type:      EventRedo,
		DocID:     s.id,
		Timestamp: time.Now().Unix(),
	})

	return nil
}

// CanUndo returns true if undo is possible.
func (s *SimpleSession) CanUndo() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.doc.CanUndo()
}

// CanRedo returns true if redo is possible.
func (s *SimpleSession) CanRedo() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.doc.CanRedo()
}

// Subscribe subscribes to session events.
func (s *SimpleSession) Subscribe() <-chan *SessionEvent {
	s.subMu.Lock()
	defer s.subMu.Unlock()

	ch := make(chan *SessionEvent, 100)
	s.subscribers[ch] = true
	return ch
}

// Unsubscribe unsubscribes from session events.
func (s *SimpleSession) Unsubscribe(ch <-chan *SessionEvent) {
	s.subMu.Lock()
	defer s.subMu.Unlock()

	for subCh := range s.subscribers {
		if subCh == ch {
			delete(s.subscribers, subCh)
			close(subCh)
			break
		}
	}
}

// publishEvent publishes an event to all subscribers.
func (s *SimpleSession) publishEvent(event *SessionEvent) {
	s.subMu.RLock()
	defer s.subMu.RUnlock()

	for ch := range s.subscribers {
		select {
		case ch <- event:
		default:
			// Channel full, skip
		}
	}
}

// Close closes the session.
func (s *SimpleSession) Close() error {
	s.cancel()

	s.subMu.Lock()
	defer s.subMu.Unlock()

	// Close all subscriber channels
	for ch := range s.subscribers {
		close(ch)
		delete(s.subscribers, ch)
	}

	return nil
}
