package e2e

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coreseekdev/texere/pkg/session"
	"github.com/coreseekdev/texere/pkg/transport"
)

// ========== Test Framework ==========

// TestFramework provides e2e testing infrastructure for collaborative editing.
type TestFramework struct {
	mu              sync.RWMutex
	clients         []*SimulatedClient
	sessions        map[string]*session.SimpleSession
	wsServer        *transport.WebSocketServer
	sseServer       *transport.SSEServer
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewTestFramework creates a new e2e test framework.
func NewTestFramework() *TestFramework {
	ctx, cancel := context.WithCancel(context.Background())
	return &TestFramework{
		clients:  make([]*SimulatedClient, 0),
		sessions: make(map[string]*session.SimpleSession),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// StartServer starts a test server with the given transport.
func (tf *TestFramework) StartServer(addr string, transportType string) error {
	switch transportType {
	case "websocket":
		wsServer := transport.NewWebSocketServer(addr)
		tf.wsServer = wsServer
		return wsServer.Start(tf.ctx)
	case "sse":
		sseServer := transport.NewSSEServer(addr)
		tf.sseServer = sseServer
		return sseServer.Start()
	default:
		return fmt.Errorf("unsupported transport type: %s", transportType)
	}
}

// StopServer stops the test server.
func (tf *TestFramework) StopServer() error {
	tf.cancel()

	// Close all clients
	for _, client := range tf.clients {
		_ = client.Close()
	}

	if tf.wsServer != nil {
		return tf.wsServer.Close()
	}
	if tf.sseServer != nil {
		return tf.sseServer.Close()
	}
	return nil
}

// CreateSession creates a new session for testing.
func (tf *TestFramework) CreateSession(docID, initialContent string, docType session.DocumentType) (*session.SimpleSession, error) {
	config := session.SessionConfig{
		DocID:          docID,
		InitialContent: initialContent,
		DocType:        docType,
		EnableUndo:     true,
		MaxHistory:     50,
		Auth:           session.NewTokenAuthenticator(),
		Content:        session.NewMemoryContentStorage(),
	}

	sess, err := session.NewSimpleSession(tf.ctx, config)
	if err != nil {
		return nil, err
	}

	tf.mu.Lock()
	tf.sessions[docID] = sess
	tf.mu.Unlock()

	return sess, nil
}

// AddClient adds a simulated client to the test.
func (tf *TestFramework) AddClient(clientID, docID string, transportType string) (*SimulatedClient, error) {
	client := NewSimulatedClient(clientID, docID)

	if err := client.Connect(tf.ctx, "ws://localhost:8080", transportType); err != nil {
		return nil, err
	}

	tf.mu.Lock()
	tf.clients = append(tf.clients, client)
	tf.mu.Unlock()

	return client, nil
}

// RunConcurrentTest runs a concurrent editing test with multiple clients.
func (tf *TestFramework) RunConcurrentTest(testSpec *TestSpec) *TestResult {
	result := &TestResult{
		StartTime: time.Now(),
		Errors:    make([]error, 0),
	}

	// Create wait group for all clients
	var wg sync.WaitGroup
	clientCount := len(testSpec.Clients)
	result.ClientCount = clientCount

	// Channel to collect errors
	errCh := make(chan error, clientCount*10)

	// Start all clients
	for _, clientSpec := range testSpec.Clients {
		wg.Add(1)
		go func(spec *ClientSpec) {
			defer wg.Done()

			client, err := tf.AddClient(spec.ID, testSpec.DocID, testSpec.TransportType)
			if err != nil {
				errCh <- fmt.Errorf("client %s failed to connect: %w", spec.ID, err)
				return
			}
			defer client.Close()

			// Run client operations
			if err := client.RunOperations(spec.Operations); err != nil {
				errCh <- fmt.Errorf("client %s operations failed: %w", spec.ID, err)
			}
		}(clientSpec)
	}

	// Wait for all clients to complete
	wg.Wait()
	close(errCh)

	// Collect errors
	for err := range errCh {
		result.Errors = append(result.Errors, err)
		atomic.AddInt32(&result.FailureCount, 1)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.SuccessCount = int32(clientCount) - result.FailureCount

	// Verify document consistency
	if testSpec.VerifyConsistency {
		result.ConsistencyCheck = tf.VerifyDocumentConsistency(testSpec.DocID)
	}

	return result
}

// VerifyDocumentConsistency verifies that all clients have the same document state.
func (tf *TestFramework) VerifyDocumentConsistency(docID string) bool {
	tf.mu.RLock()
	defer tf.mu.RUnlock()

	// Get reference content from first connected client
	var referenceContent string
	for _, client := range tf.clients {
		if client.docID == docID && client.IsConnected() {
			referenceContent = client.GetContent()
			break
		}
	}

	// Compare all clients
	for _, client := range tf.clients {
		if client.docID == docID && client.IsConnected() {
			if client.GetContent() != referenceContent {
				return false
			}
		}
	}

	return true
}

// ========== Test Specification ==========

// TestSpec defines a concurrent editing test.
type TestSpec struct {
	DocID               string
	InitialContent      string
	DocType             session.DocumentType
	TransportType       string
	Clients             []*ClientSpec
	VerifyConsistency   bool
	ExpectedResult      string
	Timeout             time.Duration
}

// ClientSpec defines a simulated client's behavior.
type ClientSpec struct {
	ID          string
	Operations  []*ClientOperation
	Delay       time.Duration
	AutoReplay  bool
}

// ClientOperation defines an operation a client will perform.
type ClientOperation struct {
	Type      OpType // Insert, Delete, Retain
	Position  int
	Content   string
	Length    int           // For delete operations
	Delay     time.Duration // Delay before this operation
	Timestamp time.Time
}

// OpType represents the type of operation.
type OpType int

const (
	OpInsert OpType = iota
	OpDelete
	OpRetain
)

// ========== Test Result ==========

// TestResult represents the results of a concurrent test.
type TestResult struct {
	StartTime           time.Time
	EndTime             time.Time
	Duration            time.Duration
	ClientCount         int
	SuccessCount        int32
	FailureCount        int32
	Errors              []error
	ConsistencyCheck    bool
	MessagesSent        int32
	MessagesReceived    int32
	OperationsApplied   int32
}

// Success returns true if all tests passed.
func (r *TestResult) Success() bool {
	return r.FailureCount == 0 && r.ConsistencyCheck
}

// String returns a string representation of the result.
func (r *TestResult) String() string {
	return fmt.Sprintf(
		"TestResult{Duration: %v, Clients: %d, Success: %d, Failures: %d, Consistent: %v}",
		r.Duration, r.ClientCount, r.SuccessCount, r.FailureCount, r.ConsistencyCheck,
	)
}
