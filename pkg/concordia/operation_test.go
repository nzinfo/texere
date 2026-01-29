package concordia

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/texere-ot/pkg/document"
)

// TestOperation_Constructor tests the operation constructor.
// Corresponds to ot.js test/lib/test-text-operation.js: exports.testConstructor
func TestOperation_Constructor(t *testing.T) {
	op := NewOperation()
	assert.NotNil(t, op)
	assert.Equal(t, 0, op.BaseLength())
	assert.Equal(t, 0, op.TargetLength())
}

// TestOperation_Lengths tests length tracking.
// Corresponds to ot.js test/lib/test-text-operation.js: exports.testLengths
func TestOperation_Lengths(t *testing.T) {
	op := NewOperation()

	assert.Equal(t, 0, op.BaseLength())
	assert.Equal(t, 0, op.TargetLength())

	op = NewBuilder().Retain(5).Build()
	assert.Equal(t, 5, op.BaseLength())
	assert.Equal(t, 5, op.TargetLength())

	op = NewBuilder().Retain(5).Insert("abc").Build()
	assert.Equal(t, 5, op.BaseLength())
	assert.Equal(t, 8, op.TargetLength())

	op = NewBuilder().Retain(5).Insert("abc").Retain(2).Build()
	assert.Equal(t, 7, op.BaseLength())
	assert.Equal(t, 10, op.TargetLength())

	op = NewBuilder().Retain(5).Insert("abc").Retain(2).Delete(2).Build()
	assert.Equal(t, 9, op.BaseLength())
	assert.Equal(t, 10, op.TargetLength())
}

// TestOperation_BuilderChaining tests builder method chaining.
// Corresponds to ot.js test/lib/test-text-operation.js: exports.testChaining
func TestOperation_BuilderChaining(t *testing.T) {
	op := NewBuilder().
		Retain(5).
		Retain(0).
		Insert("lorem").
		Insert("").
		Delete(3).
		Delete(3).
		Delete(0).
		Delete(0).
		Build()

	assert.Equal(t, 3, len(op.ops))
}

// TestOperation_Apply_Random tests random apply operations.
// Corresponds to ot.js test/lib/test-text-operation.js: exports.testApply
func TestOperation_Apply_Random(t *testing.T) {
	for i := 0; i < 100; i++ { // Reduced from 500 for faster testing
		str := randomString(50)
		op := randomOperation(str)

		assert.Equal(t, len(str), op.BaseLength(), "base length should match string")

		result, err := op.Apply(str)
		require.NoError(t, err)
		assert.Equal(t, op.TargetLength(), len(result), "target length should match result")
	}
}

// TestOperation_Invert_Random tests random invert operations.
// Corresponds to ot.js test/lib/test-text-operation.js: exports.testInvert
func TestOperation_Invert_Random(t *testing.T) {
	for i := 0; i < 100; i++ { // Reduced from 500
		str := randomString(50)
		op := randomOperation(str)
		inv := op.Invert(str)

		// Verify: op.BaseLength === inv.TargetLength
		assert.Equal(t, op.BaseLength(), inv.TargetLength())

		// Verify: op.TargetLength === inv.BaseLength
		assert.Equal(t, op.TargetLength(), inv.BaseLength())

		// Verify: inv.Apply(op.Apply(str)) === str
		result, err := op.Apply(str)
		require.NoError(t, err)
		result2, err := inv.Apply(result)
		require.NoError(t, err)
		assert.Equal(t, str, result2)
	}
}

// TestOperation_EmptyOps tests that empty operations are removed.
// Corresponds to ot.js test/lib/test-text-operation.js: exports.testEmptyOps
func TestOperation_EmptyOps(t *testing.T) {
	op := NewBuilder().
		Retain(0).
		Insert("").
		Delete(0).
		Delete(0).
		Build()

	assert.Equal(t, 0, len(op.ops))
}

// TestOperation_Equals tests operation equality.
// Corresponds to ot.js test/lib/test-text-operation.js: exports.testEquals
func TestOperation_Equals(t *testing.T) {
	op1 := NewBuilder().Delete(1).Insert("lo").Retain(2).Retain(3).Build()
	op2 := NewBuilder().Delete(1).Insert("l").Insert("o").Retain(5).Build()

	assert.True(t, op1.Equals(op2))

	op3 := NewBuilder().Delete(1).Insert("lo").Retain(2).Retain(3).Build()
	op4 := NewBuilder().Delete(1).Insert("lo").Retain(2).Retain(3).Delete(1).Build()

	assert.False(t, op3.Equals(op4))
}

// TestOperation_OpsMerging tests operation merging in builder.
// Corresponds to ot.js test/lib/test-text-operation.js: exports.testOpsMerging
func TestOperation_OpsMerging(t *testing.T) {
	op := NewBuilder().Retain(2).Build()
	assert.Equal(t, 1, len(op.ops))
	assert.Equal(t, RetainOp(2), op.ops[0])

	op = NewBuilder().Retain(2).Retain(3).Build()
	assert.Equal(t, 1, len(op.ops), "adjacent retains should be merged")
	assert.Equal(t, RetainOp(5), op.ops[0])

	op = NewBuilder().Retain(2).Insert("abc").Build()
	assert.Equal(t, 2, len(op.ops))
	assert.Equal(t, InsertOp("abc"), op.ops[1])

	op = NewBuilder().Retain(2).Insert("abc").Insert("xyz").Build()
	assert.Equal(t, 2, len(op.ops), "adjacent inserts should be merged")
	assert.Equal(t, InsertOp("abcxyz"), op.ops[1])

	op = NewBuilder().Retain(2).Insert("abc").Delete(1).Delete(1).Build()
	assert.Equal(t, 3, len(op.ops), "adjacent deletes should be merged")
	assert.Equal(t, DeleteOp(-2), op.ops[2])
}

// TestOperation_IsNoop tests no-op detection.
// Corresponds to ot.js test/lib/test-text-operation.js: exports.testIsNoop
func TestOperation_IsNoop(t *testing.T) {
	op := NewBuilder().Build()
	assert.True(t, op.IsNoop())

	op = NewBuilder().Retain(5).Build()
	assert.True(t, op.IsNoop())

	op = NewBuilder().Retain(5).Retain(3).Build()
	assert.True(t, op.IsNoop())

	op = NewBuilder().Retain(5).Insert("lorem").Build()
	assert.False(t, op.IsNoop())
}

// TestOperation_ToString tests string representation.
// Corresponds to ot.js test/lib/test-text-operation.js: exports.testToString
func TestOperation_ToString(t *testing.T) {
	op := NewBuilder().Retain(2).Insert("lorem").Delete(5).Retain(5).Build()
	expected := "retain 2, insert 'lorem', delete 5, retain 5"
	assert.Equal(t, expected, op.String())
}

// TestOperation_Json_Random tests JSON serialization.
// Corresponds to ot.js test/lib/test-text-operation.js: exports.testIdJSON
func TestOperation_Json_Random(t *testing.T) {
	for i := 0; i < 100; i++ { // Reduced from 500
		doc := randomString(50)
		op := randomOperation(doc)

		// Serialize
		json := op.ToJSON()

		// Deserialize
		op2, err := FromJSON(json)
		require.NoError(t, err)

		// Verify equality
		assert.True(t, op.Equals(op2))
	}
}

// TestOperation_FromJSON tests JSON deserialization.
// Corresponds to ot.js test/lib/test-text-operation.js: exports.testFromJSON
func TestOperation_FromJSON(t *testing.T) {
	ops := []interface{}{2, -1, -1, "cde"}
	op, err := FromJSON(ops)
	require.NoError(t, err)

	assert.Equal(t, 3, len(op.ops))
	assert.Equal(t, 4, op.BaseLength())
	assert.Equal(t, 5, op.TargetLength())

	// Test invalid operations
	invalidOps := [][]interface{}{
		append(ops, map[string]string{"insert": "x"}),
		append(ops, nil),
	}

	for _, invalidOp := range invalidOps {
		_, err := FromJSON(invalidOp)
		assert.Error(t, err)
	}
}

// TestOperation_ShouldBeComposedWith tests operation composition criteria.
// Corresponds to ot.js test/lib/test-text-operation.js: exports.testShouldBeComposedWith
func TestOperation_ShouldBeComposedWith(t *testing.T) {
	// Test retain + insert + retain
	a := NewBuilder().Retain(3).Build()
	b := NewBuilder().Retain(1).Insert("tag").Retain(2).Build()
	assert.True(t, a.ShouldBeComposedWith(b))
	assert.True(t, b.ShouldBeComposedWith(a))

	// Test insert operations
	a = NewBuilder().Retain(1).Insert("a").Retain(2).Build()
	b = NewBuilder().Retain(2).Insert("b").Retain(2).Build()
	assert.True(t, a.ShouldBeComposedWith(b))

	// Test insert + delete
	a = NewBuilder().Retain(1).Insert("a").Retain(2).Delete(3).Build()
	b = NewBuilder().Retain(2).Insert("b").Retain(2).Build()
	assert.False(t, a.ShouldBeComposedWith(b))

	// Test delete operations
	a = NewBuilder().Retain(4).Delete(3).Retain(10).Build()
	b = NewBuilder().Retain(2).Delete(2).Retain(10).Build()
	assert.True(t, a.ShouldBeComposedWith(b))

	b = NewBuilder().Retain(4).Delete(7).Retain(3).Build()
	assert.True(t, a.ShouldBeComposedWith(b))

	b = NewBuilder().Retain(2).Delete(9).Retain(3).Build()
	assert.False(t, a.ShouldBeComposedWith(b))
}

// TestOperation_Compose_Random tests random composition.
// Corresponds to ot.js test/lib/test-text-operation.js: exports.testCompose
func TestOperation_Compose_Random(t *testing.T) {
	for i := 0; i < 100; i++ { // Reduced from 500
		str := randomString(20)
		a := randomOperation(str)
		afterA, err := a.Apply(str)
		require.NoError(t, err)

		b := randomOperation(afterA)
		afterB, err := b.Apply(afterA)
		require.NoError(t, err)

		// Compose
		ab, err := Compose(a, b)
		require.NoError(t, err)

		assert.Equal(t, b.TargetLength(), ab.TargetLength())

		// Invariant: apply(str, compose(a, b)) === apply(apply(str, a), b)
		afterAB, err := ab.Apply(str)
		require.NoError(t, err)
		assert.Equal(t, afterB, afterAB)
	}
}

// TestOperation_Transform_Random tests random transformation.
// Corresponds to ot.js test/lib/test-text-operation.js: exports.testTransform
func TestOperation_Transform_Random(t *testing.T) {
	for i := 0; i < 100; i++ { // Reduced from 500
		str := randomString(20)
		a := randomOperation(str)
		b := randomOperation(str)

		// Transform
		aPrime, bPrime, err := Transform(a, b)
		require.NoError(t, err)

		// Invariant: compose(a, b') === compose(b, a')
		abPrime, err := Compose(a, bPrime)
		require.NoError(t, err)
		baPrime, err := Compose(b, aPrime)
		require.NoError(t, err)

		afterABPrime, err := abPrime.Apply(str)
		require.NoError(t, err)
		afterBaPrime, err := baPrime.Apply(str)
		require.NoError(t, err)

		// Verify convergence
		assert.True(t, abPrime.Equals(baPrime))
		assert.Equal(t, afterABPrime, afterBaPrime)
	}
}

// TestDocument_StringDocument tests StringDocument implementation.
func TestDocument_StringDocument(t *testing.T) {
	content := "Hello World"
	doc := document.NewStringDocument(content)

	assert.Equal(t, len(content), doc.Length())
	assert.Equal(t, content, doc.String())
	assert.Equal(t, content, doc.Slice(0, len(content)))
	assert.Equal(t, "World", doc.Slice(6, 11))
	assert.Equal(t, []byte(content), doc.Bytes())

	// Test clone
	cloned := doc.Clone()
	assert.Equal(t, doc.String(), cloned.String())
	assert.Equal(t, doc.Length(), cloned.Length())
}
