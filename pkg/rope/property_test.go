package rope

import (
	"math/rand"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// Property-based tests inspired by ropey's proptest_tests.rs
// These tests verify properties that should hold true for ALL inputs
// rather than testing specific cases.

// Helper: performs random inserts on a rope and verifies integrity
func randomInserts(t *testing.T, numOps int) {
	r := Empty()
	rng := rand.New(rand.NewSource(rand.Int63()))

	strings := []string{
		"Hello ",
		"world! ",
		"How are ",
		"you ",
		"doing?\r\n",
		"Let's ",
		"keep ",
		"inserting ",
		"more ",
		"items.\r\n",
		"こんいちは、",
		"みんなさん！",
		"🌍🌎🌏",
		"Test",
	}

	for i := 0; i < numOps; i++ {
		ropeLen := r.Length()
		if ropeLen == 0 {
			ropeLen = 1
		}
		pos := rng.Intn(ropeLen)
		s := strings[rng.Intn(len(strings))]
		r = r.Insert(pos, s)
	}

	// Verify integrity
	assert.True(t, utf8.ValidString(r.String()))
	assert.True(t, r.Length() >= 0)
}

// TestProperty_RandomInserts_Small tests many random inserts
func TestProperty_RandomInserts_Small(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}
	randomInserts(t, 1<<10) // 1024 inserts
}

// TestProperty_RandomInserts_Large tests many random inserts
func TestProperty_RandomInserts_Large(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}
	randomInserts(t, 1<<13) // 8192 inserts
}

// Helper: performs random mutations (insert/delete) on a rope
func randomMutations(t *testing.T, numOps int) {
	r := Empty()
	rng := rand.New(rand.NewSource(rand.Int63()))

	// Start with some content
	r = r.Insert(0, "Hello World!")

	strings := []string{
		"foo",
		"bar",
		"baz",
		"qux",
		"🌍",
		"こんにちは",
		"Test",
		"\r\n",
	}

	for i := 0; i < numOps; i++ {
		op := rng.Intn(3)

		switch op {
		case 0: // Insert
			if r.Length() < 10000 { // Cap size
				ropeLen := r.Length()
				if ropeLen == 0 {
					ropeLen = 1
				}
				pos := rng.Intn(ropeLen)
				s := strings[rng.Intn(len(strings))]
				r = r.Insert(pos, s)
			}

		case 1: // Delete
			if r.Length() > 1 {
				start := rng.Intn(r.Length())
				end := start + rng.Intn(r.Length()-start+1)
				if end > r.Length() {
					end = r.Length()
				}
				r = r.Delete(start, end)
			}

		case 2: // Append
			if r.Length() < 10000 {
				s := strings[rng.Intn(len(strings))]
				r = r.Append(s)
			}
		}

		// Verify integrity every 100 operations
		if i%100 == 0 {
			assert.True(t, utf8.ValidString(r.String()))
			assert.True(t, r.Length() >= 0)
		}
	}

	// Final integrity check
	assert.True(t, utf8.ValidString(r.String()))
	assert.True(t, r.Length() >= 0)
}

// TestProperty_RandomMutations tests random mutations
func TestProperty_RandomMutations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}
	randomMutations(t, 1000)
}

// TestProperty_SplitMergeRoundtrip verifies split+merge returns original
func TestProperty_SplitMergeRoundtrip(t *testing.T) {
	texts := []string{
		"Hello World",
		"Hello 世界 🌍",
		"Line 1\r\nLine 2\r\nLine 3",
		"",
		"a",
		strings.Repeat("a", 1000),
	}

	for _, text := range texts {
		r := New(text)

		if r.Length() == 0 {
			continue
		}

		// Split at random position
		pos := rand.Intn(r.Length() + 1)
		left, right := r.Split(pos)

		// Merge back
		merged := left.AppendRope(right)

		// Should equal original
		assert.Equal(t, text, merged.String())
	}
}

// TestProperty_InsertDeleteRoundtrip verifies insert+delete roundtrip
func TestProperty_InsertDeleteRoundtrip(t *testing.T) {
	// Start with a string
	original := "Hello World"
	r := New(original)

	// Insert at position 6
	r = r.Insert(6, "XXX")
	assert.Contains(t, r.String(), "XXX")

	// Now delete the XXX we just inserted
	// Since we know it was inserted at position 6, and it's 3 characters long
	r = r.Delete(6, 9)

	// Should be back to original
	assert.Equal(t, original, r.String())
}

// TestProperty_IteratorConsistency verifies iterator consistency
func TestProperty_IteratorConsistency(t *testing.T) {
	texts := []string{
		"Hello World",
		"Hello 世界",
		"",
		"a",
		strings.Repeat("a", 100),
		"Line 1\nLine 2\nLine 3",
	}

	for _, text := range texts {
		r := New(text)

		// Count via iterator
		it := r.NewIterator()
		count := 0
		for it.Next() {
			count++
		}

		// Should equal rune count
		expectedCount := utf8.RuneCountInString(text)
		assert.Equal(t, expectedCount, count)
	}
}

// TestProperty_SliceConsistency verifies slice consistency
func TestProperty_SliceConsistency(t *testing.T) {
	r := New("Hello World Test")

	// All possible slices should be valid
	for i := 0; i <= r.Length(); i++ {
		for j := i; j <= r.Length(); j++ {
			slice := r.Slice(i, j)

			// Should be valid UTF-8
			assert.True(t, utf8.ValidString(slice))

			// Length should match
			expected := []rune("Hello World Test")
			if i <= len(expected) && j <= len(expected) {
				var result string
				for k := i; k < j; k++ {
					result += string(expected[k])
				}
				assert.Equal(t, result, slice)
			}
		}
	}
}

// TestProperty_AppendConsistency verifies append consistency
func TestProperty_AppendConsistency(t *testing.T) {
	r1 := New("Hello")
	r2 := New(" World")

	// Append
	r3 := r1.AppendRope(r2)

	// Should equal concatenation
	assert.Equal(t, "Hello World", r3.String())

	// Original ropes should be unchanged
	assert.Equal(t, "Hello", r1.String())
	assert.Equal(t, " World", r2.String())
}

// TestProperty_DeepTreeIntegrity verifies integrity after many operations
func TestProperty_DeepTreeIntegrity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	r := Empty()

	// Create deep tree
	for i := 0; i < 1000; i++ {
		r = r.Append("x")
	}

	// Verify we can iterate correctly
	it := r.NewIterator()
	count := 0
	for it.Next() {
		count++
	}
	assert.Equal(t, 1000, count)

	// Verify string
	assert.Equal(t, strings.Repeat("x", 1000), r.String())

	// Verify UTF-8 validity
	assert.True(t, utf8.ValidString(r.String()))
}

// TestProperty_RandomSplitsConsistency verifies split consistency
func TestProperty_RandomSplitsConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	text := "Hello World this is a test string for splitting"
	r := New(text)
	rng := rand.New(rand.NewSource(rand.Int63()))

	for i := 0; i < 100; i++ {
		if r.Length() == 0 {
			break
		}

		pos := rng.Intn(r.Length())
		left, right := r.Split(pos)

		// Verify split was correct
		combined := left.String() + right.String()
		assert.Equal(t, r.String(), combined)

		// Randomly choose which part to continue with
		if rng.Intn(2) == 0 {
			r = left
		} else {
			r = right
		}
	}
}

// TestProperty_UnicodeHandling verifies Unicode handling
func TestProperty_UnicodeHandling(t *testing.T) {
	texts := []string{
		"Hello 世界 🌍",
		"こんにちは",
		"مرحبا",
		"🌍🌎🌏",
		"\r\n",
		"Mix of ASCII 世界 and Emoji 🌍",
	}

	for _, text := range texts {
		r := New(text)

		// Should be valid UTF-8
		assert.True(t, utf8.ValidString(r.String()))

		// Length should match rune count
		expectedLen := utf8.RuneCountInString(text)
		assert.Equal(t, expectedLen, r.Length())

		// Iterator should iterate correct number of times
		it := r.NewIterator()
		count := 0
		for it.Next() {
			count++
		}
		assert.Equal(t, expectedLen, count)
	}
}

// TestProperty_CRLFHandling verifies CRLF handling
func TestProperty_CRLFHandling(t *testing.T) {
	texts := []string{
		"Line 1\r\nLine 2\r\n",
		"\r\n",
		"a\r\nb\r\nc",
	}

	for _, text := range texts {
		r := New(text)

		// Should be valid UTF-8
		assert.True(t, utf8.ValidString(r.String()))

		// String should match
		assert.Equal(t, text, r.String())

		// Lines should split correctly
		lines := r.Lines()
		assert.True(t, len(lines) > 0)
	}
}

// TestProperty_EmptyRopeOperations verifies operations on empty rope
func TestProperty_EmptyRopeOperations(t *testing.T) {
	r := Empty()

	// Empty rope should have length 0
	assert.Equal(t, 0, r.Length())
	assert.Equal(t, "", r.String())

	// Iterator should have no elements
	it := r.NewIterator()
	assert.False(t, it.Next())

	// Slicing should return empty
	assert.Equal(t, "", r.Slice(0, 0))

	// Append should work
	r2 := r.Append("Hello")
	assert.Equal(t, "Hello", r2.String())

	// Original should still be empty
	assert.Equal(t, "", r.String())
}

// TestProperty_SingleCharRope verifies single character rope
func TestProperty_SingleCharRope(t *testing.T) {
	r := New("a")

	assert.Equal(t, 1, r.Length())
	assert.Equal(t, "a", r.String())

	// CharAt should work
	assert.Equal(t, 'a', r.CharAt(0))

	// Iterator should work
	it := r.NewIterator()
	assert.True(t, it.Next())
	assert.Equal(t, 'a', it.Current())
	assert.False(t, it.Next())
}
