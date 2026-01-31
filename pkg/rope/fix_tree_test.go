package rope

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// TestFixTree_DeleteAtChunkBoundary tests deletion at exact chunk boundaries
// This is ported from ropey's fix_tree.rs to verify tree seam handling
func TestFixTree_DeleteAtChunkBoundary(t *testing.T) {
	// Build a rope with known chunk structure
	text := strings.Repeat("Hello World! ", 1000)
	r := New(text)

	// Get initial length
	initialLen := r.Length()

	// Delete a range that might be at a chunk boundary
	// This should trigger the fix_tree_seam edge case
	start := initialLen / 3
	end := (initialLen / 3) * 2
	r = r.Delete(start, end)

	// Verify rope integrity
	assert.True(t, r.Length() < initialLen)
	assert.True(t, r.Length() >= 0)

	// Verify string is valid
	result := r.String()
	assert.True(t, len(result) > 0)

	// Verify UTF-8 validity
	assert.True(t, utf8.ValidString(result))
}

// TestFixTree_MultipleDeletesAtBoundaries tests multiple boundary deletions
func TestFixTree_MultipleDeletesAtBoundaries(t *testing.T) {
	// Build a rope with many small chunks
	builder := NewBuilder()
	for i := 0; i < 100; i++ {
		builder.Append("Chunk" + string(rune('0'+i%10)))
	}
	r := builder.Build()

	// Perform multiple deletions at different positions
	for i := 0; i < 10; i++ {
		if r.Length() == 0 {
			break
		}

		start := r.Length() / 4
		end := (r.Length() / 4) * 3
		if end > r.Length() {
			end = r.Length()
		}

		r = r.Delete(start, end)

		// Verify integrity after each deletion
		result := r.String()
		assert.True(t, utf8.ValidString(result))
	}
}

// TestFixTree_DeleteMiddleOfRope tests deletion in middle of rope
func TestFixTree_DeleteMiddleOfRope(t *testing.T) {
	// Build a rope
	builder := NewBuilder()
	for i := 0; i < 50; i++ {
		builder.Append("Line " + string(rune('0'+i%10)) + "\n")
	}
	r := builder.Build()

	originalLen := r.Length()

	// Delete from the middle
	start := originalLen / 2
	end := start + originalLen/4
	if end > originalLen {
		end = originalLen
	}

	r = r.Delete(start, end)

	// Verify integrity
	assert.True(t, r.Length() < originalLen)
	assert.True(t, utf8.ValidString(r.String()))
}

// TestFixTree_DeleteFromBeginning tests deletion from beginning
func TestFixTree_DeleteFromBeginning(t *testing.T) {
	builder := NewBuilder()
	for i := 0; i < 100; i++ {
		builder.Append("Test ")
	}
	r := builder.Build()

	originalLen := r.Length()

	// Delete from beginning
	r = r.Delete(0, originalLen/10)

	// Verify integrity
	assert.True(t, r.Length() < originalLen)
	assert.True(t, utf8.ValidString(r.String()))
}

// TestFixTree_DeleteToEnd tests deletion to end
func TestFixTree_DeleteToEnd(t *testing.T) {
	builder := NewBuilder()
	for i := 0; i < 100; i++ {
		builder.Append("Test ")
	}
	r := builder.Build()

	originalLen := r.Length()

	// Delete to end
	start := originalLen - originalLen/10
	r = r.Delete(start, originalLen)

	// Verify integrity
	assert.True(t, r.Length() < originalLen)
	assert.True(t, utf8.ValidString(r.String()))
}

// TestFixTree_SplitAtChunkBoundary tests splitting at chunk boundaries
func TestFixTree_SplitAtChunkBoundary(t *testing.T) {
	// Build a rope with known chunk structure
	builder := NewBuilder()
	for i := 0; i < 100; i++ {
		builder.Append("Chunk")
	}
	r := builder.Build()

	// Split at various positions
	positions := []int{0, r.Length() / 4, r.Length() / 2, r.Length() - 1}

	for _, pos := range positions {
		if pos >= r.Length() {
			continue
		}

		left, right := r.Split(pos)

		// Verify both parts are valid
		assert.True(t, utf8.ValidString(left.String()))
		assert.True(t, utf8.ValidString(right.String()))

		// Verify combined equals original
		combined := left.String() + right.String()
		assert.Equal(t, r.String(), combined)

		// Re-merge for next iteration
		r = left.AppendRope(right)
	}
}

// TestFixTree_InsertAfterDelete tests insert after delete
func TestFixTree_InsertAfterDelete(t *testing.T) {
	builder := NewBuilder()
	for i := 0; i < 50; i++ {
		builder.Append("Line " + string(rune('0'+i%10)) + "\n")
	}
	r := builder.Build()

	// Delete a range
	start := r.Length() / 3
	end := (r.Length() / 3) * 2
	if end > r.Length() {
		end = r.Length()
	}
	r = r.Delete(start, end)

	// Insert at the same position
	r = r.Insert(start, "New Content\n")

	// Verify integrity
	assert.True(t, utf8.ValidString(r.String()))
	assert.Contains(t, r.String(), "New Content")
}

// TestFixTree_ComplexMutations tests complex mutation sequences
func TestFixTree_ComplexMutations(t *testing.T) {
	builder := NewBuilder()
	for i := 0; i < 100; i++ {
		builder.Append("Test Line ")
		builder.Append(string(rune('0'+i%10)))
		builder.Append("\n")
	}
	r := builder.Build()

	// Perform complex sequence of operations
	operations := []struct {
		op string
		arg1, arg2 int
		text string
	}{
		{"delete", 10, 20, ""},
		{"insert", 15, 0, "NEW "},
		{"delete", 30, 40, ""},
		{"insert", 25, 0, "MORE "},
		{"delete", 5, 10, ""},
	}

	for _, op := range operations {
		switch op.op {
		case "delete":
			if op.arg2 <= r.Length() {
				r = r.Delete(op.arg1, op.arg2)
			}
		case "insert":
			if op.arg1 <= r.Length() {
				r = r.Insert(op.arg1, op.text)
			}
		}

		// Verify integrity after each operation
		assert.True(t, utf8.ValidString(r.String()))
	}
}

// TestFixTree_DeleteEntireRope tests deleting entire rope
func TestFixTree_DeleteEntireRope(t *testing.T) {
	r := New("Hello World Test")

	// Delete entire rope
	r = r.Delete(0, r.Length())

	// Should be empty
	assert.Equal(t, 0, r.Length())
	assert.Equal(t, "", r.String())
}

// TestFixTree_DeleteLargeRange tests deleting very large ranges
func TestFixTree_DeleteLargeRange(t *testing.T) {
	// Build a large rope
	builder := NewBuilder()
	for i := 0; i < 1000; i++ {
		builder.Append("Chunk")
	}
	r := builder.Build()

	originalLen := r.Length()

	// Delete most of it
	r = r.Delete(100, originalLen - 100)

	// Verify integrity
	assert.True(t, r.Length() < originalLen)
	assert.True(t, utf8.ValidString(r.String()))
	assert.True(t, r.Length() >= 0)
}
