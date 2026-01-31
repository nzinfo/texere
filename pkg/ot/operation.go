package ot

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrInvalidBaseLength is returned when an operation is applied to a document
	// with an incompatible length.
	ErrInvalidBaseLength = errors.New("operation base length does not match document length")

	// ErrCannotUndo is returned when trying to undo but the undo stack is empty.
	ErrCannotUndo = errors.New("cannot undo: undo stack is empty")

	// ErrCannotRedo is returned when trying to redo but the redo stack is empty.
	ErrCannotRedo = errors.New("cannot redo: redo stack is empty")
)

// Operation represents an immutable sequence of OT operations.
//
// An operation is a list of ops (retain, insert, delete) that transforms
// a document from one state to another. Operations are immutable and
// safe for concurrent use.
//
// The structure corresponds to ot.js's TextOperation class.
type Operation struct {
	ops          []Op
	baseLength   int
	targetLength int
}

// NewOperation creates a new empty operation.
//
// Returns:
//   - an empty operation ready to be built using Retain/Insert/Delete
//
// Example:
//
//	op := NewOperation()
//	op.Retain(5).Insert("Hello")
func NewOperation() *Operation {
	return &Operation{
		ops:          make([]Op, 0, 16),
		baseLength:   0,
		targetLength: 0,
	}
}

// Retain appends a retain operation to this operation.
// This is a convenience method for test helpers.
// For production code, use the Builder pattern instead.
func (op *Operation) Retain(n int) *Operation {
	if n == 0 {
		return op
	}
	op.ops = append(op.ops, RetainOp(n))
	op.baseLength += n
	op.targetLength += n
	return op
}

// Insert appends an insert operation to this operation.
// This is a convenience method for test helpers.
// For production code, use the Builder pattern instead.
func (op *Operation) Insert(str string) *Operation {
	if str == "" {
		return op
	}
	op.ops = append(op.ops, InsertOp(str))
	op.targetLength += len(str)
	return op
}

// Delete appends a delete operation to this operation.
// This is a convenience method for test helpers.
// For production code, use the Builder pattern instead.
func (op *Operation) Delete(n int) *Operation {
	if n == 0 {
		return op
	}
	op.ops = append(op.ops, DeleteOp(-n))
	op.baseLength += n
	return op
}

// BaseLength returns the length of the document this operation operates on.
//
// This is the length of the document before the operation is applied.
func (op *Operation) BaseLength() int {
	return op.baseLength
}

// TargetLength returns the length of the document after applying this operation.
func (op *Operation) TargetLength() int {
	return op.targetLength
}

// IsNoop returns true if this operation has no effect.
//
// An operation is a no-op if it's empty or only contains retain operations.
func (op *Operation) IsNoop() bool {
	if len(op.ops) == 0 {
		return true
	}
	if len(op.ops) == 1 && IsRetain(op.ops[0]) {
		return true
	}
	return false
}

// Equals checks if two operations are equal.
//
// Two operations are equal if they have the same baseLength, targetLength,
// and the same sequence of ops.
func (op *Operation) Equals(other *Operation) bool {
	if op.baseLength != other.baseLength {
		return false
	}
	if op.targetLength != other.targetLength {
		return false
	}
	if len(op.ops) != len(other.ops) {
		return false
	}
	for i := range op.ops {
		if op.ops[i] != other.ops[i] {
			return false
		}
	}
	return true
}

// String returns a string representation of the operation for debugging.
//
// Example output: "retain 5, insert 'Hello', delete 3, retain 2"
func (op *Operation) String() string {
	parts := make([]string, len(op.ops))
	for i, op := range op.ops {
		parts[i] = op.String()
	}
	return strings.Join(parts, ", ")
}

// Apply applies this operation to a string document.
//
// This is a convenience method for applying operations to string documents.
// For more control over the document type, use ApplyToDocument instead.
//
// Parameters:
//   - str: the document string to apply the operation to
//
// Returns:
//   - the transformed document string
//   - an error if the operation cannot be applied
//
// Example:
//
//	op := ot.NewBuilder().Retain(6).Insert("Go ").Delete(6).Build()
//	newDoc, err := op.Apply("Hello World")
//	// newDoc == "Hello Go "
func (op *Operation) Apply(str string) (string, error) {
	doc := NewStringDocument(str)
	result, err := op.ApplyToDocument(doc)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}

// ApplyToDocument applies this operation to any document type.
//
// This is the core apply method that works with any Document implementation.
// It validates that the operation's baseLength matches the document length,
// then applies each operation in sequence.
//
// IMPORTANT: Operations use UTF-16 code unit positions (to match JavaScript),
// but Go strings use UTF-8 encoding with rune indexing. This method handles
// the conversion between UTF-16 positions and rune positions.
//
// Parameters:
//   - doc: the document to apply the operation to
//
// Returns:
//   - the transformed document
//   - an error if the operation cannot be applied
//
// Example:
//
//	doc := ot.NewStringDocument("Hello World")
//	op := ot.NewBuilder().Retain(6).Insert("Go ").Build()
//	newDoc, err := op.ApplyToDocument(doc)
func (op *Operation) ApplyToDocument(doc Document) (Document, error) {
	// Validate operation
	if op.baseLength != doc.Length() {
		return nil, ErrInvalidBaseLength
	}

	// Convert string to rune slice for proper Unicode handling
	// Operations use UTF-16 code units, but we need rune positions in Go
	str := doc.String()
	runes := []rune(str)

	// Build mapping from UTF-16 position to rune position
	utf16ToRunePos := make([]int, 0, len(runes)*2) // Upper bound
	runePos := 0
	utf16Pos := 0

	for _, r := range runes {
		utf16ToRunePos = append(utf16ToRunePos, runePos)
		if r >= 0x10000 {
			// Surrogate pair: 2 UTF-16 code units for 1 rune
			utf16ToRunePos = append(utf16ToRunePos, runePos)
			utf16Pos += 2
		} else {
			// BMP character: 1 UTF-16 code unit
			utf16Pos += 1
		}
		runePos++
	}
	// Add sentinel value for end of string
	utf16ToRunePos = append(utf16ToRunePos, runePos)

	// Track position in UTF-16 code units
	currentUTF16Pos := 0

	var builder strings.Builder
	builder.Grow(op.targetLength) // Pre-allocate for efficiency

	for _, op := range op.ops {
		switch v := op.(type) {
		case RetainOp:
			// Retain: copy UTF-16 code units from the original document
			count := int(v)
			endUTF16Pos := currentUTF16Pos + count

			// Check bounds
			if endUTF16Pos > len(utf16ToRunePos)-1 {
				return nil, fmt.Errorf("operation can't retain more characters than are left in the string")
			}

			// Convert UTF-16 positions to rune positions
			startRunePos := utf16ToRunePos[currentUTF16Pos]
			endRunePos := utf16ToRunePos[endUTF16Pos]

			// Copy runes
			for i := startRunePos; i < endRunePos; i++ {
				builder.WriteRune(runes[i])
			}

			currentUTF16Pos = endUTF16Pos

		case InsertOp:
			// Insert: add new characters
			builder.WriteString(string(v))
			// currentUTF16Pos stays the same

		case DeleteOp:
			// Delete: skip UTF-16 code units
			count := -int(v)
			endUTF16Pos := currentUTF16Pos + count

			// Check bounds
			if endUTF16Pos > len(utf16ToRunePos)-1 {
				return nil, fmt.Errorf("operation can't delete more characters than are left in the string")
			}

			// Just advance position, don't copy anything
			currentUTF16Pos = endUTF16Pos
		}
	}

	// Verify we consumed the entire document (in UTF-16 code units)
	if currentUTF16Pos != utf16Pos {
		return nil, fmt.Errorf("the operation didn't operate on the whole string")
	}

	// Return the result as a StringDocument
	// The caller can convert to other document types if needed
	return &StringDocument{content: builder.String()}, nil
}

// Invert creates the inverse of this operation.
//
// The inverse operation, when applied to the result of this operation,
// returns the original concordia. This is used for implementing undo.
//
// Parameters:
//   - str: the document string before this operation was applied
//
// Returns:
//   - the inverse operation
//
// Example:
//
//	op := NewBuilder().Insert("Hello ").Build()
//	inverse := op.Invert("")
//	// inverse is an operation that deletes "Hello "
func (op *Operation) Invert(str string) *Operation {
	inverse := NewBuilder()
	strIndex := 0

	for _, op := range op.ops {
		switch v := op.(type) {
		case RetainOp:
			inverse.Retain(int(v))
			strIndex += int(v)

		case InsertOp:
			// Inverse of insert is delete
			inverse.Delete(len(v))
			// strIndex stays the same

		case DeleteOp:
			// Inverse of delete is insert
			// DeleteOp stores negative value, so negate it to get length
			deleteLen := -int(v)
			endIndex := strIndex + deleteLen
			if endIndex > len(str) {
				endIndex = len(str)
			}
			deletedStr := str[strIndex:endIndex]
			inverse.Insert(deletedStr)
			strIndex += deleteLen
		}
	}

	return inverse.Build()
}

// ToJSON converts this operation to a JSON-serializable format.
//
// The format is compatible with ot.js's toJSON method.
// Returns a slice where:
//   - positive integers represent retain operations
//   - strings represent insert operations
//   - negative integers represent delete operations
//
// Example:
//
//	op := NewBuilder().Retain(2).Insert("Hello").Delete(3).Build()
//	json := op.ToJSON()
//	// json == []interface{}{2, "Hello", -3}
func (op *Operation) ToJSON() []interface{} {
	result := make([]interface{}, len(op.ops))
	for i, op := range op.ops {
		switch v := op.(type) {
		case RetainOp:
			result[i] = int(v)
		case InsertOp:
			result[i] = string(v)
		case DeleteOp:
			result[i] = int(v)
		}
	}
	return result
}

// FromJSON creates an operation from a JSON-serializable format.
//
// This is the inverse of ToJSON and is compatible with ot.js's fromJSON method.
//
// Parameters:
//   - ops: a slice in the format produced by ToJSON
//
// Returns:
//   - an operation
//   - an error if the format is invalid
//
// Example:
//
//	ops := []interface{}{2, "Hello", -3}
//	op, err := FromJSON(ops)
func FromJSON(ops []interface{}) (*Operation, error) {
	builder := NewBuilder()

	for _, op := range ops {
		switch v := op.(type) {
		case int:
			if v > 0 {
				builder.Retain(v)
			} else if v < 0 {
				builder.Delete(-v)
			}
			// v == 0 is a no-op, skip it
		case string:
			builder.Insert(v)
		default:
			return nil, fmt.Errorf("unknown operation type: %T", op)
		}
	}

	return builder.Build(), nil
}

// ShouldBeComposedWith determines if this operation should be composed with another.
//
// This is used by UndoManager to decide whether to merge consecutive operations.
// Operations should be composed if they are consecutive insertions or deletions
// at the same position.
//
// Parameters:
//   - other: the operation to check
//
// Returns:
//   - true if the operations should be composed
func (op *Operation) ShouldBeComposedWith(other *Operation) bool {
	if op.IsNoop() || other.IsNoop() {
		return true
	}

	startA := getStartIndex(op)
	startB := getStartIndex(other)
	simpleA := getSimpleOp(op)
	simpleB := getSimpleOp(other)

	if simpleA == nil || simpleB == nil {
		return false
	}

	if IsInsert(simpleA) && IsInsert(simpleB) {
		return startA+simpleA.Length() == startB
	}

	if IsDelete(simpleA) && IsDelete(simpleB) {
		// Two ways to delete: backspace and delete key
		// DeleteOp values are negative, so simpleB is negative (e.g., -3)
		// In ot.js: startB - simpleB means startB - (-3) = startB + 3
		// In Go: we need to negate the value to get the length
		return (startB+simpleB.Length() == startA) || startA == startB
	}

	return false
}

// getStartIndex returns the starting position of the operation.
func getStartIndex(op *Operation) int {
	if len(op.ops) > 0 && IsRetain(op.ops[0]) {
		return int(op.ops[0].(RetainOp))
	}
	return 0
}

// getSimpleOp returns the "simple" operation (the main insert/delete).
func getSimpleOp(op *Operation) Op {
	switch len(op.ops) {
	case 1:
		return op.ops[0]
	case 2:
		if IsRetain(op.ops[0]) {
			return op.ops[1]
		}
		if IsRetain(op.ops[1]) {
			return op.ops[0]
		}
	case 3:
		if IsRetain(op.ops[0]) && IsRetain(op.ops[2]) {
			return op.ops[1]
		}
	}
	return nil
}
