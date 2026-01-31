package transport

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// ExampleProtocol demonstrates the WebSocket protocol usage.
func ExampleProtocol() {
	// ========== Server Setup ==========

	handler := NewProtocolHandler(nil, nil)
	server := NewWebSocketServer(":8080")
	handler.SetServer(server)

	// ========== Client Connection ==========

	// 1. Client connects and receives welcome
	welcomeMsg := &ProtocolMessage{
		Type:      MessageTypeWelcome,
		Timestamp: time.Now().Unix(),
	}
	welcomeData, _ := json.Marshal(welcomeMsg)
	fmt.Println(string(welcomeData))

	// 2. Client subscribes to a file (read-only)
	subscribeMsg := &ProtocolMessage{
		Type:      MessageTypeSubscribe,
		Timestamp: time.Now().Unix(),
	}
	subscribeData := &SubscribeData{
		FilePath: "/path/to/file.txt",
		ReadOnly: true,
		UseSSE:   true,
	}
	subscribeMsg.Data, _ = json.Marshal(subscribeData)
	fmt.Println("Client → Server: subscribe")
	printJSON(subscribeMsg)

	// 3. Server responds with snapshot
	snapshotMsg := &ProtocolMessage{
		Type:      MessageTypeSnapshot,
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		Timestamp: time.Now().Unix(),
	}
	snapshotData := &SnapshotData{
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		FilePath:  "/path/to/file.txt",
		Content:   "Hello World",
		Revision:  0,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
		Clients:   []ClientInfo{},
		ReadOnly:  true,
	}
	snapshotMsg.Data, _ = json.Marshal(snapshotData)
	fmt.Println("Server → Client: snapshot")
	printJSON(snapshotMsg)

	// 4. Another client starts editing
	startEditingMsg := &ProtocolMessage{
		Type:      MessageTypeStartEditing,
		Timestamp: time.Now().Unix(),
	}
	startEditingData := &StartEditingData{
		FilePath: "/path/to/file.txt",
	}
	startEditingMsg.Data, _ = json.Marshal(startEditingData)
	fmt.Println("Client → Server: start_editing")
	printJSON(startEditingMsg)

	// 5. Client sends an operation
	operationMsg := &ProtocolMessage{
		Type:      MessageTypeOperation,
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		Timestamp: time.Now().Unix(),
	}
	operationData := &OperationData{
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		Revision:  0,
		Operation: []interface{}{5, " Beautiful"},
		Selection: &CursorData{
			Position:     15,
			SelectionEnd: 15,
		},
	}
	operationMsg.Data, _ = json.Marshal(operationData)
	fmt.Println("Client → Server: operation")
	printJSON(operationMsg)

	// 6. Server acknowledges
	ackMsg := &ProtocolMessage{
		Type:      MessageTypeAck,
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		Timestamp: time.Now().Unix(),
	}
	ackData := &AckData{
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		Revision:  1,
		Timestamp: time.Now().Unix(),
	}
	ackMsg.Data, _ = json.Marshal(ackData)
	fmt.Println("Server → Client: ack")
	printJSON(ackMsg)

	// 7. Server broadcasts to other clients
	remoteOpMsg := &ProtocolMessage{
		Type:      MessageTypeRemoteOperation,
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		Timestamp: time.Now().Unix(),
	}
	remoteOpData := &RemoteOperationData{
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		ClientID:  "client-2",
		Revision:  1,
		Operation: []interface{}{5, " Beautiful"},
		Selection: &CursorData{
			Position:     15,
			SelectionEnd: 15,
		},
	}
	remoteOpMsg.Data, _ = json.Marshal(remoteOpData)
	fmt.Println("Server → Other Clients: remote_operation")
	printJSON(remoteOpMsg)

	// 8. Client stops editing
	stopEditingMsg := &ProtocolMessage{
		Type:      MessageTypeStopEditing,
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		Timestamp: time.Now().Unix(),
	}
	stopEditingData := &StopEditingData{
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
	}
	stopEditingMsg.Data, _ = json.Marshal(stopEditingData)
	fmt.Println("Client → Server: stop_editing")
	printJSON(stopEditingMsg)
}

func printJSON(v interface{}) {
	bytes, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(bytes))
}

// TestProtocolMessages tests protocol message creation.
func TestProtocolMessages(t *testing.T) {
	// Test OT operation format
	operationData := &OperationData{
		SessionID: "test-session-id",
		Revision:  10,
		Operation: []interface{}{5, "Hello", 10, -3},
		Selection: &CursorData{
			Position:     15,
			SelectionEnd: 15,
		},
	}

	bytes, err := json.MarshalIndent(operationData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	expected := `{
  "session_id": "test-session-id",
  "revision": 10,
  "operation": [5, "Hello", 10, -3],
  "selection": {
    "position": 15,
    "selection_end": 15
  }
}`

	if string(bytes) != expected {
		t.Errorf("Unexpected JSON:\nGot:\n%s\n\nExpected:\n%s", string(bytes), expected)
	}
}

// TestSessionRefCount tests reference counting logic.
func TestSessionRefCount(t *testing.T) {
	rc := & SessionRefCount{
		SessionID:   "test-session",
		FilePath:    "/test.txt",
		ReaderCount: 0,
		WriterCount: 0,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	// Test: Add reader
	rc.AddReader()
	if rc.ReaderCount != 1 {
		t.Errorf("Expected ReaderCount=1, got %d", rc.ReaderCount)
	}
	if !rc.IsActive() {
		t.Error("Expected session to be active")
	}

	// Test: Add writer
	rc.AddWriter()
	if rc.WriterCount != 1 {
		t.Errorf("Expected WriterCount=1, got %d", rc.WriterCount)
	}
	if !rc.HasWriters() {
		t.Error("Expected session to have writers")
	}

	// Test: Remove writer
	rc.RemoveWriter()
	if rc.WriterCount != 0 {
		t.Errorf("Expected WriterCount=0, got %d", rc.WriterCount)
	}
	if rc.HasWriters() {
		t.Error("Expected session to have no writers")
	}

	// Test: Should not destroy (still has reader)
	if rc.ShouldDestroy() {
		t.Error("Expected session should not be destroyed")
	}

	// Test: Remove reader
	rc.RemoveReader()
	if !rc.ShouldDestroy() {
		t.Error("Expected session should be destroyed")
	}
}

// TestParseOperationData tests parsing OT operations from different formats.
func TestParseOperationData(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []interface{}
	}{
		{
			name:     "Array format",
			input:    []interface{}{5, "Hello", 10, -3},
			expected: []interface{}{5, "Hello", 10, -3},
		},
		{
			name:  "Object format",
			input: map[string]interface{}{
				"retain": 5,
				"insert": "Hello",
				"delete": 3,
			},
			expected: []interface{}{5, "Hello", -3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err :=  ParseOperationData(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
			}

			for i := range tt.expected {
				if result[i] != tt.expected[i] {
					t.Errorf("Index %d: expected %v, got %v", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

// TestSessionManager tests session management.
func TestSessionManager(t *testing.T) {
	sm :=  NewSessionManager()

	// Test: Create new session
	session, isNew := sm.GetOrCreateSession("/test.txt")
	if !isNew {
		t.Error("Expected new session to be created")
	}
	if session == nil {
		t.Fatal("Expected session to be created")
	}
	if session.SessionID == "" {
		t.Error("Expected session ID to be set")
	}

	// Test: Get existing session
	session2, isNew2 := sm.GetOrCreateSession("/test.txt")
	if isNew2 {
		t.Error("Expected existing session")
	}
	if session2.SessionID != session.SessionID {
		t.Error("Expected same session ID")
	}

	// Test: Get by path
	session3 := sm.GetSessionByPath("/test.txt")
	if session3 == nil {
		t.Error("Expected session to be found")
	}
	if session3.SessionID != session.SessionID {
		t.Error("Expected same session")
	}

	// Test: Destroy session
	sm.DestroySession(session.SessionID)
	session4 := sm.GetSession(session.SessionID)
	if session4 != nil {
		t.Error("Expected session to be destroyed")
	}

	session5 := sm.GetSessionByPath("/test.txt")
	if session5 != nil {
		t.Error("Expected session to be destroyed")
	}
}

// TestEditSession tests edit session operations.
func TestEditSession(t *testing.T) {
	es := NewEditSession("test-session", "/test.txt", "Hello World")

	// Test: Add client
	client := &SessionClient{
		ClientID:  "client-1",
		FilePath:  "/test.txt",
		ReadOnly:  false,
		IsEditing: true,
		Connected: true,
	}
	es.AddClient("client-1", client)

	// Test: Get client
	retrievedClient := es.GetClient("client-1")
	if retrievedClient == nil {
		t.Error("Expected client to be found")
	}
	if retrievedClient.ClientID != "client-1" {
		t.Error("Expected correct client ID")
	}

	// Test: Get client infos
	infos := es.GetClientInfos()
	if len(infos) != 1 {
		t.Errorf("Expected 1 client info, got %d", len(infos))
	}

	// Test: Add operation
	es.AddOperation([]interface{}{5, " Hello"}, "client-1")
	recentOps := es.GetRecentOperations()
	if len(recentOps) != 1 {
		t.Error("Expected 1 recent operation")
	}

	// Test: Get content
	content := es.GetContent()
	if content != "Hello World" {
		t.Errorf("Expected content 'Hello World', got '%s'", content)
	}

	// Test: Get version
	version := es.GetCurrentVersion()
	if version != 1 {
		t.Errorf("Expected version 1, got %d", version)
	}

	// Test: Remove client
	es.RemoveClient("client-1")
	if es.GetClient("client-1") != nil {
		t.Error("Expected client to be removed")
	}
}
