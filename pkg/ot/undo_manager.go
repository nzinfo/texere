package ot

import (
	"sync"
)

// UndoManagerState represents the current state of the undo manager.
type UndoManagerState int

const (
	// StateNormal is the default state when not undoing or redoing.
	StateNormal UndoManagerState = iota
	// StateUndoing indicates that an undo operation is in progress.
	StateUndoing
	// StateRedoing indicates that a redo operation is in progress.
	StateRedoing
)

// UndoManager manages undo/redo stacks for OT operations.
//
// The UndoManager is compatible with collaborative editing and supports
// operation transformation when remote operations are received. It's based
// on the ot.js UndoManager implementation.
//
// Example usage:
//
//	um := NewUndoManager(50)
//
//	// Apply an operation
//	op := NewBuilder().Insert("Hello").Build()
//	doc, _ := op.Apply(doc)
//
//	// Add the inverse to the undo stack
//	inverse, _ := op.Invert(doc)
//	um.Add(inverse, true) // compose = true to merge consecutive operations
//
//	// Undo
//	um.PerformUndo(func(op *Operation) {
//	    doc, _ = op.Apply(doc)
//	})
type UndoManager struct {
	mu          sync.RWMutex
	maxItems    int
	state       UndoManagerState
	dontCompose bool
	undoStack   []*Operation
	redoStack   []*Operation
}

// NewUndoManager creates a new undo manager.
//
// Parameters:
//   - maxItems: maximum number of items to keep in each stack (0 for unlimited)
//
// Returns:
//   - a new UndoManager
//
// Example:
//
//	um := NewUndoManager(50) // Keep up to 50 operations
func NewUndoManager(maxItems int) *UndoManager {
	if maxItems <= 0 {
		maxItems = 50 // Default
	}
	return &UndoManager{
		maxItems:  maxItems,
		state:     StateNormal,
		undoStack: make([]*Operation, 0, maxItems),
		redoStack: make([]*Operation, 0, maxItems),
	}
}

// Add adds an operation to the undo or redo stack.
//
// The behavior depends on the current state:
//   - StateNormal: adds to undo stack, clears redo stack
//   - StateUndoing: adds to redo stack
//   - StateRedoing: adds to undo stack
//
// Parameters:
//   - operation: the operation to add
//   - compose: if true, try to compose with the previous operation
//
// Example:
//
//	// When applying an operation locally
//	op := NewBuilder().Insert("Hello").Build()
//	inverse, _ := op.Invert(doc)
//	um.Add(inverse, true) // Compose with previous operation
func (um *UndoManager) Add(operation *Operation, compose bool) {
	um.mu.Lock()
	defer um.mu.Unlock()

	switch um.state {
	case StateUndoing:
		// Add to redo stack
		um.redoStack = append(um.redoStack, operation)
		um.dontCompose = true

	case StateRedoing:
		// Add to undo stack
		um.undoStack = append(um.undoStack, operation)
		um.dontCompose = true

	case StateNormal:
		// Add to undo stack
		if !um.dontCompose && compose && len(um.undoStack) > 0 {
			// Try to compose with the last operation
			lastOp := um.undoStack[len(um.undoStack)-1]
			// Fix: check lastOp.ShouldBeComposedWith(operation) not the reverse
			if lastOp.ShouldBeComposedWith(operation) {
				composedOp, err := Compose(lastOp, operation)
				if err == nil {
					um.undoStack[len(um.undoStack)-1] = composedOp
				} else {
					um.undoStack = append(um.undoStack, operation)
				}
			} else {
				um.undoStack = append(um.undoStack, operation)
			}
		} else {
			// Don't compose, just add
			um.undoStack = append(um.undoStack, operation)

			// Limit stack size
			if len(um.undoStack) > um.maxItems {
				// Remove oldest operation
				um.undoStack = um.undoStack[1:]
			}
		}

		um.dontCompose = false
		// Clear redo stack (new operation makes redo invalid)
		um.redoStack = um.redoStack[:0]
	}
}

// Transform transforms both undo and redo stacks against a remote operation.
//
// This should be called when a remote operation is received, before applying
// it to the concordia. This ensures that the undo/redo history remains
// consistent with the document state.
//
// Parameters:
//   - operation: the remote operation to transform against
//
// Returns:
//   - an error if transformation fails
//
// Example:
//
//	// Receive remote operation
//	remoteOp := // ... from network
//
//	// Transform undo/redo stacks
//	err := um.Transform(remoteOp)
//	if err != nil {
//	    // Handle error
//	}
//
//	// Now apply the remote operation
//	doc, err = remoteOp.Apply(doc)
func (um *UndoManager) Transform(operation *Operation) error {
	um.mu.Lock()
	defer um.mu.Unlock()

	var err error
	um.undoStack, err = transformStack(um.undoStack, operation)
	if err != nil {
		return err
	}

	um.redoStack, err = transformStack(um.redoStack, operation)
	return err
}

// transformStack transforms a stack of operations against a single operation.
//
// This is a helper function for Transform. It transforms each operation
// in the stack against the given operation.
func transformStack(stack []*Operation, operation *Operation) ([]*Operation, error) {
	newStack := make([]*Operation, 0, len(stack))

	// Transform from newest to oldest (reverse order)
	for i := len(stack) - 1; i >= 0; i-- {
		opPrime, operationPrime, err := Transform(stack[i], operation)
		if err != nil {
			return nil, err
		}

		// Only add if the transformed operation is not a no-op
		if !opPrime.IsNoop() {
			newStack = append(newStack, opPrime)
		}

		// Update operation for next iteration
		operation = operationPrime
	}

	// Reverse to get correct order (oldest first)
	for i, j := 0, len(newStack)-1; i < j; i, j = i+1, j-1 {
		newStack[i], newStack[j] = newStack[j], newStack[i]
	}

	return newStack, nil
}

// PerformUndo executes an undo operation.
//
// The callback function receives the operation to undo. The UndoManager
// automatically manages the redo stack.
//
// Parameters:
//   - fn: callback function that receives the operation to undo
//
// Returns:
//   - an error if undo is not possible
//
// Example:
//
//	err := um.PerformUndo(func(op *Operation) {
//	    // Apply the inverse operation
//	    doc, _ = op.Apply(doc)
//	})
func (um *UndoManager) PerformUndo(fn func(op *Operation)) error {
	um.mu.Lock()

	if len(um.undoStack) == 0 {
		um.mu.Unlock()
		return ErrCannotUndo
	}

	um.state = StateUndoing

	// Pop the last operation
	op := um.undoStack[len(um.undoStack)-1]
	um.undoStack = um.undoStack[:len(um.undoStack)-1]

	// Release lock before calling callback to allow um.Add() inside callback
	um.mu.Unlock()

	// Call the callback (may call um.Add())
	fn(op)

	// Re-acquire lock to reset state
	um.mu.Lock()
	um.state = StateNormal
	um.mu.Unlock()

	return nil
}

// PerformRedo executes a redo operation.
//
// The callback function receives the operation to redo. The caller should
// apply this operation to the document and then add the inverse back to
// the UndoManager.
//
// Parameters:
//   - fn: callback function that receives the operation to redo
//
// Returns:
//   - an error if redo is not possible
//
// Example:
//
//	err := um.PerformRedo(func(op *Operation) {
//	    // Apply the operation
//	    doc, _ = op.Apply(doc)
//
//	    // Add the inverse to the undo stack
//	    undoOp, _ := op.Invert(doc)
//	    um.Add(undoOp, false)
//	})
func (um *UndoManager) PerformRedo(fn func(op *Operation)) error {
	um.mu.Lock()

	if len(um.redoStack) == 0 {
		um.mu.Unlock()
		return ErrCannotRedo
	}

	um.state = StateRedoing

	// Pop the last operation
	op := um.redoStack[len(um.redoStack)-1]
	um.redoStack = um.redoStack[:len(um.redoStack)-1]

	// Release lock before calling callback to allow um.Add() inside callback
	um.mu.Unlock()

	// Call the callback (may call um.Add())
	fn(op)

	// Re-acquire lock to reset state
	um.mu.Lock()
	um.state = StateNormal
	um.mu.Unlock()

	return nil
}

// CanUndo returns true if undo is possible.
func (um *UndoManager) CanUndo() bool {
	um.mu.RLock()
	defer um.mu.RUnlock()
	return len(um.undoStack) > 0
}

// CanRedo returns true if redo is possible.
func (um *UndoManager) CanRedo() bool {
	um.mu.RLock()
	defer um.mu.RUnlock()
	return len(um.redoStack) > 0
}

// IsUndoing returns true if an undo operation is in progress.
func (um *UndoManager) IsUndoing() bool {
	um.mu.RLock()
	defer um.mu.RUnlock()
	return um.state == StateUndoing
}

// IsRedoing returns true if a redo operation is in progress.
func (um *UndoManager) IsRedoing() bool {
	um.mu.RLock()
	defer um.mu.RUnlock()
	return um.state == StateRedoing
}

// Clear clears both undo and redo stacks.
//
// This can be useful when resetting the document or when the history
// becomes invalid.
func (um *UndoManager) Clear() {
	um.mu.Lock()
	defer um.mu.Unlock()

	um.undoStack = um.undoStack[:0]
	um.redoStack = um.redoStack[:0]
}

// UndoStackLength returns the number of operations in the undo stack.
func (um *UndoManager) UndoStackLength() int {
	um.mu.RLock()
	defer um.mu.RUnlock()
	return len(um.undoStack)
}

// RedoStackLength returns the number of operations in the redo stack.
func (um *UndoManager) RedoStackLength() int {
	um.mu.RLock()
	defer um.mu.RUnlock()
	return len(um.redoStack)
}
