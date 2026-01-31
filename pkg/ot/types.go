package ot

import "fmt"

// OperationType represents the type of an OT operation.
type OperationType int

const (
	// OpRetain retains (skips over) characters without modification.
	OpRetain OperationType = iota
	// OpInsert inserts new text at the current position.
	OpInsert
	// OpDelete removes characters from the current position.
	OpDelete
)

// Op is the interface for all operation types.
//
// In the ot.js implementation, operations are represented as:
//   - positive numbers: retain operations
//   - strings: insert operations
//   - negative numbers: delete operations
//
// In Go, we use a type-based approach with explicit Op types for better
// type safety and performance.
type Op interface {
	// Type returns the operation type.
	Type() OperationType
	// Length returns the length of the operation.
	// For retain: number of characters retained
	// For insert: length of inserted string
	// For delete: number of characters deleted
	Length() int
	// String returns a string representation for debugging.
	String() string
}

// RetainOp retains (skips over) characters without modification.
//
// Represented as a positive integer in the original ot.js implementation.
// Example: RetainOp(5) means "skip over the next 5 characters"
type RetainOp int

// Type returns OpRetain for RetainOp.
func (o RetainOp) Type() OperationType {
	return OpRetain
}

// Length returns the number of characters to retain.
func (o RetainOp) Length() int {
	return int(o)
}

// String returns a string representation for debugging.
func (o RetainOp) String() string {
	return fmt.Sprintf("retain %d", int(o))
}

// InsertOp inserts new text at the current position.
//
// Represented as a string in the original ot.js implementation.
// Example: InsertOp("Hello") means "insert 'Hello' at the current position"
type InsertOp string

// Type returns OpInsert for InsertOp.
func (o InsertOp) Type() OperationType {
	return OpInsert
}

// Length returns the length of the string to be inserted.
func (o InsertOp) Length() int {
	return len(o)
}

// String returns a string representation for debugging.
func (o InsertOp) String() string {
	return fmt.Sprintf("insert '%s'", string(o))
}

// DeleteOp removes characters from the current position.
//
// Represented as a negative integer in the original ot.js implementation.
// Example: DeleteOp(-3) means "delete the next 3 characters"
type DeleteOp int

// Type returns OpDelete for DeleteOp.
func (o DeleteOp) Type() OperationType {
	return OpDelete
}

// Length returns the number of characters to delete (absolute value).
func (o DeleteOp) Length() int {
	return -int(o)
}

// String returns a string representation for debugging.
func (o DeleteOp) String() string {
	return fmt.Sprintf("delete %d", -int(o))
}

// Helper functions for working with Op interface

// IsRetain returns true if the op is a RetainOp.
func IsRetain(op Op) bool {
	return op.Type() == OpRetain
}

// IsInsert returns true if the op is an InsertOp.
func IsInsert(op Op) bool {
	return op.Type() == OpInsert
}

// IsDelete returns true if the op is a DeleteOp.
func IsDelete(op Op) bool {
	return op.Type() == OpDelete
}
