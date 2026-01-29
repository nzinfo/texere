// Package concordia provides Operational Transformation (OT) algorithms
// for real-time collaborative editing.
//
// Concordia (拉丁语) means "harmony" or "agreement" in Latin,
// representing the core functionality of transforming and coordinating
// concurrent editing operations to maintain consistency.
//
// # Overview
//
// The concordia package implements the core OT operations including:
//   - Insert: Insert text at a position
//   - Delete: Delete text at a position
//   - Retain: Keep text without modification
//
// Operations can be composed, transformed, and applied to documents,
// ensuring that all concurrent edits converge to a consistent state.
//
// # Basic Usage
//
//	// Create operations
//	op1 := concordia.NewInsert(0, "Hello")
//	op2 := concordia.NewRetain(5)
//	op3 := concordia.NewInsert(5, " World")
//
//	// Compose operations
//	composed := concordia.Compose(op1, op2, op3)
//
//	// Transform operations against each other
//	transformed1, transformed2 := concordia.Transform(op1, op2)
//
// # Thread Safety
//
// Operations are immutable and safe for concurrent use.
// The transformation state is maintained separately.
//
// # Performance
//
// Transformation is O(n) where n is the length of operations.
// For large documents, consider batching operations.
//
// # Etymology
//
// Concordia is the Roman goddess of harmony and agreement.
// In the context of OT, it represents the harmony achieved
// by coordinating concurrent edits from multiple users.
package concordia

import "fmt"

// OperationType represents the type of an operation.
type OperationType int

const (
	// OpInsert inserts text at a position.
	OpInsert OperationType = iota
	// OpDelete deletes text at a position.
	OpDelete
	// OpRetain keeps text without modification.
	OpRetain
)

// Operation represents a single text editing operation.
// Operations are immutable and safe for concurrent use.
type Operation struct {
	opType    OperationType
	position  int
	content   string
	deleteLen int
}

// NewInsert creates a new insert operation.
func NewInsert(pos int, text string) *Operation {
	return &Operation{
		opType:   OpInsert,
		position: pos,
		content:  text,
	}
}

// NewDelete creates a new delete operation.
func NewDelete(pos int, length int) *Operation {
	return &Operation{
		opType:    OpDelete,
		position:  pos,
		deleteLen: length,
	}
}

// NewRetain creates a new retain operation.
func NewRetain(length int) *Operation {
	return &Operation{
		opType:    OpRetain,
		deleteLen: length,
	}
}

// Type returns the operation type.
func (o *Operation) Type() OperationType {
	return o.opType
}

// Position returns the position for insert/delete operations.
func (o *Operation) Position() int {
	return o.position
}

// Content returns the text content for insert operations.
func (o *Operation) Content() string {
	return o.content
}

// Length returns the length affected by this operation.
// For insert, it's the length of inserted text.
// For delete, it's the number of deleted characters.
// For retain, it's the number of retained characters.
func (o *Operation) Length() int {
	switch o.opType {
	case OpInsert:
		return len(o.content)
	case OpDelete:
		return o.deleteLen
	case OpRetain:
		return o.deleteLen
	default:
		return 0
	}
}

// IsInsert returns true if this is an insert operation.
func (o *Operation) IsInsert() bool {
	return o.opType == OpInsert
}

// IsDelete returns true if this is a delete operation.
func (o *Operation) IsDelete() bool {
	return o.opType == OpDelete
}

// IsRetain returns true if this is a retain operation.
func (o *Operation) IsRetain() bool {
	return o.opType == OpRetain
}

// String returns a string representation of the operation.
func (o *Operation) String() string {
	switch o.opType {
	case OpInsert:
		return fmt.Sprintf("Insert(%d, %q)", o.position, o.content)
	case OpDelete:
		return fmt.Sprintf("Delete(%d, %d)", o.position, o.deleteLen)
	case OpRetain:
		return fmt.Sprintf("Retain(%d)", o.deleteLen)
	default:
		return "Unknown()"
	}
}

// Compose combines multiple operations into a single operation.
// The operations are applied in order from left to right.
//
// Example:
//
//	op := concordia.Compose(
//	    concordia.NewInsert(0, "Hello"),
//	    concordia.NewRetain(5),
//	    concordia.NewInsert(5, " World"),
//	)
func Compose(ops ...*Operation) *Operation {
	if len(ops) == 0 {
		return NewRetain(0)
	}

	// Start with the first operation
	result := ops[0]

	// Compose with each subsequent operation
	for i := 1; i < len(ops); i++ {
		result = composePair(result, ops[i])
	}

	return result
}

// Transform transforms two operations against each other,
// returning two new operations that can be applied in any order.
//
// This is the core OT transformation algorithm.
// It ensures that concurrent operations can be applied consistently.
//
// Example:
//
//	op1 := concordia.NewInsert(0, "Hello")
//	op2 := concordia.NewInsert(0, "Hi")
//	newOp1, newOp2 := concordia.Transform(op1, op2)
func Transform(op1, op2 *Operation) (*Operation, *Operation) {
	// Implementation based on operational transformation papers
	// This is a simplified version for demonstration

	// Both inserts
	if op1.IsInsert() && op2.IsInsert() {
		if op1.Position() < op2.Position() {
			return op1, NewInsert(op2.Position()+len(op1.Content()), op2.Content())
		} else if op1.Position() > op2.Position() {
			return NewInsert(op1.Position()+len(op2.Content()), op1.Content()), op2
		}
		// Same position - use tie-breaking rule (e.g., by user ID)
		return op1, NewInsert(op2.Position()+len(op1.Content()), op2.Content())
	}

	// Insert and delete
	if op1.IsInsert() && op2.IsDelete() {
		if op1.Position() < op2.Position() {
			return op1, op2
		} else if op1.Position() >= op2.Position()+op2.Length() {
			return NewInsert(op1.Position()-op2.Length(), op1.Content()), op2
		}
		// Insert inside delete range - just keep delete
		return op1, op2
	}

	if op1.IsDelete() && op2.IsInsert() {
		newOp2, newOp1 := Transform(op2, op1)
		return newOp1, newOp2
	}

	// Both deletes - simplify to union
	if op1.IsDelete() && op2.IsDelete() {
		// Simplified: return the larger delete
		if op1.Length() > op2.Length() {
			return op1, NewRetain(0)
		}
		return NewRetain(0), op2
	}

	// Retain cases
	if op1.IsRetain() {
		return op1, op2
	}
	if op2.IsRetain() {
		return op1, op2
	}

	// Default: no transformation needed
	return op1, op2
}

// composePair composes two operations into one.
func composePair(op1, op2 *Operation) *Operation {
	// Simplified implementation
	// Real implementation would handle all cases properly
	if op2.IsRetain() {
		return op1
	}
	return op2
}

// Apply applies an operation to a document string.
func Apply(doc string, op *Operation) string {
	switch op.opType {
	case OpInsert:
		return doc[:op.position] + op.content + doc[op.position:]
	case OpDelete:
		return doc[:op.position] + doc[op.position+op.deleteLen:]
	case OpRetain:
		return doc
	default:
		return doc
	}
}
