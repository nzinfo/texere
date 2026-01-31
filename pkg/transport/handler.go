package transport

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"unicode/utf8"

	"github.com/coreseekdev/texere/pkg/session"
	"github.com/coreseekdev/texere/pkg/ot"
)

// ProtocolHandler handles WebSocket protocol messages.
type ProtocolHandler struct {
	mu               sync.RWMutex
	sessionManager   *SessionManager
	contentStorage   session.ContentStorage
	authenticator    session.Authenticator
	server           *WebSocketServer
}

// NewProtocolHandler creates a new protocol handler.
func NewProtocolHandler(storage session.ContentStorage, auth session.Authenticator) *ProtocolHandler {
	sm := NewSessionManager()
	sm.SetContentStorage(storage)

	return &ProtocolHandler{
		sessionManager: sm,
		contentStorage: storage,
		authenticator:  auth,
	}
}

// SetServer sets the WebSocket server.
func (h *ProtocolHandler) SetServer(server *WebSocketServer) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.server = server

	// Set raw message handler (for new protocol)
	server.SetRawMessageHandler(h.handleRawMessage)
}

// handleRawMessage handles incoming raw WebSocket messages (new protocol).
func (h *ProtocolHandler) handleRawMessage(clientID string, messageBytes []byte) {
	// Parse client message format
	var clientMsg struct {
		Type      string                 `json:"type"`
		ClientID  string                 `json:"client_id"`
		DocID     string                 `json:"doc_id,omitempty"`
		Timestamp int64                  `json:"timestamp"`
		Metadata  map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := json.Unmarshal(messageBytes, &clientMsg); err != nil {
		log.Printf("[Handler] Failed to parse client message: %v", err)
		return
	}

	// Extract protocol message from metadata
	protocolMsgBytes, ok := clientMsg.Metadata["protocol_message"]
	if !ok {
		log.Printf("[Handler] No protocol_message in metadata")
		return
	}

	protocolMsgJSON, err := json.Marshal(protocolMsgBytes)
	if err != nil {
		log.Printf("[Handler] Failed to marshal protocol message: %v", err)
		return
	}

	var protocolMsg ProtocolMessage
	if err := json.Unmarshal(protocolMsgJSON, &protocolMsg); err != nil {
		log.Printf("[Handler] Failed to parse protocol message: %v", err)
		return
	}

	log.Printf("[Handler] %s: Received %s", clientID, protocolMsg.Type)

	// Create a simple Message wrapper for compatibility
	msg := &Message{
		Type:      0, // Not used in new protocol
		DocID:     clientMsg.DocID,
		ClientID:  clientMsg.ClientID,
		Timestamp: clientMsg.Timestamp,
		Metadata:  clientMsg.Metadata,
	}

	// Handle based on protocol message type
	switch protocolMsg.Type {
	case MessageTypeSubscribe:
		h.handleSubscribe(msg, &protocolMsg)
	case MessageTypeUnsubscribe:
		h.handleUnsubscribe(msg, &protocolMsg)
	case MessageTypeStartEditing:
		h.handleStartEditing(msg, &protocolMsg)
	case MessageTypeStopEditing:
		h.handleStopEditing(msg, &protocolMsg)
	case MessageTypeOperation:
		h.handleOperation(msg, &protocolMsg)
	case MessageTypeCursor:
		h.handleCursor(msg, &protocolMsg)
	case MessageTypeHeartbeat:
		h.handleHeartbeat(msg, &protocolMsg)
	default:
		log.Printf("[Handler] Unknown message type: %s", protocolMsg.Type)
	}
}

// handleMessage handles incoming WebSocket messages (legacy).
func (h *ProtocolHandler) handleMessage(msg *Message) {
	var protocolMsg ProtocolMessage
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", msg.Metadata)), &protocolMsg); err != nil {
		log.Printf("Failed to parse protocol message: %v", err)
		return
	}

	switch protocolMsg.Type {
	case MessageTypeSubscribe:
		h.handleSubscribe(msg, &protocolMsg)
	case MessageTypeUnsubscribe:
		h.handleUnsubscribe(msg, &protocolMsg)
	case MessageTypeStartEditing:
		h.handleStartEditing(msg, &protocolMsg)
	case MessageTypeStopEditing:
		h.handleStopEditing(msg, &protocolMsg)
	case MessageTypeOperation:
		h.handleOperation(msg, &protocolMsg)
	case MessageTypeCursor:
		h.handleCursor(msg, &protocolMsg)
	case MessageTypeHeartbeat:
		h.handleHeartbeat(msg, &protocolMsg)
	default:
		log.Printf("Unknown message type: %s", protocolMsg.Type)
	}
}

// handleSubscribe handles file subscription.
func (h *ProtocolHandler) handleSubscribe(msg *Message, pm *ProtocolMessage) {
	var data SubscribeData
	if err := json.Unmarshal(pm.Data, &data); err != nil {
		h.sendError(msg.ClientID, pm.SessionID, "invalid_subscribe_data", err.Error())
		return
	}

	// Get or create edit session
	sessionInfo, isNew := h.sessionManager.GetOrCreateSession(data.FilePath)

	// Add client to session
	client := &SessionClient{
		ClientID:  msg.ClientID,
		FilePath:  data.FilePath,
		ReadOnly:  data.ReadOnly,
		Connected: true,
	}

	if data.ReadOnly {
		sessionInfo.RefCount.AddReader()
	} else {
		sessionInfo.RefCount.AddWriter()
	}

	sessionInfo.AddClient(msg.ClientID, client)

	// Send snapshot to client
	snapshotData := &SnapshotData{
		SessionID:  sessionInfo.SessionID,
		FilePath:   data.FilePath,
		Content:    sessionInfo.GetContent(),
		Revision:   sessionInfo.GetCurrentVersion(),
		CreatedAt:  sessionInfo.CreatedAt,
		UpdatedAt:  sessionInfo.UpdatedAt,
		Clients:    sessionInfo.GetClientInfos(),
		ReadOnly:   data.ReadOnly,
	}

	if isNew {
		// New session, no operations yet
		snapshotData.Operations = nil
	} else {
		// Existing session, send recent operations
		snapshotData.Operations = sessionInfo.GetRecentOperations()
	}

	h.sendMessage(msg.ClientID, MessageTypeSnapshot, snapshotData)

	// Notify other clients
	h.notifyUserJoined(sessionInfo, msg.ClientID)
}

// handleUnsubscribe handles file unsubscription.
func (h *ProtocolHandler) handleUnsubscribe(msg *Message, pm *ProtocolMessage) {
	var data UnsubscribeData
	if err := json.Unmarshal(pm.Data, &data); err != nil {
		h.sendError(msg.ClientID, pm.SessionID, "invalid_unsubscribe_data", err.Error())
		return
	}

	sessionInfo := h.sessionManager.GetSession(data.SessionID)
	if sessionInfo == nil {
		h.sendError(msg.ClientID, data.SessionID, "session_not_found", "Session not found")
		return
	}

	// Remove client
	client := sessionInfo.RemoveClient(msg.ClientID)
	if client == nil {
		return
	}

	// Update ref count
	if client.ReadOnly {
		sessionInfo.RefCount.RemoveReader()
	} else {
		sessionInfo.RefCount.RemoveWriter()
	}

	// Notify other clients
	h.notifyUserLeft(sessionInfo, msg.ClientID)

	// Check if session should be destroyed
	if sessionInfo.RefCount.ShouldDestroy() {
		h.sessionManager.DestroySession(data.SessionID)
	}
}

// handleStartEditing handles start editing request.
func (h *ProtocolHandler) handleStartEditing(msg *Message, pm *ProtocolMessage) {
	var data StartEditingData
	if err := json.Unmarshal(pm.Data, &data); err != nil {
		h.sendError(msg.ClientID, pm.SessionID, "invalid_start_editing_data", err.Error())
		return
	}

	// Get or create edit session
	sessionInfo, _ := h.sessionManager.GetOrCreateSession(data.FilePath)

	// Add as writer
	sessionInfo.RefCount.AddWriter()

	// Add or update client
	client := &SessionClient{
		ClientID:  msg.ClientID,
		FilePath:  data.FilePath,
		ReadOnly:  false,
		IsEditing: true,
		Connected: true,
	}

	sessionInfo.AddClient(msg.ClientID, client)

	// Send snapshot
	snapshotData := &SnapshotData{
		SessionID: sessionInfo.SessionID,
		FilePath:  data.FilePath,
		Content:   sessionInfo.GetContent(),
		Revision:  sessionInfo.GetCurrentVersion(),
		CreatedAt: sessionInfo.CreatedAt,
		UpdatedAt: sessionInfo.UpdatedAt,
		Clients:   sessionInfo.GetClientInfos(),
		ReadOnly:  false,
	}

	h.sendMessage(msg.ClientID, MessageTypeSnapshot, snapshotData)

	// Notify other clients
	h.notifySessionInfo(sessionInfo)
}

// handleStopEditing handles stop editing request.
func (h *ProtocolHandler) handleStopEditing(msg *Message, pm *ProtocolMessage) {
	var data StopEditingData
	if err := json.Unmarshal(pm.Data, &data); err != nil {
		h.sendError(msg.ClientID, pm.SessionID, "invalid_stop_editing_data", err.Error())
		return
	}

	sessionInfo := h.sessionManager.GetSession(data.SessionID)
	if sessionInfo == nil {
		h.sendError(msg.ClientID, data.SessionID, "session_not_found", "Session not found")
		return
	}

	// Update client
	client := sessionInfo.GetClient(msg.ClientID)
	if client != nil {
		client.IsEditing = false
	}

	// Decrease writer count
	sessionInfo.RefCount.RemoveWriter()

	// Notify other clients
	h.notifySessionInfo(sessionInfo)

	// Check if session should be destroyed (only if no readers either)
	if sessionInfo.RefCount.ShouldDestroy() {
		h.sessionManager.DestroySession(data.SessionID)
	}
}

// handleOperation handles OT operation.
func (h *ProtocolHandler) handleOperation(msg *Message, pm *ProtocolMessage) {
	var data OperationData
	if err := json.Unmarshal(pm.Data, &data); err != nil {
		h.sendError(msg.ClientID, pm.SessionID, "invalid_operation_data", err.Error())
		return
	}

	sessionInfo := h.sessionManager.GetSession(data.SessionID)
	if sessionInfo == nil {
		h.sendError(msg.ClientID, data.SessionID, "session_not_found", "Session not found")
		return
	}

	// Parse operation
	opData, err := ParseOperationData(data.Operation)
	if err != nil {
		h.sendError(msg.ClientID, data.SessionID, "invalid_operation", err.Error())
		return
	}

	// Convert array format to OT operation
	op := h.arrayToOperation(opData)
	if op == nil {
		h.sendError(msg.ClientID, data.SessionID, "invalid_operation", "operation is nil")
		return
	}

	// Debug: log operation details
	docContent := sessionInfo.GetContent()
	maxPreview := 50
	if len(docContent) < maxPreview {
		maxPreview = len(docContent)
	}
	log.Printf("[Handler] Operation details: op.BaseLength()=%d, doc rune count=%d, doc byte len=%d, preview=%q",
		op.BaseLength(), utf8.RuneCountInString(docContent), len(docContent), docContent[:maxPreview])

	// Apply operation to document
	newContent, err := op.Apply(sessionInfo.GetContent())
	if err != nil {
		h.sendError(msg.ClientID, data.SessionID, "operation_failed", err.Error())
		return
	}

	// Update session content snapshot
	sessionInfo.SetContent(newContent)

	// Add operation to history (creates new version)
	if err := sessionInfo.AddOperation(opData, msg.ClientID); err != nil {
		h.sendError(msg.ClientID, data.SessionID, "history_error", err.Error())
		return
	}

	// Send acknowledgment with new version
	ackData := &AckData{
		SessionID: data.SessionID,
		Revision:  sessionInfo.GetCurrentVersion(),
		Timestamp: pm.Timestamp,
	}
	h.sendMessage(msg.ClientID, MessageTypeAck, ackData)

	// Broadcast to other clients
	remoteOpData := &RemoteOperationData{
		SessionID: data.SessionID,
		ClientID:  msg.ClientID,
		Revision:  sessionInfo.GetCurrentVersion(),
		Operation: opData,
		Selection: data.Selection,
	}

	h.broadcastToSession(data.SessionID, msg.ClientID, MessageTypeRemoteOperation, remoteOpData)
}

// handleCursor handles cursor position updates.
func (h *ProtocolHandler) handleCursor(msg *Message, pm *ProtocolMessage) {
	var data CursorData
	if err := json.Unmarshal(pm.Data, &data); err != nil {
		return
	}

	// TODO: Broadcast cursor position to other clients in session
	log.Printf("Client %s cursor moved to %d", msg.ClientID, data.Position)
}

// handleHeartbeat handles heartbeat messages.
func (h *ProtocolHandler) handleHeartbeat(msg *Message, pm *ProtocolMessage) {
	var data HeartbeatData
	if err := json.Unmarshal(pm.Data, &data); err != nil {
		return
	}

	// Update client last seen time
	for _, sessionID := range data.SessionIDs {
		sessionInfo := h.sessionManager.GetSession(sessionID)
		if sessionInfo != nil {
			client := sessionInfo.GetClient(msg.ClientID)
			if client != nil {
				client.LastSeen = pm.Timestamp
			}
		}
	}
}

// ========== Helper Methods ==========

// sendMessage sends a message to a specific client.
func (h *ProtocolHandler) sendMessage(clientID string, msgType MessageType, data interface{}) error {
	log.Printf("[Handler] Sending %s to %s", msgType, clientID)

	pm, err := NewProtocolMessage(msgType, "", data)
	if err != nil {
		log.Printf("[Handler] Failed to create protocol message: %v", err)
		return err
	}

	// Create response message in new protocol format
	response := map[string]interface{}{
		"type":      string(msgType),
		"client_id": clientID,
		"timestamp": pm.Timestamp,
		"metadata": map[string]interface{}{
			"protocol_message": pm,
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Printf("[Handler] Failed to marshal response: %v", err)
		return err
	}

	log.Printf("[Handler] Sending JSON: %s", string(jsonData))

	// Send via WebSocket server using new SendJSON method
	if err := h.server.SendJSON(clientID, jsonData); err != nil {
		log.Printf("[Handler] Failed to send: %v", err)
		return err
	}

	return nil
}

// broadcastToSession broadcasts a message to all clients in a session except sender.
func (h *ProtocolHandler) broadcastToSession(sessionID, excludeClientID string, msgType MessageType, data interface{}) {
	sessionInfo := h.sessionManager.GetSession(sessionID)
	if sessionInfo == nil {
		return
	}

	_, err := NewProtocolMessage(msgType, sessionID, data)
	if err != nil {
		log.Printf("Failed to create protocol message: %v", err)
		return
	}

	for clientID := range sessionInfo.Clients {
		if clientID == excludeClientID {
			continue
		}

		h.sendMessage(clientID, msgType, data)
	}
}

// sendError sends an error message to client.
func (h *ProtocolHandler) sendError(clientID, sessionID, code, message string) {
	errorData := &ErrorData{
		SessionID: sessionID,
		Code:      code,
		Message:   message,
	}
	h.sendMessage(clientID, MessageTypeError, errorData)
}

// notifyUserJoined notifies other clients that a user joined.
func (h *ProtocolHandler) notifyUserJoined(sessionInfo *EditSession, clientID string) {
	client := sessionInfo.GetClient(clientID)
	if client == nil {
		return
	}

	joinData := &UserJoinedData{
		SessionID: sessionInfo.SessionID,
		ClientID:  clientID,
		Client: ClientInfo{
			ClientID:  clientID,
			IsEditing: client.IsEditing,
			UpdatedAt: sessionInfo.UpdatedAt,
		},
	}

	h.broadcastToSession(sessionInfo.SessionID, clientID, MessageTypeUserJoined, joinData)
}

// notifyUserLeft notifies other clients that a user left.
func (h *ProtocolHandler) notifyUserLeft(sessionInfo *EditSession, clientID string) {
	leftData := &UserLeftData{
		SessionID: sessionInfo.SessionID,
		ClientID:  clientID,
	}

	h.broadcastToSession(sessionInfo.SessionID, clientID, MessageTypeUserLeft, leftData)
}

// notifySessionInfo broadcasts session info to all clients.
func (h *ProtocolHandler) notifySessionInfo(sessionInfo *EditSession) {
	infoData := &SessionInfoData{
		SessionID:   sessionInfo.SessionID,
		FilePath:    sessionInfo.FilePath,
		ReaderCount: sessionInfo.RefCount.ReaderCount,
		WriterCount: sessionInfo.RefCount.WriterCount,
		Clients:     sessionInfo.GetClientInfos(),
		IsEditing:   sessionInfo.RefCount.HasWriters(),
	}

	h.broadcastToSession(sessionInfo.SessionID, "", MessageTypeSessionInfo, infoData)
}

// arrayToOperation converts array format to OT Operation.
func (h *ProtocolHandler) arrayToOperation(data []interface{}) *ot.Operation {
	if len(data) == 0 {
		return nil
	}

	builder := ot.NewBuilder()

	for _, item := range data {
		switch v := item.(type) {
		case float64:
			// Retain or delete
			if v > 0 {
				builder.Retain(int(v))
			} else {
				builder.Delete(-int(v))
			}
		case int:
			if v > 0 {
				builder.Retain(v)
			} else {
				builder.Delete(-v)
			}
		case string:
			// Insert
			builder.Insert(v)
		}
	}

	return builder.Build()
}
