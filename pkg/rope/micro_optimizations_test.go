package rope

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInsertFast_BasicInsertion tests basic InsertFast operations
func TestInsertFast_BasicInsertion(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		pos      int
		text     string
		expected string
	}{
		{
			name:     "Insert at beginning",
			initial:  "World",
			pos:      0,
			text:     "Hello ",
			expected: "Hello World",
		},
		{
			name:     "Insert at end",
			initial:  "Hello",
			pos:      5,
			text:     " World",
			expected: "Hello World",
		},
		{
			name:     "Insert in middle",
			initial:  "HeWorld",
			pos:      2,
			text:     "llo ",
			expected: "Hello World",
		},
		{
			name:     "Insert empty string",
			initial:  "Hello",
			pos:      2,
			text:     "",
			expected: "Hello",
		},
		{
			name:     "Insert into empty rope",
			initial:  "",
			pos:      0,
			text:     "Hello",
			expected: "Hello",
		},
		{
			name:     "Insert Unicode",
			initial:  "世界",
			pos:      1,
			text:     "你好",
			expected: "世你好界",
		},
		{
			name:     "Insert emoji",
			initial:  "Hello",
			pos:      5,
			text:     "🌍",
			expected: "Hello🌍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.initial)
			result := r.InsertFast(tt.pos, tt.text)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

// TestInsertFast_NilRope tests InsertFast with nil rope
func TestInsertFast_NilRope(t *testing.T) {
	var r *Rope
	result := r.InsertFast(0, "Hello")
	assert.Equal(t, "Hello", result.String())
}

// TestInsertFast_SingleLeafOptimization tests single leaf optimization
func TestInsertFast_SingleLeafOptimization(t *testing.T) {
	r := New("Hello World")
	// This should use single leaf optimization
	result := r.InsertFast(5, " Beautiful")
	assert.Equal(t, "Hello Beautiful World", result.String())
}

// TestDeleteFast_BasicDeletion tests basic DeleteFast operations
func TestDeleteFast_BasicDeletion(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		start    int
		end      int
		expected string
	}{
		{
			name:     "Delete from beginning",
			initial:  "Hello World",
			start:    0,
			end:      6,
			expected: "World",
		},
		{
			name:     "Delete from end",
			initial:  "Hello World",
			start:    5,
			end:      11,
			expected: "Hello",
		},
		{
			name:     "Delete from middle",
			initial:  "Hello Beautiful World",
			start:    5,
			end:      15,
			expected: "Hello World",
		},
		{
			name:     "Delete empty range",
			initial:  "Hello",
			start:    2,
			end:      2,
			expected: "Hello",
		},
		{
			name:     "Delete all",
			initial:  "Hello",
			start:    0,
			end:      5,
			expected: "",
		},
		{
			name:     "Delete Unicode",
			initial:  "你好世界",
			start:    1,
			end:      3,
			expected: "你界",
		},
		{
			name:     "Delete emoji",
			initial:  "Hello🌍World",
			start:    5,
			end:      6,
			expected: "HelloWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.initial)
			result := r.DeleteFast(tt.start, tt.end)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

// TestDeleteFast_NilRope tests DeleteFast with nil rope
func TestDeleteFast_NilRope(t *testing.T) {
	var r *Rope
	result := r.DeleteFast(0, 5)
	assert.Nil(t, result)
}

// TestDeleteFast_SingleLeafOptimization tests single leaf optimization for deletion
func TestDeleteFast_SingleLeafOptimization(t *testing.T) {
	r := New("Hello World")
	// This should use single leaf optimization
	result := r.DeleteFast(5, 6)
	assert.Equal(t, "HelloWorld", result.String())
}

// TestSliceFast_BasicSlicing tests basic SliceFast operations
func TestSliceFast_BasicSlicing(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		start    int
		end      int
		expected string
	}{
		{
			name:     "Full slice",
			text:     "Hello World",
			start:    0,
			end:      11,
			expected: "Hello World",
		},
		{
			name:     "Slice from beginning",
			text:     "Hello World",
			start:    0,
			end:      5,
			expected: "Hello",
		},
		{
			name:     "Slice from end",
			text:     "Hello World",
			start:    6,
			end:      11,
			expected: "World",
		},
		{
			name:     "Slice from middle",
			text:     "Hello World",
			start:    3,
			end:      8,
			expected: "lo Wo",
		},
		{
			name:     "Empty slice",
			text:     "Hello",
			start:    2,
			end:      2,
			expected: "",
		},
		{
			name:     "Slice Unicode",
			text:     "你好世界",
			start:    1,
			end:      3,
			expected: "好世",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.SliceFast(tt.start, tt.end)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSliceFast_SingleLeafOptimization tests single leaf optimization for slicing
func TestSliceFast_SingleLeafOptimization(t *testing.T) {
	r := New("Hello World")
	// This should use single leaf optimization
	result := r.SliceFast(0, 5)
	assert.Equal(t, "Hello", result)
}

// TestSliceToRope tests SliceToRope method
func TestSliceToRope(t *testing.T) {
	r := New("Hello World")

	result := r.SliceToRope(0, 5)
	assert.Equal(t, "Hello", result.String())
	assert.Equal(t, 5, result.Length())
}

// TestInsertIntoSingleLeaf tests insertIntoSingleLeaf helper
func TestInsertIntoSingleLeaf(t *testing.T) {
	r := New("Hello")

	// Insert at position 0 (prepend)
	result := r.InsertFast(0, "Hi ")
	assert.Equal(t, "Hi Hello", result.String())

	// Insert at end (append) - result now has length 8
	result = result.InsertFast(8, " World")
	assert.Equal(t, "Hi Hello World", result.String())

	// Insert in middle of original rope (position 2 in "Hello")
	result = r.InsertFast(2, "XX")
	assert.Equal(t, "HeXXllo", result.String())
}

// TestDeleteFromSingleLeaf tests deleteFromSingleLeaf helper
func TestDeleteFromSingleLeaf(t *testing.T) {
	r := New("Hello World")

	// Delete from beginning
	result := r.DeleteFast(0, 6)
	assert.Equal(t, "World", result.String())

	// Delete from end
	result = r.DeleteFast(5, 11)
	assert.Equal(t, "Hello", result.String())

	// Delete from middle
	result = r.DeleteFast(2, 4)
	assert.Equal(t, "Heo World", result.String())
}

// TestSliceSingleLeaf tests sliceSingleLeaf helper
func TestSliceSingleLeaf(t *testing.T) {
	r := New("Hello World")

	result := r.SliceFast(6, 11)
	assert.Equal(t, "World", result)
}

// TestFindBytePosInString tests findBytePosInString helper
func TestFindBytePosInString(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		charPos    int
		expectByte int
	}{
		{
			name:       "ASCII string",
			text:       "Hello",
			charPos:    2,
			expectByte: 2,
		},
		{
			name:       "Unicode string",
			text:       "你好",
			charPos:    1,
			expectByte: 3,
		},
		{
			name:       "Mixed ASCII and Unicode",
			text:       "Hi你好",
			charPos:    3,
			expectByte: 5,
		},
		{
			name:       "Position 0",
			text:       "Hello",
			charPos:    0,
			expectByte: 0,
		},
		{
			name:       "All ASCII",
			text:       "ABC",
			charPos:    2,
			expectByte: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findBytePosInString(tt.text, tt.charPos)
			assert.Equal(t, tt.expectByte, result)
		})
	}
}

// TestBatchInsert_SingleInsertion tests single insertion via BatchInsert
func TestBatchInsert_SingleInsertion(t *testing.T) {
	r := New("Hello World")

	inserts := []Insertion{
		{Pos: 5, Text: " Beautiful"},
	}

	result := r.BatchInsert(inserts)
	assert.Equal(t, "Hello Beautiful World", result.String())
}

// TestBatchInsert_MultipleInsertions tests multiple insertions
func TestBatchInsert_MultipleInsertions(t *testing.T) {
	r := New("The rope")

	inserts := []Insertion{
		{Pos: 8, Text: " is fast"},
		{Pos: 3, Text: " quick"},
	}

	result := r.BatchInsert(inserts)
	assert.Equal(t, "The quick rope is fast", result.String())
}

// TestBatchInsert_EmptySlice tests empty insertions slice
func TestBatchInsert_EmptySlice(t *testing.T) {
	r := New("Hello")

	inserts := []Insertion{}
	result := r.BatchInsert(inserts)

	assert.Equal(t, "Hello", result.String())
	assert.Equal(t, r, result) // Should return same rope
}

// TestBatchInsert_Ordering tests that inserts are applied in correct order
func TestBatchInsert_Ordering(t *testing.T) {
	r := New("ACE")

	inserts := []Insertion{
		{Pos: 2, Text: "D"},
		{Pos: 1, Text: "B"},
		{Pos: 0, Text: "A"},
	}

	result := r.BatchInsert(inserts)
	assert.Equal(t, "AABCDE", result.String())
}

// TestBatchInsert_WithUnicode tests batch insert with Unicode
func TestBatchInsert_WithUnicode(t *testing.T) {
	r := New("ABC")

	inserts := []Insertion{
		{Pos: 1, Text: "你好"},
		{Pos: 2, Text: "🌍"},
	}

	result := r.BatchInsert(inserts)
	assert.Equal(t, "A你好B🌍C", result.String())
}

// TestBatchDelete_SingleDeletion tests single deletion via BatchDelete
func TestBatchDelete_SingleDeletion(t *testing.T) {
	r := New("Hello Beautiful World")

	ranges := []Range{
		NewRange(5, 15),
	}

	result := r.BatchDelete(ranges)
	assert.Equal(t, "Hello World", result.String())
}

// TestBatchDelete_MultipleDeletions tests multiple deletions
func TestBatchDelete_MultipleDeletions(t *testing.T) {
	r := New("The quick brown fox jumps")

	ranges := []Range{
		NewRange(4, 9),   // "quick " (including trailing space)
		NewRange(10, 16), // "brown " (including trailing space)
		NewRange(16, 20), // "fox " (including trailing space)
	}

	result := r.BatchDelete(ranges)
	assert.Equal(t, "The  jumps", result.String())
}

// TestBatchDelete_EmptySlice tests empty ranges slice
func TestBatchDelete_EmptySlice(t *testing.T) {
	r := New("Hello")

	ranges := []Range{}
	result := r.BatchDelete(ranges)

	assert.Equal(t, "Hello", result.String())
}

// TestBatchDelete_Ordering tests that deletions are applied correctly
func TestBatchDelete_Ordering(t *testing.T) {
	r := New("ABCDEFG")

	ranges := []Range{
		NewRange(2, 4), // "CD"
		NewRange(0, 1), // "A"
		NewRange(5, 7), // "FG"
	}

	result := r.BatchDelete(ranges)
	assert.Equal(t, "BE", result.String())
}

// TestBatchDelete_WithUnicode tests batch delete with Unicode
func TestBatchDelete_WithUnicode(t *testing.T) {
	r := New("A你好B🌍C世界D")

	ranges := []Range{
		NewRange(1, 3), // "你好"
		NewRange(4, 5), // "🌍"
		NewRange(6, 8), // "世界"
	}

	result := r.BatchDelete(ranges)
	assert.Equal(t, "ABCD", result.String())
}

// TestIsASCII tests IsASCII helper function
func TestIsASCII(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{
			name:     "Pure ASCII",
			text:     "Hello World",
			expected: true,
		},
		{
			name:     "Unicode Chinese",
			text:     "你好",
			expected: false,
		},
		{
			name:     "Unicode emoji",
			text:     "🌍",
			expected: false,
		},
		{
			name:     "Empty string",
			text:     "",
			expected: true,
		},
		{
			name:     "Mixed ASCII and Unicode",
			text:     "Hello你好",
			expected: false,
		},
		{
			name:     "ASCII with special chars",
			text:     "Hello\nWorld\t!",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsASCII(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestASCIIStringLength tests ASCIIStringLength helper
func TestASCIIStringLength(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "ASCII string",
			text:     "Hello",
			expected: 5,
		},
		{
			name:     "Unicode string",
			text:     "你好",
			expected: -1,
		},
		{
			name:     "Empty string",
			text:     "",
			expected: 0,
		},
		{
			name:     "Mixed",
			text:     "Hi你好",
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ASCIIStringLength(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRuneCountInStringFast tests RuneCountInStringFast helper
func TestRuneCountInStringFast(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "ASCII string",
			text:     "Hello",
			expected: 5,
		},
		{
			name:     "Unicode Chinese",
			text:     "你好",
			expected: 2,
		},
		{
			name:     "Unicode emoji",
			text:     "🌍🌎",
			expected: 2,
		},
		{
			name:     "Empty string",
			text:     "",
			expected: 0,
		},
		{
			name:     "Mixed",
			text:     "Hi🌍",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RuneCountInStringFast(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestInsertFast_PreservesOriginal tests that InsertFast doesn't modify original
func TestInsertFast_PreservesOriginal(t *testing.T) {
	r := New("Hello")
	originalStr := r.String()

	result := r.InsertFast(2, "XX")

	assert.Equal(t, "HeXXllo", result.String())
	assert.Equal(t, originalStr, r.String())
}

// TestDeleteFast_PreservesOriginal tests that DeleteFast doesn't modify original
func TestDeleteFast_PreservesOriginal(t *testing.T) {
	r := New("Hello World")
	originalStr := r.String()

	result := r.DeleteFast(5, 6)

	assert.Equal(t, "HelloWorld", result.String())
	assert.Equal(t, originalStr, r.String())
}

// TestBatchInsert_LargeNumber tests batch insert with many insertions
func TestBatchInsert_LargeNumber(t *testing.T) {
	r := New("Start")

	inserts := make([]Insertion, 100)
	for i := 0; i < 100; i++ {
		inserts[i] = Insertion{
			Pos:  5, // All insert at the end
			Text: "X",
		}
	}

	result := r.BatchInsert(inserts)
	assert.Equal(t, 105, result.Length())
}

// TestBatchDelete_LargeNumber tests batch delete with many deletions
func TestBatchDelete_LargeNumber(t *testing.T) {
	r := New("AAAAABBBBBCCCCCDDDDDEEEEE")

	ranges := make([]Range, 5)
	ranges[0] = NewRange(0, 5)
	ranges[1] = NewRange(5, 10)
	ranges[2] = NewRange(10, 15)
	ranges[3] = NewRange(15, 20)
	ranges[4] = NewRange(20, 25)

	result := r.BatchDelete(ranges)
	assert.Equal(t, 0, result.Length())
}

// TestInsertFast_EdgeCases tests edge cases for InsertFast
func TestInsertFast_EdgeCases(t *testing.T) {
	t.Run("Insert at position beyond length", func(t *testing.T) {
		r := New("Hi")
		// This should panic or handle gracefully
		assert.Panics(t, func() {
			r.InsertFast(10, "Test")
		})
	})

	t.Run("Insert at negative position", func(t *testing.T) {
		r := New("Hi")
		assert.Panics(t, func() {
			r.InsertFast(-1, "Test")
		})
	})
}

// TestRange_FromTo tests Range methods
func TestRange_FromTo(t *testing.T) {
	r := NewRange(5, 10)

	assert.Equal(t, 5, r.From())
	assert.Equal(t, 10, r.To())
}
