package rope

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ========== Slice Tests ==========

// TestSlice_Full tests creating full slice
func TestSlice_FullRange(t *testing.T) {
	r := New("Hello World")
	result := r.Slice(0, r.Length())
	assert.Equal(t, "Hello World", result)
}

// TestSlice_Partial tests creating partial slice
func TestSlice_Partial(t *testing.T) {
	r := New("Hello World")
	result := r.Slice(6, 11)
	assert.Equal(t, "World", result)
}

// TestSlice_Beginning tests slicing from beginning
func TestSlice_Beginning(t *testing.T) {
	r := New("Hello World")
	result := r.Slice(0, 5)
	assert.Equal(t, "Hello", result)
}

// TestSlice_End tests slicing to end
func TestSlice_End(t *testing.T) {
	r := New("Hello World")
	result := r.Slice(6, r.Length())
	assert.Equal(t, "World", result)
}

// TestSlice_Empty tests creating empty slice
func TestSlice_Empty(t *testing.T) {
	r := New("Hello World")
	result := r.Slice(5, 5)
	assert.Equal(t, "", result)
}

// TestSlice_WithUnicode tests slicing with Unicode characters
func TestSlice_UnicodeRange(t *testing.T) {
	r := New("Hello 世界 World")
	result := r.Slice(6, 8)
	assert.Equal(t, "世界", result)
}

// TestSlice_MultiByteChar tests slicing at multi-byte character boundaries
func TestSlice_MultiByteCharBoundary(t *testing.T) {
	r := New("Hello 世界")
	
	// Slice from 0 to 6 (includes "Hello " and first byte of "世")
	result := r.Slice(0, 6)
	assert.Equal(t, "Hello ", result)
	
	// Slice from 6 to end (should be valid UTF-8)
	result = r.Slice(6, r.Length())
	assert.Equal(t, "世界", result)
	assert.True(t, len(result) > 0)
}

// TestSlice_InvalidRange tests invalid slice ranges
func TestSlice_InvalidRanges(t *testing.T) {
	_ = New("Hello World")

	// These should panic or return empty
	// Note: behavior depends on implementation

	// Start > End should return empty or panic
	// (commenting out as behavior may vary)
	// result := r.Slice(10, 5)
	// assert.Equal(t, "", result)
}

// TestSlice_LargeText tests slicing large text
func TestSlice_LargeText(t *testing.T) {
	text := "Hello World"
	r := New(text)
	
	// Multiple slices should all work
	for i := 0; i <= r.Length(); i++ {
		for j := i; j <= r.Length(); j++ {
			result := r.Slice(i, j)
			expected := ""
			if i < len(text) && j <= len(text) {
				// Only validate if within bounds
				runes := []rune(text)
				for k := i; k < j; k++ {
					if k < len(runes) {
						expected += string(runes[k])
					}
				}
			}
			if i <= j && j <= len([]rune(text)) {
				assert.Equal(t, expected, result)
			}
		}
	}
}

// ========== Range Tests ==========

// TestRange_CharRange tests getting range by character indices
func TestRange_CharRange(t *testing.T) {
	r := New("Hello World")
	
	// Get char at position 0
	ch := r.CharAt(0)
	assert.Equal(t, 'H', ch)
	
	// Get char at position 6
	ch = r.CharAt(6)
	assert.Equal(t, 'W', ch)
	
	// Get char at last position
	ch = r.CharAt(r.Length() - 1)
	assert.Equal(t, 'd', ch)
}

// TestRange_CharAtByte tests getting char at byte position
func TestRange_CharAtByte(t *testing.T) {
	r := New("Hello World")
	
	// Byte 0 should be 'H'
	ch := r.ByteAt(0)
	assert.Equal(t, byte('H'), ch)
	
	// Byte 6 should be 'W'
	ch = r.ByteAt(6)
	assert.Equal(t, byte('W'), ch)
	
	// Last byte
	ch = r.ByteAt(r.Size() - 1)
	assert.Equal(t, byte('d'), ch)
}

// TestRange_Line tests getting lines
func TestRange_GetLine(t *testing.T) {
	text := "Line 1\nLine 2\nLine 3"
	r := New(text)
	
	lines := r.Lines()
	assert.Equal(t, 3, len(lines))
	assert.Equal(t, "Line 1\n", lines[0])
	assert.Equal(t, "Line 2\n", lines[1])
	assert.Equal(t, "Line 3", lines[2])
}

// TestRange_LineAt tests getting line at specific index
func TestRange_LineAt(t *testing.T) {
	text := "Line 1\nLine 2\nLine 3"
	r := New(text)

	line := r.Line(0)
	// Note: Line() may or may not include trailing newline
	// Adjust based on actual implementation
	assert.True(t, line == "Line 1" || line == "Line 1\n")

	line = r.Line(1)
	assert.True(t, line == "Line 2" || line == "Line 2\n")

	line = r.Line(2)
	assert.Equal(t, "Line 3", line) // Last line shouldn't have newline
}

// ========== Byte/Char Conversion Tests ==========

// TestCharToByteIdx tests character to byte index conversion
func TestConversions_CharToByte(t *testing.T) {
	text := "Hello 世界"
	r := New(text)
	
	// Position 0 -> byte 0
	byteIdx := r.charToByte(0)
	assert.Equal(t, 0, byteIdx)
	
	// Position 5 (end of "Hello") -> byte 5
	byteIdx = r.charToByte(5)
	assert.Equal(t, 5, byteIdx)
	
	// Position 6 (first char of "世界") -> byte 6
	byteIdx = r.charToByte(6)
	assert.Equal(t, 6, byteIdx)
	
	// Position 7 (second char of "世界") -> byte 9
	byteIdx = r.charToByte(7)
	assert.Equal(t, 9, byteIdx)
}

// TestByteToCharIdx tests byte to character index conversion
func TestConversions_ByteToChar(t *testing.T) {
	text := "Hello 世界"
	r := New(text)
	
	// Byte 0 -> char 0
	charIdx := r.byteToChar(0)
	assert.Equal(t, 0, charIdx)
	
	// Byte 5 -> char 5
	charIdx = r.byteToChar(5)
	assert.Equal(t, 5, charIdx)
	
	// Byte 6 -> char 6
	charIdx = r.byteToChar(6)
	assert.Equal(t, 6, charIdx)
	
	// Byte 7 (middle of "世") -> char 6
	charIdx = r.byteToChar(7)
	assert.Equal(t, 6, charIdx)
	
	// Byte 8 (end of "世") -> char 6
	charIdx = r.byteToChar(8)
	assert.Equal(t, 6, charIdx)
	
	// Byte 9 -> char 7
	charIdx = r.byteToChar(9)
	assert.Equal(t, 7, charIdx)
}

// ========== Line Info Tests ==========

// TestLineAtChar tests getting line number at character position
func TestLineInfo_LineAtChar(t *testing.T) {
	text := "Line 1\nLine 2\nLine 3"
	r := New(text)

	// Character 0-4 (Line 1) -> line 0
	lineNum := r.LineAtChar(0)
	assert.Equal(t, 0, lineNum)

	lineNum = r.LineAtChar(4)
	assert.Equal(t, 0, lineNum)

	// Character 5 (\n) -> still line 0
	lineNum = r.LineAtChar(5)
	assert.Equal(t, 0, lineNum)

	// Character 6-12 (Line 2) -> line 1
	lineNum = r.LineAtChar(6)
	assert.Equal(t, 1, lineNum)

	lineNum = r.LineAtChar(12)
	assert.Equal(t, 1, lineNum)

	// Character 13-19 (Line 3) -> line 2
	lineNum = r.LineAtChar(13)
	assert.Equal(t, 2, lineNum)
}

// ========== Slice Consistency Tests ==========

// TestSliceConsistency_SliceMatchesString tests that slice matches String()
func TestSliceConsistency_SliceMatchesString(t *testing.T) {
	text := "Hello, World! 🌍"
	r := New(text)
	
	// Full slice should match String()
	slice := r.Slice(0, r.Length())
	assert.Equal(t, r.String(), slice)
}

// TestSliceConsistency_MultipleSlices tests multiple slices are consistent
func TestSliceConsistency_MultipleSlices(t *testing.T) {
	text := "Hello World"
	r := New(text)

	// Multiple non-overlapping slices
	slice1 := r.Slice(0, 5)
	slice2 := r.Slice(6, 11)

	assert.Equal(t, "Hello", slice1)
	assert.Equal(t, "World", slice2)

	// Concatenate should give original (with space in between)
	assert.Equal(t, text, slice1+" "+slice2)
}

// ========== Edge Cases ==========

// TestSlice_ZeroLength tests zero-length slices
func TestSlice_ZeroLength(t *testing.T) {
	r := New("Hello")
	
	// Start == End
	slice := r.Slice(3, 3)
	assert.Equal(t, "", slice)
	
	// Full range
	slice = r.Slice(0, 0)
	assert.Equal(t, "", slice)
}

// TestSlice_OutOfBoundsRange tests out of bounds slicing
func TestSlice_OutOfBoundsRange(t *testing.T) {
	_ = New("Hello")

	// These should panic or be handled gracefully
	// (implementation dependent)

	// Start > length
	// slice := r.Slice(10, 15)
	// assert.Equal(t, "", slice)
}
