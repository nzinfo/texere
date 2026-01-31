package concordia

import (
	"testing"

	"github.com/coreseekdev/texere/pkg/ot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUndoManager_Basic tests basic undo/redo functionality.
func TestUndoManager_Basic(t *testing.T) {
	um := NewUndoManager(50)

	// Initial state
	assert.False(t, um.CanUndo())
	assert.False(t, um.CanRedo())
	assert.Equal(t, 0, um.UndoStackLength())
	assert.Equal(t, 0, um.RedoStackLength())

	// Add an operation
	op := ot.NewBuilder().Insert("Hello").Build()
	um.Add(op, true)

	// Should be able to undo
	assert.True(t, um.CanUndo())
	assert.False(t, um.CanRedo())
	assert.Equal(t, 1, um.UndoStackLength())

	// Undo
	var undoneOp *ot.Operation
	err := um.PerformUndo(func(op *ot.Operation) {
		undoneOp = op
		// Add the inverse operation to redo stack
		// In this case, op is Insert("Hello"), so we add it directly
		um.Add(op, false)
	})
	require.NoError(t, err)
	assert.Equal(t, op, undoneOp)

	// Should be able to redo
	assert.False(t, um.CanUndo())
	assert.True(t, um.CanRedo())
	assert.Equal(t, 0, um.UndoStackLength())
	assert.Equal(t, 1, um.RedoStackLength())

	// Redo
	var redoneOp *ot.Operation
	err = um.PerformRedo(func(op *ot.Operation) {
		redoneOp = op
		// Add back to undo stack
		um.Add(op, false)
	})
	require.NoError(t, err)
	assert.Equal(t, op, redoneOp)

	// Should be able to undo again
	assert.True(t, um.CanUndo())
	assert.False(t, um.CanRedo())
}

// TestUndoManager_Compose tests operation composition.
func TestUndoManager_Compose(t *testing.T) {
	um := NewUndoManager(50)

	// Add consecutive insert operations
	op1 := ot.NewBuilder().Retain(0).Insert("H").Build()
	um.Add(op1, true)

	op2 := ot.NewBuilder().Retain(1).Insert("e").Build()
	um.Add(op2, true)

	op3 := ot.NewBuilder().Retain(2).Insert("l").Build()
	um.Add(op3, true)

	// Should be composed into a single operation
	assert.Equal(t, 1, um.UndoStackLength())

	// Undo should undo all three inserts at once
	var undoneOp *ot.Operation
	err := um.PerformUndo(func(op *ot.Operation) {
		undoneOp = op
	})
	require.NoError(t, err)

	// The undone operation should be the composition
	assert.NotNil(t, undoneOp)
}

// TestUndoManager_Transform tests stack transformation.
func TestUndoManager_Transform(t *testing.T) {
	um := NewUndoManager(50)

	// Add an operation to the undo stack
	// This represents: apply to a doc of length 5, retain 5, then insert "Hello"
	op1 := ot.NewBuilder().Retain(5).Insert("Hello").Build()
	um.Add(op1, true)

	// Simulate a remote operation that also operates on a doc of length 5
	// For example, retain 2 then insert "Hi" (position 2 insertion)
	remoteOp := ot.NewBuilder().Retain(2).Insert("Hi").Retain(3).Build()

	// Transform the undo stack
	err := um.Transform(remoteOp)
	require.NoError(t, err)

	// The undo stack should still have one operation
	assert.Equal(t, 1, um.UndoStackLength())

	// The operation should have been transformed
	var undoneOp *ot.Operation
	err = um.PerformUndo(func(op *ot.Operation) {
		undoneOp = op
	})
	require.NoError(t, err)
	assert.NotNil(t, undoneOp)
}

// TestUndoManager_Concurrent tests concurrent safety.
func TestUndoManager_Concurrent(t *testing.T) {
	um := NewUndoManager(50)

	// Add operations from multiple goroutines
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			op := ot.NewBuilder().Insert(string(rune('A' + n))).Build()
			um.Add(op, false)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 10 operations in the undo stack
	assert.Equal(t, 10, um.UndoStackLength())
}

// TestUndoManager_Clear tests clearing the stacks.
func TestUndoManager_Clear(t *testing.T) {
	um := NewUndoManager(50)

	// Add some operations
	for i := 0; i < 5; i++ {
		op := ot.NewBuilder().Insert("test").Build()
		um.Add(op, false)
	}

	assert.Equal(t, 5, um.UndoStackLength())

	// Clear
	um.Clear()

	assert.Equal(t, 0, um.UndoStackLength())
	assert.Equal(t, 0, um.RedoStackLength())
	assert.False(t, um.CanUndo())
	assert.False(t, um.CanRedo())
}

// TestUndoManager_MaxItems tests stack size limiting.
func TestUndoManager_MaxItems(t *testing.T) {
	um := NewUndoManager(5) // Max 5 items

	// Add 10 operations
	for i := 0; i < 10; i++ {
		op := ot.NewBuilder().Insert(string(rune('A' + i))).Build()
		um.Add(op, false)
	}

	// Should only have 5 items
	assert.Equal(t, 5, um.UndoStackLength())
}

// TestUndoManager_State tests undo/redo state tracking.
func TestUndoManager_State(t *testing.T) {
	um := NewUndoManager(50)

	// Normal state
	assert.False(t, um.IsUndoing())
	assert.False(t, um.IsRedoing())

	// Add an operation
	op := ot.NewBuilder().Insert("Hello").Build()
	um.Add(op, true)

	// Start undo
	err := um.PerformUndo(func(op *ot.Operation) {
		assert.True(t, um.IsUndoing())
		assert.False(t, um.IsRedoing())
		um.Add(op, false) // Add to redo stack
	})
	require.NoError(t, err)

	// Back to normal
	assert.False(t, um.IsUndoing())

	// Start redo
	err = um.PerformRedo(func(op *ot.Operation) {
		assert.True(t, um.IsRedoing())
		assert.False(t, um.IsUndoing())
		um.Add(op, false) // Add back to undo stack
	})
	require.NoError(t, err)

	// Back to normal
	assert.False(t, um.IsRedoing())
}

// TestUndoManager_EmptyStack tests undo/redo on empty stacks.
func TestUndoManager_EmptyStack(t *testing.T) {
	um := NewUndoManager(50)

	// Try to undo when empty
	err := um.PerformUndo(func(op *ot.Operation) {
		t.Fatal("Should not call callback on empty stack")
	})
	assert.Equal(t, ot.ErrCannotUndo, err)

	// Try to redo when empty
	err = um.PerformRedo(func(op *ot.Operation) {
		t.Fatal("Should not call callback on empty stack")
	})
	assert.Equal(t, ot.ErrCannotRedo, err)
}

// TestUndoManager_DontCompose tests that dontCompose prevents composition.
func TestUndoManager_DontCompose(t *testing.T) {
	um := NewUndoManager(50)

	// Add operations with compose=false
	op1 := ot.NewBuilder().Insert("H").Build()
	um.Add(op1, false)

	op2 := ot.NewBuilder().Insert("e").Build()
	um.Add(op2, false)

	// Should not be composed
	assert.Equal(t, 2, um.UndoStackLength())
}

// TestUndoManager_RedoStackCleared tests that redo stack is cleared on new operation.
func TestUndoManager_RedoStackCleared(t *testing.T) {
	um := NewUndoManager(50)

	// Add and undo an operation
	op1 := ot.NewBuilder().Insert("Hello").Build()
	um.Add(op1, true)

	err := um.PerformUndo(func(op *ot.Operation) {
		um.Add(op, false) // Add to redo stack
	})
	require.NoError(t, err)

	assert.True(t, um.CanRedo())

	// Add a new operation
	op2 := ot.NewBuilder().Insert("World").Build()
	um.Add(op2, true)

	// Redo stack should be cleared
	assert.False(t, um.CanRedo())
}
