package rope

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ========== Constructor Tests ==========

func TestNew_Empty(t *testing.T) {
	r := New("")
	assert.Equal(t, 0, r.Length())
	assert.Equal(t, 0, r.Size())
	assert.Equal(t, "", r.String())
	assert.Equal(t, []byte{}, r.Bytes())
}

func TestNew_FromString(t *testing.T) {
	r := New("Hello World")
	assert.Equal(t, 11, r.Length())
	assert.Equal(t, 11, r.Size())
	assert.Equal(t, "Hello World", r.String())
	assert.Equal(t, []byte("Hello World"), r.Bytes())
}

func TestNew_UTF8(t *testing.T) {
	r := New("Hello 世界")
	assert.Equal(t, 8, r.Length()) // 5 + 1 + 2 Chinese chars
	assert.Equal(t, 12, r.Size())   // 5 + 1 + 6 bytes for Chinese
	assert.Equal(t, "Hello 世界", r.String())
}

func TestEmpty(t *testing.T) {
	r := Empty()
	assert.Equal(t, 0, r.Length())
	assert.Equal(t, 0, r.Size())
	assert.Equal(t, "", r.String())
}

// ========== Basic Query Tests ==========

func TestLength(t *testing.T) {
	r := New("Hello")
	assert.Equal(t, 5, r.Length())
}

func TestSize(t *testing.T) {
	r := New("Hello")
	assert.Equal(t, 5, r.Size())
}

func TestSlice_Basic(t *testing.T) {
	r := New("Hello World")

	assert.Equal(t, "Hello", r.Slice(0, 5))
	assert.Equal(t, "World", r.Slice(6, 11))
	assert.Equal(t, "Hello World", r.Slice(0, 11))
}

func TestSlice_OutOfBounds(t *testing.T) {
	r := New("Hello")

	assert.Panics(t, func() { r.Slice(-1, 3) })
	assert.Panics(t, func() { r.Slice(0, 10) })
	assert.Panics(t, func() { r.Slice(3, 2) })
}

func TestCharAt(t *testing.T) {
	r := New("Hello")

	assert.Equal(t, 'H', r.CharAt(0))
	assert.Equal(t, 'e', r.CharAt(1))
	assert.Equal(t, 'o', r.CharAt(4))
}

func TestCharAt_OutOfBounds(t *testing.T) {
	r := New("Hello")
	assert.Panics(t, func() { r.CharAt(-1) })
	assert.Panics(t, func() { r.CharAt(10) })
}

func TestByteAt(t *testing.T) {
	r := New("Hello")

	assert.Equal(t, byte('H'), r.ByteAt(0))
	assert.Equal(t, byte('e'), r.ByteAt(1))
	assert.Equal(t, byte('o'), r.ByteAt(4))
}

// ========== Insert Tests ==========

func TestInsert_Start(t *testing.T) {
	r := New("World")
	r2 := r.Insert(0, "Hello ")

	assert.Equal(t, "Hello World", r2.String())
	assert.Equal(t, "World", r.String()) // Original unchanged
}

func TestInsert_Middle(t *testing.T) {
	r := New("Hello World")
	r2 := r.Insert(5, " Beautiful")

	assert.Equal(t, "Hello Beautiful World", r2.String())
}

func TestInsert_End(t *testing.T) {
	r := New("Hello")
	r2 := r.Insert(5, " World")

	assert.Equal(t, "Hello World", r2.String())
}

func TestInsert_EmptyString(t *testing.T) {
	r := New("Hello")
	r2 := r.Insert(2, "")

	assert.Same(t, r, r2)
}

func TestInsert_OutOfBounds(t *testing.T) {
	r := New("Hello")
	assert.Panics(t, func() { r.Insert(-1, "X") })
	assert.Panics(t, func() { r.Insert(10, "X") })
}

// ========== Delete Tests ==========

func TestDelete_Start(t *testing.T) {
	r := New("Hello World")
	r2 := r.Delete(0, 6)

	assert.Equal(t, "World", r2.String())
	assert.Equal(t, "Hello World", r.String()) // Original unchanged
}

func TestDelete_Middle(t *testing.T) {
	r := New("Hello World")
	r2 := r.Delete(5, 6)

	assert.Equal(t, "HelloWorld", r2.String())
}

func TestDelete_End(t *testing.T) {
	r := New("Hello World")
	r2 := r.Delete(5, 11)

	assert.Equal(t, "Hello", r2.String())
}

func TestDelete_All(t *testing.T) {
	r := New("Hello World")
	r2 := r.Delete(0, r.Length())

	assert.Equal(t, "", r2.String())
	assert.Equal(t, 0, r2.Length())
}

func TestDelete_EmptyRange(t *testing.T) {
	r := New("Hello")
	r2 := r.Delete(2, 2)

	assert.Same(t, r, r2)
}

func TestDelete_OutOfBounds(t *testing.T) {
	r := New("Hello")
	assert.Panics(t, func() { r.Delete(-1, 3) })
	assert.Panics(t, func() { r.Delete(0, 10) })
	assert.Panics(t, func() { r.Delete(3, 2) })
}

// ========== Replace Tests ==========

func TestReplace_Basic(t *testing.T) {
	r := New("Hello World")
	r2 := r.Replace(6, 11, "Go")

	assert.Equal(t, "Hello Go", r2.String())
}

func TestReplace_SameLength(t *testing.T) {
	r := New("Hello World")
	r2 := r.Replace(0, 5, "World")

	assert.Equal(t, "World World", r2.String())
}

// ========== Split Tests ==========

func TestSplit_Basic(t *testing.T) {
	r := New("Hello World")
	left, right := r.Split(5)

	assert.Equal(t, "Hello", left.String())
	assert.Equal(t, " World", right.String())
}

func TestSplit_Start(t *testing.T) {
	r := New("Hello World")
	left, right := r.Split(0)

	assert.Equal(t, "", left.String())
	assert.Equal(t, "Hello World", right.String())
}

func TestSplit_End(t *testing.T) {
	r := New("Hello World")
	left, right := r.Split(r.Length())

	assert.Equal(t, "Hello World", left.String())
	assert.Equal(t, "", right.String())
}

func TestSplit_OutOfBounds(t *testing.T) {
	r := New("Hello")
	assert.Panics(t, func() { r.Split(-1) })
	assert.Panics(t, func() { r.Split(10) })
}

// ========== Concat Tests ==========

func TestConcat_Basic(t *testing.T) {
	r1 := New("Hello")
	r2 := New(" World")
	r3 := r1.Concat(r2)

	assert.Equal(t, "Hello World", r3.String())
	assert.Equal(t, "Hello", r1.String())
	assert.Equal(t, " World", r2.String())
}

func TestConcat_Empty(t *testing.T) {
	r1 := New("Hello")
	r2 := Empty()
	r3 := r1.Concat(r2)

	assert.Same(t, r1, r3)
}

func TestConcat_Multiple(t *testing.T) {
	r1 := New("Hello")
	r2 := New(" ")
	r3 := New("World")

	result := r1.Concat(r2).Concat(r3)
	assert.Equal(t, "Hello World", result.String())
}

// ========== Clone Tests ==========

func TestClone(t *testing.T) {
	r := New("Hello World")
	r2 := r.Clone()

	assert.Equal(t, r.String(), r2.String())
	assert.Same(t, r, r2) // Same instance due to immutability
}

// ========== Utility Tests ==========

func TestContains(t *testing.T) {
	r := New("Hello World")

	assert.True(t, r.Contains("Hello"))
	assert.True(t, r.Contains("World"))
	assert.True(t, r.Contains("lo Wo"))
	assert.False(t, r.Contains("xyz"))
}

func TestIndex(t *testing.T) {
	r := New("Hello World")

	assert.Equal(t, 0, r.Index("H"))
	assert.Equal(t, 6, r.Index("W"))
	assert.Equal(t, -1, r.Index("z"))
}

func TestLastIndex(t *testing.T) {
	r := New("Hello Hello")

	assert.Equal(t, 6, r.LastIndex("H"))
	assert.Equal(t, -1, r.LastIndex("z"))
}

func TestCompare(t *testing.T) {
	r1 := New("Apple")
	r2 := New("Banana")
	r3 := New("Apple")

	assert.Equal(t, -1, r1.Compare(r2))
	assert.Equal(t, 1, r2.Compare(r1))
	assert.Equal(t, 0, r1.Compare(r3))
}

func TestEquals(t *testing.T) {
	r1 := New("Hello")
	r2 := New("Hello")
	r3 := New("World")

	assert.True(t, r1.Equals(r2))
	assert.False(t, r1.Equals(r3))
}

// ========== UTF-8 Tests ==========

func TestUTF8_Chinese(t *testing.T) {
	r := New("你好世界")

	assert.Equal(t, 4, r.Length())
	assert.Equal(t, 12, r.Size())
	assert.Equal(t, "你好世界", r.String())

	// Test slicing with UTF-8
	assert.Equal(t, "你好", r.Slice(0, 2))
	assert.Equal(t, "世界", r.Slice(2, 4))
}

func TestUTF8_Emoji(t *testing.T) {
	r := New("Hello 👋 World")

	assert.Equal(t, 13, r.Length()) // 5 + 1 + 1 + 1 + 5
	assert.Equal(t, "Hello 👋 World", r.String())

	// Test slicing with emoji
	assert.Equal(t, "Hello", r.Slice(0, 5))
	assert.Equal(t, "👋", r.Slice(6, 7))
}

func TestUTF8_Mixed(t *testing.T) {
	r := New("Hello世界World")

	assert.Equal(t, 12, r.Length()) // 5 + 2 + 5
	assert.Equal(t, "Hello世界World", r.String())
}

// ========== Large Text Tests ==========

func TestLargeText_Insert(t *testing.T) {
	// Create a large text (1MB)
	large := strings.Repeat("a", 1024*1024)
	r := New(large)

	// Insert in the middle
	r2 := r.Insert(512*1024, "INSERTED")

	assert.Equal(t, 1024*1024+8, r2.Length()) // "INSERTED" has 8 characters
	assert.Contains(t, r2.String(), "INSERTED")
}

func TestLargeText_Delete(t *testing.T) {
	// Create a large text (1MB)
	large := strings.Repeat("a", 1024*1024)
	r := New(large)

	// Delete a chunk from the middle
	r2 := r.Delete(512*1024, 512*1024+1024)

	assert.Equal(t, 1024*1024-1024, r2.Length())
}

func TestLargeText_Split(t *testing.T) {
	// Create a large text (1MB)
	large := strings.Repeat("a", 1024*1024)
	r := New(large)

	// Split in half
	left, right := r.Split(512 * 1024)

	assert.Equal(t, 512*1024, left.Length())
	assert.Equal(t, 512*1024, right.Length())
}

// ========== Immutability Tests ==========

func TestImmutability_Insert(t *testing.T) {
	r1 := New("Hello")
	r2 := r1.Insert(5, " World")

	assert.Equal(t, "Hello", r1.String())
	assert.Equal(t, "Hello World", r2.String())
}

func TestImmutability_Delete(t *testing.T) {
	r1 := New("Hello World")
	r2 := r1.Delete(5, 11)

	assert.Equal(t, "Hello World", r1.String())
	assert.Equal(t, "Hello", r2.String())
}

func TestImmutability_Replace(t *testing.T) {
	r1 := New("Hello World")
	r2 := r1.Replace(6, 11, "Go")

	assert.Equal(t, "Hello World", r1.String())
	assert.Equal(t, "Hello Go", r2.String())
}

// ========== Edge Cases ==========

func TestEdgeCase_EmptyRope(t *testing.T) {
	r := Empty()

	assert.Equal(t, 0, r.Length())
	assert.Equal(t, "", r.String())
	assert.Panics(t, func() { r.CharAt(0) })
	assert.Panics(t, func() { r.Slice(0, 1) })
}

func TestEdgeCase_SingleChar(t *testing.T) {
	r := New("a")

	assert.Equal(t, 1, r.Length())
	assert.Equal(t, "a", r.String())
	assert.Equal(t, 'a', r.CharAt(0))
}

func TestEdgeCase_ManyInserts(t *testing.T) {
	r := Empty()
	for i := 0; i < 1000; i++ {
		r = r.Insert(i, "a")
	}

	assert.Equal(t, 1000, r.Length())
	assert.Equal(t, strings.Repeat("a", 1000), r.String())
}

func TestEdgeCase_ManyDeletes(t *testing.T) {
	r := New(strings.Repeat("a", 1000))
	for i := 0; i < 1000; i++ {
		r = r.Delete(0, 1)
	}

	assert.Equal(t, 0, r.Length())
	assert.Equal(t, "", r.String())
}

// ========== Line Operations Tests ==========

func TestLineCount(t *testing.T) {
	tests := []struct {
		text     string
		expected int
	}{
		{"", 0},
		{"Hello", 1},
		{"Hello\n", 1},
		{"Hello\nWorld", 2},
		{"Hello\nWorld\n", 2},
		{"Line1\nLine2\nLine3", 3},
	}

	for _, tt := range tests {
		r := New(tt.text)
		assert.Equal(t, tt.expected, r.LineCount(), "Text: %q", tt.text)
	}
}

func TestLine(t *testing.T) {
	r := New("Line1\nLine2\nLine3")

	assert.Equal(t, "Line1", r.Line(0))
	assert.Equal(t, "Line2", r.Line(1))
	assert.Equal(t, "Line3", r.Line(2))
}

func TestLineStart(t *testing.T) {
	r := New("Line1\nLine2\nLine3")

	assert.Equal(t, 0, r.LineStart(0))
	assert.Equal(t, 6, r.LineStart(1)) // After "Line1\n"
	assert.Equal(t, 12, r.LineStart(2)) // After "Line2\n"
}

func TestLineEnd(t *testing.T) {
	r := New("Line1\nLine2\nLine3")

	assert.Equal(t, 5, r.LineEnd(0)) // "Line1"
	assert.Equal(t, 11, r.LineEnd(1)) // "Line2"
	assert.Equal(t, 17, r.LineEnd(2)) // "Line3"
}

func TestLineAtChar(t *testing.T) {
	r := New("Line1\nLine2\nLine3")

	assert.Equal(t, 0, r.LineAtChar(0))
	assert.Equal(t, 0, r.LineAtChar(4))
	assert.Equal(t, 1, r.LineAtChar(5)) // After \n
	assert.Equal(t, 1, r.LineAtChar(6))
	assert.Equal(t, 2, r.LineAtChar(11)) // After \n
}

// ========== Builder Tests ==========

func TestBuilder_Append(t *testing.T) {
	b := NewBuilder()
	b.Append("Hello")
	b.Append(" ")
	b.Append("World")

	r := b.Build()
	assert.Equal(t, "Hello World", r.String())
}

func TestBuilder_Insert(t *testing.T) {
	b := NewBuilder()
	b.Append("Hello World")
	b.Insert(5, " Beautiful")

	r := b.Build()
	assert.Equal(t, "Hello Beautiful World", r.String())
}

func TestBuilder_Delete(t *testing.T) {
	b := NewBuilder()
	b.Append("Hello Beautiful World")
	b.Delete(6, 16) // Delete " Beautiful" (keep the first space)

	r := b.Build()
	assert.Equal(t, "Hello World", r.String())
}

func TestBuilder_Reuse(t *testing.T) {
	b := NewBuilder()

	// First build
	b.Append("Hello")
	r1 := b.Build()
	assert.Equal(t, "Hello", r1.String())

	// Reuse for second build
	b.Append(" World")
	r2 := b.Build()
	assert.Equal(t, "Hello World", r2.String())

	// First rope should be unchanged
	assert.Equal(t, "Hello", r1.String())
}

func TestBuilder_WriteInterface(t *testing.T) {
	b := NewBuilder()

	b.Write([]byte("Hello"))
	b.WriteString(" World")

	r := b.Build()
	assert.Equal(t, "Hello World", r.String())
}

// ========== Iterator Tests ==========

func TestIterator_Basic(t *testing.T) {
	r := New("Hello")
	it := r.NewIterator()

	result := ""
	for it.Next() {
		result += string(it.Current())
	}

	assert.Equal(t, "Hello", result)
}

func TestIterator_Position(t *testing.T) {
	r := New("Hello")
	it := r.NewIterator()

	assert.Equal(t, 0, it.Position())

	it.Next()
	assert.Equal(t, 1, it.Position())

	it.Next()
	assert.Equal(t, 2, it.Position())
}

func TestIterator_Seek(t *testing.T) {
	r := New("Hello World")
	it := r.IteratorAt(6)

	// IteratorAt(6) positions us so Next() will return character at position 6
	assert.Equal(t, 6, it.Position())

	// Call Next() to get the character
	it.Next()
	assert.Equal(t, 'W', it.Current())
	assert.Equal(t, 7, it.Position())

	// Next() advances to next character
	it.Next()
	assert.Equal(t, 'o', it.Current())
	assert.Equal(t, 8, it.Position())

	// Seek(0) positions us at character 0
	it.Seek(0)
	assert.Equal(t, 0, it.Position())
	it.Next()
	assert.Equal(t, 'H', it.Current())
}

func TestIterator_Peek(t *testing.T) {
	r := New("Hello")
	it := r.NewIterator()

	ch, ok := it.Peek()
	assert.True(t, ok)
	assert.Equal(t, 'H', ch)
	assert.Equal(t, 0, it.Position()) // Position unchanged
}

func TestIterator_Skip(t *testing.T) {
	r := New("Hello World")
	it := r.NewIterator()

	skipped := it.Skip(6)
	assert.Equal(t, 6, skipped)
	assert.Equal(t, 6, it.Position())

	// Call Next() to get character at position 6
	it.Next()
	assert.Equal(t, 'W', it.Current())
	assert.Equal(t, 7, it.Position())

	// Next() advances to next character
	it.Next()
	assert.Equal(t, 'o', it.Current())
	assert.Equal(t, 8, it.Position())
}

func TestIterator_Collect(t *testing.T) {
	r := New("Hello World")
	it := r.IteratorAt(6)

	collected := it.Collect()
	// Should collect from position 6 onwards: "World"
	assert.Equal(t, "World", string(collected))
}

// ========== ForEach Tests ==========

func TestForEach(t *testing.T) {
	r := New("Hello")
	result := ""

	r.ForEach(func(ch rune) {
		result += string(ch)
	})

	assert.Equal(t, "Hello", result)
}

func TestForEachWithIndex(t *testing.T) {
	r := New("Hello")
	indices := []int{}

	r.ForEachWithIndex(func(i int, ch rune) {
		indices = append(indices, i)
	})

	assert.Equal(t, []int{0, 1, 2, 3, 4}, indices)
}

func TestMap(t *testing.T) {
	r := New("hello")
	r2 := r.Map(func(ch rune) rune {
		if ch >= 'a' && ch <= 'z' {
			return ch - 32
		}
		return ch
	})

	assert.Equal(t, "HELLO", r2.String())
}

func TestFilter(t *testing.T) {
	r := New("Hello World")
	r2 := r.Filter(func(ch rune) bool {
		return ch != ' '
	})

	assert.Equal(t, "HelloWorld", r2.String())
}

func TestCount(t *testing.T) {
	r := New("Hello World")
	count := r.Count(func(ch rune) bool {
		return ch == 'l'
	})

	assert.Equal(t, 3, count)
}

// ========== Balance Tests ==========

func TestBalance_Simple(t *testing.T) {
	r := New("Hello World")
	r2 := r.Balance()

	assert.Equal(t, r.String(), r2.String())
	assert.True(t, r2.IsBalanced())
}

func testIsBalanced_Empty(t *testing.T) {
	r := Empty()
	assert.True(t, r.IsBalanced())
}

func TestDepth(t *testing.T) {
	// Empty rope has depth 0
	r := Empty()
	assert.Equal(t, 0, r.Depth())

	// Single node rope has depth 0 (by tree height definition)
	r = New("Hello World")
	assert.Equal(t, 0, r.Depth())

	// Depth calculation works correctly
	assert.GreaterOrEqual(t, r.Depth(), 0)
}

func TestStats(t *testing.T) {
	r := New("Hello World")
	stats := r.Stats()

	assert.Greater(t, stats.NodeCount, 0)
	assert.Greater(t, stats.LeafCount, 0)
}

// ========== RopeDocument Tests ==========

func TestRopeDocument_Basic(t *testing.T) {
	doc := NewRopeDocument("Hello World")

	assert.Equal(t, 11, doc.Length())
	assert.Equal(t, "Hello World", doc.String())
	assert.Equal(t, []byte("Hello World"), doc.Bytes())
}

func TestRopeDocument_Slice(t *testing.T) {
	doc := NewRopeDocument("Hello World")

	assert.Equal(t, "Hello", doc.Slice(0, 5))
	assert.Equal(t, "World", doc.Slice(6, 11))
}

func TestRopeDocument_Clone(t *testing.T) {
	doc := NewRopeDocument("Hello World")
	doc2 := doc.Clone().(*RopeDocument)

	assert.Equal(t, doc.String(), doc2.String())
}

func TestRopeDocument_Insert(t *testing.T) {
	doc := NewRopeDocument("World")
	doc2 := doc.Insert(0, "Hello ")

	assert.Equal(t, "Hello World", doc2.String())
	assert.Equal(t, "World", doc.String())
}

func TestRopeDocument_Delete(t *testing.T) {
	doc := NewRopeDocument("Hello World")
	doc2 := doc.Delete(5, 11)

	assert.Equal(t, "Hello", doc2.String())
}

func TestRopeDocument_Concat(t *testing.T) {
	doc1 := NewRopeDocument("Hello")
	doc2 := NewRopeDocument(" World")
	doc3 := doc1.Concat(doc2)

	assert.Equal(t, "Hello World", doc3.String())
}

func TestRopeDocument_Equals(t *testing.T) {
	doc1 := NewRopeDocument("Hello")
	doc2 := NewRopeDocument("Hello")
	doc3 := NewRopeDocument("World")

	assert.True(t, doc1.Equals(doc2))
	assert.False(t, doc1.Equals(doc3))
}

// ========== DocumentBuilder Tests ==========

func TestDocumentBuilder_Basic(t *testing.T) {
	b := NewDocumentBuilder()
	b.Append("Hello")
	b.Append(" ")
	b.Append("World")

	doc := b.Build()
	assert.Equal(t, "Hello World", doc.String())
}

func TestDocumentBuilder_AppendLine(t *testing.T) {
	b := NewDocumentBuilder()
	b.AppendLine("Line1")
	b.AppendLine("Line2")

	doc := b.Build()
	assert.Equal(t, "Line1\nLine2\n", doc.String())
}

func TestDocumentBuilder_Reset(t *testing.T) {
	b := NewDocumentBuilder()
	b.Append("Hello")
	b.Reset()
	b.Append("World")

	doc := b.Build()
	assert.Equal(t, "World", doc.String())
}

// ========== Property Tests (Manual) ==========

func TestProperty_InsertThenDelete(t *testing.T) {
	tests := []string{
		"Hello",
		"Hello World",
		"你好世界",
		"Hello 👋 World",
		strings.Repeat("a", 100),
	}

	for _, original := range tests {
		r := New(original)
		r2 := r.Insert(2, "XX")
		r3 := r2.Delete(2, 4)

		assert.Equal(t, r.String(), r3.String(),
			"Insert then delete should return original")
	}
}

func TestProperty_SplitConcat(t *testing.T) {
	tests := []string{
		"Hello World",
		"你好世界",
		"Line1\nLine2\nLine3",
		strings.Repeat("a", 100),
	}

	for _, text := range tests {
		r := New(text)
		pos := r.Length() / 2
		left, right := r.Split(pos)
		merged := left.Concat(right)

		assert.Equal(t, r.String(), merged.String(),
			"Split then concat should return original")
	}
}

func TestProperty_MultipleInserts(t *testing.T) {
	r := Empty()
	expected := ""

	for i := 0; i < 100; i++ {
		char := string(rune('a' + (i % 26)))
		r = r.Insert(i, char)
		expected += char
	}

	assert.Equal(t, expected, r.String())
}

// ========== Benchmark Tests ==========

func BenchmarkRope_New(b *testing.B) {
	text := strings.Repeat("a", 1000)
	for i := 0; i < b.N; i++ {
		_ = New(text)
	}
}

func BenchmarkRope_Insert_Small(b *testing.B) {
	r := New("Hello World")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r = r.Insert(5, "X")
	}
}

func BenchmarkRope_Delete_Small(b *testing.B) {
	r := New(strings.Repeat("a", 1000))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r = r.Delete(0, 1)
	}
}

func BenchmarkRope_Slice(b *testing.B) {
	r := New(strings.Repeat("a", 10000))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = r.Slice(100, 9000)
	}
}

func BenchmarkRope_Concat(b *testing.B) {
	r1 := New(strings.Repeat("a", 1000))
	r2 := New(strings.Repeat("b", 1000))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = r1.Concat(r2)
	}
}

func BenchmarkRope_Iterator(b *testing.B) {
	r := New(strings.Repeat("a", 10000))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		it := r.NewIterator()
		for it.Next() {
			_ = it.Current()
		}
	}
}

