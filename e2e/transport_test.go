package e2e

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// MockWebSocket simulates a WebSocket connection for testing
type MockWebSocket struct {
	conn       *websocket.Conn
	url        string
	clientID   string
	vm         *goja.Runtime
	onMessage  goja.Value
	onError    goja.Value
	onOpen     goja.Value
	onClose    goja.Value
	messageCh  chan []byte
	connected  bool
	t          *testing.T
}

// NewMockWebSocket creates a new mock WebSocket connection
func NewMockWebSocket(t *testing.T, vm *goja.Runtime, url, clientID string) *MockWebSocket {
	return &MockWebSocket{
		url:       url,
		clientID:  clientID,
		vm:        vm,
		messageCh: make(chan []byte, 100),
		t:         t,
	}
}

// Connect establishes a real WebSocket connection to the server
func (m *MockWebSocket) Connect() error {
	wsURL := m.url + "?client_id=" + m.clientID
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	m.conn = conn
	m.connected = true

	// Start receiving messages in background
	go m.receiveMessages()

	// Call onOpen callback if set
	if m.onOpen != nil && !goja.IsUndefined(m.onOpen) && !goja.IsNull(m.onOpen) {
		if fn, ok := goja.AssertFunction(m.onOpen); ok {
			if _, err := fn(goja.Undefined()); err != nil {
				m.t.Logf("Warning: onOpen callback failed: %v", err)
			}
		}
	}

	return nil
}

// receiveMessages continuously reads messages from the WebSocket connection
func (m *MockWebSocket) receiveMessages() {
	for {
		_, message, err := m.conn.ReadMessage()
		if err != nil {
			if m.onClose != nil && !goja.IsUndefined(m.onClose) && !goja.IsNull(m.onClose) {
				if fn, ok := goja.AssertFunction(m.onClose); ok {
					_, _ = fn(goja.Undefined())
				}
			}
			break
		}

		// Send to channel for async processing
		select {
		case m.messageCh <- message:
		default:
			m.t.Logf("Warning: message channel full, dropping message")
		}

		// Call onMessage callback if set
		if m.onMessage != nil && !goja.IsUndefined(m.onMessage) && !goja.IsNull(m.onMessage) {
			if fn, ok := goja.AssertFunction(m.onMessage); ok {
				if _, err := fn(goja.Undefined(), m.vm.ToValue(string(message))); err != nil {
					m.t.Logf("Warning: onMessage callback failed: %v", err)
				}
			}
		}
	}
}

// Send sends a message through the WebSocket
func (m *MockWebSocket) Send(message string) error {
	if !m.connected || m.conn == nil {
		return fmt.Errorf("WebSocket not connected")
	}

	return m.conn.WriteMessage(websocket.TextMessage, []byte(message))
}

// Close closes the WebSocket connection
func (m *MockWebSocket) Close() error {
	if !m.connected || m.conn == nil {
		return nil
	}

	m.connected = false
	return m.conn.Close()
}

// WaitForMessage waits for a message with a specific type
func (m *MockWebSocket) WaitForMessage(msgType string, timeout time.Duration) (map[string]interface{}, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case msg := <-m.messageCh:
			var data map[string]interface{}
			if err := json.Unmarshal(msg, &data); err != nil {
				return nil, fmt.Errorf("failed to parse message: %w", err)
			}

			// Extract type from metadata
			if metadata, ok := data["metadata"].(map[string]interface{}); ok {
				if protocolMsg, ok := metadata["protocol_message"].(map[string]interface{}); ok {
					if t, ok := protocolMsg["type"].(string); ok && t == msgType {
						return data, nil
					}
				}
			}

		case <-timer.C:
			return nil, fmt.Errorf("timeout waiting for %s message", msgType)
		}
	}
}

// SetupWebSocketEnvironment sets up the WebSocket mock in the Goja runtime
func SetupWebSocketEnvironment(t *testing.T, vm *goja.Runtime, serverURL string) {
	// Set up console.log
	console := vm.NewObject()
	_ = console.Set("log", func(call goja.FunctionCall) goja.Value {
		args := make([]interface{}, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = arg.Export()
		}
		t.Log(args...)
		return goja.Undefined()
	})
	_ = vm.Set("console", console)

	// Set up sleep function
	_ = vm.Set("sleep", func(call goja.FunctionCall) goja.Value {
		millis := int(call.Argument(0).ToInteger())
		time.Sleep(time.Duration(millis) * time.Millisecond)
		return goja.Undefined()
	})

	// Set up JSON
	_ = vm.Set("JSON", jsonMarshaler{
		stringify: func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			return string(b), err
		},
		parse: func(s string) (interface{}, error) {
			var v interface{}
			err := json.Unmarshal([]byte(s), &v)
			return v, err
		},
	})

	wsConstructor := func(call goja.FunctionCall) goja.Value {
		url := call.Argument(0).ToString().String()
		clientID := call.Argument(1).ToString().String()

		ws := NewMockWebSocket(t, vm, url, clientID)

		// Return WebSocket object with methods
		wsObj := vm.NewObject()
		_ = wsObj.Set("connect", func(call goja.FunctionCall) goja.Value {
			err := ws.Connect()
			if err != nil {
				panic(vm.ToValue(err.Error()))
			}
			return goja.Undefined()
		})

		_ = wsObj.Set("send", func(call goja.FunctionCall) goja.Value {
			msg := call.Argument(0).ToString().String()
			err := ws.Send(msg)
			if err != nil {
				panic(vm.ToValue(err.Error()))
			}
			return goja.Undefined()
		})

		_ = wsObj.Set("close", func(call goja.FunctionCall) goja.Value {
			err := ws.Close()
			if err != nil {
				panic(vm.ToValue(err.Error()))
			}
			return goja.Undefined()
		})

		_ = wsObj.Set("onMessage", nil)
		_ = wsObj.Set("onError", nil)
		_ = wsObj.Set("onOpen", nil)
		_ = wsObj.Set("onClose", nil)

		_ = wsObj.Set("setOnMessage", func(call goja.FunctionCall) goja.Value {
			ws.onMessage = call.Argument(0)
			return goja.Undefined()
		})

		_ = wsObj.Set("setOnError", func(call goja.FunctionCall) goja.Value {
			ws.onError = call.Argument(0)
			return goja.Undefined()
		})

		_ = wsObj.Set("setOnOpen", func(call goja.FunctionCall) goja.Value {
			ws.onOpen = call.Argument(0)
			return goja.Undefined()
		})

		_ = wsObj.Set("setOnClose", func(call goja.FunctionCall) goja.Value {
			ws.onClose = call.Argument(0)
			return goja.Undefined()
		})

		_ = wsObj.Set("waitForMessage", func(call goja.FunctionCall) goja.Value {
			msgType := call.Argument(0).ToString().String()
			timeout := int(call.Argument(1).ToInteger())

			msg, err := ws.WaitForMessage(msgType, time.Duration(timeout)*time.Millisecond)
			if err != nil {
				panic(vm.ToValue(err.Error()))
			}

			msgJSON, _ := json.Marshal(msg)
			return vm.ToValue(string(msgJSON))
		})

		_ = wsObj.Set("getClientID", func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(ws.clientID)
		})

		return wsObj
	}

	_ = vm.Set("WebSocket", wsConstructor)
}

// jsonMarshaler provides JSON.stringify and JSON.parse
type jsonMarshaler struct {
	stringify func(v interface{}) (string, error)
	parse     func(s string) (interface{}, error)
}

// TestWebSocketConnection tests basic WebSocket connection
func TestWebSocketConnection(t *testing.T) {
	// Skip if server is not running
	if !isServerAvailable() {
		t.Skip("Server not available")
	}

	vm := goja.New()
	SetupWebSocketEnvironment(t, vm, "ws://localhost:8080/ws")

	_, err := vm.RunString(`
		const ws = WebSocket("ws://localhost:8080/ws", "test-connection-" + Date.now());
		ws.connect();
		sleep(100);
		ws.close();
	`)

	require.NoError(t, err)
}

// TestCollaborativeEditing tests basic collaborative editing scenario
func TestCollaborativeEditing(t *testing.T) {
	if !isServerAvailable() {
		t.Skip("Server not available")
	}

	// Use unique IDs for each test run to avoid state pollution
	uniqueID := time.Now().UnixMilli()
	user1ID := fmt.Sprintf("e2e-user1-%d", uniqueID)
	user2ID := fmt.Sprintf("e2e-user2-%d", uniqueID)
	docID := fmt.Sprintf("test-e2e-%d.txt", uniqueID)

	// Create WebSocket connections directly (without Goja for now)
	ws1 := NewMockWebSocket(t, nil, "ws://localhost:8080/ws", user1ID)
	err := ws1.Connect()
	require.NoError(t, err)
	defer ws1.Close()

	ws2 := NewMockWebSocket(t, nil, "ws://localhost:8080/ws", user2ID)
	err = ws2.Connect()
	require.NoError(t, err)
	defer ws2.Close()

	// Subscribe user1
	subscribeMsg1 := map[string]interface{}{
		"type":      "subscribe",
		"client_id": user1ID,
		"doc_id":    docID,
		"metadata": map[string]interface{}{
			"protocol_message": map[string]interface{}{
				"type":      "subscribe",
				"timestamp": time.Now().UnixMilli(),
				"data": map[string]interface{}{
					"file_path": docID,
					"read_only": false,
				},
			},
		},
	}
	subscribeJSON1, _ := json.Marshal(subscribeMsg1)
	err = ws1.Send(string(subscribeJSON1))
	require.NoError(t, err)

	// Subscribe user2
	subscribeMsg2 := map[string]interface{}{
		"type":      "subscribe",
		"client_id": user2ID,
		"doc_id":    docID,
		"metadata": map[string]interface{}{
			"protocol_message": map[string]interface{}{
				"type":      "subscribe",
				"timestamp": time.Now().UnixMilli(),
				"data": map[string]interface{}{
					"file_path": docID,
					"read_only": false,
				},
			},
		},
	}
	subscribeJSON2, _ := json.Marshal(subscribeMsg2)
	err = ws2.Send(string(subscribeJSON2))
	require.NoError(t, err)

	// Wait for snapshot for user1
	snapshot1, err := ws1.WaitForMessage("snapshot", 2*time.Second)
	require.NoError(t, err)
	t.Logf("User1 received snapshot")

	// Wait for snapshot for user2
	_, err = ws2.WaitForMessage("snapshot", 2*time.Second)
	require.NoError(t, err)
	t.Logf("User2 received snapshot")

	// Extract session_id from snapshot
	sessionID := snapshot1["metadata"].(map[string]interface{})["protocol_message"].(map[string]interface{})["data"].(map[string]interface{})["session_id"].(string)

	// User2 sends an operation
	operationMsg := map[string]interface{}{
		"type":      "operation",
		"client_id": user2ID,
		"doc_id":    docID,
		"timestamp": time.Now().UnixMilli(),
		"metadata": map[string]interface{}{
			"protocol_message": map[string]interface{}{
				"type":       "operation",
				"session_id": sessionID,
				"timestamp":  time.Now().UnixMilli(),
				"data": map[string]interface{}{
					"session_id": sessionID,
					"operation":  []interface{}{0, "Hello", 0},
					"selection":  nil,
				},
			},
		},
	}
	opJSON, _ := json.Marshal(operationMsg)
	err = ws2.Send(string(opJSON))
	require.NoError(t, err)

	// User1 should receive remote_operation
	remoteOp, err := ws1.WaitForMessage("remote_operation", 2*time.Second)
	require.NoError(t, err)
	require.NotNil(t, remoteOp)
	t.Logf("User1 received remote operation")

	t.Log("Collaborative editing test passed")
}

// isServerAvailable checks if the WebSocket server is running
func isServerAvailable() bool {
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws?client_id=health-check", nil)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
