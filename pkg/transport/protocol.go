package transport

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ========== Message Types ==========

// MessageType represents the type of WebSocket message.
type MessageType string

const (
	// Client → Server messages
	MessageTypeSubscribe         MessageType = "subscribe"          // 关注文件
	MessageTypeUnsubscribe       MessageType = "unsubscribe"        // 取消关注
	MessageTypeStartEditing      MessageType = "start_editing"     // 开始编辑
	MessageTypeStopEditing       MessageType = "stop_editing"      // 停止编辑
	MessageTypeOperation         MessageType = "operation"          // 发送 OT 操作
	MessageTypeCursor            MessageType = "cursor"             // 光标位置
	MessageTypeHeartbeat         MessageType = "heartbeat"          // 心跳

	// Server → Client messages
	MessageTypeWelcome           MessageType = "welcome"            // 连接成功
	MessageTypeSnapshot          MessageType = "snapshot"           // 文档快照
	MessageTypeSnapshotCreated   MessageType = "snapshot_created"  // 快照已创建（通知Redis）
	MessageTypeRemoteOperation   MessageType = "remote_operation"   // 远程操作
	MessageTypeAck               MessageType = "ack"                // 操作确认
	MessageTypeError             MessageType = "error"              // 错误
	MessageTypeUserJoined        MessageType = "user_joined"        // 用户加入
	MessageTypeUserLeft          MessageType = "user_left"          // 用户离开
	MessageTypeSessionInfo       MessageType = "session_info"       // 会话信息
)

// ========== Protocol Messages ==========

// ProtocolMessage is the base structure for all WebSocket messages.
type ProtocolMessage struct {
	Type      MessageType          `json:"type"`
	SessionID string               `json:"session_id,omitempty"` // Edit session UUID
	Timestamp int64                `json:"timestamp"`
	Data      json.RawMessage      `json:"data,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ========== Client Messages ==========

// SubscribeData represents subscribe request data.
type SubscribeData struct {
	FilePath   string `json:"file_path"`
	ReadOnly   bool   `json:"read_only"`   // true = 只读（可用SSE）
	UseSSE     bool   `json:"use_sse"`     // true = 优先使用SSE推送
	ClientID   string `json:"client_id,omitempty"`
}

// UnsubscribeData represents unsubscribe request data.
type UnsubscribeData struct {
	SessionID string `json:"session_id"` // Edit session UUID
	FilePath  string `json:"file_path,omitempty"`
}

// StartEditingData represents start editing request data.
type StartEditingData struct {
	FilePath    string  `json:"file_path"`
	ContentType string  `json:"content_type,omitempty"` // "text", "markdown", etc.
	InitialText string  `json:"initial_text,omitempty"` // 如果文件不存在，创建时的初始内容
	ClientID    string  `json:"client_id,omitempty"`
}

// StopEditingData represents stop editing request data.
type StopEditingData struct {
	SessionID string `json:"session_id"` // Edit session UUID
}

// OperationData represents OT operation data.
type OperationData struct {
	SessionID string      `json:"session_id"` // Edit session UUID
	Revision  int64       `json:"revision"`    // Document version
	Operation interface{} `json:"operation"`   // OT operation: [5, "Hello", 10, -3]
	Selection *CursorData `json:"selection,omitempty"`
}

// CursorData represents cursor/selection data.
type CursorData struct {
	Position     int `json:"position"`
	SelectionEnd int `json:"selection_end"`
}

// HeartbeatData represents heartbeat data.
type HeartbeatData struct {
	SessionIDs []string `json:"session_ids"` // All sessions client is subscribed to
}

// ========== Server Messages ==========

// WelcomeData represents welcome message data.
type WelcomeData struct {
	ClientID  string `json:"client_id"`
	ServerID  string `json:"server_id"`
	Timestamp int64  `json:"timestamp"`
}

// SnapshotData represents document snapshot data.
type SnapshotData struct {
	SessionID   string      `json:"session_id"`           // Edit session UUID
	FilePath    string      `json:"file_path"`
	Content     string      `json:"content"`               // Current document content
	Revision    int64       `json:"revision"`              // Current version
	CreatedAt   int64       `json:"created_at"`
	UpdatedAt   int64       `json:"updated_at"`
	Operations  interface{} `json:"operations,omitempty"`  // Recent OT operations since last sync
	Clients     []ClientInfo `json:"clients"`               // Other clients in this session
	ReadOnly    bool        `json:"read_only"`             // Whether client has write permission
}

// RemoteOperationData represents remote operation data.
type RemoteOperationData struct {
	SessionID   string      `json:"session_id"`   // Edit session UUID
	ClientID    string      `json:"client_id"`    // Who sent this operation
	Revision    int64       `json:"revision"`     // New document version
	Operation   interface{} `json:"operation"`    // OT operation: [5, "Hello", 10, -3]
	Selection   *CursorData `json:"selection,omitempty"`
}

// AckData represents acknowledgment data.
type AckData struct {
	SessionID string `json:"session_id"` // Edit session UUID
	Revision  int64 `json:"revision"`    // Acknowledged revision
	Timestamp int64 `json:"timestamp"`
}

// ErrorData represents error data.
type ErrorData struct {
	SessionID string `json:"session_id,omitempty"`
	Code      string `json:"code"`       // Error code
	Message   string `json:"message"`    // Human-readable message
	Details   map[string]interface{} `json:"details,omitempty"`
}

// ClientInfo represents information about a connected client.
type ClientInfo struct {
	ClientID   string     `json:"client_id"`
	Name       string     `json:"name,omitempty"`
	Color      string     `json:"color,omitempty"`
	IsEditing  bool       `json:"is_editing"`  // Whether this client is editing (vs just viewing)
	Selection  *CursorData `json:"selection,omitempty"`
	UpdatedAt  int64      `json:"updated_at"`
}

// UserJoinedData represents user joined notification.
type UserJoinedData struct {
	SessionID string      `json:"session_id"`
	ClientID  string      `json:"client_id"`
	Client    ClientInfo  `json:"client"`
}

// UserLeftData represents user left notification.
type UserLeftData struct {
	SessionID string `json:"session_id"`
	ClientID  string `json:"client_id"`
}

// SessionInfoData represents session information.
type SessionInfoData struct {
	SessionID    string       `json:"session_id"`
	FilePath     string       `json:"file_path"`
	ReaderCount  int          `json:"reader_count"`  // Number of read-only subscribers
	WriterCount  int          `json:"writer_count"`  // Number of editors
	Clients      []ClientInfo `json:"clients"`       // All connected clients
	IsEditing    bool         `json:"is_editing"`    // Whether this file is being edited
}

// SnapshotCreatedData represents snapshot creation notification (sent to Redis/History service).
type SnapshotCreatedData struct {
	SessionID   string       `json:"session_id"`   // Edit session UUID
	FilePath    string       `json:"file_path"`    // File path
	VersionID   int64        `json:"version_id"`   // Snapshot version ID
	Content     string       `json:"content"`      // Full text content at snapshot
	Operations  []interface{} `json:"operations"`  // Operations since last snapshot
	CreatedAt   int64        `json:"created_at"`   // Creation timestamp
	CreatedBy   string       `json:"created_by"`   // Client ID who triggered snapshot
}

// ========== Helper Functions ==========

// NewProtocolMessage creates a new protocol message.
func NewProtocolMessage(msgType MessageType, sessionID string, data interface{}) (*ProtocolMessage, error) {
	var rawData json.RawMessage
	if data != nil {
		bytes, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		rawData = bytes
	}

	return &ProtocolMessage{
		Type:      msgType,
		SessionID: sessionID,
		Timestamp: time.Now().Unix(),
		Data:      rawData,
	}, nil
}

// ParseOperationData parses OT operation from JSON.
func ParseOperationData(data interface{}) ([]interface{}, error) {
	// Expected format: [5, "Hello", 10, -3]
	// Or: {"retain": 5, "insert": "Hello", "delete": 3}

	switch v := data.(type) {
	case []interface{}:
		// Array format: [5, "Hello", 10, -3]
		return v, nil
	case map[string]interface{}:
		// Object format: convert to array
		return convertObjectToArray(v)
	default:
		// Try JSON unmarshal
		bytes, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		var result []interface{}
		if err := json.Unmarshal(bytes, &result); err != nil {
			return nil, err
		}
		return result, nil
	}
}

// convertObjectToArray converts object format to array format.
func convertObjectToArray(obj map[string]interface{}) ([]interface{}, error) {
	var result []interface{}

	// Order: retain, insert, delete
	if retain, ok := obj["retain"]; ok {
		result = append(result, retain)
	}
	if insert, ok := obj["insert"]; ok {
		result = append(result, insert)
	}
	if delete, ok := obj["delete"]; ok {
		// Convert to negative number
		if d, ok := delete.(float64); ok {
			result = append(result, -int(d))
		} else if d, ok := delete.(int); ok {
			result = append(result, -d)
		}
	}

	return result, nil
}

// GenerateSessionID generates a new UUID for edit session.
func GenerateSessionID() string {
	return uuid.New().String()
}

// ========== Session Reference Counting ==========

// SessionRefCount manages reference counts for edit sessions.
type SessionRefCount struct {
	SessionID   string `json:"session_id"`
	FilePath    string `json:"file_path"`
	ReaderCount int    `json:"reader_count"` // Read-only subscribers
	WriterCount int    `json:"writer_count"` // Active editors
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// AddReader adds a read-only subscriber.
func (rc *SessionRefCount) AddReader() {
	rc.ReaderCount++
	rc.UpdatedAt = time.Now().Unix()
}

// RemoveReader removes a read-only subscriber.
func (rc *SessionRefCount) RemoveReader() {
	if rc.ReaderCount > 0 {
		rc.ReaderCount--
	}
	rc.UpdatedAt = time.Now().Unix()
}

// AddWriter adds an editor.
func (rc *SessionRefCount) AddWriter() {
	rc.WriterCount++
	rc.UpdatedAt = time.Now().Unix()
}

// RemoveWriter removes an editor.
func (rc *SessionRefCount) RemoveWriter() {
	if rc.WriterCount > 0 {
		rc.WriterCount--
	}
	rc.UpdatedAt = time.Now().Unix()
}

// IsActive returns true if there are any readers or writers.
func (rc *SessionRefCount) IsActive() bool {
	return rc.ReaderCount > 0 || rc.WriterCount > 0
}

// ShouldDestroy returns true if session should be destroyed.
func (rc *SessionRefCount) ShouldDestroy() bool {
	return rc.ReaderCount == 0 && rc.WriterCount == 0
}

// HasWriters returns true if there are active editors.
func (rc *SessionRefCount) HasWriters() bool {
	return rc.WriterCount > 0
}
