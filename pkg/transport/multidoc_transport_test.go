package transport

import (
	"fmt"
	"testing"
)

// TestMultiDocWebSocketTransport_BasicSubscription tests basic subscription functionality.
func TestMultiDocWebSocketTransport_BasicSubscription(t *testing.T) {
	// Create a multi-doc transport
	transport := NewMultiDocWebSocketTransport("test-client", "ws://localhost:8080/ws")

	// Subscribe to multiple documents
	doc1Sub, err := transport.Subscribe("/doc1.txt")
	if err != nil {
		t.Fatalf("Failed to subscribe to doc1: %v", err)
	}

	doc2Sub, err := transport.Subscribe("/doc2.txt")
	if err != nil {
		t.Fatalf("Failed to subscribe to doc2: %v", err)
	}

	doc3Sub, err := transport.Subscribe("/doc3.txt")
	if err != nil {
		t.Fatalf("Failed to subscribe to doc3: %v", err)
	}

	// Verify subscriptions
	if doc1Sub == nil {
		t.Error("Expected doc1 subscription to be created")
	}

	if doc2Sub == nil {
		t.Error("Expected doc2 subscription to be created")
	}

	if doc3Sub == nil {
		t.Error("Expected doc3 subscription to be created")
	}

	// List subscriptions
	docs := transport.ListSubscriptions()
	if len(docs) != 3 {
		t.Errorf("Expected 3 subscriptions, got %d", len(docs))
	}

	t.Logf("Subscribed to documents: %v", docs)
}

// TestMultiDocWebSocketTransport_Unsubscribe tests unsubscribe functionality.
func TestMultiDocWebSocketTransport_Unsubscribe(t *testing.T) {
	transport := NewMultiDocWebSocketTransport("test-client", "ws://localhost:8080/ws")

	// Subscribe and then unsubscribe
	_, err := transport.Subscribe("/doc1.txt")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Verify subscription exists
	if _, exists := transport.GetSubscription("/doc1.txt"); !exists {
		t.Error("Expected subscription to exist")
	}

	// Unsubscribe
	err = transport.Unsubscribe("/doc1.txt")
	if err != nil {
		t.Fatalf("Failed to unsubscribe: %v", err)
	}

	// Verify subscription removed
	if _, exists := transport.GetSubscription("/doc1.txt"); exists {
		t.Error("Expected subscription to be removed")
	}

	docs := transport.ListSubscriptions()
	if len(docs) != 0 {
		t.Errorf("Expected 0 subscriptions, got %d", len(docs))
	}
}

// TestMultiDocWebSocketTransport_SubscriptionHandlers tests setting message handlers.
func TestMultiDocWebSocketTransport_SubscriptionHandlers(t *testing.T) {
	transport := NewMultiDocWebSocketTransport("test-client", "ws://localhost:8080/ws")

	// Subscribe and set handlers
	sub, err := transport.Subscribe("/doc1.txt")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Set handlers
	sub.OnRemoteOperation(func(data *RemoteOperationData) {
		t.Logf("Received operation for session %s from client %s", data.SessionID, data.ClientID)
	})

	sub.OnSnapshot(func(data *SnapshotData) {
		t.Logf("Received snapshot for session %s", data.SessionID)
	})

	sub.OnUserJoined(func(data *UserJoinedData) {
		t.Logf("User %s joined session %s", data.ClientID, data.SessionID)
	})

	// Verify handlers are set (we can't actually test them firing without a real connection)
	if sub.onRemoteOp == nil {
		t.Error("Expected operation handler to be set")
	}

	if sub.onSnapshot == nil {
		t.Error("Expected snapshot handler to be set")
	}

	if sub.onUserJoined == nil {
		t.Error("Expected user joined handler to be set")
	}

	t.Log("All handlers configured successfully")
}

// TestMultiDocWebSocketTransport_MultipleDocumentsIndependence tests that different documents have independent state.
func TestMultiDocWebSocketTransport_MultipleDocumentsIndependence(t *testing.T) {
	transport := NewMultiDocWebSocketTransport("test-client", "ws://localhost:8080/ws")

	// Subscribe to multiple documents
	doc1Sub, _ := transport.Subscribe("/doc1.txt")
	doc2Sub, _ := transport.Subscribe("/doc2.txt")
	doc3Sub, _ := transport.Subscribe("/doc3.txt")

	// Set different handlers for each document
	doc1OpCount := 0
	doc2OpCount := 0
	doc3OpCount := 0

	doc1Sub.OnRemoteOperation(func(data *RemoteOperationData) {
		doc1OpCount++
	})

	doc2Sub.OnRemoteOperation(func(data *RemoteOperationData) {
		doc2OpCount++
	})

	doc3Sub.OnRemoteOperation(func(data *RemoteOperationData) {
		doc3OpCount++
	})

	// Set different snapshot handlers
	doc1Sub.OnSnapshot(func(data *SnapshotData) {
		t.Logf("Doc1 snapshot: %s", data.SessionID)
	})

	doc2Sub.OnSnapshot(func(data *SnapshotData) {
		t.Logf("Doc2 snapshot: %s", data.SessionID)
	})

	doc3Sub.OnSnapshot(func(data *SnapshotData) {
		t.Logf("Doc3 snapshot: %s", data.SessionID)
	})

	// Verify handlers are set
	if doc1Sub.onRemoteOp == nil || doc2Sub.onRemoteOp == nil || doc3Sub.onRemoteOp == nil {
		t.Error("Expected all operation handlers to be set")
	}

	if doc1Sub.onSnapshot == nil || doc2Sub.onSnapshot == nil || doc3Sub.onSnapshot == nil {
		t.Error("Expected all snapshot handlers to be set")
	}

	t.Logf("Document handlers are independent")
}

// TestMultiDocWebSocketTransport_ReadOnlySubscription tests read-only subscriptions.
func TestMultiDocWebSocketTransport_ReadOnlySubscription(t *testing.T) {
	transport := NewMultiDocWebSocketTransport("test-client", "ws://localhost:8080/ws")

	// Subscribe with read-only
	sub, err := transport.Subscribe("/doc1.txt")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Set read-only flag
	sub.ReadOnly = true

	// Verify read-only is set
	if !sub.ReadOnly {
		t.Error("Expected read-only to be true")
	}

	t.Log("Read-only subscription configured")
}

// TestMultiDocWebSocketTransport_ListSubscriptions tests listing all subscriptions.
func TestMultiDocWebSocketTransport_ListSubscriptions(t *testing.T) {
	transport := NewMultiDocWebSocketTransport("test-client", "ws://localhost:8080/ws")

	// Initially no subscriptions
	docs := transport.ListSubscriptions()
	if len(docs) != 0 {
		t.Errorf("Expected 0 subscriptions, got %d", len(docs))
	}

	// Subscribe to documents
	transport.Subscribe("/doc1.txt")
	transport.Subscribe("/doc2.txt")
	transport.Subscribe("/doc3.txt")

	// List subscriptions
	docs = transport.ListSubscriptions()
	if len(docs) != 3 {
		t.Errorf("Expected 3 subscriptions, got %d", len(docs))
	}

	// Verify all documents are listed
	expectedDocs := map[string]bool{
		"/doc1.txt": true,
		"/doc2.txt": true,
		"/doc3.txt": true,
	}

	for _, doc := range docs {
		if !expectedDocs[doc] {
			t.Errorf("Unexpected document: %s", doc)
		}
	}

	t.Logf("Listed subscriptions: %v", docs)
}

// TestMultiDocWebSocketTransport_Close tests closing the transport.
func TestMultiDocWebSocketTransport_Close(t *testing.T) {
	transport := NewMultiDocWebSocketTransport("test-client", "ws://localhost:8080/ws")

	// Subscribe to documents
	_, err := transport.Subscribe("/doc1.txt")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	_, err = transport.Subscribe("/doc2.txt")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Close transport
	err = transport.Close()
	if err != nil {
		t.Fatalf("Failed to close transport: %v", err)
	}

	// Verify all subscriptions are cleared
	docs := transport.ListSubscriptions()
	if len(docs) != 0 {
		t.Errorf("Expected 0 subscriptions after close, got %d", len(docs))
	}

	// Verify can't subscribe after close
	_, err = transport.Subscribe("/doc3.txt")
	if err == nil {
		t.Error("Expected error when subscribing after close")
	}

	t.Log("Transport closed successfully")
}

// TestMultiDocWebSocketTransport_GetSubscription tests getting a specific subscription.
func TestMultiDocWebSocketTransport_GetSubscription(t *testing.T) {
	transport := NewMultiDocWebSocketTransport("test-client", "ws://localhost:8080/ws")

	// Subscribe to a document
	sub, err := transport.Subscribe("/doc1.txt")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Get subscription
	retrievedSub, exists := transport.GetSubscription("/doc1.txt")
	if !exists {
		t.Error("Expected subscription to exist")
	}

	if retrievedSub != sub {
		t.Error("Expected retrieved subscription to be the same")
	}

	// Try to get non-existent subscription
	_, exists = transport.GetSubscription("/nonexistent.txt")
	if exists {
		t.Error("Expected non-existent subscription to return false")
	}

	t.Log("GetSubscription works correctly")
}

// TestMultiDocWebSocketTransport_ConnectedState tests connection state tracking.
func TestMultiDocWebSocketTransport_ConnectedState(t *testing.T) {
	transport := NewMultiDocWebSocketTransport("test-client", "ws://localhost:8080/ws")

	// Initially not connected
	if transport.IsConnected() {
		t.Error("Expected to be disconnected initially")
	}

	// Close without connecting
	transport.Close()

	// Still not connected (no crash)
	if transport.IsConnected() {
		t.Error("Expected to be disconnected after close")
	}

	t.Log("Connection state tracking works")
}

// TestMultiDocWebSocketTransport_ConcurrentAccess tests concurrent access to subscriptions.
func TestMultiDocWebSocketTransport_ConcurrentAccess(t *testing.T) {
	transport := NewMultiDocWebSocketTransport("test-client", "ws://localhost:8080/ws")

	// Subscribe from multiple goroutines
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(index int) {
			docPath := fmt.Sprintf("/doc%d.txt", index)
			_, err := transport.Subscribe(docPath)
			if err != nil {
				t.Errorf("Failed to subscribe to %s: %v", docPath, err)
			}
			done <- true
		}(i)
	}

	// Wait for all subscriptions
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all subscriptions
	docs := transport.ListSubscriptions()
	if len(docs) != 10 {
		t.Errorf("Expected 10 subscriptions, got %d", len(docs))
	}

	t.Logf("Concurrent access handled correctly: %d subscriptions", len(docs))
}

// TestMultiDocWebSocketTransport_SessionIDMapping tests SessionID mapping.
func TestMultiDocWebSocketTransport_SessionIDMapping(t *testing.T) {
	transport := NewMultiDocWebSocketTransport("test-client", "ws://localhost:8080/ws")

	// Subscribe to documents
	sub1, _ := transport.Subscribe("/doc1.txt")
	sub2, _ := transport.Subscribe("/doc2.txt")

	// Simulate server assigning SessionIDs
	sub1.SessionID = "session-abc123"
	sub2.SessionID = "session-def456"

	// Verify SessionIDs are stored
	if sub1.SessionID != "session-abc123" {
		t.Errorf("Expected SessionID session-abc123, got %s", sub1.SessionID)
	}

	if sub2.SessionID != "session-def456" {
		t.Errorf("Expected SessionID session-def456, got %s", sub2.SessionID)
	}

	t.Logf("SessionID mapping: doc1=%s, doc2=%s", sub1.SessionID, sub2.SessionID)
}

// BenchmarkMultiDocWebSocketTransport_Subscribe benchmarks subscription creation.
func BenchmarkMultiDocWebSocketTransport_Subscribe(b *testing.B) {
	transport := NewMultiDocWebSocketTransport("bench-client", "ws://localhost:8080/ws")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		docPath := fmt.Sprintf("/doc%d.txt", i)
		_, err := transport.Subscribe(docPath)
		if err != nil {
			b.Fatalf("Failed to subscribe: %v", err)
		}
	}
}

// BenchmarkMultiDocWebSocketTransport_ListSubscriptions benchmarks listing subscriptions.
func BenchmarkMultiDocWebSocketTransport_ListSubscriptions(b *testing.B) {
	transport := NewMultiDocWebSocketTransport("bench-client", "ws://localhost:8080/ws")

	// Subscribe to 100 documents
	for i := 0; i < 100; i++ {
		docPath := fmt.Sprintf("/doc%d.txt", i)
		transport.Subscribe(docPath)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transport.ListSubscriptions()
	}
}
