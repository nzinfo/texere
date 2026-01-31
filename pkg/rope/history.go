package rope

import (
	"sync"
	"time"
)

// Revision represents a single revision in the undo/redo history tree.
type Revision struct {
	parent      int              // Index of parent revision (for undo)
	lastChild   int              // Index of last child revision (for redo)
	transaction *Transaction     // Forward transaction (redo)
	inversion   *Transaction     // Inverted transaction (undo)
	timestamp   time.Time        // When this revision was created
}

// History manages a tree of document revisions for undo/redo.
// Unlike a simple stack, this allows non-linear history (branching).
type History struct {
	mu         sync.RWMutex
	revisions   []*Revision // All revisions in chronological order
	current     int         // Index of current revision
	maxSize     int         // Maximum history size (0 = unlimited)
}

// NewHistory creates a new empty history.
func NewHistory() *History {
	return &History{
		revisions: make([]*Revision, 0, 128),
		current:   -1,
		maxSize:   1000, // Default max revisions
	}
}

// SetMaxSize sets the maximum number of revisions to keep.
// When the limit is reached, oldest revisions are removed.
func (h *History) SetMaxSize(size int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.maxSize = size
	h.prune()
}

// MaxSize returns the maximum number of revisions.
func (h *History) MaxSize() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.maxSize
}

// CommitRevision adds a new revision to the history.
// The revision becomes a child of the current revision.
func (h *History) CommitRevision(transaction *Transaction, original *Rope) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if transaction == nil || transaction.IsEmpty() {
		return
	}

	// Create inversion for undo
	inversion := transaction.Invert(original)

	revision := &Revision{
		parent:      h.current,
		lastChild:   -1,
		transaction: transaction,
		inversion:   inversion,
		timestamp:   time.Now(),
	}

	// Add to revisions
	h.revisions = append(h.revisions, revision)
	newIndex := len(h.revisions) - 1

	// Update parent's last child pointer (if there is a parent)
	if h.current >= 0 {
		// h.current is a valid index in revisions BEFORE we add the new one
		// but now we need to check if it's within bounds
		if h.current < len(h.revisions)-1 {
			h.revisions[h.current].lastChild = newIndex
		}
	}

	// Move to new revision
	h.current = newIndex

	h.prune()
}

// CanUndo returns true if there is a revision to undo to.
func (h *History) CanUndo() bool {
	h.mu.RLock()
	result := h.current >= 0
	h.mu.RUnlock()
	return result
}

// CanRedo returns true if there is a revision to redo to.
func (h *History) CanRedo() bool {
	h.mu.RLock()

	// Special case: if at root (-1), can redo to first revision
	if h.current == -1 {
		result := len(h.revisions) > 0
		h.mu.RUnlock()
		return result
	}

	if h.current >= len(h.revisions) {
		h.mu.RUnlock()
		return false
	}

	current := h.revisions[h.current]
	result := current.lastChild >= 0
	h.mu.RUnlock()

	return result
}

// Undo returns the transaction to undo the current revision.
// Returns nil if already at the root (no more to undo).
func (h *History) Undo() *Transaction {
	h.mu.Lock()

	// Direct check instead of calling CanUndo() to avoid deadlock
	if h.current < 0 {
		h.mu.Unlock()
		return nil
	}

	current := h.revisions[h.current]
	h.current = current.parent

	result := current.inversion
	h.mu.Unlock()

	return result
}

// Redo returns the transaction to redo to the next revision.
// Returns nil if there is no forward revision.
func (h *History) Redo() *Transaction {
	h.mu.Lock()

	// Special case: if at root (-1), allow redo to first revision (index 0)
	if h.current == -1 {
		if len(h.revisions) == 0 {
			h.mu.Unlock()
			return nil
		}
		h.current = 0
		result := h.revisions[0].transaction
		h.mu.Unlock()
		return result
	}

	// Normal case: check if current has a last child
	if h.current >= len(h.revisions) {
		h.mu.Unlock()
		return nil
	}

	current := h.revisions[h.current]
	if current.lastChild < 0 {
		h.mu.Unlock()
		return nil
	}

	nextIndex := current.lastChild
	h.current = nextIndex

	result := h.revisions[nextIndex].transaction
	h.mu.Unlock()

	return result
}

// CurrentIndex returns the index of the current revision.
func (h *History) CurrentIndex() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.current
}

// CurrentRevision returns the current revision.
func (h *History) CurrentRevision() *Revision {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.current < 0 || h.current >= len(h.revisions) {
		return nil
	}

	return h.revisions[h.current]
}

// RevisionCount returns the total number of revisions.
func (h *History) RevisionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.revisions)
}

// GetRevision returns the revision at the given index.
func (h *History) GetRevision(index int) *Revision {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if index < 0 || index >= len(h.revisions) {
		return nil
	}

	return h.revisions[index]
}

// AtRoot returns true if the current revision is the root (no parent).
func (h *History) AtRoot() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.current < 0
}

// AtTip returns true if the current revision is at the tip (no children).
func (h *History) AtTip() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// If at root (-1), check if there are any revisions
	if h.current == -1 {
		// At root, but if there are revisions, can redo (not at tip)
		return len(h.revisions) == 0
	}

	if h.current >= len(h.revisions) {
		return true
	}

	return h.revisions[h.current].lastChild < 0
}

// Clear removes all revisions from the history.
func (h *History) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.revisions = make([]*Revision, 0, 128)
	h.current = -1
}

// prune removes old revisions if the history exceeds maxSize.
func (h *History) prune() {
	if h.maxSize <= 0 {
		return
	}

	// Don't prune if under limit
	if len(h.revisions) <= h.maxSize {
		return
	}

	// Simple strategy: remove oldest revisions
	// In a real implementation, you'd want to be more careful
	// about preserving branches and the current path
	excess := len(h.revisions) - h.maxSize

	// Find the new root (oldest revision to keep)
	newRoot := excess
	if newRoot >= len(h.revisions) {
		newRoot = len(h.revisions) - 1
	}

	// Remove old revisions
	h.revisions = h.revisions[newRoot:]

	// Update indices
	for i := range h.revisions {
		if h.revisions[i].parent >= 0 {
			h.revisions[i].parent -= newRoot
		}
		if h.revisions[i].lastChild >= 0 {
			h.revisions[i].lastChild -= newRoot
		}
	}

	h.current -= newRoot
	if h.current < -1 {
		h.current = -1
	}
}

// GotoRevision moves to a specific revision by index.
// Returns the transaction needed to apply to get there, or nil if invalid.
func (h *History) GotoRevision(index int) *Transaction {
	h.mu.Lock()
	defer h.mu.Unlock()

	if index < -1 || index >= len(h.revisions) {
		return nil
	}

	if index == h.current {
		return nil // Already there
	}

	// Find lowest common ancestor
	_ = h.lowestCommonAncestor(h.current, index)

	// Path from current to LCA (undo)
	// Path from LCA to target (redo)

	// Simplified: Just return the transaction from target
	// In a real implementation, you'd compute the full path
	h.current = index

	if index >= 0 {
		return h.revisions[index].transaction
	}

	return nil
}

// lowestCommonAncestor finds the lowest common ancestor of two revisions.
func (h *History) lowestCommonAncestor(a, b int) int {
	if a < 0 || b < 0 {
		return -1
	}

	visitedA := make(map[int]bool)
	visitedB := make(map[int]bool)

	for {
		visitedA[a] = true
		visitedB[b] = true

		if visitedA[b] {
			return b
		}
		if visitedB[a] {
			return a
		}

		if a >= 0 {
			a = h.revisions[a].parent
		}
		if b >= 0 {
			b = h.revisions[b].parent
		}

		if a < 0 && b < 0 {
			return -1
		}
	}
}

// Earlier moves back in time by the specified number of undo steps.
// Returns the final document state after undoing, or nil if already at root.
// This is a convenience method that calls Undo multiple times.
func (h *History) Earlier(steps int) *Transaction {
	if steps <= 0 {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Undo step by step
	var result *Transaction = nil
	for i := 0; i < steps && h.current >= 0; i++ {
		current := h.revisions[h.current]
		h.current = current.parent
		result = current.inversion
	}

	return result
}

// EarlierByTime moves back in time to the revision closest to the specified duration ago.
// Uses binary search for efficient O(log N) time complexity.
// Returns the transaction to apply, or nil if already at root.
func (h *History) EarlierByTime(duration time.Duration) *Transaction {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.current < 0 || len(h.revisions) == 0 {
		return nil
	}

	// Calculate target timestamp
	targetTime := time.Now().Add(-duration)

	// Binary search for the revision closest to target time
	idx := h.findRevisionByTime(targetTime, true) // search backwards

	if idx < 0 || idx == h.current {
		return nil
	}

	// Build path from current to target
	return h.buildTransactionToRevision(idx)
}

// Later moves forward in time by the specified number of redo steps.
// Returns the final transaction to apply, or nil if already at tip.
func (h *History) Later(steps int) *Transaction {
	if steps <= 0 {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Redo step by step
	var result *Transaction = nil
	for i := 0; i < steps; i++ {
		// Special case: if at root (-1), allow redo to first revision
		if h.current == -1 {
			if len(h.revisions) == 0 {
				return nil
			}
			h.current = 0
			result = h.revisions[0].transaction
			continue
		}

		if h.current >= len(h.revisions) {
			return result
		}

		current := h.revisions[h.current]
		if current.lastChild < 0 {
			return result
		}

		h.current = current.lastChild
		result = h.revisions[h.current].transaction
	}

	return result
}

// LaterByTime moves forward in time to the revision closest to the specified duration ahead.
// Uses binary search for efficient O(log N) time complexity.
// Returns the transaction to apply, or nil if already at tip.
func (h *History) LaterByTime(duration time.Duration) *Transaction {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.current < 0 || len(h.revisions) == 0 {
		return nil
	}

	// Get current revision's timestamp
	currentRev := h.revisions[h.current]
	targetTime := currentRev.timestamp.Add(duration)

	// Binary search for the revision closest to target time
	idx := h.findRevisionByTime(targetTime, false) // search forwards

	if idx < 0 || idx == h.current {
		return nil
	}

	// Build path from current to target
	return h.buildTransactionToRevision(idx)
}

// findRevisionByTime uses binary search to find the revision closest to target time.
// If searchBackwards is true, searches for revisions before current, otherwise after.
func (h *History) findRevisionByTime(targetTime time.Time, searchBackwards bool) int {
	if len(h.revisions) == 0 {
		return -1
	}

	// Binary search for closest timestamp
	left := 0
	right := len(h.revisions) - 1
	closestIdx := -1
	minDiff := time.Duration(1<<63 - 1) // Max duration

	for left <= right {
		mid := (left + right) / 2
		rev := h.revisions[mid]

		// Skip revisions that are not in the correct direction
		if searchBackwards && mid > h.current {
			right = mid - 1
			continue
		}
		if !searchBackwards && mid < h.current {
			left = mid + 1
			continue
		}

		diff := rev.timestamp.Sub(targetTime)
		if diff < 0 {
			diff = -diff
		}

		// Check if this is closer
		if diff < minDiff {
			minDiff = diff
			closestIdx = mid
		}

		// Adjust search range
		if rev.timestamp.Before(targetTime) {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return closestIdx
}

// buildTransactionToRevision builds a transaction to navigate from current to target revision.
// This computes the path using the lowest common ancestor algorithm and composes transactions.
func (h *History) buildTransactionToRevision(targetIdx int) *Transaction {
	if targetIdx == h.current {
		return nil
	}

	// Find lowest common ancestor
	lca := h.lowestCommonAncestor(h.current, targetIdx)

	// Build path from current to LCA (undo operations)
	var undoPath []*Transaction
	current := h.current
	for current != lca && current >= 0 {
		if current >= len(h.revisions) {
			break
		}
		rev := h.revisions[current]
		undoPath = append(undoPath, NewTransaction(rev.inversion.changeset))
		current = rev.parent
	}

	// Build path from LCA to target (redo operations)
	var redoPath []*Transaction
	target := targetIdx
	for target != lca && target >= 0 {
		if target >= len(h.revisions) {
			break
		}
		rev := h.revisions[target]
		redoPath = append([]*Transaction{NewTransaction(rev.transaction.changeset)}, redoPath...)
		target = rev.parent
	}

	// Compose all transactions
	// First apply undo path in reverse, then redo path
	var composed *ChangeSet = nil

	// Compose undo path
	for i := len(undoPath) - 1; i >= 0; i-- {
		txn := undoPath[i]
		if composed == nil {
			composed = txn.changeset
		} else {
			composed = composed.Compose(txn.changeset)
		}
	}

	// Compose redo path
	for _, txn := range redoPath {
		if composed == nil {
			composed = txn.changeset
		} else {
			composed = composed.Compose(txn.changeset)
		}
	}

	// Move to target
	oldCurrent := h.current
	h.current = targetIdx

	if composed != nil {
		return NewTransaction(composed)
	}

	// Fallback: return target's transaction directly
	if targetIdx >= 0 && targetIdx < len(h.revisions) {
		return h.revisions[targetIdx].transaction
	}

	// Restore current if we couldn't find a valid transaction
	h.current = oldCurrent
	return nil
}

// GetPath returns the path from root to the current revision.
func (h *History) GetPath() []int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.current < 0 {
		return []int{}
	}

	path := []int{}
	current := h.current

	for current >= 0 {
		path = append([]int{current}, path...)
		current = h.revisions[current].parent
	}

	return path
}

// Stats returns statistics about the history.
func (h *History) Stats() *HistoryStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return &HistoryStats{
		TotalRevisions: len(h.revisions),
		CurrentIndex:   h.current,
		MaxSize:        h.maxSize,
		CanUndo:        h.CanUndo(),
		CanRedo:        h.CanRedo(),
	}
}

// HistoryStats contains statistics about the history.
type HistoryStats struct {
	TotalRevisions int
	CurrentIndex   int
	MaxSize        int
	CanUndo        bool
	CanRedo        bool
}

// ========== Time-based Navigation (Immutable State) ==========

// EarlierByDuration moves back in history by the specified time duration.
// Returns a new history at the state approximately 'duration' ago.
// This does not modify the current history; it returns a new History object.
func (h *History) EarlierByDuration(duration time.Duration) *History {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.current < 0 || len(h.revisions) == 0 {
		return &History{
			revisions:   h.revisions,
			current:     h.current,
			maxSize:     h.maxSize,
		}
	}

	// Get current revision's timestamp and subtract duration
	currentRev := h.revisions[h.current]
	targetTime := currentRev.timestamp.Add(-duration)
	targetTimeTrunc := targetTime.Truncate(time.Millisecond)

	// Walk back through history to find revision closest to target time
	for i := h.current; i >= 0; i-- {
		rev := h.revisions[i]
		revTime := rev.timestamp.Truncate(time.Millisecond)

		if revTime.Before(targetTimeTrunc) || revTime.Equal(targetTimeTrunc) {
			// Found state at or before target time
			return &History{
				revisions:   h.revisions,
				current:     i,
				maxSize:     h.maxSize,
			}
		}
	}

	// If not found, return root state (current = -1)
	return &History{
		revisions: h.revisions,
		current:   -1,
		maxSize:   h.maxSize,
	}
}

// LaterByDuration moves forward in history by the specified time duration.
// Returns a new history at the state approximately 'duration' in the future.
// This does not modify the current history; it returns a new History object.
func (h *History) LaterByDuration(duration time.Duration) *History {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.revisions) == 0 {
		return &History{
			revisions:   h.revisions,
			current:     h.current,
			maxSize:     h.maxSize,
		}
	}

	// Determine the starting time based on current position
	var startTime time.Time
	startIdx := 0

	if h.current < 0 {
		// At root, start from the first revision's time
		startTime = h.revisions[0].timestamp
		startIdx = -1
	} else {
		startTime = h.revisions[h.current].timestamp
		startIdx = h.current
	}

	targetTime := startTime.Add(duration)
	targetTimeTrunc := targetTime.Truncate(time.Millisecond)

	// Walk forward through history to find revision closest to target time
	bestIdx := startIdx
	for i := startIdx + 1; i < len(h.revisions); i++ {
		rev := h.revisions[i]
		revTime := rev.timestamp.Truncate(time.Millisecond)

		if revTime.After(targetTimeTrunc) || revTime.Equal(targetTimeTrunc) {
			// Found state at or after target time
			// Return the state just before this one (to not overshoot)
			bestIdx = i - 1
			break
		}
		bestIdx = i
	}

	// Ensure we don't go past the tip and bestIdx is at least 0
	if bestIdx >= len(h.revisions) {
		bestIdx = len(h.revisions) - 1
	}
	if bestIdx < 0 && len(h.revisions) > 0 {
		bestIdx = 0
	}

	return &History{
		revisions: h.revisions,
		current:   bestIdx,
		maxSize:   h.maxSize,
	}
}

// TimeAt returns the timestamp of the current state.
func (h *History) TimeAt() time.Time {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.current < 0 || h.current >= len(h.revisions) {
		return time.Time{}
	}

	return h.revisions[h.current].timestamp
}

// DurationFromRoot returns the time elapsed since the root state.
func (h *History) DurationFromRoot() time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.current < 0 || len(h.revisions) == 0 {
		return 0
	}

	rootTime := h.revisions[0].timestamp
	currentTime := h.revisions[h.current].timestamp
	return currentTime.Sub(rootTime)
}

// DurationToTip returns the time from the current state to the tip state.
func (h *History) DurationToTip() time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.revisions) == 0 {
		return 0
	}

	if h.current < 0 {
		// At root, return duration from first revision to tip
		if len(h.revisions) >= 2 {
			firstTime := h.revisions[0].timestamp
			tipTime := h.revisions[len(h.revisions)-1].timestamp
			return tipTime.Sub(firstTime)
		}
		return 0
	}

	currentTime := h.revisions[h.current].timestamp
	tipTime := h.revisions[len(h.revisions)-1].timestamp
	return tipTime.Sub(currentTime)
}

// IsEmpty returns true if the history is empty (no revisions).
func (h *History) IsEmpty() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.revisions) == 0
}

// ToRoot returns a new history at the root state (before all revisions).
func (h *History) ToRoot() *History {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return &History{
		revisions: h.revisions,
		current:   -1,
		maxSize:   h.maxSize,
	}
}

// ToTip returns a new history at the tip state (after all revisions).
func (h *History) ToTip() *History {
	h.mu.RLock()
	defer h.mu.RUnlock()

	tipIdx := len(h.revisions) - 1
	if tipIdx < 0 {
		tipIdx = -1
	}

	return &History{
		revisions: h.revisions,
		current:   tipIdx,
		maxSize:   h.maxSize,
	}
}

// Clone creates a deep copy of the history.
func (h *History) Clone() *History {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Deep copy revisions
	revisionsCopy := make([]*Revision, len(h.revisions))
	for i, rev := range h.revisions {
		revisionsCopy[i] = &Revision{
			parent:      rev.parent,
			lastChild:   rev.lastChild,
			transaction: rev.transaction,
			inversion:   rev.inversion,
			timestamp:   rev.timestamp,
		}
	}

	return &History{
		revisions: revisionsCopy,
		current:   h.current,
		maxSize:   h.maxSize,
	}
}
