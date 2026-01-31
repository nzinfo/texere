package transport

import (
	"context"
	"fmt"
	"sync"
)

// MemoryHistoryService provides an in-memory history service implementation.
// Useful for testing and single-instance deployments.
type MemoryHistoryService struct {
	mu            sync.RWMutex
	snapshots     map[string]map[int64]*HistoryEvent // sessionID -> versionID -> event
	operations    map[string][]*HistoryEvent         // sessionID -> operations
	eventChan     chan *HistoryEvent
	closed        bool
	wg            sync.WaitGroup
	closeChan     chan struct{}
	usePatchMode  bool
	patchManager  *PatchManager
}

// NewMemoryHistoryService creates a new in-memory history service.
func NewMemoryHistoryService(usePatchMode bool) *MemoryHistoryService {
	return &MemoryHistoryService{
		snapshots:    make(map[string]map[int64]*HistoryEvent),
		operations:   make(map[string][]*HistoryEvent),
		eventChan:    make(chan *HistoryEvent, 1000),
		closeChan:    make(chan struct{}),
		usePatchMode: usePatchMode,
		patchManager: NewPatchManager(),
	}
}

// OnSnapshot handles snapshot events.
func (s *MemoryHistoryService) OnSnapshot(event *HistoryEvent) error {
	if s.closed {
		return fmt.Errorf("history service is closed")
	}

	select {
	case s.eventChan <- event:
		return nil
	default:
		return fmt.Errorf("event channel full")
	}
}

// OnOperation handles operation events.
func (s *MemoryHistoryService) OnOperation(event *HistoryEvent) error {
	if s.closed {
		return fmt.Errorf("history service is closed")
	}

	select {
	case s.eventChan <- event:
		return nil
	default:
		return fmt.Errorf("event channel full")
	}
}

// processEvents processes history events in the background.
func (s *MemoryHistoryService) processEvents() {
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
func (s *MemoryHistoryService) handleEvent(event *HistoryEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	switch event.EventType {
	case "snapshot":
		// Initialize snapshot map for session if needed
		if s.snapshots[event.SessionID] == nil {
			s.snapshots[event.SessionID] = make(map[int64]*HistoryEvent)
		}

		// Store snapshot
		if s.usePatchMode && len(s.snapshots[event.SessionID]) > 0 {
			// Compute patch from previous snapshot
			lastVersion := s.findLastSnapshotVersion(event.SessionID)
			if lastSnapshot, ok := s.snapshots[event.SessionID][lastVersion]; ok {
				// Compute and store patch
				patchResult := s.patchManager.ComputePatch(lastSnapshot.Content, event.Content)
				event.Content = "" // Clear content
				// Store patch in metadata
				if event.Metadata == nil {
					event.Metadata = make(map[string]interface{})
				}
				event.Metadata["patch"] = patchResult.Patch
				event.Metadata["last_content"] = lastSnapshot.Content
			}
		}

		s.snapshots[event.SessionID][event.VersionID] = event

	case "operation":
		s.operations[event.SessionID] = append(s.operations[event.SessionID], event)
	}
}

// findLastSnapshotVersion finds the last snapshot version for a session.
func (s *MemoryHistoryService) findLastSnapshotVersion(sessionID string) int64 {
	maxVersion := int64(-1)
	for version := range s.snapshots[sessionID] {
		if version > maxVersion {
			maxVersion = version
		}
	}
	return maxVersion
}

// GetSnapshot retrieves a specific snapshot.
func (s *MemoryHistoryService) GetSnapshot(ctx context.Context, sessionID string, versionID int64) (*HistoryEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshots, ok := s.snapshots[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	event, ok := snapshots[versionID]
	if !ok {
		return nil, fmt.Errorf("snapshot not found")
	}

	// Reconstruct content if using patch mode
	if s.usePatchMode && event.Content == "" {
		content, err := s.ReconstructSnapshot(ctx, sessionID, versionID)
		if err != nil {
			return nil, err
		}
		// Return a copy with reconstructed content
		result := *event
		result.Content = content
		return &result, nil
	}

	return event, nil
}

// GetSessionHistory retrieves history for a session.
func (s *MemoryHistoryService) GetSessionHistory(ctx context.Context, sessionID string, limit int64) ([]*HistoryEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	operations, ok := s.operations[sessionID]
	if !ok {
		return []*HistoryEvent{}, nil
	}

	if limit > 0 && int64(len(operations)) > limit {
		operations = operations[:limit]
	}

	return operations, nil
}

// ReconstructSnapshot reconstructs the content of a specific version.
func (s *MemoryHistoryService) ReconstructSnapshot(ctx context.Context, sessionID string, targetVersionID int64) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshots, ok := s.snapshots[sessionID]
	if !ok {
		return "", fmt.Errorf("session not found")
	}

	content := ""
	for version := int64(0); version <= targetVersionID; version++ {
		event, ok := snapshots[version]
		if !ok {
			continue
		}

		if event.Content != "" {
			// Full content snapshot
			content = event.Content
		} else if event.Metadata != nil {
			// Patch-based snapshot
			if patch, ok := event.Metadata["patch"].(string); ok && patch != "" {
				result := s.patchManager.ApplyPatch(content, patch)
				if !result.Success {
					return "", fmt.Errorf("failed to apply patch for version %d", version)
				}
				content = result.Content
			}
		}
	}

	return content, nil
}

// ListSnapshots lists all snapshots for a session.
func (s *MemoryHistoryService) ListSnapshots(ctx context.Context, sessionID string) ([]*SnapshotInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshots, ok := s.snapshots[sessionID]
	if !ok {
		return []*SnapshotInfo{}, nil
	}

	infos := make([]*SnapshotInfo, 0, len(snapshots))
	for _, event := range snapshots {
		infos = append(infos, &SnapshotInfo{
			SnapshotVersion: event.VersionID,
			LastSnapshotTime: event.CreatedAt,
		})
	}

	return infos, nil
}

// Close closes the history service.
func (s *MemoryHistoryService) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()

	close(s.closeChan)
	s.wg.Wait()
	close(s.eventChan)

	return nil
}
