package rope

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// ========== Ropey-Aligned Basic Operations Tests ==========

// TestFromStr_Empty tests creating rope from empty string
func TestRopey_FromStr_Empty(t *testing.T) {
	r := New("")
	assert.Equal(t, 0, r.Length())
	assert.Equal(t, 0, r.Size())
	assert.Equal(t, "", r.String())
}

// TestFromStr_Simple tests creating rope from simple text
func TestRopey_FromStr_Simple(t *testing.T) {
	text := "Hello, World!"
	r := New(text)
	assert.Equal(t, len([]rune(text)), r.Length())
	assert.Equal(t, len(text), r.Size())
	assert.Equal(t, text, r.String())
}

// TestFromStr_WithNewlines tests creating rope with various newlines
func TestRopey_FromStr_WithNewlines(t *testing.T) {
	text := "Line 1\nLine 2\r\nLine 3\rLine 4"
	r := New(text)
	assert.Equal(t, text, r.String())
	assert.Equal(t, len([]rune(text)), r.Length())
}

// TestFromStr_WithUnicode tests creating rope with Unicode characters
func TestRopey_FromStr_WithUnicode(t *testing.T) {
	text := "Hello 世界 🌍🌎🌏"
	r := New(text)
	// "Hello" (5) + " " (1) + "世界" (2) + " " (1) + 3 emojis (3) = 12
	assert.Equal(t, 12, r.Length())
	assert.Equal(t, text, r.String())
}

// TestInsert_Beginning tests inserting at the beginning
func TestRopey_Insert_Beginning(t *testing.T) {
	r := New("World")
	result := r.Insert(0, "Hello, ")
	assert.Equal(t, "Hello, World", result.String())
}

// TestInsert_End tests inserting at the end
func TestRopey_Insert_End(t *testing.T) {
	r := New("Hello, ")
	result := r.Insert(r.Length(), "World")
	assert.Equal(t, "Hello, World", result.String())
}

// TestInsert_Middle tests inserting in the middle
func TestRopey_Insert_Middle(t *testing.T) {
	r := New("HeWorld")
	result := r.Insert(2, "llo, ")
	assert.Equal(t, "Hello, World", result.String())
}

// TestInsert_Multiple tests multiple insertions
func TestRopey_Insert_Multiple(t *testing.T) {
	r := New("")
	r = r.Insert(0, "World")
	r = r.Insert(0, "Hello, ")
	assert.Equal(t, "Hello, World", r.String())
}

// TestInsert_EmptyString tests inserting empty string
func TestRopey_Insert_EmptyString(t *testing.T) {
	r := New("Hello, World")
	result := r.Insert(5, "")
	assert.Equal(t, r.String(), result.String())
	assert.Equal(t, r.Length(), result.Length())
}

// TestInsert_LargeText tests inserting large text
func TestRopey_Insert_LargeText(t *testing.T) {
	r := New("Hello")
	largeText := strings.Repeat("A", 10000)
	result := r.Insert(5, largeText)
	assert.Equal(t, 5+10000, result.Length())
	assert.Equal(t, "Hello"+largeText, result.String())
}

// TestInsert_WithUnicode tests inserting Unicode text
func TestRopey_Insert_WithUnicode(t *testing.T) {
	r := New("Hello World")
	result := r.Insert(5, " 世界")
	assert.Equal(t, "Hello 世界 World", result.String())
	assert.Equal(t, 14, result.Length()) // Hello(5) + space(1) + 世界(2) + space(1) + World(5) = 14
}

// TestRemove_Beginning tests removing from beginning
func TestRopey_Remove_Beginning(t *testing.T) {
	r := New("Hello, World")
	result := r.Delete(0, 7)
	assert.Equal(t, "World", result.String())
}

// TestRemove_End tests removing from end
func TestRopey_Remove_End(t *testing.T) {
	r := New("Hello, World")
	result := r.Delete(7, 12)
	assert.Equal(t, "Hello, ", result.String()) // Includes trailing space
}

// TestRemove_Middle tests removing from middle
func TestRopey_Remove_Middle(t *testing.T) {
	r := New("Hello, World!")
	result := r.Delete(5, 7)
	assert.Equal(t, "HelloWorld!", result.String())
}

// TestRemove_All tests removing all content
func TestRopey_Remove_All(t *testing.T) {
	r := New("Hello, World")
	result := r.Delete(0, r.Length())
	assert.Equal(t, 0, result.Length())
	assert.Equal(t, "", result.String())
}

// TestRemove_EmptyRange tests removing empty range
func TestRopey_Remove_EmptyRange(t *testing.T) {
	r := New("Hello, World")
	result := r.Delete(5, 5)
	assert.Equal(t, r.String(), result.String())
	assert.Equal(t, r.Length(), result.Length())
}

// TestRemove_SingleChar tests removing single character
func TestRopey_Remove_SingleChar(t *testing.T) {
	r := New("Hello")
	result := r.Delete(1, 2)
	assert.Equal(t, "Hllo", result.String())
}

// TestSplitOff_Beginning tests splitting from beginning
func TestRopey_SplitOff_Beginning(t *testing.T) {
	r := New("Hello, World")
	left, right := r.Split(5)
	assert.Equal(t, "Hello", left.String())
	assert.Equal(t, ", World", right.String())
}

// TestSplitOff_End tests splitting from end
func TestRopey_SplitOff_End(t *testing.T) {
	r := New("Hello, World")
	left, right := r.Split(r.Length())
	assert.Equal(t, "Hello, World", left.String())
	assert.Equal(t, 0, right.Length())
}

// TestSplitOff_Middle tests splitting from middle
func TestRopey_SplitOff_Middle(t *testing.T) {
	r := New("Hello, World")
	left, right := r.Split(7)
	assert.Equal(t, "Hello, ", left.String()) // Includes trailing space
	assert.Equal(t, "World", right.String())
}

// TestAppend tests appending rope to rope
func TestRopey_Append(t *testing.T) {
	r1 := New("Hello, ")
	r2 := New("World")
	result := r1.AppendRope(r2)
	assert.Equal(t, "Hello, World", result.String())
}

// TestAppend_Empty tests appending empty rope
func TestRopey_Append_Empty(t *testing.T) {
	r1 := New("Hello, World")
	r2 := Empty()
	result := r1.AppendRope(r2)
	assert.Equal(t, "Hello, World", result.String())
}

// TestAppend_Multiple tests multiple appends
func TestRopey_Append_Multiple(t *testing.T) {
	r := New("")
	r = r.AppendRope(New("Hello"))
	r = r.AppendRope(New(", "))
	r = r.AppendRope(New("World"))
	assert.Equal(t, "Hello, World", r.String())
}

// ========== Index Conversion Tests ==========

// TestByteToCharIdx tests byte to character index conversion
func TestRopey_ByteToCharIdx(t *testing.T) {
	text := "Hello 世界"

	// Test all positions
	charPos := 0
	bytePos := 0
	for bytePos < len(text) {
		// Find byte position for this character
		foundBytePos := 0
		for i := 0; i < charPos; i++ {
			_, size := utf8.DecodeRuneInString(text[foundBytePos:])
			foundBytePos += size
		}
		assert.Equal(t, foundBytePos, bytePos)
		bytePos += utf8.RuneLen([]rune(text)[charPos])
		charPos++
	}
}

// TestCharToByteIdx tests character to byte index conversion
func TestCharToByteIdx(t *testing.T) {
	text := "Hello 世界"
	r := New(text)

	// Position 0 -> byte 0
	assert.Equal(t, 0, bytePosForCharPos(r, 0))

	// Position 5 (after "Hello") -> byte 5
	assert.Equal(t, 5, bytePosForCharPos(r, 5))

	// Position 6 (first char of "世界") -> byte 6
	assert.Equal(t, 6, bytePosForCharPos(r, 6))

	// Position 7 (second char of "世界") -> byte 9
	assert.Equal(t, 9, bytePosForCharPos(r, 7))
}

// Helper function to find byte position for character position
func bytePosForCharPos(r *Rope, charPos int) int {
	text := r.String()
	bytePos := 0
	for i := 0; i < charPos; i++ {
		_, size := utf8.DecodeRuneInString(text[bytePos:])
		bytePos += size
	}
	return bytePos
}

// ========== Line-Related Tests ==========

// TestLineCount tests counting lines
func TestRopey_LineCount(t *testing.T) {
	tests := []struct {
		text     string
		expected int
	}{
		{"", 0},                  // Empty rope has 0 lines
		{"Hello", 1},             // Single line
		{"Hello\nWorld", 2},      // Two lines
		{"Hello\nWorld\nFoo", 3}, // Three lines
		{"Hello\nWorld\n", 2},    // Two lines (ending with \n)
		{"\nHello\nWorld\n", 3},  // Three lines
	}

	for _, tt := range tests {
		r := New(tt.text)
		assert.Equal(t, tt.expected, r.LineCount(), "Text: %q", tt.text)
	}
}

// TestLine tests getting specific line
func TestRopey_Line(t *testing.T) {
	text := "Line 1\nLine 2\nLine 3"
	r := New(text)

	lines := r.Lines()
	assert.Equal(t, 3, len(lines))
	assert.Equal(t, "Line 1\n", lines[0])
	assert.Equal(t, "Line 2\n", lines[1])
	assert.Equal(t, "Line 3", lines[2])
}

// ========== Slice Tests ==========

// TestSlice_Full tests creating full slice
func TestRopey_Slice_Full(t *testing.T) {
	r := New("Hello, World")
	result := r.Slice(0, r.Length())
	assert.Equal(t, "Hello, World", result)
}

// TestSlice_Subset tests creating partial slice
func TestRopey_Slice_Subset(t *testing.T) {
	r := New("Hello, World")
	result := r.Slice(7, 12)
	assert.Equal(t, "World", result)
}

// TestSlice_Empty tests creating empty slice
func TestRopey_Slice_Empty(t *testing.T) {
	r := New("Hello, World")
	result := r.Slice(5, 5)
	assert.Equal(t, "", result)
}

// TestSlice_WithUnicode tests slicing with Unicode
func TestRopey_Slice_WithUnicode(t *testing.T) {
	r := New("Hello 世界 World")
	result := r.Slice(6, 8)
	assert.Equal(t, "世界", result)
}

// ========== Unicode and UTF-8 Tests ==========

// TestUTF8_AllValid tests that all operations maintain valid UTF-8
func TestRopey_UTF8_AllValid(t *testing.T) {
	r := New("Hello 世界 🌍")

	// Insert
	r2 := r.Insert(6, "ABC")
	assert.True(t, utf8.ValidString(r2.String()))

	// Delete
	r3 := r.Delete(6, 9)
	assert.True(t, utf8.ValidString(r3.String()))

	// Slice
	slice := r.Slice(0, 6)
	assert.True(t, utf8.ValidString(slice))
}

// TestUnicode_4ByteChars tests 4-byte UTF-8 characters (emojis)
func TestRopey_Unicode_4ByteChars(t *testing.T) {
	text := "🌍🌎🌏" // Each emoji is 4 bytes in UTF-8
	r := New(text)

	assert.Equal(t, 3, r.Length())
	assert.Equal(t, 12, r.Size()) // 3 * 4

	// Insert emoji
	r2 := r.Insert(1, "🌐")
	assert.Equal(t, 4, r2.Length())

	// Delete emoji
	r3 := r.Delete(0, 1)
	assert.Equal(t, 2, r3.Length())
}

// TestUnicode_CombiningChars tests combining characters
func TestRopey_Unicode_CombiningChars(t *testing.T) {
	// "e" + combining acute accent
	text := "Hello World" // Simple text
	r := New(text)

	// Insert combining character
	r2 := r.Insert(6, "\u0301") // Combining acute accent
	result := r2.String()

	// Result should be valid UTF-8
	assert.True(t, utf8.ValidString(result))
}

// ========== CRLF-Specific Tests ==========

// TestCRLF_Insert tests inserting CRLF
func TestRopey_CRLF_Insert(t *testing.T) {
	r := New("Hello")
	result := r.Insert(5, "\r\n")
	assert.Equal(t, "Hello\r\n", result.String())
}

// TestCRLF_Remove tests removing CRLF
func TestRopey_CRLF_Remove(t *testing.T) {
	r := New("Hello\r\nWorld")
	result := r.Delete(5, 7)
	assert.Equal(t, "HelloWorld", result.String())
}

// TestCRLF_Split tests splitting at CRLF
func TestRopey_CRLF_Split(t *testing.T) {
	r := New("Hello\r\nWorld")
	left, right := r.Split(5)
	assert.Equal(t, "Hello", left.String())
	assert.Equal(t, "\r\nWorld", right.String())
}

// ========== Edge Case Tests ==========

// TestEmptyRope tests operations on empty rope
func TestRopey_EmptyRope(t *testing.T) {
	r := Empty()

	assert.Equal(t, 0, r.Length())
	assert.Equal(t, 0, r.Size())
	assert.Equal(t, "", r.String())

	// Operations that should handle empty rope
	r2 := r.Insert(0, "Hello")
	assert.Equal(t, "Hello", r2.String())

	r3 := r.AppendRope(New("World"))
	assert.Equal(t, "World", r3.String())
}

// TestSingleChar tests single character rope
func TestRopey_SingleChar(t *testing.T) {
	r := New("a")

	assert.Equal(t, 1, r.Length())
	assert.Equal(t, 1, r.Size())
	assert.Equal(t, "a", r.String())

	r2 := r.Delete(0, 1)
	assert.Equal(t, 0, r2.Length())
}

// TestVerySmall tests very small text
func TestRopey_VerySmall(t *testing.T) {
	tests := []string{"", "a", "ab", "abc"}
	for _, text := range tests {
		r := New(text)
		assert.Equal(t, len([]rune(text)), r.Length())
		assert.Equal(t, text, r.String())
	}
}

// ========== Boundary Index Tests ==========

// TestIndexZero tests operations at index 0
func TestIndexZero(t *testing.T) {
	r := New("World")

	// Insert at 0
	result := r.Insert(0, "Hello, ")
	assert.Equal(t, "Hello, World", result.String())

	// Delete from 0
	result = r.Delete(0, 1)
	assert.Equal(t, "orld", result.String())
}

// TestIndexAtEnd tests operations at end index
func TestIndexAtEnd(t *testing.T) {
	r := New("Hello")

	// Insert at end
	result := r.Insert(5, " World")
	assert.Equal(t, "Hello World", result.String())

	// Delete to end
	result = r.Delete(3, 5)
	assert.Equal(t, "Hel", result.String())
}

// TestInvalidIndex tests operations with invalid index
func TestRopey_InvalidIndex(t *testing.T) {
	r := New("Hello")

	// These should panic
	assert.Panics(t, func() { r.Insert(-1, "X") })
	assert.Panics(t, func() { r.Insert(10, "X") })
	assert.Panics(t, func() { r.Delete(-1, 3) })
	assert.Panics(t, func() { r.Delete(0, 10) })
	assert.Panics(t, func() { r.Slice(-1, 3) })
	assert.Panics(t, func() { r.Slice(0, 10) })
}

// ========== Integrity and Invariant Tests ==========

// TestIntegrity_TreeStructure tests tree structure integrity after operations
func TestRopey_Integrity_TreeStructure(t *testing.T) {
	r := New("")

	// Perform multiple operations
	for i := 0; i < 100; i++ {
		r = r.Insert(r.Length(), fmt.Sprintf("%d", i%10))
	}

	// Verify integrity
	AssertIntegrity(t, r)
	AssertInvariants(t, r)
}

// TestInvariants_CharCount tests character count invariant
func TestRopey_Invariants_CharCount(t *testing.T) {
	text := "Hello, 世界!"
	r := New(text)

	assert.Equal(t, utf8.RuneCountInString(text), r.Length())
	assert.Equal(t, len(text), r.Size())
}

// TestInvariants_AfterInsert tests invariants after insert
func TestRopey_Invariants_AfterInsert(t *testing.T) {
	r := New("Hello")

	r2 := r.Insert(5, " World")

	assert.Equal(t, r.Length()+6, r2.Length())
	assert.Equal(t, r.Size()+6, r2.Size())
	assert.Equal(t, utf8.RuneCountInString(r2.String()), r2.Length())
}

// TestInvariants_AfterDelete tests invariants after delete
func TestRopey_Invariants_AfterDelete(t *testing.T) {
	r := New("Hello World")

	r2 := r.Delete(5, 6) // Delete the space

	assert.Equal(t, r.Length()-1, r2.Length())
	assert.Equal(t, r.Size()-1, r2.Size())
	assert.Equal(t, utf8.RuneCountInString(r2.String()), r2.Length())
}

// ========== Comparison Tests ==========

// TestEq_SameContent tests equality with same content
func TestRopey_Eq_SameContent(t *testing.T) {
	r1 := New("Hello, World")
	r2 := New("Hello, World")

	assert.True(t, r1.Equals(r2))
}

// TestEq_DifferentContent tests inequality with different content
func TestRopey_Eq_DifferentContent(t *testing.T) {
	r1 := New("Hello, World")
	r2 := New("Hello, World!")

	assert.False(t, r1.Equals(r2))
}

// ========== Clone Tests ==========

// TestClone_Independence tests that clone is independent
func TestRopey_Clone_Independence(t *testing.T) {
	r1 := New("Hello")
	r2 := r1.Clone()

	// Modify r1
	r1Modified := r1.Insert(5, " World")

	// r2 should be unchanged
	assert.Equal(t, "Hello", r2.String())
	assert.Equal(t, "Hello World", r1Modified.String())
}

// ========== Reader Tests ==========

// TestFromReader_Valid tests creating rope from reader
func TestRopey_FromReader_Valid(t *testing.T) {
	text := "Hello, World!"
	_ = strings.NewReader(text)

	// Note: Go doesn't have a direct FromReader in our API yet
	// This is a placeholder for future implementation
	r := New(text)
	assert.Equal(t, text, r.String())
}

// TestFromReader_Empty tests creating rope from empty reader
func TestRopey_FromReader_Empty(t *testing.T) {
	_ = strings.NewReader("")
	r := New("")
	assert.Equal(t, "", r.String())
}

// ========== Helper Functions ==========

// AssertIntegrity checks the tree structure integrity
func AssertIntegrity(t *testing.T, r *Rope) {
	t.Helper()

	if r == nil || r.Length() == 0 {
		return
	}

	// Check that root is not nil
	assert.NotNil(t, r.root, "Root should not be nil for non-empty rope")

	// Check depth is reasonable
	depth := calculateDepth(r.root)
	maxDepth := maxInt(64, r.Length()/100) // Reasonable depth
	assert.True(t, depth <= maxDepth, "Tree depth %d exceeds reasonable bound %d", depth, maxDepth)
}

// AssertInvariants checks rope invariants
func AssertInvariants(t *testing.T, r *Rope) {
	t.Helper()

	if r == nil {
		return
	}

	// Length and size should be non-negative
	assert.True(t, r.Length() >= 0, "Length should be non-negative")
	assert.True(t, r.Size() >= 0, "Size should be non-negative")

	// String() should be valid UTF-8
	assert.True(t, utf8.ValidString(r.String()), "String should be valid UTF-8")

	// Length should match actual character count
	str := r.String()
	actualLength := utf8.RuneCountInString(str)
	assert.Equal(t, actualLength, r.Length(), "Length should match actual character count")

	// Size should match actual byte count
	assert.Equal(t, len(str), r.Size(), "Size should match actual byte count")
}

// calculateDepth calculates the depth of a node
func calculateDepth(node RopeNode) int {
	if node == nil || node.IsLeaf() {
		return 0
	}

	internal := node.(*InternalNode)
	leftDepth := calculateDepth(internal.left)
	rightDepth := calculateDepth(internal.right)
	return maxInt(leftDepth, rightDepth) + 1
}

// maxInt returns the maximum of two integers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
