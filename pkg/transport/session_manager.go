package transport

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/coreseekdev/texere/pkg/session"
)

// ContentStorage interface for loading file contents.
type ContentStorage interface {
	Get(ctx context.Context, contentPath string, options *session.GetOptions) (*session.ContentModel, error)
}

// ========== History Listener Interface ==========

// HistoryEvent represents a history event that can be sent to Redis/History service.
type HistoryEvent struct {
	SessionID  string        `json:"session_id"`
	FilePath   string        `json:"file_path"`
	EventType  string        `json:"event_type"` // "snapshot" or "operation"
	VersionID  int64         `json:"version_id"`
	Content    string        `json:"content,omitempty"`      // Full content for snapshot
	Operations []interface{} `json:"operations,omitempty"`  // OT operations
	CreatedAt  int64         `json:"created_at"`
	CreatedBy  string        `json:"created_by"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"` // Additional metadata (patches, etc.)
}

// HistoryListener listens to edit session events and forwards to Redis/History service.
type HistoryListener interface {
	// OnSnapshot is called when a new snapshot is created.
	OnSnapshot(event *HistoryEvent) error

	// OnOperation is called when a new operation is applied.
	OnOperation(event *HistoryEvent) error

	// Close closes the history listener.
	Close() error
}

// ========== Edit Session ==========

// EditSession represents an active editing session for a file.
// Only keeps: 1 snapshot + recent changes (older history forwarded to Redis).
type EditSession struct {
	SessionID   string                       // UUID
	FilePath    string                       // File path
	RefCount    *SessionRefCount              // Reader/Writer counts
	CreatedAt   int64                        // Session creation time
	UpdatedAt   int64                        // Last update time
	Clients     map[string]*SessionClient    // Connected clients (clientID -> client)
	mu          sync.RWMutex

	// Current snapshot (always exactly 1)
	snapshotContent string   // Current full content snapshot
	snapshotVersion int64    // Version ID of current snapshot

	// Recent changes (in-memory only, forwarded to Redis)
	recentChanges []interface{} // Recent OT operations since last snapshot
	currentVersion int64        // Current version number

	// History listener (forwards to Redis/History service)
	historyListener HistoryListener

	// Snapshot creation settings
	maxChangesBeforeSnapshot int // Max changes before forcing snapshot creation
	lastSnapshotTime          int64 // Timestamp of last snapshot
	maxSnapshotInterval       int64 // Max time between snapshots (seconds)
}

const (
	// DefaultMaxChangesBeforeSnapshot is the default max changes before creating a new snapshot.
	DefaultMaxChangesBeforeSnapshot = 200
	// DefaultMaxSnapshotInterval is the default max time between snapshots (5 minutes)
	DefaultMaxSnapshotInterval = 300 // 5 minutes = 300 seconds
)

// NewEditSession creates a new edit session with snapshot + changes structure.
func NewEditSession(sessionID, filePath string, initialContent string) *EditSession {
	now := time.Now().Unix()
	return &EditSession{
		SessionID:                 sessionID,
		FilePath:                  filePath,
		RefCount: &SessionRefCount{
			SessionID:   sessionID,
			FilePath:    filePath,
			ReaderCount: 0,
			WriterCount: 0,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		CreatedAt:                 now,
		UpdatedAt:                 now,
		Clients:                   make(map[string]*SessionClient),
		snapshotContent:           initialContent,
		snapshotVersion:           0,
		recentChanges:             make([]interface{}, 0),
		currentVersion:            0,
		maxChangesBeforeSnapshot:  DefaultMaxChangesBeforeSnapshot,
		lastSnapshotTime:          now,
		maxSnapshotInterval:       DefaultMaxSnapshotInterval,
	}
}

// GetContent returns the current document content (from snapshot).
func (es *EditSession) GetContent() string {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.snapshotContent
}

// SetContent sets the current document content and updates snapshot.
func (es *EditSession) SetContent(content string) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.snapshotContent = content
	es.UpdatedAt = time.Now().Unix()
}

// SetHistoryListener sets the history listener for forwarding to Redis.
func (es *EditSession) SetHistoryListener(listener HistoryListener) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.historyListener = listener
}

// AddOperation adds an operation to recent changes and forwards to history listener.
func (es *EditSession) AddOperation(operation interface{}, clientID string) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	es.currentVersion++
	es.UpdatedAt = time.Now().Unix()

	// Add to recent changes
	es.recentChanges = append(es.recentChanges, operation)

	// Forward to history listener (Redis/History service)
	if es.historyListener != nil {
		event := &HistoryEvent{
			SessionID:  es.SessionID,
			FilePath:   es.FilePath,
			EventType:  "operation",
			VersionID:  es.currentVersion,
			Operations: []interface{}{operation},
			CreatedAt:  es.UpdatedAt,
			CreatedBy:  clientID,
		}
		// Non-blocking send to avoid blocking the editing operation
		go es.historyListener.OnOperation(event)
	}

	// Check if we need to create a new snapshot
	// Condition 1: Operation count threshold
	// Condition 2: Time threshold (timeout snapshot)
	if len(es.recentChanges) >= es.maxChangesBeforeSnapshot ||
		es.shouldCreateTimeoutSnapshot() {
		es.createSnapshot(clientID)
	}

	return nil
}

// shouldCreateTimeoutSnapshot checks if enough time has passed to create a timeout snapshot.
func (es *EditSession) shouldCreateTimeoutSnapshot() bool {
	if es.lastSnapshotTime == 0 {
		return false
	}

	elapsed := es.UpdatedAt - es.lastSnapshotTime
	return elapsed >= es.maxSnapshotInterval
}

// createSnapshot creates a new snapshot from the current content and clears recent changes.
func (es *EditSession) createSnapshot(clientID string) {
	// Current snapshot content becomes the new snapshot
	es.snapshotVersion = es.currentVersion
	es.lastSnapshotTime = time.Now().Unix()

	// Store snapshot content and operations for forwarding
	snapshotContent := es.snapshotContent
	operationsSinceSnapshot := make([]interface{}, len(es.recentChanges))
	copy(operationsSinceSnapshot, es.recentChanges)

	// Forward snapshot event to history listener with FULL TEXT CONTENT
	if es.historyListener != nil {
		event := &HistoryEvent{
			SessionID:  es.SessionID,
			FilePath:   es.FilePath,
			EventType:  "snapshot",
			VersionID:  es.snapshotVersion,
			Content:    snapshotContent, // Full text content
			Operations: operationsSinceSnapshot,
			CreatedAt:  es.lastSnapshotTime,
			CreatedBy:  clientID,
		}
		go es.historyListener.OnSnapshot(event)
	}

	// Clear recent changes after snapshot
	es.recentChanges = make([]interface{}, 0)
}

// GetRecentOperations returns recent operations since last snapshot.
func (es *EditSession) GetRecentOperations() []interface{} {
	es.mu.RLock()
	defer es.mu.RUnlock()

	ops := make([]interface{}, len(es.recentChanges))
	copy(ops, es.recentChanges)
	return ops
}

// GetSnapshotVersion returns the current snapshot version ID.
func (es *EditSession) GetSnapshotVersion() int64 {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.snapshotVersion
}

// GetCurrentVersion returns the current version number.
func (es *EditSession) GetCurrentVersion() int64 {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.currentVersion
}

// SetMaxChangesBeforeSnapshot sets the max changes before creating a new snapshot.
func (es *EditSession) SetMaxChangesBeforeSnapshot(max int) {
	es.mu.Lock()
	defer es.mu.Unlock()

	if max < 1 {
		max = 1
	}
	es.maxChangesBeforeSnapshot = max
}

// SetMaxSnapshotInterval sets the max time interval between snapshots (in seconds).
func (es *EditSession) SetMaxSnapshotInterval(interval int64) {
	es.mu.Lock()
	defer es.mu.Unlock()

	if interval < 60 {
		interval = 60 // Minimum 1 minute
	}
	es.maxSnapshotInterval = interval
}

// GetSnapshotInfo returns information about snapshot status.
func (es *EditSession) GetSnapshotInfo() *SnapshotInfo {
	es.mu.RLock()
	defer es.mu.RUnlock()

	var timeUntilSnapshot int64
	if es.lastSnapshotTime > 0 {
		elapsed := time.Now().Unix() - es.lastSnapshotTime
		remaining := es.maxSnapshotInterval - elapsed
		if remaining < 0 {
			remaining = 0
		}
		timeUntilSnapshot = remaining
	}

	return &SnapshotInfo{
		SnapshotVersion:        es.snapshotVersion,
		LastSnapshotTime:      es.lastSnapshotTime,
		RecentChangeCount:      len(es.recentChanges),
		MaxChangesBeforeSnapshot: es.maxChangesBeforeSnapshot,
		MaxSnapshotInterval:    es.maxSnapshotInterval,
		TimeUntilSnapshot:      timeUntilSnapshot,
	}
}

// GetSessionInfo returns information about the session.
func (es *EditSession) GetSessionInfo() *SessionInfo {
	es.mu.RLock()
	defer es.mu.RUnlock()

	return &SessionInfo{
		SessionID:         es.SessionID,
		FilePath:          es.FilePath,
		Content:           es.snapshotContent,
		Revision:          es.currentVersion,
		SnapshotVersion:   es.snapshotVersion,
		RecentChangeCount: len(es.recentChanges),
		ReaderCount:       es.RefCount.ReaderCount,
		WriterCount:       es.RefCount.WriterCount,
		CreatedAt:         es.CreatedAt,
		UpdatedAt:         es.UpdatedAt,
	}
}

// SessionInfo contains session information.
type SessionInfo struct {
	SessionID         string `json:"session_id"`
	FilePath          string `json:"file_path"`
	Content           string `json:"content"`
	Revision          int64  `json:"revision"`
	SnapshotVersion   int64  `json:"snapshot_version"`
	RecentChangeCount int    `json:"recent_change_count"`
	ReaderCount       int    `json:"reader_count"`
	WriterCount       int    `json:"writer_count"`
	CreatedAt         int64  `json:"created_at"`
	UpdatedAt         int64  `json:"updated_at"`
}

// ========== Client Management ==========

// AddClient adds a client to the session.
func (es *EditSession) AddClient(clientID string, client *SessionClient) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.Clients[clientID] = client
	es.UpdatedAt = time.Now().Unix()
}

// RemoveClient removes a client from the session.
func (es *EditSession) RemoveClient(clientID string) *SessionClient {
	es.mu.Lock()
	defer es.mu.Unlock()

	client := es.Clients[clientID]
	if client != nil {
		delete(es.Clients, clientID)
		es.UpdatedAt = time.Now().Unix()
	}
	return client
}

// GetClient retrieves a client by ID.
func (es *EditSession) GetClient(clientID string) *SessionClient {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.Clients[clientID]
}

// GetClientInfos returns information about all connected clients.
func (es *EditSession) GetClientInfos() []ClientInfo {
	es.mu.RLock()
	defer es.mu.RUnlock()

	infos := make([]ClientInfo, 0, len(es.Clients))
	for _, client := range es.Clients {
		infos = append(infos, ClientInfo{
			ClientID:  client.ClientID,
			IsEditing: client.IsEditing,
			UpdatedAt: client.LastSeen,
		})
	}
	return infos
}

// ========== Session Manager ==========

// SessionManager manages multiple edit sessions.
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*EditSession // sessionID -> EditSession
	byPath   map[string]string       // filePath -> sessionID

	// Content storage for loading file contents
	contentStorage ContentStorage

	// Global history listener for all sessions
	historyListener HistoryListener
}

// NewSessionManager creates a new session manager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*EditSession),
		byPath:   make(map[string]string),
	}
}

// SetContentStorage sets the content storage for loading files.
func (sm *SessionManager) SetContentStorage(storage ContentStorage) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.contentStorage = storage
}

// NewSessionManagerWithHistory creates a new session manager with a history service.
// This is the preferred way to create a session manager with history support.
//
// Example:
//   // Using Redis history service with patch mode
//   historySvc := NewRedisHistoryServiceWithOpts(redisClient, true)
//   sm := NewSessionManagerWithHistory(historySvc)
//
//   // Using in-memory history service
//   historySvc := NewMemoryHistoryService(false)
//   sm := NewSessionManagerWithHistory(historySvc)
//
//   // Using history service factory
//   historySvc := NewHistoryService(&HistoryOptions{
//       StorageBackend: "redis",
//       UsePatchMode: true,
//   })
//   sm := NewSessionManagerWithHistory(historySvc)
func NewSessionManagerWithHistory(history HistoryListener) *SessionManager {
	return &SessionManager{
		sessions:        make(map[string]*EditSession),
		byPath:          make(map[string]string),
		historyListener: history,
	}
}

// SetHistoryListener sets the history listener for all sessions.
func (sm *SessionManager) SetHistoryListener(listener HistoryListener) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.historyListener = listener

	// Set listener for existing sessions
	for _, session := range sm.sessions {
		session.SetHistoryListener(listener)
	}
}

// GetOrCreateSession gets an existing session or creates a new one.
func (sm *SessionManager) GetOrCreateSession(filePath string) (*EditSession, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if session exists for this file
	if sessionID, ok := sm.byPath[filePath]; ok {
		if session, ok := sm.sessions[sessionID]; ok {
			return session, false
		}
		// Cleanup orphaned byPath entry
		delete(sm.byPath, filePath)
	}

	// Load content from storage if available
	content := ""
	if sm.contentStorage != nil {
		// Try to load from ContentStorage
		ctx := context.Background()
		var options *session.GetOptions
		if model, err := sm.contentStorage.Get(ctx, filePath, options); err == nil && model != nil {
			content = model.Content
		}
	}

	// Create new session with UUID
	sessionID := uuid.New().String()
	session := NewEditSession(sessionID, filePath, content)

	// Set history listener if available
	if sm.historyListener != nil {
		session.SetHistoryListener(sm.historyListener)
	}

	sm.sessions[sessionID] = session
	sm.byPath[filePath] = sessionID

	return session, true
}

// GetSession retrieves a session by ID.
func (sm *SessionManager) GetSession(sessionID string) *EditSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sessions[sessionID]
}

// GetSessionByPath retrieves a session by file path.
func (sm *SessionManager) GetSessionByPath(filePath string) *EditSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sessionID, ok := sm.byPath[filePath]; ok {
		return sm.sessions[sessionID]
	}
	return nil
}

// DestroySession destroys a session.
func (sm *SessionManager) DestroySession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return
	}

	// Remove from path index
	delete(sm.byPath, session.FilePath)

	// Remove from sessions
	delete(sm.sessions, sessionID)
}

// ListSessions returns all active sessions.
func (sm *SessionManager) ListSessions() []*EditSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*EditSession, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// ========== Session Client ==========

// SessionClient represents a client in an edit session.
type SessionClient struct {
	ClientID  string       // Client ID
	FilePath  string       // File path
	ReadOnly  bool         // Whether client is read-only
	IsEditing bool         // Whether client is actively editing
	Connected bool         // Whether client is connected
	Selection *CursorData // Current cursor/selection
	LastSeen  int64        // Last activity timestamp
}

// GetClientID returns the client ID.
func (sc *SessionClient) GetID() string {
	return sc.ClientID
}
