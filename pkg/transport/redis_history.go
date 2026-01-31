package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// ========== Redis History Service ==========

// RedisHistoryService implements HistoryService and forwards events to Redis.
// Falls back to MiniRedis if Redis is not available.
type RedisHistoryService struct {
	mu            sync.RWMutex
	redisClient   RedisClient // Can be real Redis or MiniRedis
	sessionEvents map[string][]*HistoryEvent // sessionID -> events
	eventChan     chan *HistoryEvent
	closed        bool
	wg            sync.WaitGroup
	closeChan     chan struct{}
	usePatchMode  bool // If true, use patch-based storage (like HedgeDoc)
	patchManager  *PatchManager // Handles diff-match-patch operations
}

// RedisClient defines the interface for Redis operations.
type RedisClient interface {
	// Set stores a key-value pair with optional TTL.
	Set(key string, value interface{}, ttl time.Duration) error

	// Get retrieves a value by key.
	Get(key string) (string, error)

	// LPush adds an element to the left of a list.
	LPush(key string, values ...interface{}) error

	// LRange retrieves a range of elements from a list.
	LRange(key string, start, stop int64) ([]string, error)

	// Publish publishes a message to a channel.
	Publish(channel string, message interface{}) error

	// Close closes the Redis connection.
	Close() error
}

// NewRedisHistoryService creates a new Redis history service.
// If redisClient is nil, uses MiniRedis as fallback.
// By default, does NOT use patch mode (stores full content for simplicity).
// To enable patch mode, use NewRedisHistoryServiceWithOpts().
func NewRedisHistoryService(redisClient RedisClient) *RedisHistoryService {
	if redisClient == nil {
		redisClient = NewMiniRedis()
	}

	service := &RedisHistoryService{
		redisClient:   redisClient,
		sessionEvents: make(map[string][]*HistoryEvent),
		eventChan:     make(chan *HistoryEvent, 1000),
		closeChan:     make(chan struct{}),
		usePatchMode:  false, // Default: full content mode
		patchManager:  NewPatchManager(),
	}

	// Start event processor
	service.wg.Add(1)
	go service.processEvents()

	return service
}

// NewRedisHistoryServiceWithOpts creates a new Redis history service with options.
func NewRedisHistoryServiceWithOpts(redisClient RedisClient, usePatchMode bool) *RedisHistoryService {
	if redisClient == nil {
		redisClient = NewMiniRedis()
	}

	service := &RedisHistoryService{
		redisClient:   redisClient,
		sessionEvents: make(map[string][]*HistoryEvent),
		eventChan:     make(chan *HistoryEvent, 1000),
		closeChan:     make(chan struct{}),
		usePatchMode:  usePatchMode,
		patchManager:  NewPatchManager(),
	}

	// Start event processor
	service.wg.Add(1)
	go service.processEvents()

	return service
}

// OnSnapshot handles snapshot events from edit sessions.
func (s *RedisHistoryService) OnSnapshot(event *HistoryEvent) error {
	if s.closed {
		return fmt.Errorf("history service is closed")
	}

	// Send to event channel for async processing
	select {
	case s.eventChan <- event:
		return nil
	default:
		log.Printf("Warning: history service event channel full, dropping snapshot event for session %s", event.SessionID)
		return fmt.Errorf("event channel full")
	}
}

// OnOperation handles operation events from edit sessions.
func (s *RedisHistoryService) OnOperation(event *HistoryEvent) error {
	if s.closed {
		return fmt.Errorf("history service is closed")
	}

	// Send to event channel for async processing
	select {
	case s.eventChan <- event:
		return nil
	default:
		log.Printf("Warning: history service event channel full, dropping operation event for session %s", event.SessionID)
		return fmt.Errorf("event channel full")
	}
}

// processEvents processes history events in the background.
func (s *RedisHistoryService) processEvents() {
	defer s.wg.Done()

	for {
		select {
		case <-s.closeChan:
			return
		case event := <-s.eventChan:
			s.handleEvent(event)
		}
	}
}

// handleEvent handles a single history event.
func (s *RedisHistoryService) handleEvent(event *HistoryEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Skip processing if service is closed
	if s.closed {
		return
	}

	switch event.EventType {
	case "snapshot":
		if s.usePatchMode {
			s.storeSnapshotWithPatch(event)
		} else {
			s.storeSnapshot(event)
		}
	case "operation":
		s.storeOperation(event)
	default:
		log.Printf("Warning: unknown event type: %s", event.EventType)
	}
}

// storeSnapshot stores a snapshot in Redis.
func (s *RedisHistoryService) storeSnapshot(event *HistoryEvent) {
	// Store snapshot content
	snapshotKey := fmt.Sprintf("snapshot:%s:%d", event.SessionID, event.VersionID)

	snapshotData := map[string]interface{}{
		"version_id":  event.VersionID,
		"content":     event.Content,
		"operations":  event.Operations,
		"created_at":  event.CreatedAt,
		"created_by":  event.CreatedBy,
	}

	if err := s.redisClient.Set(snapshotKey, snapshotData, 0); err != nil {
		log.Printf("Error storing snapshot in Redis: %v", err)
		return
	}

	// Add to session's snapshot list
	listKey := fmt.Sprintf("snapshots:%s", event.SessionID)
	if err := s.redisClient.LPush(listKey, snapshotData); err != nil {
		log.Printf("Error adding snapshot to list: %v", err)
	}

	// Publish snapshot event for real-time notifications
	pubKey := fmt.Sprintf("session:%s:snapshots", event.SessionID)
	if err := s.redisClient.Publish(pubKey, snapshotData); err != nil {
		log.Printf("Error publishing snapshot event: %v", err)
	}

	// Store in memory for quick access
	s.sessionEvents[event.SessionID] = append(s.sessionEvents[event.SessionID], event)

	log.Printf("Stored snapshot for session %s, version %d", event.SessionID, event.VersionID)
}

// storeSnapshotWithPatch stores a snapshot using patch-based storage (HedgeDoc-style).
// Only stores the diff (patch) from the previous snapshot, not the full content.
func (s *RedisHistoryService) storeSnapshotWithPatch(event *HistoryEvent) {
	// 1. Get last snapshot to compute patch
	lastSnapshotKey := fmt.Sprintf("snapshot:%s:%d", event.SessionID, event.VersionID-1)
	lastSnapshotData, err := s.redisClient.Get(lastSnapshotKey)
	var lastContent string

	if err == nil && lastSnapshotData != "" {
		// Try to parse as last snapshot
		var lastSnapshot map[string]interface{}
		if err := json.Unmarshal([]byte(lastSnapshotData), &lastSnapshot); err == nil {
			if content, ok := lastSnapshot["content"].(string); ok && content != "" {
				// Previous snapshot has full content
				lastContent = content
			} else if patch, ok := lastSnapshot["patch"].(string); ok && patch != "" {
				// Previous snapshot is in patch mode, reconstruct content
				if lastContentStr, ok := lastSnapshot["last_content"].(string); ok {
					// Reconstruct from previous patch
					result := s.patchManager.ApplyPatch(lastContentStr, patch)
					if result.Success {
						lastContent = result.Content
					}
				}
			}
		}
	}

	// 2. Compute patch using diff-match-patch if we have previous content
	var patch string
	var savedBytes int
	if lastContent != "" && lastContent != event.Content {
		// Use real diff-match-patch algorithm
		patchResult := s.patchManager.ComputePatch(lastContent, event.Content)
		patch = patchResult.Patch
		savedBytes = patchResult.SavedBytes
	} else if event.Content != "" {
		// First snapshot - no patch needed
		patch = ""
		savedBytes = 0
	}

	// 3. Store snapshot with patch or full content
	snapshotData := map[string]interface{}{
		"version_id":  event.VersionID,
		"patch":       patch,
		"content":     "", // Don't store full content in patch mode (except first snapshot)
		"last_content": lastContent, // Keep reference to previous content
		"operations":  event.Operations,
		"created_at":  event.CreatedAt,
		"created_by":  event.CreatedBy,
	}

	// If this is the first snapshot or we don't have previous content, store full content
	if lastContent == "" {
		snapshotData["content"] = event.Content
		snapshotData["patch"] = ""
		savedBytes = 0
	}

	snapshotKey := fmt.Sprintf("snapshot:%s:%d", event.SessionID, event.VersionID)

	if err := s.redisClient.Set(snapshotKey, snapshotData, 0); err != nil {
		log.Printf("Error storing patch-based snapshot in Redis: %v", err)
		return
	}

	// Add to session's snapshot list
	listKey := fmt.Sprintf("snapshots:%s", event.SessionID)
	snapshotJSON, _ := json.Marshal(snapshotData)
	if err := s.redisClient.LPush(listKey, snapshotJSON); err != nil {
		log.Printf("Error adding snapshot to list: %v", err)
	}

	// Publish snapshot event
	pubKey := fmt.Sprintf("session:%s:snapshots", event.SessionID)
	if err := s.redisClient.Publish(pubKey, snapshotData); err != nil {
		log.Printf("Error publishing snapshot event: %v", err)
	}

	// Store in memory
	s.sessionEvents[event.SessionID] = append(s.sessionEvents[event.SessionID], event)

	log.Printf("Stored patch-based snapshot for session %s, version %d (patch: %d bytes, saved ~%d bytes)",
		event.SessionID, event.VersionID, len(patch), savedBytes)
}

// storeOperation stores an operation in Redis.
func (s *RedisHistoryService) storeOperation(event *HistoryEvent) {
	// Store operation data
	opKey := fmt.Sprintf("operation:%s:%d", event.SessionID, event.VersionID)

	opData := map[string]interface{}{
		"version_id": event.VersionID,
		"operations": event.Operations,
		"created_at": event.CreatedAt,
		"created_by": event.CreatedBy,
	}

	if err := s.redisClient.Set(opKey, opData, 0); err != nil {
		log.Printf("Error storing operation in Redis: %v", err)
		return
	}

	// Add to session's operation list
	listKey := fmt.Sprintf("operations:%s", event.SessionID)
	if err := s.redisClient.LPush(listKey, opData); err != nil {
		log.Printf("Error adding operation to list: %v", err)
	}

	// Publish operation event for real-time notifications
	pubKey := fmt.Sprintf("session:%s:operations", event.SessionID)
	if err := s.redisClient.Publish(pubKey, opData); err != nil {
		log.Printf("Error publishing operation event: %v", err)
	}

	// Store in memory for quick access
	s.sessionEvents[event.SessionID] = append(s.sessionEvents[event.SessionID], event)
}

// GetSessionHistory retrieves history for a session from Redis.
func (s *RedisHistoryService) GetSessionHistory(ctx context.Context, sessionID string, limit int64) ([]*HistoryEvent, error) {
	listKey := fmt.Sprintf("operations:%s", sessionID)

	values, err := s.redisClient.LRange(listKey, 0, limit-1)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}

	events := make([]*HistoryEvent, 0, len(values))
	for _, value := range values {
		var event HistoryEvent
		if err := json.Unmarshal([]byte(value), &event); err != nil {
			log.Printf("Error unmarshaling history event: %v", err)
			continue
		}
		events = append(events, &event)
	}

	return events, nil
}

// GetSnapshot retrieves a specific snapshot from Redis.
func (s *RedisHistoryService) GetSnapshot(ctx context.Context, sessionID string, versionID int64) (*HistoryEvent, error) {
	snapshotKey := fmt.Sprintf("snapshot:%s:%d", sessionID, versionID)

	value, err := s.redisClient.Get(snapshotKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	var event HistoryEvent
	if err := json.Unmarshal([]byte(value), &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}

	return &event, nil
}

// ReconstructSnapshot reconstructs the content of a specific version by applying patches.
// This is necessary when using patch mode, where only the first snapshot has full content.
// Starting from version 0, it applies all patches up to the target version.
func (s *RedisHistoryService) ReconstructSnapshot(ctx context.Context, sessionID string, targetVersionID int64) (string, error) {
	// Start with empty content
	content := ""

	// Iterate from version 0 to target version
	for version := int64(0); version <= targetVersionID; version++ {
		snapshotKey := fmt.Sprintf("snapshot:%s:%d", sessionID, version)
		snapshotData, err := s.redisClient.Get(snapshotKey)
		if err != nil {
			return "", fmt.Errorf("failed to get snapshot version %d: %w", version, err)
		}

		// Parse snapshot data
		var snapshot map[string]interface{}
		if err := json.Unmarshal([]byte(snapshotData), &snapshot); err != nil {
			return "", fmt.Errorf("failed to parse snapshot version %d: %w", version, err)
		}

		// Check if snapshot has full content
		if contentValue, ok := snapshot["content"].(string); ok && contentValue != "" {
			// Use full content directly
			content = contentValue
		} else if patchValue, ok := snapshot["patch"].(string); ok && patchValue != "" {
			// Apply patch to current content
			result := s.patchManager.ApplyPatch(content, patchValue)
			if !result.Success {
				return "", fmt.Errorf("failed to apply patch for version %d", version)
			}
			content = result.Content
		}
	}

	return content, nil
}

// ListSnapshots lists all snapshots for a session.
func (s *RedisHistoryService) ListSnapshots(ctx context.Context, sessionID string) ([]*SnapshotInfo, error) {
	listKey := fmt.Sprintf("snapshots:%s", sessionID)

	values, err := s.redisClient.LRange(listKey, 0, -1)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	infos := make([]*SnapshotInfo, 0, len(values))
	for _, value := range values {
		var event HistoryEvent
		if err := json.Unmarshal([]byte(value), &event); err != nil {
			log.Printf("Error unmarshaling snapshot: %v", err)
			continue
		}

		infos = append(infos, &SnapshotInfo{
			SnapshotVersion:   event.VersionID,
			LastSnapshotTime:  event.CreatedAt,
			RecentChangeCount: len(event.Operations),
		})
	}

	return infos, nil
}

// Close closes the history service.
func (s *RedisHistoryService) Close() error {
	// Set closed flag and close channel to stop goroutine
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()

	// Signal goroutine to stop (must be done before Wait())
	close(s.closeChan)

	// Wait for event processor to finish
	// Note: This must be done WITHOUT holding the mutex, otherwise
	// the goroutine can't acquire the mutex to finish processing
	s.wg.Wait()

	// Now it's safe to do cleanup while holding the mutex
	s.mu.Lock()
	defer s.mu.Unlock()

	// Close Redis connection
	if err := s.redisClient.Close(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}

	close(s.eventChan)

	return nil
}

// ========== Mini-Redis Implementation ==========

// MiniRedis is an in-memory implementation of RedisClient for testing/fallback.
type MiniRedis struct {
	mu       sync.RWMutex
	data     map[string]string              // String values
	lists    map[string][]string            // List values
	subs     map[string][]chan string       // Pub/Sub subscribers
	closed   bool
}

// NewMiniRedis creates a new MiniRedis instance.
func NewMiniRedis() *MiniRedis {
	return &MiniRedis{
		data:  make(map[string]string),
		lists: make(map[string][]string),
		subs:  make(map[string][]chan string),
	}
}

// Set stores a key-value pair.
func (m *MiniRedis) Set(key string, value interface{}, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return fmt.Errorf("miniredis is closed")
	}

	// Serialize value to JSON
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	m.data[key] = string(jsonBytes)
	return nil
}

// Get retrieves a value by key.
func (m *MiniRedis) Get(key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return "", fmt.Errorf("miniredis is closed")
	}

	value, ok := m.data[key]
	if !ok {
		return "", fmt.Errorf("key not found: %s", key)
	}

	return value, nil
}

// LPush adds an element to the left of a list.
func (m *MiniRedis) LPush(key string, values ...interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return fmt.Errorf("miniredis is closed")
	}

	// Serialize values to JSON
	for _, value := range values {
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		m.lists[key] = append([]string{string(jsonBytes)}, m.lists[key]...)
	}

	return nil
}

// LRange retrieves a range of elements from a list.
func (m *MiniRedis) LRange(key string, start, stop int64) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, fmt.Errorf("miniredis is closed")
	}

	list, ok := m.lists[key]
	if !ok {
		return []string{}, nil
	}

	// Handle negative indices (like Redis)
	if stop < 0 {
		stop = int64(len(list)) + stop + 1
	}
	if stop >= int64(len(list)) {
		stop = int64(len(list)) - 1
	}

	if start >= int64(len(list)) {
		return []string{}, nil
	}

	return list[start:stop+1], nil
}

// Publish publishes a message to a channel.
func (m *MiniRedis) Publish(channel string, message interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return fmt.Errorf("miniredis is closed")
	}

	// Serialize message to JSON
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send to all subscribers
	subs, ok := m.subs[channel]
	if !ok {
		return nil // No subscribers, that's fine
	}

	for _, ch := range subs {
		select {
		case ch <- string(jsonBytes):
		default:
			// Channel full, skip
		}
	}

	return nil
}

// Subscribe subscribes to a channel (MiniRedis extension).
func (m *MiniRedis) Subscribe(channel string) <-chan string {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan string, 100)
	m.subs[channel] = append(m.subs[channel], ch)
	return ch
}

// Close closes the MiniRedis connection.
func (m *MiniRedis) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}

	m.closed = true

	// Close all subscriber channels
	for _, subs := range m.subs {
		for _, ch := range subs {
			close(ch)
		}
	}

	return nil
}

// GetData returns all stored data (for testing).
func (m *MiniRedis) GetData() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data := make(map[string]string)
	for k, v := range m.data {
		data[k] = v
	}
	return data
}

// GetLists returns all stored lists (for testing).
func (m *MiniRedis) GetLists() map[string][]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lists := make(map[string][]string)
	for k, v := range m.lists {
		listCopy := make([]string, len(v))
		copy(listCopy, v)
		lists[k] = listCopy
	}
	return lists
}
