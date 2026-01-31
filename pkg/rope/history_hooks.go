package rope

import (
	"sync"
)

// ============================================================================
// History Hook System
// ============================================================================

// HookEventType represents the type of hook event.
type HookEventType int

const (
	// HookBeforeEdit is called before an edit operation
	HookBeforeEdit HookEventType = iota
	// HookAfterEdit is called after an edit operation
	HookAfterEdit
	// HookBeforeUndo is called before an undo operation
	HookBeforeUndo
	// HookAfterUndo is called after an undo operation
	HookAfterUndo
	// HookBeforeRedo is called before a redo operation
	HookBeforeRedo
	// HookAfterRedo is called after a redo operation
	HookAfterRedo
	// HookOnBranch is called when a new branch is created
	HookOnBranch
	// HookOnSavepoint is called when a savepoint is created
	HookOnSavepoint
	// HookOnError is called when an error occurs
	HookOnError
)

// HookContext provides context about the hook event.
type HookContext struct {
	EventType HookEventType
	Rope      *Rope
	Edit      *EditInfo
	UndoInfo  *UndoInfo
	RedoInfo  *RedoInfo
	Error     error
	Metadata  map[string]interface{}
}

// EditInfo contains information about an edit operation.
type EditInfo struct {
	Operation string // "insert", "delete", "replace"
	StartPos  int
	EndPos    int
	Text      string
	Length    int
}

// UndoInfo contains information about an undo operation.
type UndoInfo struct {
	RevisionID int
	FromNode   *HistoryNode
	ToNode     *HistoryNode
}

// RedoInfo contains information about a redo operation.
type RedoInfo struct {
	RevisionID int
	FromNode   *HistoryNode
	ToNode     *HistoryNode
}

// HistoryNode represents a node in the history tree for hook context.
type HistoryNode struct {
	RevisionID int
	Parent     *HistoryNode
	Children   []*HistoryNode
}

// HookFunc is a function that can be called as a hook.
// Returning an error will cancel the operation if it's a "Before" hook.
type HookFunc func(ctx *HookContext) error

// Hook represents a registered hook with priority.
type Hook struct {
	ID       string
	Name     string
	Priority int // Higher priority hooks run first
	Func     HookFunc
	Enabled  bool
	mu       sync.Mutex
}

// NewHook creates a new hook with the given parameters.
func NewHook(id, name string, priority int, fn HookFunc) *Hook {
	return &Hook{
		ID:       id,
		Name:     name,
		Priority: priority,
		Func:     fn,
		Enabled:  true,
	}
}

// Enable enables the hook.
func (h *Hook) Enable() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Enabled = true
}

// Disable disables the hook.
func (h *Hook) Disable() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Enabled = false
}

// IsEnabled returns whether the hook is enabled.
func (h *Hook) IsEnabled() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.Enabled
}

// HookManager manages registered hooks for history events.
type HookManager struct {
	hooks  map[HookEventType][]*Hook
	nextID int
	mu     sync.RWMutex
}

// NewHookManager creates a new hook manager.
func NewHookManager() *HookManager {
	return &HookManager{
		hooks:  make(map[HookEventType][]*Hook),
		nextID: 0,
	}
}

// Register registers a new hook for the specified event type.
// Returns the hook object that can be used to enable/disable it.
func (hm *HookManager) Register(eventType HookEventType, name string, priority int, fn HookFunc) *Hook {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hook := &Hook{
		ID:       hm.generateHookID(),
		Name:     name,
		Priority: priority,
		Func:     fn,
		Enabled:  true,
	}

	hm.hooks[eventType] = append(hm.hooks[eventType], hook)
	hm.sortHooks(eventType)

	return hook
}

// Unregister removes a hook by its ID.
func (hm *HookManager) Unregister(hookID string) bool {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	for eventType, hooks := range hm.hooks {
		for i, hook := range hooks {
			if hook.ID == hookID {
				// Remove hook from slice
				hm.hooks[eventType] = append(hooks[:i], hooks[i+1:]...)
				return true
			}
		}
	}

	return false
}

// Trigger executes all hooks for the specified event type.
// Returns an error if any hook returns an error (for "Before" hooks, this cancels the operation).
// For "After" hooks, errors are collected but don't affect the operation.
func (hm *HookManager) Trigger(ctx *HookContext) error {
	hm.mu.RLock()
	hooks := hm.hooks[ctx.EventType]
	hm.mu.RUnlock()

	// Hooks are already sorted by priority (highest first)
	for _, hook := range hooks {
		if !hook.IsEnabled() {
			continue
		}

		if err := hook.Func(ctx); err != nil {
			// For "Before" hooks, error cancels the operation
			if isBeforeHook(ctx.EventType) {
				return err
			}
			// For "After" hooks, log but don't cancel
			// Could add logging here
		}
	}

	return nil
}

// TriggerBeforeEdit triggers all before-edit hooks.
// Returns an error if any hook cancels the operation.
func (hm *HookManager) TriggerBeforeEdit(rope *Rope, edit *EditInfo) error {
	return hm.Trigger(&HookContext{
		EventType: HookBeforeEdit,
		Rope:      rope,
		Edit:      edit,
		Metadata:  make(map[string]interface{}),
	})
}

// TriggerAfterEdit triggers all after-edit hooks.
func (hm *HookManager) TriggerAfterEdit(rope *Rope, edit *EditInfo) {
	_ = hm.Trigger(&HookContext{
		EventType: HookAfterEdit,
		Rope:      rope,
		Edit:      edit,
		Metadata:  make(map[string]interface{}),
	})
}

// TriggerBeforeUndo triggers all before-undo hooks.
// Returns an error if any hook cancels the operation.
func (hm *HookManager) TriggerBeforeUndo(rope *Rope, undoInfo *UndoInfo) error {
	return hm.Trigger(&HookContext{
		EventType: HookBeforeUndo,
		Rope:      rope,
		UndoInfo:  undoInfo,
		Metadata:  make(map[string]interface{}),
	})
}

// TriggerAfterUndo triggers all after-undo hooks.
func (hm *HookManager) TriggerAfterUndo(rope *Rope, undoInfo *UndoInfo) {
	_ = hm.Trigger(&HookContext{
		EventType: HookAfterUndo,
		Rope:      rope,
		UndoInfo:  undoInfo,
		Metadata:  make(map[string]interface{}),
	})
}

// TriggerBeforeRedo triggers all before-redo hooks.
// Returns an error if any hook cancels the operation.
func (hm *HookManager) TriggerBeforeRedo(rope *Rope, redoInfo *RedoInfo) error {
	return hm.Trigger(&HookContext{
		EventType: HookBeforeRedo,
		Rope:      rope,
		RedoInfo:  redoInfo,
		Metadata:  make(map[string]interface{}),
	})
}

// TriggerAfterRedo triggers all after-redo hooks.
func (hm *HookManager) TriggerAfterRedo(rope *Rope, redoInfo *RedoInfo) {
	_ = hm.Trigger(&HookContext{
		EventType: HookAfterRedo,
		Rope:      rope,
		RedoInfo:  redoInfo,
		Metadata:  make(map[string]interface{}),
	})
}

// TriggerOnBranch triggers all on-branch hooks.
func (hm *HookManager) TriggerOnBranch(rope *Rope) {
	_ = hm.Trigger(&HookContext{
		EventType: HookOnBranch,
		Rope:      rope,
		Metadata:  make(map[string]interface{}),
	})
}

// TriggerOnSavepoint triggers all on-savepoint hooks.
func (hm *HookManager) TriggerOnSavepoint(rope *Rope, savepointID int) {
	ctx := &HookContext{
		EventType: HookOnSavepoint,
		Rope:      rope,
		Metadata:  make(map[string]interface{}),
	}
	ctx.Metadata["savepoint_id"] = savepointID
	_ = hm.Trigger(ctx)
}

// TriggerOnError triggers all on-error hooks.
func (hm *HookManager) TriggerOnError(rope *Rope, err error) {
	_ = hm.Trigger(&HookContext{
		EventType: HookOnError,
		Rope:      rope,
		Error:     err,
		Metadata:  make(map[string]interface{}),
	})
}

// GetHooks returns all hooks registered for the specified event type.
func (hm *HookManager) GetHooks(eventType HookEventType) []*Hook {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	hooks := hm.hooks[eventType]
	result := make([]*Hook, len(hooks))
	copy(result, hooks)
	return result
}

// GetAllHooks returns all registered hooks grouped by event type.
func (hm *HookManager) GetAllHooks() map[HookEventType][]*Hook {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	result := make(map[HookEventType][]*Hook)
	for eventType, hooks := range hm.hooks {
		result[eventType] = make([]*Hook, len(hooks))
		copy(result[eventType], hooks)
	}
	return result
}

// Clear removes all registered hooks.
func (hm *HookManager) Clear() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.hooks = make(map[HookEventType][]*Hook)
}

// Count returns the total number of registered hooks.
func (hm *HookManager) Count() int {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	count := 0
	for _, hooks := range hm.hooks {
		count += len(hooks)
	}
	return count
}

// EnableAll enables all hooks.
func (hm *HookManager) EnableAll() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	for _, hooks := range hm.hooks {
		for _, hook := range hooks {
			hook.mu.Lock()
			hook.Enabled = true
			hook.mu.Unlock()
		}
	}
}

// DisableAll disables all hooks.
func (hm *HookManager) DisableAll() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	for _, hooks := range hm.hooks {
		for _, hook := range hooks {
			hook.mu.Lock()
			hook.Enabled = false
			hook.mu.Unlock()
		}
	}
}

// sortHooks sorts hooks for an event type by priority (highest first).
// Must be called with write lock held.
func (hm *HookManager) sortHooks(eventType HookEventType) {
	hooks := hm.hooks[eventType]
	// Sort by priority (highest first)
	for i := 0; i < len(hooks); i++ {
		for j := i + 1; j < len(hooks); j++ {
			if hooks[j].Priority > hooks[i].Priority {
				hooks[i], hooks[j] = hooks[j], hooks[i]
			}
		}
	}
}

// generateHookID generates a unique hook ID.
// Must be called with write lock held.
func (hm *HookManager) generateHookID() string {
	id := hm.nextID
	hm.nextID++
	return HookID(id)
}

// HookID converts a number to a hook ID string.
func HookID(id int) string {
	return "hook_" + string(rune('0'+id))
}

// isBeforeHook returns true if the event type is a "Before" hook.
func isBeforeHook(eventType HookEventType) bool {
	switch eventType {
	case HookBeforeEdit, HookBeforeUndo, HookBeforeRedo:
		return true
	default:
		return false
	}
}

// ============================================================================
// Built-in Hooks
// ============================================================================

// BuiltinHooks provides common built-in hooks for convenience.
type BuiltinHooks struct {
	// LimitEditSize creates a hook that limits edit size
	LimitEditSize func(maxSize int) HookFunc
	// LogEdit creates a hook that logs all edits
	LogEdit func(logger func(string)) HookFunc
	// ValidateEdit creates a hook that validates edits
	ValidateEdit func(validator func(*EditInfo) error) HookFunc
	// TrackMetrics creates a hook that tracks edit metrics
	TrackMetrics func(metrics *EditMetrics) HookFunc
}

// EditMetrics tracks statistics about edits.
type EditMetrics struct {
	TotalEdits         int64
	TotalInserts       int64
	TotalDeletes       int64
	TotalReplaces      int64
	TotalCharsInserted int64
	TotalCharsDeleted  int64
	mu                 sync.RWMutex
}

// RecordEdit records an edit operation.
func (em *EditMetrics) RecordEdit(edit *EditInfo) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.TotalEdits++

	switch edit.Operation {
	case "insert":
		em.TotalInserts++
		em.TotalCharsInserted += int64(edit.Length)
	case "delete":
		em.TotalDeletes++
		em.TotalCharsDeleted += int64(edit.Length)
	case "replace":
		em.TotalReplaces++
		em.TotalCharsInserted += int64(edit.Length)
	}
}

// Stats returns a snapshot of the metrics.
func (em *EditMetrics) Stats() map[string]int64 {
	em.mu.RLock()
	defer em.mu.RUnlock()

	return map[string]int64{
		"total_edits":          em.TotalEdits,
		"total_inserts":        em.TotalInserts,
		"total_deletes":        em.TotalDeletes,
		"total_replaces":       em.TotalReplaces,
		"total_chars_inserted": em.TotalCharsInserted,
		"total_chars_deleted":  em.TotalCharsDeleted,
	}
}

// Reset resets all metrics to zero.
func (em *EditMetrics) Reset() {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.TotalEdits = 0
	em.TotalInserts = 0
	em.TotalDeletes = 0
	em.TotalReplaces = 0
	em.TotalCharsInserted = 0
	em.TotalCharsDeleted = 0
}

// DefaultBuiltinHooks returns the default built-in hooks collection.
func DefaultBuiltinHooks() BuiltinHooks {
	return BuiltinHooks{
		LimitEditSize: func(maxSize int) HookFunc {
			return func(ctx *HookContext) error {
				if ctx.Edit != nil && ctx.Edit.Length > maxSize {
					return &HookError{
						Type:    HookBeforeEdit,
						Message: "edit size exceeds maximum allowed",
						Details: map[string]interface{}{
							"max_size":  maxSize,
							"edit_size": ctx.Edit.Length,
						},
					}
				}
				return nil
			}
		},

		LogEdit: func(logger func(string)) HookFunc {
			return func(ctx *HookContext) error {
				if ctx.Edit != nil {
					logger(ctx.Edit.String())
				}
				return nil
			}
		},

		ValidateEdit: func(validator func(*EditInfo) error) HookFunc {
			return func(ctx *HookContext) error {
				if ctx.Edit != nil {
					return validator(ctx.Edit)
				}
				return nil
			}
		},

		TrackMetrics: func(metrics *EditMetrics) HookFunc {
			return func(ctx *HookContext) error {
				if ctx.Edit != nil {
					metrics.RecordEdit(ctx.Edit)
				}
				return nil
			}
		},
	}
}

// String returns a string representation of the edit info.
func (ei *EditInfo) String() string {
	return ei.Operation + " at [" + string(rune(ei.StartPos)) + ":" + string(rune(ei.EndPos)) + "] len=" + string(rune(ei.Length))
}

// HookError represents an error that occurred during hook execution.
type HookError struct {
	Type    HookEventType
	Message string
	Details map[string]interface{}
}

func (he *HookError) Error() string {
	return he.Message
}
