package rope

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// ========== Random Insert/Delete Stress Tests ==========

// TestStress_RandomInsertDelete tests random insert and delete operations
func TestStress_RandomInsertDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	rand.Seed(time.Now().UnixNano())
	r := Empty()
	originalText := ""

	// Perform 1000 random insert operations
	for i := 0; i < 1000; i++ {
		// Generate random text
		text := randomString(rand.Intn(20))
		pos := 0
		if r.Length() > 0 {
			pos = rand.Intn(r.Length() + 1)
		}

		r = r.Insert(pos, text)

		// Update expected text
		if pos == 0 {
			originalText = text + originalText
		} else if pos >= len([]rune(originalText)) {
			originalText = originalText + text
		} else {
			runes := []rune(originalText)
			before := string(runes[:pos])
			after := string(runes[pos:])
			originalText = before + text + after
		}
	}

	// Verify integrity
	assert.Equal(t, len([]rune(originalText)), r.Length())
	assert.Equal(t, originalText, r.String())

	// Perform 500 random delete operations
	for i := 0; i < 500; i++ {
		if r.Length() == 0 {
			break
		}

		start := rand.Intn(r.Length())
		end := start + rand.Intn(r.Length()-start+1)

		// Update expected text
		runes := []rune(originalText)
		if start < len(runes) && end <= len(runes) {
			_ = string(runes[start:end])
			originalText = string(runes[:start]) + string(runes[end:])
		}

		r = r.Delete(start, end)

		// Verify integrity
		assert.Equal(t, len([]rune(originalText)), r.Length())
		assert.Equal(t, originalText, r.String())
	}
}

// TestStress_LargeInsertAtBeginning tests many inserts at beginning
func TestStress_LargeInsertAtBeginning(t *testing.T) {
	r := Empty()

	// Insert 100 times at beginning
	for i := 0; i < 100; i++ {
		r = r.Insert(0, "x")
		expected := strings.Repeat("x", i+1)
		assert.Equal(t, expected, r.String())
	}

	assert.Equal(t, 100, r.Length())
}

// TestStress_LargeInsertAtEnd tests many inserts at end
func TestStress_LargeInsertAtEnd(t *testing.T) {
	r := Empty()

	// Insert 100 times at end
	for i := 0; i < 100; i++ {
		r = r.Insert(r.Length(), "x")
	}

	assert.Equal(t, 100, r.Length())
	assert.Equal(t, strings.Repeat("x", 100), r.String())
}

// TestStress_AlternatingInsertDelete tests alternating insert and delete
func TestStress_AlternatingInsertDelete(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	r := New("Hello World")

	// Alternately insert and delete
	for i := 0; i < 100; i++ {
		// Insert at random position
		pos := rand.Intn(r.Length() + 1)
		r = r.Insert(pos, "X")

		// Delete from random position
		if r.Length() > 5 {
			start := rand.Intn(r.Length() - 5)
			end := start + rand.Intn(6)
			if end > r.Length() {
				end = r.Length()
			}
			r = r.Delete(start, end)
		}
	}

	// Just verify it doesn't crash and maintains valid UTF-8
	assert.True(t, utf8.ValidString(r.String()))
	assert.True(t, r.Length() >= 0)
}

// ========== Split Stress Tests ==========

// TestStress_RandomSplits tests random split operations
func TestStress_RandomSplits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	rand.Seed(time.Now().UnixNano())
	text := "Hello World this is a test string for splitting"
	r := New(text)

	// Perform 100 random split operations
	for i := 0; i < 100; i++ {
		if r.Length() == 0 {
			break
		}

		pos := rand.Intn(r.Length())
		left, right := r.Split(pos)

		// Verify split was correct
		combined := left.String() + right.String()
		assert.Equal(t, r.String(), combined)

		// Randomly choose which part to continue with
		if rand.Intn(2) == 0 {
			r = left
		} else {
			r = right
		}
	}
}

// TestStress_SplitAndMerge tests split followed by merge
func TestStress_SplitAndMerge(t *testing.T) {
	r := New("Hello World Test String")

	// Split at position 6
	left, right := r.Split(6)
	assert.Equal(t, "Hello ", left.String())
	assert.Equal(t, "World Test String", right.String())

	// Merge back
	merged := left.AppendRope(right)
	assert.Equal(t, "Hello World Test String", merged.String())
}

// ========== Append Stress Tests ==========

// TestStress_SequentialAppends tests many sequential appends
func TestStress_SequentialAppends(t *testing.T) {
	r := Empty()

	// Append 1000 characters one at a time
	for i := 0; i < 1000; i++ {
		r = r.Append("x")
		assert.Equal(t, i+1, r.Length())
	}

	assert.Equal(t, 1000, r.Length())
	assert.Equal(t, strings.Repeat("x", 1000), r.String())
}

// TestStress_LargeAppends tests appending large chunks
func TestStress_LargeAppends(t *testing.T) {
	r := Empty()

	// Append 100 chunks
	for i := 0; i < 100; i++ {
		text := fmt.Sprintf("Chunk%03d", i)
		r = r.Append(text)
	}

	assert.Equal(t, 100*8, r.Length()) // Each chunk is 8 chars ("Chunk" + 3 digits)
}

// ========== Deep Tree Stress Tests ==========

// TestStress_DeepTreeCreation tests creating very deep trees
func TestStress_DeepTreeCreation(t *testing.T) {
	r := Empty()

	// Create deep tree through many appends
	for i := 0; i < 1000; i++ {
		r = r.Append(fmt.Sprintf("%d", i%10))
	}

	// Verify we can still iterate correctly
	it := r.NewIterator()
	count := 0
	for it.Next() {
		count++
	}

	assert.Equal(t, 1000, count)
	assert.True(t, utf8.ValidString(r.String()))
}

// TestStress_DeepTreeRandomAccess tests random access on deep tree
func TestStress_DeepTreeRandomAccess(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	r := Empty()

	// Create deep tree
	for i := 0; i < 100; i++ {
		r = r.Append(fmt.Sprintf("str%d", i))
	}

	// Perform random char access
	for i := 0; i < 100; i++ {
		pos := rand.Intn(r.Length())
		ch := r.CharAt(pos)
		assert.True(t, ch != 0)
	}
}

// ========== Mutation Integrity Tests ==========

// TestIntegrity_AfterManyMutations tests integrity after many mutations
func TestIntegrity_AfterManyMutations(t *testing.T) {
	r := New("Hello")

	// Perform 500 random mutations
	for i := 0; i < 500; i++ {
		op := rand.Intn(3)

		switch op {
		case 0: // Insert
			if r.Length() < 10000 { // Cap size
				text := randomString(rand.Intn(10))
				pos := rand.Intn(r.Length() + 1)
				r = r.Insert(pos, text)
			}

		case 1: // Delete
			if r.Length() > 1 {
				start := rand.Intn(r.Length())
				end := start + rand.Intn(r.Length()-start+1)
				if end > r.Length() {
					end = r.Length()
				}
				r = r.Delete(start, end)
			}

		case 2: // Append
			if r.Length() < 10000 {
				text := randomString(rand.Intn(10))
				r = r.Append(text)
			}
		}
	}

	// Verify final integrity
	assert.True(t, utf8.ValidString(r.String()))
	assert.True(t, r.Length() >= 0)
}

// ========== Unicode Stress Tests ==========

// TestStress_UnicodeMutations tests mutations with unicode text
func TestStress_UnicodeMutations(t *testing.T) {
	r := New("Hello 世界 🌍")

	// Perform mutations with unicode
	for i := 0; i < 100; i++ {
		op := rand.Intn(2)

		if op == 0 {
			// Insert unicode
			unicodeText := "🌍🌎🌏"
			pos := rand.Intn(r.Length() + 1)
			r = r.Insert(pos, unicodeText)
		} else {
			// Delete random range
			if r.Length() > 5 {
				start := rand.Intn(r.Length() - 4)
				end := start + rand.Intn(r.Length()-start)
				r = r.Delete(start, end)
			}
		}

		// Always maintain valid UTF-8
		assert.True(t, utf8.ValidString(r.String()))
	}
}

// ========== Helper Functions ==========

// randomString generates a random string of given length
func randomString(length int) string {
	if length == 0 {
		return ""
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
