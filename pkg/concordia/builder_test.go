package concordia

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBuilder_OptimizeRetain tests that adjacent retains are merged.
func TestBuilder_OptimizeRetain(t *testing.T) {
	op := NewBuilder().
		Retain(5).
		Retain(3).
		Build()

	assert.Equal(t, 1, len(op.ops))
	assert.Equal(t, RetainOp(8), op.ops[0])
	assert.Equal(t, 8, op.BaseLength())
	assert.Equal(t, 8, op.TargetLength())
}

// TestBuilder_OptimizeInsert tests that adjacent inserts are merged.
func TestBuilder_OptimizeInsert(t *testing.T) {
	op := NewBuilder().
		Insert("Hello").
		Insert(" ").
		Insert("World").
		Build()

	assert.Equal(t, 1, len(op.ops))
	assert.Equal(t, InsertOp("Hello World"), op.ops[0])
	assert.Equal(t, 0, op.BaseLength())
	assert.Equal(t, 11, op.TargetLength())
}

// TestBuilder_OptimizeDelete tests that adjacent deletes are merged.
func TestBuilder_OptimizeDelete(t *testing.T) {
	op := NewBuilder().
		Delete(2).
		Delete(3).
		Build()

	assert.Equal(t, 1, len(op.ops))
	assert.Equal(t, DeleteOp(-5), op.ops[0])
	assert.Equal(t, 5, op.BaseLength())
	assert.Equal(t, 0, op.TargetLength())
}

// TestBuilder_Complex tests a complex operation sequence.
func TestBuilder_Complex(t *testing.T) {
	op := NewBuilder().
		Retain(5).
		Insert("Hello").
		Retain(3).
		Delete(2).
		Insert("World").
		Build()

	assert.Equal(t, 5, len(op.ops))
	assert.Equal(t, RetainOp(5), op.ops[0])
	assert.Equal(t, InsertOp("Hello"), op.ops[1])
	assert.Equal(t, RetainOp(3), op.ops[2])
	// NOTE: ot.js normalization rule swaps Insert after Delete
	// So Insert("World") comes before Delete(2)
	assert.Equal(t, InsertOp("World"), op.ops[3])
	assert.Equal(t, DeleteOp(-2), op.ops[4])
}

// TestBuilder_Apply tests building and applying an operation.
func TestBuilder_Apply(t *testing.T) {
	doc := "Hello World"
	op := NewBuilder().
		Retain(6).
		Insert("Go ").
		Delete(5).
		Build()

	result, err := op.Apply(doc)
	assert.NoError(t, err)
	assert.Equal(t, "Hello Go ", result)
}

// TestBuilder_Empty tests building an empty operation.
func TestBuilder_Empty(t *testing.T) {
	op := NewBuilder().Build()

	assert.Equal(t, 0, len(op.ops))
	assert.Equal(t, 0, op.BaseLength())
	assert.Equal(t, 0, op.TargetLength())
	assert.True(t, op.IsNoop())
}

// TestBuilder_OnlyRetain tests building an operation with only retains.
func TestBuilder_OnlyRetain(t *testing.T) {
	op := NewBuilder().
		Retain(5).
		Retain(3).
		Retain(2).
		Build()

	assert.Equal(t, 1, len(op.ops))
	assert.Equal(t, RetainOp(10), op.ops[0])
	assert.Equal(t, 10, op.BaseLength())
	assert.Equal(t, 10, op.TargetLength())
	assert.True(t, op.IsNoop())
}

// TestBuilder_OnlyInsert tests building an operation with only inserts.
func TestBuilder_OnlyInsert(t *testing.T) {
	op := NewBuilder().
		Insert("Hello").
		Insert(" ").
		Insert("World").
		Build()

	assert.Equal(t, 1, len(op.ops))
	assert.Equal(t, InsertOp("Hello World"), op.ops[0])
	assert.Equal(t, 0, op.BaseLength())
	assert.Equal(t, 11, op.TargetLength())
	assert.False(t, op.IsNoop())
}

// TestBuilder_OnlyDelete tests building an operation with only deletes.
func TestBuilder_OnlyDelete(t *testing.T) {
	op := NewBuilder().
		Delete(3).
		Delete(2).
		Delete(5).
		Build()

	assert.Equal(t, 1, len(op.ops))
	assert.Equal(t, DeleteOp(-10), op.ops[0])
	assert.Equal(t, 10, op.BaseLength())
	assert.Equal(t, 0, op.TargetLength())
	assert.False(t, op.IsNoop())
}

// TestBuilder_Mixed tests building a mixed operation.
func TestBuilder_Mixed(t *testing.T) {
	op := NewBuilder().
		Retain(3).
		Insert("abc").
		Delete(2).
		Retain(5).
		Insert("xyz").
		Build()

	assert.Equal(t, 5, len(op.ops))
	assert.Equal(t, RetainOp(3), op.ops[0])
	assert.Equal(t, InsertOp("abc"), op.ops[1])
	assert.Equal(t, DeleteOp(-2), op.ops[2])
	assert.Equal(t, RetainOp(5), op.ops[3])
	assert.Equal(t, InsertOp("xyz"), op.ops[4])

	assert.Equal(t, 10, op.BaseLength())
	assert.Equal(t, 14, op.TargetLength())  // Fixed: was 16, ot.js returns 14
}

// TestBuilder_NoopRemoval tests that no-ops are removed.
func TestBuilder_NoopRemoval(t *testing.T) {
	op := NewBuilder().
		Retain(0).
		Insert("").
		Delete(0).
		Retain(5).
		Insert("Hello").
		Delete(0).
		Build()

	assert.Equal(t, 2, len(op.ops))
	assert.Equal(t, RetainOp(5), op.ops[0])
	assert.Equal(t, InsertOp("Hello"), op.ops[1])
}
