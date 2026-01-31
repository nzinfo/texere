package rope

import (
	"sync"
	"time"
)

// SavePoint represents a snapshot of the document at a specific point in time.
// It is reference-counted and will be automatically cleaned up when no longer referenced.
type SavePoint struct {
	rope       *Rope
	timestamp  time.Time
	revisionID int
	refCount   int
	mu         sync.Mutex
}

// NewSavePoint creates a new savepoint from the current document state.
func NewSavePoint(rope *Rope, revisionID int) *SavePoint {
	return &SavePoint{
		rope:       rope,
		timestamp:  time.Now(),
		revisionID: revisionID,
		refCount:   1, // Initial reference from creator
	}
}

// Rope returns the rope snapshot.
func (sp *SavePoint) Rope() *Rope {
	return sp.rope
}

// Timestamp returns when the savepoint was created.
func (sp *SavePoint) Timestamp() time.Time {
	return sp.timestamp
}

// RevisionID returns the revision ID this savepoint points to.
func (sp *SavePoint) RevisionID() int {
	return sp.revisionID
}

// Increment increases the reference count.
func (sp *SavePoint) Increment() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.refCount++
}

// Decrement decreases the reference count.
// Returns true if the savepoint should be cleaned up (refCount == 0).
func (sp *SavePoint) Decrement() bool {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.refCount--
	return sp.refCount <= 0
}

// RefCount returns the current reference count.
func (sp *SavePoint) RefCount() int {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	return sp.refCount
}

// SavePointManager manages document savepoints with automatic cleanup.
type SavePointManager struct {
	savepoints map[int]*SavePoint
	nextID     int
	mu         sync.RWMutex
}

// NewSavePointManager creates a new savepoint manager.
func NewSavePointManager() *SavePointManager {
	return &SavePointManager{
		savepoints: make(map[int]*SavePoint),
		nextID:     0,
	}
}

// Create creates a new savepoint from the current document state.
// Returns a savepoint ID that can be used to restore or cleanup.
func (sm *SavePointManager) Create(rope *Rope, revisionID int) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	id := sm.nextID
	sm.nextID++

	sm.savepoints[id] = NewSavePoint(rope, revisionID)

	return id
}

// Get retrieves a savepoint by ID and increments its reference count.
// Returns nil if the savepoint doesn't exist.
func (sm *SavePointManager) Get(id int) *SavePoint {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sp, exists := sm.savepoints[id]
	if !exists {
		return nil
	}

	sp.Increment()
	return sp
}

// Release decrements the reference count for a savepoint.
// If the reference count reaches 0, the savepoint is removed.
func (sm *SavePointManager) Release(id int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sp, exists := sm.savepoints[id]
	if !exists {
		return
	}

	if sp.Decrement() {
		// RefCount is 0, remove the savepoint
		delete(sm.savepoints, id)
	}
}

// Restore restores the document to the state saved in the savepoint.
// Returns the rope snapshot, or nil if the savepoint doesn't exist.
func (sm *SavePointManager) Restore(id int) *Rope {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sp, exists := sm.savepoints[id]
	if !exists {
		return nil
	}

	// Return a clone of the rope to avoid mutating the savepoint
	return sp.rope.Clone()
}

// HasSavepoint returns true if a savepoint with the given ID exists.
func (sm *SavePointManager) HasSavepoint(id int) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	_, exists := sm.savepoints[id]
	return exists
}

// Clear removes all savepoints.
func (sm *SavePointManager) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.savepoints = make(map[int]*SavePoint)
}

// Count returns the number of active savepoints.
func (sm *SavePointManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.savepoints)
}

// CleanOlderThan removes all savepoints older than the specified duration.
// Returns the number of savepoints removed.
func (sm *SavePointManager) CleanOlderThan(duration time.Duration) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	cutoff := time.Now().Add(-duration)
	removed := 0

	for id, sp := range sm.savepoints {
		if sp.timestamp.Before(cutoff) {
			delete(sm.savepoints, id)
			removed++
		}
	}

	return removed
}
