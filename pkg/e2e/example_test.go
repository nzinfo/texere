package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/coreseekdev/texere/pkg/session"
)

// ExampleTestFramework demonstrates basic usage of the e2e testing framework.
func ExampleTestFramework() {
	framework := NewTestFramework()
	defer framework.StopServer()

	// Start server
	_ = framework.StartServer(":8085", "websocket")

	// Create session
	_, _ = framework.CreateSession("example-doc", "Hello", session.DocTypeString)

	// Define test
	testSpec := &TestSpec{
		DocID:          "example-doc",
		InitialContent: "Hello",
		DocType:        session.DocTypeString,
		TransportType:  "websocket",
		VerifyConsistency: true,
		Clients: []*ClientSpec{
			{
				ID: "user1",
				Operations: []*ClientOperation{
					{
						Type:     OpInsert,
						Position: 5,
						Content:  " Alice",
					},
				},
			},
			{
				ID: "user2",
				Operations: []*ClientOperation{
					{
						Type:     OpInsert,
						Position: 5,
						Content:  " Bob",
					},
				},
			},
		},
	}

	// Run test
	result := framework.RunConcurrentTest(testSpec)

	fmt.Println(result.String())
	// Output: TestResult{Duration: ..., Clients: 2, Success: 2, Failures: 0, Consistent: true}
}

// TestBasicConcurrentEditing tests concurrent editing with multiple clients.
func TestBasicConcurrentEditing(t *testing.T) {
	// Create test framework
	framework := NewTestFramework()
	defer framework.StopServer()

	// Start WebSocket server
	if err := framework.StartServer(":8086", "websocket"); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Create a test session
	_, err := framework.CreateSession("test-doc", "Hello World", session.DocTypeString)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Define test spec: 3 clients concurrently editing
	testSpec := &TestSpec{
		DocID:          "test-doc",
		InitialContent: "Hello World",
		DocType:        session.DocTypeString,
		TransportType:  "websocket",
		VerifyConsistency: true,
		Timeout:        30 * time.Second,
		Clients: []*ClientSpec{
			{
				ID: "client-1",
				Operations: []*ClientOperation{
					{
						Type:     OpInsert,
						Position: 5,
						Content:  " Beautiful",
					},
					{
						Type:     OpInsert,
						Position: 16,
						Content:  " Day",
						Delay:    100 * time.Millisecond,
					},
				},
			},
			{
				ID: "client-2",
				Operations: []*ClientOperation{
					{
						Type:     OpInsert,
						Position: 11,
						Content:  " Wonderful",
						Delay:    50 * time.Millisecond,
					},
				},
			},
			{
				ID: "client-3",
				Operations: []*ClientOperation{
					{
						Type:     OpDelete,
						Position: 0,
						Length:   5,
						Delay:    75 * time.Millisecond,
					},
					{
						Type:     OpInsert,
						Position: 0,
						Content:  "Goodbye",
						Delay:    50 * time.Millisecond,
					},
				},
			},
		},
	}

	// Run concurrent test
	result := framework.RunConcurrentTest(testSpec)

	// Verify results
	if !result.Success() {
		t.Errorf("Test failed: %s", result.String())
		for _, err := range result.Errors {
			t.Errorf("Error: %v", err)
		}
	}

	t.Logf("Test completed successfully: %s", result.String())
}

// TestSingleClient tests a single client editing.
func TestSingleClient(t *testing.T) {
	framework := NewTestFramework()
	defer framework.StopServer()

	if err := framework.StartServer(":8087", "websocket"); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	_, err := framework.CreateSession("single-doc", "Hello", session.DocTypeString)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	testSpec := &TestSpec{
		DocID:          "single-doc",
		InitialContent: "Hello",
		DocType:        session.DocTypeString,
		TransportType:  "websocket",
		VerifyConsistency: true,
		Clients: []*ClientSpec{
			{
				ID: "single-client",
				Operations: []*ClientOperation{
					{
						Type:     OpInsert,
						Position: 5,
						Content:  " World",
					},
				},
			},
		},
	}

	result := framework.RunConcurrentTest(testSpec)

	if !result.Success() {
		t.Errorf("Single client test failed: %s", result.String())
	}

	t.Logf("Single client test: %s", result.String())
}

// TestClientPool tests the client pool functionality.
func TestClientPool(t *testing.T) {
	pool := NewClientPool()
	defer pool.Close()

	// Add simulated clients
	for i := 0; i < 10; i++ {
		client := NewSimulatedClient(fmt.Sprintf("pool-client-%d", i), "pool-doc")
		pool.Add(client)
	}

	// Verify all clients are in pool
	clients := pool.GetAll()
	if len(clients) != 10 {
		t.Errorf("Expected 10 clients, got %d", len(clients))
	}

	// Get specific client
	client, ok := pool.Get("pool-client-5")
	if !ok {
		t.Error("Failed to get client from pool")
	}

	if client.GetID() != "pool-client-5" {
		t.Errorf("Expected client ID 'pool-client-5', got '%s'", client.GetID())
	}

	// Remove client
	pool.Remove("pool-client-5")

	// Verify client is removed
	_, ok = pool.Get("pool-client-5")
	if ok {
		t.Error("Client should have been removed from pool")
	}

	clients = pool.GetAll()
	if len(clients) != 9 {
		t.Errorf("Expected 9 clients after removal, got %d", len(clients))
	}
}
