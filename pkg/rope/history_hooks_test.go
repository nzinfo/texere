package rope

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// HookManager Tests
// ============================================================================

func TestHookManager_Register(t *testing.T) {
	hm := NewHookManager()

	called := false
	hook := hm.Register(HookBeforeEdit, "test-hook", 10, func(ctx *HookContext) error {
		called = true
		return nil
	})

	r := New("Hello")
	edit := &EditInfo{Operation: "insert"}
	hm.TriggerBeforeEdit(r, edit)

	assert.True(t, called)
	assert.NotNil(t, hook)
	assert.Equal(t, "test-hook", hook.Name)
	assert.Equal(t, 10, hook.Priority)
	assert.True(t, hook.IsEnabled())
	assert.Equal(t, 1, hm.Count())
}

func TestHookManager_Trigger(t *testing.T) {
	hm := NewHookManager()
	r := New("Hello World")

	callCount := 0
	hm.Register(HookBeforeEdit, "hook1", 10, func(ctx *HookContext) error {
		callCount++
		assert.Equal(t, r, ctx.Rope)
		return nil
	})

	hm.Register(HookBeforeEdit, "hook2", 5, func(ctx *HookContext) error {
		callCount++
		return nil
	})

	edit := &EditInfo{
		Operation: "insert",
		StartPos:  5,
		EndPos:    5,
		Text:      " Beautiful",
		Length:    10,
	}

	err := hm.TriggerBeforeEdit(r, edit)
	assert.NoError(t, err)
	assert.Equal(t, 2, callCount) // Both hooks should be called
}

func TestHookManager_Priority(t *testing.T) {
	hm := NewHookManager()
	r := New("Hello")

	callOrder := []int{}

	// Register hooks with different priorities
	hm.Register(HookBeforeEdit, "low", 1, func(ctx *HookContext) error {
		callOrder = append(callOrder, 1)
		return nil
	})

	hm.Register(HookBeforeEdit, "high", 10, func(ctx *HookContext) error {
		callOrder = append(callOrder, 10)
		return nil
	})

	hm.Register(HookBeforeEdit, "medium", 5, func(ctx *HookContext) error {
		callOrder = append(callOrder, 5)
		return nil
	})

	edit := &EditInfo{Operation: "insert"}
	hm.TriggerBeforeEdit(r, edit)

	// Should be called in priority order: high (10), medium (5), low (1)
	assert.Equal(t, []int{10, 5, 1}, callOrder)
}

func TestHookManager_EnableDisable(t *testing.T) {
	hm := NewHookManager()
	r := New("Hello")

	called := false
	hook := hm.Register(HookBeforeEdit, "test", 10, func(ctx *HookContext) error {
		called = true
		return nil
	})

	// Initially enabled
	edit := &EditInfo{Operation: "insert"}
	hm.TriggerBeforeEdit(r, edit)
	assert.True(t, called)

	// Disable
	called = false
	hook.Disable()
	assert.False(t, hook.IsEnabled())

	hm.TriggerBeforeEdit(r, edit)
	assert.False(t, called) // Should not be called

	// Re-enable
	hook.Enable()
	assert.True(t, hook.IsEnabled())

	called = false
	hm.TriggerBeforeEdit(r, edit)
	assert.True(t, called) // Should be called again
}

func TestHookManager_BeforeHookCancel(t *testing.T) {
	hm := NewHookManager()
	r := New("Hello")

	hm.Register(HookBeforeEdit, "cancel-hook", 10, func(ctx *HookContext) error {
		return errors.New("operation cancelled")
	})

	calledAfter := false
	hm.Register(HookAfterEdit, "after-hook", 10, func(ctx *HookContext) error {
		calledAfter = true
		return nil
	})

	edit := &EditInfo{Operation: "insert"}

	// Before hook should cancel the operation
	err := hm.TriggerBeforeEdit(r, edit)
	assert.Error(t, err)
	assert.Equal(t, "operation cancelled", err.Error())
	assert.False(t, calledAfter) // After hook was not called automatically
}

func TestHookManager_AfterHookError(t *testing.T) {
	hm := NewHookManager()
	r := New("Hello")

	hm.Register(HookAfterEdit, "error-hook", 10, func(ctx *HookContext) error {
		return errors.New("after hook error")
	})

	edit := &EditInfo{Operation: "insert"}

	// After hook errors should not stop the operation
	hm.TriggerAfterEdit(r, edit)
	// No error returned, operation continues
}

func TestHookManager_Unregister(t *testing.T) {
	hm := NewHookManager()

	hook := hm.Register(HookBeforeEdit, "test", 10, func(ctx *HookContext) error {
		return nil
	})

	assert.Equal(t, 1, hm.Count())

	// Unregister
	success := hm.Unregister(hook.ID)
	assert.True(t, success)
	assert.Equal(t, 0, hm.Count())

	// Try to unregister again
	success = hm.Unregister(hook.ID)
	assert.False(t, success)
}

func TestHookManager_GetHooks(t *testing.T) {
	hm := NewHookManager()

	hm.Register(HookBeforeEdit, "hook1", 10, func(ctx *HookContext) error {
		return nil
	})

	hm.Register(HookBeforeEdit, "hook2", 5, func(ctx *HookContext) error {
		return nil
	})

	hm.Register(HookAfterEdit, "hook3", 10, func(ctx *HookContext) error {
		return nil
	})

	// Get hooks for specific event
	beforeHooks := hm.GetHooks(HookBeforeEdit)
	assert.Equal(t, 2, len(beforeHooks))

	afterHooks := hm.GetHooks(HookAfterEdit)
	assert.Equal(t, 1, len(afterHooks))
}

func TestHookManager_GetAllHooks(t *testing.T) {
	hm := NewHookManager()

	hm.Register(HookBeforeEdit, "before", 10, func(ctx *HookContext) error {
		return nil
	})

	hm.Register(HookAfterEdit, "after", 10, func(ctx *HookContext) error {
		return nil
	})

	allHooks := hm.GetAllHooks()
	assert.Equal(t, 2, len(allHooks))
	assert.Equal(t, 1, len(allHooks[HookBeforeEdit]))
	assert.Equal(t, 1, len(allHooks[HookAfterEdit]))
}

func TestHookManager_Clear(t *testing.T) {
	hm := NewHookManager()

	hm.Register(HookBeforeEdit, "hook1", 10, func(ctx *HookContext) error {
		return nil
	})

	hm.Register(HookAfterEdit, "hook2", 10, func(ctx *HookContext) error {
		return nil
	})

	assert.Equal(t, 2, hm.Count())

	hm.Clear()

	assert.Equal(t, 0, hm.Count())
}

func TestHookManager_EnableAll_DisableAll(t *testing.T) {
	hm := NewHookManager()

	hook1 := hm.Register(HookBeforeEdit, "hook1", 10, func(ctx *HookContext) error {
		return nil
	})

	hook2 := hm.Register(HookAfterEdit, "hook2", 10, func(ctx *HookContext) error {
		return nil
	})

	// All enabled by default
	assert.True(t, hook1.IsEnabled())
	assert.True(t, hook2.IsEnabled())

	// Disable all
	hm.DisableAll()
	assert.False(t, hook1.IsEnabled())
	assert.False(t, hook2.IsEnabled())

	// Enable all
	hm.EnableAll()
	assert.True(t, hook1.IsEnabled())
	assert.True(t, hook2.IsEnabled())
}

// ============================================================================
// Hook Context Tests
// ============================================================================

func TestHookContext_Metadata(t *testing.T) {
	hm := NewHookManager()
	r := New("Hello")

	receivedMetadata := make(map[string]interface{})

	hm.Register(HookBeforeEdit, "test", 10, func(ctx *HookContext) error {
		// Add metadata
		ctx.Metadata["key1"] = "value1"
		ctx.Metadata["key2"] = 42
		return nil
	})

	hm.Register(HookBeforeEdit, "test2", 5, func(ctx *HookContext) error {
		// Read metadata
		receivedMetadata = ctx.Metadata
		return nil
	})

	edit := &EditInfo{Operation: "insert"}
	hm.TriggerBeforeEdit(r, edit)

	assert.Equal(t, "value1", receivedMetadata["key1"])
	assert.Equal(t, 42, receivedMetadata["key2"])
}

// ============================================================================
// Built-in Hooks Tests
// ============================================================================

func TestBuiltinHook_LimitEditSize(t *testing.T) {
	hm := NewHookManager()
	r := New("Hello")
	builtin := DefaultBuiltinHooks()

	// Register hook with 100 char limit
	hm.Register(HookBeforeEdit, "limit", 10, builtin.LimitEditSize(100))

	// Small edit should pass
	smallEdit := &EditInfo{Operation: "insert", Length: 10}
	err := hm.TriggerBeforeEdit(r, smallEdit)
	assert.NoError(t, err)

	// Large edit should fail
	largeEdit := &EditInfo{Operation: "insert", Length: 150}
	err = hm.TriggerBeforeEdit(r, largeEdit)
	assert.Error(t, err)

	hookErr, ok := err.(*HookError)
	assert.True(t, ok)
	assert.Equal(t, "edit size exceeds maximum allowed", hookErr.Message)
}

func TestBuiltinHook_LogEdit(t *testing.T) {
	hm := NewHookManager()
	r := New("Hello")
	builtin := DefaultBuiltinHooks()

	logs := []string{}
	logger := func(msg string) {
		logs = append(logs, msg)
	}

	hm.Register(HookAfterEdit, "logger", 10, builtin.LogEdit(logger))

	edit := &EditInfo{
		Operation: "insert",
		StartPos:  5,
		EndPos:    5,
		Text:      " World",
		Length:    6,
	}

	hm.TriggerAfterEdit(r, edit)

	assert.Equal(t, 1, len(logs))
	assert.Contains(t, logs[0], "insert")
}

func TestBuiltinHook_ValidateEdit(t *testing.T) {
	hm := NewHookManager()
	r := New("Hello")
	builtin := DefaultBuiltinHooks()

	// Create validator that rejects empty text
	validator := func(edit *EditInfo) error {
		if edit.Operation == "insert" && edit.Text == "" {
			return errors.New("cannot insert empty text")
		}
		return nil
	}

	hm.Register(HookBeforeEdit, "validator", 10, builtin.ValidateEdit(validator))

	// Valid edit
	validEdit := &EditInfo{Operation: "insert", Text: "Hello", Length: 5}
	err := hm.TriggerBeforeEdit(r, validEdit)
	assert.NoError(t, err)

	// Invalid edit
	invalidEdit := &EditInfo{Operation: "insert", Text: "", Length: 0}
	err = hm.TriggerBeforeEdit(r, invalidEdit)
	assert.Error(t, err)
	assert.Equal(t, "cannot insert empty text", err.Error())
}

func TestBuiltinHook_TrackMetrics(t *testing.T) {
	hm := NewHookManager()
	r := New("Hello")
	builtin := DefaultBuiltinHooks()

	metrics := &EditMetrics{}

	hm.Register(HookAfterEdit, "metrics", 10, builtin.TrackMetrics(metrics))

	// Simulate edits
	edits := []*EditInfo{
		{Operation: "insert", Length: 5},
		{Operation: "delete", Length: 3},
		{Operation: "insert", Length: 10},
		{Operation: "replace", Length: 7},
	}

	for _, edit := range edits {
		hm.TriggerAfterEdit(r, edit)
	}

	stats := metrics.Stats()
	assert.Equal(t, int64(4), stats["total_edits"])
	assert.Equal(t, int64(2), stats["total_inserts"])
	assert.Equal(t, int64(1), stats["total_deletes"])
	assert.Equal(t, int64(1), stats["total_replaces"])
	assert.Equal(t, int64(22), stats["total_chars_inserted"]) // 5 + 10 + 7 (replace also inserts)
	assert.Equal(t, int64(3), stats["total_chars_deleted"])

	// Reset
	metrics.Reset()
	stats = metrics.Stats()
	assert.Equal(t, int64(0), stats["total_edits"])
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestHookManager_ComplexWorkflow(t *testing.T) {
	hm := NewHookManager()
	r := New("Hello")
	builtin := DefaultBuiltinHooks()

	// Setup metrics
	metrics := &EditMetrics{}

	// Register multiple hooks
	hm.Register(HookBeforeEdit, "validate", 30, builtin.ValidateEdit(func(edit *EditInfo) error {
		if edit.Length > 1000 {
			return errors.New("edit too large")
		}
		return nil
	}))

	hm.Register(HookBeforeEdit, "limit", 20, builtin.LimitEditSize(100))

	hm.Register(HookAfterEdit, "metrics", 10, builtin.TrackMetrics(metrics))

	// Valid edit
	validEdit := &EditInfo{Operation: "insert", Length: 50, Text: "Test"}
	err := hm.TriggerBeforeEdit(r, validEdit)
	assert.NoError(t, err)

	hm.TriggerAfterEdit(r, validEdit)

	stats := metrics.Stats()
	assert.Equal(t, int64(1), stats["total_edits"])

	// Invalid edit (too large)
	invalidEdit := &EditInfo{Operation: "insert", Length: 2000}
	err = hm.TriggerBeforeEdit(r, invalidEdit)
	assert.Error(t, err)
}

func TestHookManager_Concurrent(t *testing.T) {
	hm := NewHookManager()
	r := New("Hello")

	// Register multiple hooks
	for i := 0; i < 10; i++ {
		hm.Register(HookBeforeEdit, "hook"+string(rune('0'+i)), i, func(ctx *HookContext) error {
			time.Sleep(1 * time.Millisecond)
			return nil
		})
	}

	edit := &EditInfo{Operation: "insert"}

	// Trigger hooks
	done := make(chan bool)
	go func() {
		hm.TriggerBeforeEdit(r, edit)
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Hooks took too long")
	}
}

func TestHookManager_AllEventTypes(t *testing.T) {
	hm := NewHookManager()
	r := New("Hello")

	events := make(map[HookEventType]int)

	// Register hooks for all event types
	hm.Register(HookBeforeEdit, "before-edit", 10, func(ctx *HookContext) error {
		events[HookBeforeEdit]++
		return nil
	})

	hm.Register(HookAfterEdit, "after-edit", 10, func(ctx *HookContext) error {
		events[HookAfterEdit]++
		return nil
	})

	hm.Register(HookBeforeUndo, "before-undo", 10, func(ctx *HookContext) error {
		events[HookBeforeUndo]++
		return nil
	})

	hm.Register(HookAfterUndo, "after-undo", 10, func(ctx *HookContext) error {
		events[HookAfterUndo]++
		return nil
	})

	hm.Register(HookBeforeRedo, "before-redo", 10, func(ctx *HookContext) error {
		events[HookBeforeRedo]++
		return nil
	})

	hm.Register(HookAfterRedo, "after-redo", 10, func(ctx *HookContext) error {
		events[HookAfterRedo]++
		return nil
	})

	hm.Register(HookOnBranch, "on-branch", 10, func(ctx *HookContext) error {
		events[HookOnBranch]++
		return nil
	})

	hm.Register(HookOnSavepoint, "on-savepoint", 10, func(ctx *HookContext) error {
		events[HookOnSavepoint]++
		return nil
	})

	hm.Register(HookOnError, "on-error", 10, func(ctx *HookContext) error {
		events[HookOnError]++
		return nil
	})

	// Trigger all events
	edit := &EditInfo{Operation: "insert"}
	hm.TriggerBeforeEdit(r, edit)
	hm.TriggerAfterEdit(r, edit)

	undoInfo := &UndoInfo{RevisionID: 1}
	hm.TriggerBeforeUndo(r, undoInfo)
	hm.TriggerAfterUndo(r, undoInfo)

	redoInfo := &RedoInfo{RevisionID: 2}
	hm.TriggerBeforeRedo(r, redoInfo)
	hm.TriggerAfterRedo(r, redoInfo)

	hm.TriggerOnBranch(r)
	hm.TriggerOnSavepoint(r, 123)
	hm.TriggerOnError(r, errors.New("test error"))

	// Verify all events were triggered
	assert.Equal(t, 1, events[HookBeforeEdit])
	assert.Equal(t, 1, events[HookAfterEdit])
	assert.Equal(t, 1, events[HookBeforeUndo])
	assert.Equal(t, 1, events[HookAfterUndo])
	assert.Equal(t, 1, events[HookBeforeRedo])
	assert.Equal(t, 1, events[HookAfterRedo])
	assert.Equal(t, 1, events[HookOnBranch])
	assert.Equal(t, 1, events[HookOnSavepoint])
	assert.Equal(t, 1, events[HookOnError])
}
