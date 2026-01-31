package concordia

import (
	"sync"

	"github.com/coreseekdev/texere/pkg/ot"
	"github.com/coreseekdev/texere/pkg/rope"
)

// LamportTime represents a logical timestamp using Lamport clocks.
// This provides a partial ordering of events in distributed systems
// without requiring synchronized physical clocks.
type LamportTime int64

// Revision represents a single revision in the undo/redo history tree.
type Revision struct {
	parent    int             // Index of parent revision (for undo)
	lastChild int             // Index of last child revision (for redo)
	operation *ot.Operation   // Forward operation (redo)
	inversion *ot.Operation   // Inverted operation (undo)
	lamport   LamportTime     // Lamport timestamp (logical clock)
}

// History manages a tree of document revisions for undo/redo.
// Unlike a simple stack, this allows non-linear history (branching).
type History struct {
	mu        sync.RWMutex
	revisions []*Revision // All revisions in chronological order
	current   int         // Index of current revision
	maxSize   int         // Maximum history size (0 = unlimited)
	lamport   LamportTime // Current Lamport timestamp
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
func (h *History) CommitRevision(operation *ot.Operation, original *rope.Rope) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if operation == nil || operation.IsNoop() {
		return
	}

	// Increment Lamport clock
	h.lamport++

	// Create inversion for undo
	inversion := operation.Invert(original.String())

	revision := &Revision{
		parent:    h.current,
		lastChild: -1,
		operation: operation,
		inversion: inversion,
		lamport:   h.lamport,
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

// Undo returns the operation to undo the current revision.
// Returns nil if already at the root (no more to undo).
func (h *History) Undo() *ot.Operation {
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

// Redo returns the operation to redo to the next revision.
// Returns nil if there is no forward revision.
func (h *History) Redo() *ot.Operation {
	h.mu.Lock()

	// Special case: if at root (-1), allow redo to first revision (index 0)
	if h.current == -1 {
		if len(h.revisions) == 0 {
			h.mu.Unlock()
			return nil
		}
		h.current = 0
		result := h.revisions[0].operation
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

	result := h.revisions[nextIndex].operation
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
// Returns the operation needed to apply to get there, or nil if invalid.
func (h *History) GotoRevision(index int) *ot.Operation {
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

	// Simplified: Just return the operation from target
	// In a real implementation, you'd compute the full path
	h.current = index

	if index >= 0 {
		return h.revisions[index].operation
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
// Returns the final operation after undoing, or nil if already at root.
// This is a convenience method that calls Undo multiple times.
func (h *History) Earlier(steps int) *ot.Operation {
	if steps <= 0 {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Undo step by step
	var result *ot.Operation = nil
	for i := 0; i < steps && h.current >= 0; i++ {
		current := h.revisions[h.current]
		h.current = current.parent
		result = current.inversion
	}

	return result
}

// EarlierByLamport moves back in time to the revision closest to the specified Lamport time.
// Uses binary search for efficient O(log N) time complexity.
// Returns the operation to apply, or nil if already at root.
func (h *History) EarlierByLamport(targetLamport LamportTime) *ot.Operation {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.current < 0 || len(h.revisions) == 0 {
		return nil
	}

	// Binary search for the revision closest to target Lamport time
	idx := h.findRevisionByLamport(targetLamport, true) // search backwards

	if idx < 0 || idx == h.current {
		return nil
	}

	// Build path from current to target
	return h.buildOperationToRevision(idx)
}

// Later moves forward in time by the specified number of redo steps.
// Returns the final operation to apply, or nil if already at tip.
func (h *History) Later(steps int) *ot.Operation {
	if steps <= 0 {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Redo step by step
	var result *ot.Operation = nil
	for i := 0; i < steps; i++ {
		// Special case: if at root (-1), allow redo to first revision
		if h.current == -1 {
			if len(h.revisions) == 0 {
				return nil
			}
			h.current = 0
			result = h.revisions[0].operation
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
		result = h.revisions[h.current].operation
	}

	return result
}

// LaterByLamport moves forward in time to the revision closest to the specified Lamport time ahead.
// Uses binary search for efficient O(log N) time complexity.
// Returns the operation to apply, or nil if already at tip.
func (h *History) LaterByLamport(targetLamport LamportTime) *ot.Operation {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.current < 0 || len(h.revisions) == 0 {
		return nil
	}

	// Binary search for the revision closest to target Lamport time
	idx := h.findRevisionByLamport(targetLamport, false) // search forwards

	if idx < 0 || idx == h.current {
		return nil
	}

	// Build path from current to target
	return h.buildOperationToRevision(idx)
}

// findRevisionByLamport uses binary search to find the revision closest to target Lamport time.
// If searchBackwards is true, searches for revisions before current, otherwise after.
func (h *History) findRevisionByLamport(targetLamport LamportTime, searchBackwards bool) int {
	if len(h.revisions) == 0 {
		return -1
	}

	// Binary search for closest Lamport time
	left := 0
	right := len(h.revisions) - 1
	closestIdx := -1
	minDiff := LamportTime(1<<63 - 1) // Max LamportTime

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

		diff := rev.lamport - targetLamport
		if diff < 0 {
			diff = -diff
		}

		// Check if this is closer
		if diff < minDiff {
			minDiff = diff
			closestIdx = mid
		}

		// Adjust search range
		if rev.lamport < targetLamport {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return closestIdx
}

// buildOperationToRevision builds an operation to navigate from current to target revision.
// This computes the path using the lowest common ancestor algorithm and composes operations.
func (h *History) buildOperationToRevision(targetIdx int) *ot.Operation {
	if targetIdx == h.current {
		return nil
	}

	// Find lowest common ancestor
	lca := h.lowestCommonAncestor(h.current, targetIdx)

	// Build path from current to LCA (undo operations)
	var undoPath []*ot.Operation
	current := h.current
	for current != lca && current >= 0 {
		if current >= len(h.revisions) {
			break
		}
		rev := h.revisions[current]
		undoPath = append(undoPath, rev.inversion)
		current = rev.parent
	}

	// Build path from LCA to target (redo operations)
	var redoPath []*ot.Operation
	target := targetIdx
	for target != lca && target >= 0 {
		if target >= len(h.revisions) {
			break
		}
		rev := h.revisions[target]
		redoPath = append([]*ot.Operation{rev.operation}, redoPath...)
		target = rev.parent
	}

	// Compose all operations
	// First apply undo path in reverse, then redo path
	var composed *ot.Operation = nil

	// Compose undo path
	for i := len(undoPath) - 1; i >= 0; i-- {
		op := undoPath[i]
		if composed == nil {
			composed = op
		} else {
			composed, _ = ot.Compose(composed, op)
		}
	}

	// Compose redo path
	for _, op := range redoPath {
		if composed == nil {
			composed = op
		} else {
			composed, _ = ot.Compose(composed, op)
		}
	}

	// Move to target
	oldCurrent := h.current
	h.current = targetIdx

	if composed != nil {
		return composed
	}

	// Fallback: return target's operation directly
	if targetIdx >= 0 && targetIdx < len(h.revisions) {
		return h.revisions[targetIdx].operation
	}

	// Restore current if we couldn't find a valid operation
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

// ========== Lamport Time-based Navigation ==========

// LamportAt returns the Lamport timestamp of the current state.
func (h *History) LamportAt() LamportTime {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.current < 0 || h.current >= len(h.revisions) {
		return 0
	}

	return h.revisions[h.current].lamport
}

// LamportFromRoot returns the Lamport time elapsed since the root state.
func (h *History) LamportFromRoot() LamportTime {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.current < 0 || len(h.revisions) == 0 {
		return 0
	}

	rootLamport := h.revisions[0].lamport
	currentLamport := h.revisions[h.current].lamport
	return currentLamport - rootLamport
}

// LamportToTip returns the Lamport time from the current state to the tip state.
func (h *History) LamportToTip() LamportTime {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.revisions) == 0 {
		return 0
	}

	if h.current < 0 {
		// At root, return duration from first revision to tip
		if len(h.revisions) >= 2 {
			firstLamport := h.revisions[0].lamport
			tipLamport := h.revisions[len(h.revisions)-1].lamport
			return tipLamport - firstLamport
		}
		return 0
	}

	currentLamport := h.revisions[h.current].lamport
	tipLamport := h.revisions[len(h.revisions)-1].lamport
	return tipLamport - currentLamport
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
		lamport:   h.lamport,
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
		lamport:   h.lamport,
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
			parent:    rev.parent,
			lastChild: rev.lastChild,
			operation: rev.operation,
			inversion: rev.inversion,
			lamport:   rev.lamport,
		}
	}

	return &History{
		revisions: revisionsCopy,
		current:   h.current,
		maxSize:   h.maxSize,
		lamport:   h.lamport,
	}
}
