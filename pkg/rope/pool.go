package rope

import (
	"sync"
	"time"
)

// ObjectPool provides reusable objects to reduce allocations.
type ObjectPool struct {
	changesets sync.Pool
	transactions sync.Pool
}

// NewObjectPool creates a new object pool.
func NewObjectPool() *ObjectPool {
	return &ObjectPool{
		changesets: sync.Pool{
			New: func() interface{} {
				return &ChangeSet{
					operations: make([]Operation, 0, 8),
				}
			},
		},
		transactions: sync.Pool{
			New: func() interface{} {
				return &Transaction{}
			},
		},
	}
}

// GetChangeSet retrieves a changeset from the pool or creates a new one.
func (op *ObjectPool) GetChangeSet(lenBefore int) *ChangeSet {
	cs := op.changesets.Get().(*ChangeSet)
	cs.operations = cs.operations[:0] // Reset slice
	cs.lenBefore = lenBefore
	cs.lenAfter = lenBefore
	// Reset fused flag implicitly by clearing operations
	return cs
}

// PutChangeSet returns a changeset to the pool for reuse.
func (op *ObjectPool) PutChangeSet(cs *ChangeSet) {
	if cs != nil && cap(cs.operations) < 256 { // Only pool small slices
		op.changesets.Put(cs)
	}
}

// GetTransaction retrieves a transaction from the pool or creates a new one.
func (op *ObjectPool) GetTransaction(changeset *ChangeSet) *Transaction {
	txn := op.transactions.Get().(*Transaction)
	txn.changeset = changeset
	txn.timestamp = txn.timestamp // Reuse timestamp field
	return txn
}

// PutTransaction returns a transaction to the pool for reuse.
func (op *ObjectPool) PutTransaction(txn *Transaction) {
	if txn != nil {
		op.transactions.Put(txn)
	}
}

// Global pool for reuse
var globalPool = NewObjectPool()

// LazyTransaction is a transaction that defers expensive operations until needed.
type LazyTransaction struct {
	changeset     *ChangeSet
	original      *Rope
	inversion     *ChangeSet
	inversionCalculated bool
	originalCalculated  bool
	mu            sync.RWMutex
}

// NewLazyTransaction creates a new lazy transaction.
func NewLazyTransaction(changeset *ChangeSet) *LazyTransaction {
	return &LazyTransaction{
		changeset:          changeset,
		inversionCalculated: false,
		originalCalculated: false,
	}
}

// Changeset returns the changeset.
func (lt *LazyTransaction) Changeset() *ChangeSet {
	return lt.changeset
}

// Apply applies the transaction to a rope.
func (lt *LazyTransaction) Apply(r *Rope) *Rope {
	if lt.changeset == nil {
		return r
	}
	return lt.changeset.Apply(r)
}

// Invert creates an inverted transaction for undo.
// The inversion is calculated lazily on first call.
func (lt *LazyTransaction) Invert(original *Rope) *Transaction {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	if !lt.inversionCalculated {
		lt.original = original
		lt.inversion = lt.changeset.Invert(original)
		lt.inversionCalculated = true
	}

	return NewTransaction(lt.inversion)
}

// IsEmpty returns true if the transaction has no changes.
func (lt *LazyTransaction) IsEmpty() bool {
	return lt.changeset == nil || lt.changeset.IsEmpty()
}

// CachedInversion returns the cached inversion if available.
func (lt *LazyTransaction) CachedInversion() *ChangeSet {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	return lt.inversion
}

// LazyHistory is a history manager with lazy evaluation and caching.
type LazyHistory struct {
	history     *History
	cache       map[int]*Transaction
	cacheSize   int
	mu          sync.RWMutex
}

// NewLazyHistory creates a new lazy history manager.
func NewLazyHistory(maxSize int) *LazyHistory {
	return &LazyHistory{
		history:   NewHistory(),
		cache:     make(map[int]*Transaction),
		cacheSize: maxSize,
	}
}

// CommitRevision adds a new revision to the history.
func (lh *LazyHistory) CommitRevision(transaction *Transaction, original *Rope) {
	lh.mu.Lock()
	defer lh.mu.Unlock()

	lh.history.CommitRevision(transaction, original)

	// Clear cache if it gets too big
	if len(lh.cache) > lh.cacheSize {
		lh.clearCache()
	}
}

// Undo returns the transaction to undo the current revision.
// Uses cached value if available.
func (lh *LazyHistory) Undo() *Transaction {
	lh.mu.Lock()
	defer lh.mu.Unlock()

	current := lh.history.CurrentIndex()
	if cached, exists := lh.cache[current]; exists {
		// Move the history cursor
		lh.history.Undo()
		return cached
	}

	// Compute and cache
	txn := lh.history.Undo()
	if txn != nil && current >= 0 {
		lh.cache[current] = txn
	}

	return txn
}

// Redo returns the transaction to redo to the next revision.
// Uses cached value if available.
func (lh *LazyHistory) Redo() *Transaction {
	lh.mu.Lock()
	defer lh.mu.Unlock()

	current := lh.history.CurrentIndex()
	if cached, exists := lh.cache[current]; exists {
		// Move the history cursor
		lh.history.Redo()
		return cached
	}

	// Compute and cache
	txn := lh.history.Redo()
	if txn != nil {
		lh.cache[current] = txn
	}

	return txn
}

// CanUndo returns true if there is a revision to undo to.
func (lh *LazyHistory) CanUndo() bool {
	return lh.history.CanUndo()
}

// CanRedo returns true if there is a revision to redo to.
func (lh *LazyHistory) CanRedo() bool {
	return lh.history.CanRedo()
}

// CurrentIndex returns the index of the current revision.
func (lh *LazyHistory) CurrentIndex() int {
	return lh.history.CurrentIndex()
}

// RevisionCount returns the total number of revisions.
func (lh *LazyHistory) RevisionCount() int {
	return lh.history.RevisionCount()
}

// Clear removes all revisions from the history.
func (lh *LazyHistory) Clear() {
	lh.mu.Lock()
	defer lh.mu.Unlock()

	lh.history.Clear()
	lh.clearCache()
}

// ClearCache removes all cached transactions.
func (lh *LazyHistory) ClearCache() {
	lh.mu.Lock()
	defer lh.mu.Unlock()

	lh.clearCache()
}

// clearCache clears the cache (must be called with lock held).
func (lh *LazyHistory) clearCache() {
	lh.cache = make(map[int]*Transaction)
}

// EarlierByTime moves back in time to the revision closest to the specified duration ago.
func (lh *LazyHistory) EarlierByTime(duration time.Duration) *Transaction {
	lh.mu.Lock()
	defer lh.mu.Unlock()

	return lh.history.EarlierByTime(duration)
}

// LaterByTime moves forward in time to the revision closest to the specified duration ahead.
func (lh *LazyHistory) LaterByTime(duration time.Duration) *Transaction {
	lh.mu.Lock()
	defer lh.mu.Unlock()

	return lh.history.LaterByTime(duration)
}

// Stats returns statistics about the history including cache info.
func (lh *LazyHistory) Stats() *LazyHistoryStats {
	lh.mu.RLock()
	defer lh.mu.RUnlock()

	baseStats := lh.history.Stats()

	return &LazyHistoryStats{
		TotalRevisions: baseStats.TotalRevisions,
		CurrentIndex:   baseStats.CurrentIndex,
		MaxSize:        baseStats.MaxSize,
		CanUndo:        baseStats.CanUndo,
		CanRedo:        baseStats.CanRedo,
		CacheSize:      len(lh.cache),
		CacheCapacity:  lh.cacheSize,
	}
}

// LazyHistoryStats contains statistics about the lazy history.
type LazyHistoryStats struct {
	TotalRevisions int
	CurrentIndex   int
	MaxSize        int
	CanUndo        bool
	CanRedo        bool
	CacheSize      int
	CacheCapacity  int
}
